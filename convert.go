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
)

type convert struct {
	http.Handler
}

// ConvertFunc is an alias for the standard 'http.HandlerFunc' that can be used
// for compatibility with any built-in interface support.
type ConvertFunc http.HandlerFunc

// Func is an alias that can be used to use a function signature as a 'Handler' instead.
type Func func(context.Context, http.ResponseWriter, *Request)

// ErrorFunc is an alias that can be used to use a function signature as a 'ErrorHandler'
// instead.
type ErrorFunc func(int, string, http.ResponseWriter, *Request)

// WrapFunc is an alias that can be used to use a function signature as a 'Wrapper'
// instead.
type WrapFunc func(context.Context, http.ResponseWriter, *Request, Content)

// MarshalFunc is an alias that can be used to use a function signature as a 'Marshaler'
// instead.
type MarshalFunc func(context.Context, http.ResponseWriter, *Request, interface{})

// Convert is a warpper for the standard 'http.Handler' that can be used for
// compatibility with any built-in interface to support RouteX functions.
func Convert(h http.Handler) Handler {
	return convert{Handler: h}
}

// Handle allows this alias to fulfill the Handler interface.
func (f Func) Handle(x context.Context, w http.ResponseWriter, r *Request) {
	f(x, w, r)
}

// Handle allows this alias to fulfill the Handler interface.
func (c convert) Handle(_ context.Context, w http.ResponseWriter, r *Request) {
	c.Handler.ServeHTTP(w, r.Request)
}

// Handle allows this alias to fulfill the Handler interface.
func (f ConvertFunc) Handle(_ context.Context, w http.ResponseWriter, r *Request) {
	f(w, r.Request)
}

// HandleError allows this alias to fulfill the ErrorHandler interface.
func (f ErrorFunc) HandleError(c int, s string, w http.ResponseWriter, r *Request) {
	f(c, s, w, r)
}

// Handle allows this alias to fulfill the Wrapper interface.
func (f WrapFunc) Handle(x context.Context, w http.ResponseWriter, r *Request, c Content) {
	f(x, w, r, c)
}

// Handle allows this alias to fulfill the Marshaler interface.
func (f MarshalFunc) Handle(x context.Context, w http.ResponseWriter, r *Request, i interface{}) {
	f(x, w, r, i)
}
