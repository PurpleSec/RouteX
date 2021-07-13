// Copyright 2021 PurpleSec Team
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
	"sort"
	"strings"
	"sync"
	"time"
)

// Mux is a http Handler that can be used to handle connections based on Regex expression paths. Matching
// groups passed in the request URL values may be parse out and passed to the resulting request.
//
// This Handler supports a base context that can be used to signal closure to all running Handlers.
type Mux struct {
	Error   ErrorHandler
	Default Handler

	ctx     context.Context
	log     logger
	cancel  context.CancelFunc
	entries entries

	Timeout time.Duration
	lock    sync.RWMutex
}
type entry struct {
	base    Handler
	method  map[string]Handler
	matcher *regexp.Regexp
	name    string
}
type entries []*entry
type logger interface {
	Print(v ...interface{})
}

// Handler is a fork of the http.Handler interface. This interface supplies a base Context to be used and augments
// the supplied Request to have options for getting formatted JSON body content or getting the URL match groups.
type Handler interface {
	Handle(context.Context, http.ResponseWriter, *Request)
}
type stringer interface {
	String() string
}

// ErrorHandler is an interface that allows for handeling any error returns to be reported to the client instead
// of using the default methods. The 'HandleError' method will be called with an error status code, error message
// and the standard 'Handler' options (expcept the Context).
type ErrorHandler interface {
	HandleError(int, string, http.ResponseWriter, *Request)
}

const (
	// ErrInvalidPath is returned from the 'Add*' functions when the path is empty.
	ErrInvalidPath = strErr("supplied path is invalid")
	// ErrInvalidRegexp is returned from the 'AddExp*' functions when the Regexp expression is nil.
	ErrInvalidRegexp = strErr("cannot use a nil Regexp")
	// ErrInvalidHandler is returned from the 'Add*' functions when the Handler is nil.
	ErrInvalidHandler = strErr("cannot use a nil Handler")
)

var notAllowed = Func(func(_ context.Context, w http.ResponseWriter, _ *Request) {
	http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
})

// New returns a new Mux instance.
func New() *Mux {
	return new(Mux)
}
func (e entries) Len() int {
	return len(e)
}
func clean(p string) string {
	if len(p) == 0 {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	n := path.Clean(p)
	if len(n) > 1 && p[len(p)-1] == '/' && n != "/" {
		if len(p) == len(n)+1 && strings.HasPrefix(p, n) {
			return p
		}
		return n + "/"
	}
	return n
}

// Close will attempt to cancel the built-in Context. This will always return nil and will signal
// the remaining open connextions to close. You can omit this call if a parent Content was supplied and canceled.
func (m *Mux) Close() error {
	if m.ctx == nil {
		return nil
	}
	m.cancel()
	return nil
}

// SetLog will set the internal logger for the Mux instance. This can be used to debug any errors during runtime.
func (m *Mux) SetLog(l logger) {
	m.log = l
}
func (e entries) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}
func (e entries) Less(i, j int) bool {
	return len(e[i].matcher.String()) < len(e[j].matcher.String())
}

// NewContext creates a new Mux and applies the supplied Context as the Mux base Context.
func NewContext(x context.Context) *Mux {
	m := new(Mux)
	m.ctx, m.cancel = context.WithCancel(x)
	return m
}

// Must adds the Handler to the supplied regex expression path and gives it the supplied name. Path values must be
// unique and don't have to contain regex expressions. Regex match groups can be used to grab data out of the call
// and will be placed in the 'Values' Request map. The name will be in the 'Route' Request attribute to signal where
// the call originated from. This function panics if a duplicate path exists or the regex expression is invalid.
//
// This function will add a handler that will be considered the 'default' handler for the path and will be called
// unless a method-based Handler is also specified and that HTTP method is used.
func (m *Mux) Must(name, path string, h Handler) {
	if err := m.AddMethod(name, "", path, h); err != nil {
		panic(err)
	}
}

// Add adds the Handler to the supplied regex expression path and gives it the supplied name. Path values must be
// unique and don't have to contain regex expressions. Regex match groups can be used to grab data out of the call
// and will be placed in the 'Values' Request map. The name will be in the 'Route' Request attribute to signal where
// the call originated from. This function returns an error if a duplicate path exists or the regex expression is
// invalid.
//
// This function will add a handler that will be considered the 'default' handler for the path and will be called
// unless a method-based Handler is also specified and that HTTP method is used.
func (m *Mux) Add(name, path string, h Handler) error {
	return m.AddMethod(name, "", path, h)
}

