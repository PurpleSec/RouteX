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
	Timeout time.Duration
	Default http.Handler

	ctx     context.Context
	lock    sync.RWMutex
	cancel  context.CancelFunc
	entries entries
}
type entry struct {
	name    string
	entry   Handler
	method  map[string]Handler
	matcher *regexp.Regexp
}
type entries []*entry

// Handler is a fork of the http.Handler interface. This interface supplies a base Context to be used and augments
// the supplied Request to have options for getting formatted JSON body content or getting the URL match groups.
type Handler interface {
	Handle(context.Context, http.ResponseWriter, *Request)
}

// HandlerFunc is a function wrapper for handlers to be used as single functions instead.
type HandlerFunc func(context.Context, http.ResponseWriter, *Request)

const (
	// ErrInvalidPath is returned from the 'Handler*' functions when the path is empty.
	ErrInvalidPath = strErr("supplied path is invalid")
	// ErrInvalidHandler is returned from the 'Handler*' functions when the Handler is nil.
	ErrInvalidHandler = strErr("cannot use a nil Handler")
)

var notAllowed = HandlerFunc(func(_ context.Context, w http.ResponseWriter, _ *Request) {
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
func (e entries) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}
func (e entries) Less(i, j int) bool {
	return len(e[i].matcher.String()) < len(e[j].matcher.String())
}

// NewContext creates a new Mux and applies the supplied Context as the Mux base Context.
func NewContext(x context.Context) *Mux {
	m := &Mux{}
	m.ctx, m.cancel = context.WithCancel(x)
	return m
}

// Handle adds the Handler to the supplied regex expression path and gives it the supplied name. Path values must be
// unique and don't have to contain regex expressions. Regex match groups can be used to grab data out of the call
// and will be placed in the 'Values' Request map. The name will be in the 'Route' Request attribute to signal where
// the call originated from. This function returns an error if a duplicate path exists or the regex expression is
// invalid.
//
// This function will add a handler that will be considered the 'default' handler for the path and will be called
// unless a method-based Handler is also specified and that HTTP method is used.
func (m *Mux) Handle(name, path string, h Handler) error {
	if len(path) == 0 {
		return ErrInvalidPath
	}
	if h == nil {
		return ErrInvalidHandler
	}
	x, err := regexp.Compile(path)
	if err != nil {
		return err
	}
	return m.add(name, "", path, x, h)
}

// ServeHTTP allows RegexMux to fulfill the http.Handler interface.
func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI == "*" {
		if r.ProtoAtLeast(1, 1) {
			w.Header().Set("Connection", "close")
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if p := clean(r.URL.Path); p != r.URL.Path {
		u := *r.URL
		u.Path = p
		http.Redirect(w, r, u.String(), http.StatusMovedPermanently)
		return
	}
	y := m.ctx
	if y == nil {
		y = r.Context()
	}
	h, x := m.handler(r.URL.Path, r)
	if h != nil && x != nil {
		if m.Timeout == 0 {
			handle(y, h, w, x)
			return
		}
		t, f := context.WithTimeout(y, m.Timeout)
		handle(t, h, w, x)
		f()
		return
	}
	if h != nil {
		h.Handle(m.ctx, w, nil)
		return
	}
	if m.Default != nil {
		m.Default.ServeHTTP(w, r)
		return
	}
	http.NotFound(w, r)
}

// HandleFunc adds the handler function to the supplied regex expression path and gives it the supplied name.
// Path values must be unique and don't have to contain regex expressions. Regex match groups can be used to grab
// data out of the call and will be placed in the 'Values' Request map. The name will be in the 'Route' Request
// attribute to signal where the call originated from. This function returns an error if a duplicate path exists or
// the regex expression is invalid.
//
// This function will add a handler that will be considered the 'default' handler for the path and will be called
// unless a method-based Handler is also specified and that HTTP method is used.
func (m *Mux) HandleFunc(name, path string, h HandlerFunc) error {
	if len(path) == 0 {
		return ErrInvalidPath
	}
	if h == nil {
		return ErrInvalidHandler
	}
	x, err := regexp.Compile(path)
	if err != nil {
		return err
	}
	return m.add(name, "", path, x, HandlerFunc(h))
}
func (m *Mux) add(n, o, s string, x *regexp.Regexp, h Handler) error {
	if m.lock.Lock(); len(m.entries) > 0 {
		for i := range m.entries {
			if m.entries[i].matcher.String() != s {
				continue
			}
			if len(o) > 0 {
				if m.entries[i].method == nil {
					m.entries[i].method = make(map[string]Handler, 1)
				}
				m.entries[i].method[o] = h
				m.lock.Unlock()
				return nil
			}
			if m.entries[i].entry != nil {
				m.lock.Unlock()
				return strErr(`path matcher "` + s + `" already exists`)
			}
			m.entries[i].entry = h
			m.lock.Unlock()
			return nil
		}
	}
	e := &entry{name: n, matcher: x}
	if len(o) > 0 {
		e.method = map[string]Handler{o: h}
	} else {
		e.entry = h
	}
	m.entries = append(m.entries, e)
	sort.Sort(m.entries)
	m.lock.Unlock()
	return nil
}
func (m *Mux) handler(s string, r *http.Request) (Handler, *Request) {
	m.lock.RLock()
	var (
		l []string
		h Handler
	)
	for i := range m.entries {
		if l = m.entries[i].matcher.FindStringSubmatch(s); len(l) == 0 {
			continue
		}
		if len(m.entries[i].method) > 0 {
			h = m.entries[i].method[r.Method]
		}
		if h == nil && m.entries[i].entry == nil {
			return notAllowed, nil
		}
		if h == nil {
			h = m.entries[i].entry
		}
		o := &Request{
			ctx:     m.ctx,
			Route:   m.entries[i].name,
			Values:  make(values, len(l)),
			Request: r,
		}
		for x, n := range m.entries[i].matcher.SubexpNames() {
			if x == 0 || len(n) == 0 {
				continue
			}
			o.Values[n] = requestValue(l[x])
		}
		m.lock.RUnlock()
		return h, o
	}
	m.lock.RUnlock()
	return nil, nil
}

// HandleMethod adds the handler function to the supplied regex expression path abd HTTP method and gives it the
// supplied name. Path values must be unique and don't have to contain regex expressions. Regex match groups can be
// used to grab data out of the call and will be placed in the 'Values' Request map. The name will be in the 'Route'
// Request attribute to signal where the call originated from. This function returns an error if a duplicate path
// exists or the regex expression is invalid.
//
// This function is similar to the 'Handle' and 'HandleFunc' methods, but will also take a HTTP method name. This will
// take precendance over the non-method calls. Multiple calls to this function will allow for overriting the method
// Handler or adding another handler for a different method name.
func (m *Mux) HandleMethod(name, method, path string, h Handler) error {
	if len(path) == 0 {
		return ErrInvalidPath
	}
	if h == nil {
		return ErrInvalidHandler
	}
	x, err := regexp.Compile(path)
	if err != nil {
		return err
	}
	return m.add(name, method, path, x, h)
}
func handle(x context.Context, h Handler, w http.ResponseWriter, r *Request) {
	defer func() {
		if recover() != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}()
	h.Handle(x, w, r)
}

// HandleFuncMethod adds the handler function to the supplied regex expression path abd HTTP method and gives it the
// supplied name. Path values must be unique and don't have to contain regex expressions. Regex match groups can be
// used to grab data out of the call and will be placed in the 'Values' Request map. The name will be in the 'Route'
// Request attribute to signal where the call originated from. This function returns an error if a duplicate path
// exists or the regex expression is invalid.
//
// This function is similar to the 'Handle' and 'HandleFunc' methods, but will also take a HTTP method name. This will
// take precendance over the non-method calls. Multiple calls to this function will allow for overriting the method
// Handler or adding another handler for a different method name.
func (m *Mux) HandleFuncMethod(name, method, path string, h HandlerFunc) error {
	if len(path) == 0 {
		return ErrInvalidPath
	}
	if h == nil {
		return ErrInvalidHandler
	}
	x, err := regexp.Compile(path)
	if err != nil {
		return err
	}
	return m.add(name, method, path, x, HandlerFunc(h))
}

// Handle allows this alias to fulfill the Handler interface.
func (h HandlerFunc) Handle(x context.Context, w http.ResponseWriter, r *Request) {
	h(x, w, r)
}
