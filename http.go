// Copyright 2021 - 2022 PurpleSec Team
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package routex

import (
	"context"
	"net/http"
	"path"
	"regexp"
	"strings"
	"sync"
)

type wares struct {
	lock sync.RWMutex
	w    []Middleware
}
type entry struct {
	base    *handler
	method  map[string]*handler
	matcher *regexp.Regexp
}
type router []*entry
type handler struct {
	h     Handler
	wares *wares
}
type logger interface {
	Println(v ...interface{})
}
type stringer interface {
	String() string
}

func (r router) Len() int {
	return len(r)
}
func clean(s string) string {
	if len(s) == 0 {
		return "/"
	}
	if s[0] != '/' {
		s = "/" + s
	}
	n := path.Clean(s)
	if len(n) > 1 && s[len(s)-1] == '/' && n != "/" {
		if len(s) == len(n)+1 && strings.HasPrefix(s, n) {
			return s
		}
		return n + "/"
	}
	return n
}
func (r router) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}
func (r router) Less(i, j int) bool {
	return len(r[i].matcher.String()) < len(r[j].matcher.String())
}

// ServeHTTP allows RegexMux to fulfill the http.Handler interface.
func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI == "*" {
		if r.ProtoAtLeast(1, 1) {
			w.Header().Set("Connection", "close")
		}
		w.WriteHeader(http.StatusBadRequest)
		r.Body.Close()
		return
	}
	if p := clean(r.URL.Path); p != r.URL.Path {
		u := *r.URL
		if u.Path = p; m.log != nil {
			m.log.Println(`[RouteX] Requested "` + r.URL.String() + `" redirecting to "` + u.String() + `".`)
		}
		http.Redirect(w, r, u.String(), http.StatusMovedPermanently)
		r.Body.Close()
		return
	}
	ctx := m.ctx
	if ctx == nil || ctx == context.Background() {
		ctx = r.Context()
	}
	h, x, a, f := m.handler(r.URL.Path, r)
	if x == nil && f && len(a) > 0 {
		w.Header().Set("Allow", a)
		w.WriteHeader(http.StatusNoContent)
		r.Body.Close()
		return
	}
	if h != nil {
		m.process(ctx, h.h, h.wares, w, x)
		x.Body.Close()
		return
	}
	if f {
		m.handleError(http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed), w, x)
		x.Body.Close()
		return
	}
	if m.Default != nil {
		m.process(ctx, m.Default, nil, w, x)
		x.Body.Close()
		return
	}
	m.handleError(http.StatusNotFound, http.StatusText(http.StatusNotFound), w, x)
	x.Body.Close()
}
func (m *Mux) handleError(c int, s string, w http.ResponseWriter, r *Request) {
	switch {
	case c == http.StatusNotFound && m.Error404 != nil:
		m.Error404.HandleError(c, s, w, r)
	case c == http.StatusMethodNotAllowed && m.Error405 != nil:
		m.Error405.HandleError(c, s, w, r)
	case c == http.StatusInternalServerError && m.Error500 != nil:
		m.Error500.HandleError(c, s, w, r)
	case m.Error != nil:
		m.Error.HandleError(c, s, w, r)
	default:
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(c)
	}
}
func (m *Mux) handler(s string, r *http.Request) (*handler, *Request, string, bool) {
	var h *handler
	if m.lock.RLock(); m.log != nil {
		m.log.Println(`[RouteX] URL "` + s + `" requested..`)
	}
	for i := range m.routes {
		l := m.routes[i].matcher.FindStringSubmatch(s)
		if len(l) == 0 {
			continue
		}
		if m.log != nil {
			m.log.Println(`[RouteX] URL "` + s + `" was matched by "` + m.routes[i].matcher.String() + `".`)
		}
		if len(m.routes[i].method) > 0 {
			h = m.routes[i].method[r.Method]
		}
		if h == nil {
			if r.Method == http.MethodOptions {
				if m.lock.RUnlock(); len(m.routes[i].method) > 0 {
					var (
						v strings.Builder
						c uint
					)
					for n := range m.routes[i].method {
						if c > 0 {
							v.WriteString(", ")
						}
						v.WriteString(n)
						c++
					}
					return nil, nil, v.String(), true
				}
				return nil, nil, "*", true
			}
			if h = m.routes[i].base; h == nil {
				if m.lock.RUnlock(); m.log != nil {
					m.log.Println(`[RouteX] URL "` + s + `" was matched, but method ` + r.Method + ` was not (default == nil) returning 405!`)
				}
				return nil, &Request{ctx: m.ctx, Mux: m, Request: r}, "", true
			}
		}
		x := &Request{ctx: m.ctx, Mux: m, Values: make(values, len(l)), Request: r}
		for z, n := range m.routes[i].matcher.SubexpNames() {
			if z == 0 || len(n) == 0 {
				continue
			}
			if x.Values[n] = value(l[z]); m.log != nil {
				m.log.Println(`[RouteX] URL "` + r.URL.String() + `" "` + n + `=` + l[z] + `"`)
			}
		}
		m.lock.RUnlock()
		return h, x, "", true
	}
	m.lock.RUnlock()
	return nil, &Request{ctx: m.ctx, Mux: m, Request: r}, "", false
}
func (m *Mux) process(ctx context.Context, h Handler, v *wares, w http.ResponseWriter, r *Request) {
	defer func() {
		if err := recover(); err != nil {
			v := "unknown panic"
			switch i := err.(type) {
			case error:
				v = i.Error()
			case string:
				v = i
			case stringer:
				v = i.String()
			}
			if m.log != nil {
				m.log.Println(`[RouteX] Request "` + r.URL.String() + `" recovered from a panic caused by "` + v + `"!`)
			}
			m.handleError(http.StatusInternalServerError, v, w, r)
		}
	}()
	var (
		x = ctx
		f = func() {}
	)
	if m.Timeout > 0 {
		x, f = context.WithTimeout(x, m.Timeout)
	}
	if m.wares != nil && len(m.wares.w) > 0 {
		m.wares.lock.RLock()
		for i := range m.wares.w {
			if !m.wares.w[i](x, w, r) {
				m.wares.lock.RUnlock()
				f()
				return
			}
		}
		m.wares.lock.RUnlock()
	}
	if v != nil && len(v.w) > 0 {
		v.lock.RLock()
		for i := range v.w {
			if !v.w[i](x, w, r) {
				v.lock.RUnlock()
				f()
				return
			}
		}
		v.lock.RUnlock()
	}
	h.Handle(x, w, r)
	f()
}