// MustMethod adds the handler function to the supplied regex expression path and HTTP method and gives it the
// supplied name. Path values must be unique and don't have to contain regex expressions. Regex match groups can be
// used to grab data out of the call and will be placed in the 'Values' Request map. The name will be in the 'Route'
// Request attribute to signal where the call originated from. This function panics if a duplicate path exists or
// the regex expression is invalid.
//
// This function is similar to the 'Add' method, but will also take a HTTP method name. This will take precendance
// over the non-method calls. Multiple calls to this function will allow for overriting the method Handler or adding
// another handler for a different method name.
func (m *Mux) MustMethod(name, method, path string, h Handler) {
	if err := m.AddMethod(name, method, path, h); err != nil {
		panic(err)
	}
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
		u.Path = p
		if m.log != nil {
			m.log.Print(`[RouteX Mux] Requested "` + r.URL.String() + `" redirecting to "` + u.String() + `".`)
		}
		http.Redirect(w, r, u.String(), http.StatusMovedPermanently)
		r.Body.Close()
		return
	}
	y := m.ctx
	if y == nil || y == context.Background() {
		y = r.Context()
	}
	h, x, f := m.handler(r.URL.Path, r)
	if h != nil {
		m.process(y, h, w, x)
		x.Body.Close()
		return
	}
	if f {
		if m.Error != nil {
			m.Error.HandleError(http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed), w, x)
		} else {
			notAllowed.Handle(y, w, x)
		}
		x.Body.Close()
		return
	}
	if m.Default != nil {
		m.Default.Handle(y, w, x)
		x.Body.Close()
		return
	}
	if m.Error != nil {
		m.Error.HandleError(http.StatusNotFound, http.StatusText(http.StatusNotFound), w, x)
	} else {
		http.NotFound(w, r)
	}
	x.Body.Close()
}

// MustExp adds the Handler to the supplied regex expression and gives it the supplied name. Path values must be
// unique and don't have to contain regex expressions. Regex match groups can be used to grab data out of the call
// and will be placed in the 'Values' Request map. The name will be in the 'Route' Request attribute to signal where
// the call originated from. This function panics if a duplicate path exists or the regex expression is invalid.
//
// This function will add a handler that will be considered the 'default' handler for the path and will be called
// unless a method-based Handler is also specified and that HTTP method is used.
func (m *Mux) MustExp(name string, exp *regexp.Regexp, h Handler) {
	if err := m.AddExpMethod(name, "", exp, h); err != nil {
		panic(err)
	}
}

// AddMethod adds the handler function to the supplied regex expression path and HTTP method and gives it the
// supplied name. Path values must be unique and don't have to contain regex expressions. Regex match groups can be
// used to grab data out of the call and will be placed in the 'Values' Request map. The name will be in the 'Route'
// Request attribute to signal where the call originated from. This function returns an error if a duplicate path
// exists or the regex expression is invalid.
//
// This function is similar to the 'Add' method, but will also take a HTTP method name. This will take precendance
// over the non-method calls. Multiple calls to this function will allow for overriting the method Handler or adding
// another handler for a different method name.
func (m *Mux) AddMethod(name, method, path string, h Handler) error {
	if len(path) == 0 {
		return ErrInvalidPath
	}
	if h == nil {
		return ErrInvalidHandler
	}
	x, err := regexp.Compile(path)
	if err != nil {
		return wrap(`path "`+path+`" compile`, err)
	}
	return m.add(name, method, path, x, h)
}

