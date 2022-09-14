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
	"regexp"
	"sort"
	"sync"
	"time"
)

const (
	// ErrInvalidPath is returned from the 'Add*' functions when the path is empty.
	ErrInvalidPath = errStr("supplied path is invalid")
	// ErrInvalidRegexp is returned from the 'AddExp*' functions when the Regexp
	// expression is nil.
	ErrInvalidRegexp = errStr("cannot use a nil Regexp")
	// ErrInvalidHandler is returned from the 'Add*' functions when the Handler is
	// nil.
	ErrInvalidHandler = errStr("cannot use a nil Handler")
	// ErrInvalidMethod is an error returned when the HTTP method names provided
	// are empty.
	ErrInvalidMethod = errStr("supplied methods contains an empty method name")
)

// Mux is a http Handler that can be used to handle connections based on Regex
// expression paths. Matching groups passed in the request URL values may be parsed
// out and passed to the resulting request.
//
// This Handler supports a base context that can be used to signal closure to all
// running Handlers.
type Mux struct {
	lock sync.RWMutex

	Error, Error404    ErrorHandler
	Error405, Error500 ErrorHandler
	ctx                context.Context
	log                logger

	Default Handler
	wares   *wares
	routes  router

	Timeout time.Duration
}

// Route is an interface that allows for modification of an added HTTP route after
// being created.
//
// One example function is adding route-specific middleware.
type Route interface {
	Middleware(m ...Middleware) Route
}

// Handler is a fork of the http.Handler interface. This interface supplies a base
// Context to be used and augments the supplied Request to have options for getting
// formatted JSON body content or getting the URL match groups.
type Handler interface {
	Handle(context.Context, http.ResponseWriter, *Request)
}

// ErrorHandler is an interface that allows for handling any error returns to be
// reported to the client instead of using the default methods.
//
// The 'HandleError' method will be called with an error status code, error message
// and the standard 'Handler' options (except the Context).
type ErrorHandler interface {
	HandleError(int, string, http.ResponseWriter, *Request)
}

// New returns a new Mux instance.
func New() *Mux {
	return new(Mux)
}

// SetLog will set the internal logger for the Mux instance. This can be used to
// debug any errors during runtime.
func (m *Mux) SetLog(l logger) {
	m.log = l
}

// NewContext creates a new Mux and applies the supplied Context as the Mux base Context.
func NewContext(x context.Context) *Mux {
	return &Mux{ctx: x}
}

// Must adds the Handler to the supplied regex expression path. Path values must
// be unique and don't have to contain regex expressions.
//
// Regex match groups can be used to grab data out of the call and will be placed
// in the 'Values' Request map.
//
// This function panics if a duplicate path exists or the regex expression is invalid.
//
// This function will add a handler that will be considered the 'default' handler
// for the path and will be called unless a method-based Handler is also specified
// and that HTTP method is used.
func (m *Mux) Must(path string, h Handler, methods ...string) Route {
	v, err := m.Add(path, h, methods...)
	if err != nil {
		panic(err.Error())
	}
	return v
}

// Add adds the Handler to the supplied regex expression path. Path values must be
// unique and don't have to contain regex expressions.
//
// Regex match groups can be used to grab data out of the call and will be placed
// in the 'Values' Request map.
//
// This function returns an error if a duplicate path exists or the regex expression
// is invalid.
//
// This function will add a handler that will be considered the 'default' handler
// for the path and will be called unless a method-based Handler is also specified
// and that HTTP method is used.
func (m *Mux) Add(path string, h Handler, methods ...string) (Route, error) {
	if len(path) == 0 {
		return nil, ErrInvalidPath
	}
	if h == nil {
		return nil, ErrInvalidHandler
	}
	x, err := regexp.Compile(path)
	if err != nil {
		return nil, &errValue{s: `path "` + path + `" compile`, e: err}
	}
	return m.add(path, methods, x, h)
}

// MustExp adds the Handler to the supplied regex expression. Path values must be
// unique and don't have to contain regex expressions.
//
// Regex match groups can be used to grab data out of the call and will be placed
// in the 'Values' Request map.
//
// This function panics if a duplicate path exists or the regex expression is invalid.
//
// This function will add a handler that will be considered the 'default' handler
// for the path and will be called unless a method-based Handler is also specified
// and that HTTP method is used.
func (m *Mux) MustExp(exp *regexp.Regexp, h Handler, methods ...string) Route {
	v, err := m.AddExp(exp, h, methods...)
	if err != nil {
		panic(err.Error())
	}
	return v
}

// AddExp adds the Handler to the supplied regex expression. Path values must be
// unique and don't have to contain regex expressions.
//
// Regex match groups can be used to grab data out of the call and will be placed
// in the 'Values' Request map.
//
// This function returns an error if a duplicate path exists or the regex expression
// is invalid.
//
// This function will add a handler that will be considered the 'default' handler
// for the path and will be called unless a method-based Handler is also specified
// and that HTTP method is used.
func (m *Mux) AddExp(exp *regexp.Regexp, h Handler, methods ...string) (Route, error) {
	if exp == nil {
		return nil, ErrInvalidRegexp
	}
	v := exp.String()
	if len(v) == 0 {
		return nil, ErrInvalidPath
	}
	if h == nil {
		return nil, ErrInvalidHandler
	}
	return m.add(v, methods, exp, h)
}
func (m *Mux) add(path string, methods []string, x *regexp.Regexp, h Handler) (*handler, error) {
	for _, n := range methods {
		if len(n) == 0 {
			return nil, ErrInvalidMethod
		}
	}
	if m.lock.Lock(); len(m.routes) > 0 {
		for i := range m.routes {
			if m.routes[i].matcher.String() != path {
				continue
			}
			if len(methods) > 0 {
				if m.routes[i].method == nil {
					m.routes[i].method = make(map[string]*handler, len(methods))
				}
				v := &handler{h: h}
				for _, n := range methods {
					m.routes[i].method[n] = v
				}
				m.lock.Unlock()
				return v, nil
			}
			if m.routes[i].base != nil {
				m.lock.Unlock()
				return nil, errStr(`matcher path "` + path + `" already exists`)
			}
			v := &handler{h: h}
			m.routes[i].base = v
			m.lock.Unlock()
			return v, nil
		}
	}
	var (
		v = &handler{h: h}
		e = &entry{matcher: x}
	)
	if len(methods) > 0 {
		e.method = make(map[string]*handler, len(methods))
		for _, n := range methods {
			e.method[n] = v
		}
	} else {
		e.base = v
	}
	m.routes = append(m.routes, e)
	sort.Sort(m.routes)
	m.lock.Unlock()
	return v, nil
}