// AddExp adds the Handler to the supplied regex expression and gives it the supplied name. Path values must be
// unique and don't have to contain regex expressions. Regex match groups can be used to grab data out of the call
// and will be placed in the 'Values' Request map. The name will be in the 'Route' Request attribute to signal where
// the call originated from. This function returns an error if a duplicate path exists or the regex expression is
// invalid.
//
// This function will add a handler that will be considered the 'default' handler for the path and will be called
// unless a method-based Handler is also specified and that HTTP method is used.
func (m *Mux) AddExp(name string, exp *regexp.Regexp, h Handler) error {
	return m.AddExpMethod(name, "", exp, h)
}
func (m *Mux) handler(s string, r *http.Request) (Handler, *Request, bool) {
	var (
		l []string
		h Handler
	)
	if m.lock.RLock(); m.log != nil {
		m.log.Print(`[RouteX Mux] URL "` + s + `" requested...`)
	}
	for i := range m.entries {
		if l = m.entries[i].matcher.FindStringSubmatch(s); len(l) == 0 {
			continue
		}
		if m.log != nil {
			m.log.Print(`[RouteX Mux] URL "` + s + `" was matched by "` + m.entries[i].name + `".`)
		}
		if len(m.entries[i].method) > 0 {
			h = m.entries[i].method[r.Method]
		}
		if h == nil {
			if h = m.entries[i].base; h == nil {
				if m.log != nil {
					m.log.Print(`[RouteX Mux] URL "` + s + `" was matched, but method ` + r.Method + ` was not (default == nil) returning 405!`)
				}
				return nil, &Request{ctx: m.ctx, Request: r}, true
			}
		}
		o := &Request{ctx: m.ctx, Route: m.entries[i].name, Values: make(values, len(l)), Request: r}
		for x, n := range m.entries[i].matcher.SubexpNames() {
			if x == 0 || len(n) == 0 {
				continue
			}
			if o.Values[n] = value(l[x]); m.log != nil {
				m.log.Print(`[RouteX Mux] URL "` + r.URL.String() + `" handler "` + m.entries[i].name + ` "` + n + `=` + l[x] + `"`)
			}
		}
		m.lock.RUnlock()
		return h, o, true
	}
	m.lock.RUnlock()
	return nil, &Request{ctx: m.ctx, Request: r}, false
}

// MustExpMethod adds the handler function to the supplied regex expression and HTTP method and gives it the
// supplied name. Path values must be unique and don't have to contain regex expressions. Regex match groups can be
// used to grab data out of the call and will be placed in the 'Values' Request map. The name will be in the 'Route'
// Request attribute to signal where the call originated from. This function panics if a duplicate path exists or
// the regex expression is invalid.
//
// This function is similar to the 'MustExp'. but will also take a HTTP method name. This will take precendance
// over the non-method calls. Multiple calls to this function will allow for overriting the method Handler or adding
// another handler for a different method name.
func (m *Mux) MustExpMethod(name, method string, exp *regexp.Regexp, h Handler) {
	if err := m.AddExpMethod(name, method, exp, h); err != nil {
		panic(err)
	}
}
func (m *Mux) add(name, method, path string, x *regexp.Regexp, h Handler) error {
	if m.lock.Lock(); len(m.entries) > 0 {
		for i := range m.entries {
			if m.entries[i].matcher.String() != path {
				continue
			}
			if len(method) > 0 {
				if m.entries[i].method == nil {
					m.entries[i].method = make(map[string]Handler, 1)
				}
				m.entries[i].method[method] = h
				m.lock.Unlock()
				return nil
			}
			if m.entries[i].base != nil {
				m.lock.Unlock()
				return strErr(`matcher path "` + path + `" already exists`)
			}
			m.entries[i].base = h
			m.lock.Unlock()
			return nil
		}
	}
	e := &entry{name: name, matcher: x}
	if len(method) > 0 {
		e.method = map[string]Handler{method: h}
	} else {
		e.base = h
	}
	m.entries = append(m.entries, e)
	sort.Sort(m.entries)
	m.lock.Unlock()
	return nil
}

// AddExpMethod adds the handler function to the supplied regex expression and HTTP method and gives it the
// supplied name. Path values must be unique and don't have to contain regex expressions. Regex match groups can be
// used to grab data out of the call and will be placed in the 'Values' Request map. The name will be in the 'Route'
// Request attribute to signal where the call originated from. This function returns an error if a duplicate path
// exists or the regex expression is invalid.
//
// This function is similar to the 'AddExp' method, but will also take a HTTP method name. This will take precendance
// over the non-method calls. Multiple calls to this function will allow for overriting the method Handler or adding
// another handler for a different method name.
func (m *Mux) AddExpMethod(name, method string, exp *regexp.Regexp, h Handler) error {
	if exp == nil {
		return ErrInvalidRegexp
	}
	if len(exp.String()) == 0 {
		return ErrInvalidPath
	}
	if h == nil {
		return ErrInvalidHandler
	}
	return m.add(name, method, exp.String(), exp, h)
}
func (m *Mux) process(x context.Context, h Handler, w http.ResponseWriter, r *Request) {
	defer func() {
		if err := recover(); err != nil {
			v := "panic"
			switch i := err.(type) {
			case error:
				v = i.Error()
			case string:
				v = i
			case stringer:
				v = i.String()
			}
			if m.Error != nil {
				m.Error.HandleError(http.StatusInternalServerError, v, w, r)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
		}
	}()
	if m.Timeout == 0 {
		h.Handle(x, w, r)
	} else {
		z, f := context.WithTimeout(x, m.Timeout)
		h.Handle(z, w, r)
		f()
	}
}
