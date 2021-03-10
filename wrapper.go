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
	"reflect"
)

var errInvalid = &err{s: "object is not valid"}

type wrapper struct {
	w writer
	h Wrapper
	v Validator
}
type marshaler struct {
	w writer
	h Marshaler
	v Validator
	o reflect.Type
}

// Wrapper is an interface that can wrap a Handler to instead directly get a Content object from the
// Router instead. These can be created using the 'Wrap*' functions passed with a Validator.
type Wrapper interface {
	Handle(context.Context, http.ResponseWriter, *Request, Content)
}

// Marshaler is an interface that can wrap a Handler to instead directly get the associated struct type from the
// Router instead. These can be created using the 'Marshal*' functions passed with a Validator.
type Marshaler interface {
	Handle(context.Context, http.ResponseWriter, *Request, interface{})
}
type writer func(http.ResponseWriter, error)
type wrapperFunc func(context.Context, http.ResponseWriter, *Request, Content)
type marshalerFunc func(context.Context, http.ResponseWriter, *Request, interface{})

// Wrap will create a handler with the specified Validator that will check the content before passing control
// to the specified Handler.
func Wrap(v Validator, h Wrapper) Handler {
	return &wrapper{h: h, v: v}
}

// WrapFunc will create a handler with the specified Validator that will check the content before passing control
// to the specified Handler. This function allows for passing a function instead on an interface.
func WrapFunc(v Validator, h wrapperFunc) Handler {
	return &wrapper{h: h, v: v}
}

// WrapEx will create a handler with the specified Validator that will check the content before passing control
// to the specified Handler. The supplied writer value allows for controlling the output when an error occurs.
func WrapEx(v Validator, w writer, h Wrapper) Handler {
	return &wrapper{h: h, v: v, w: w}
}

// WrapFuncEx will create a handler with the specified Validator that will check the content before passing control
// to the specified Handler. This function allows for passing a function instead on an interface. The supplied writer
// value allows for controlling the output when an error occurs.
func WrapFuncEx(v Validator, w writer, h wrapperFunc) Handler {
	return &wrapper{h: h, v: v, w: w}
}

// Marshal will create a handler that will attempt to unmarshal a copy of the supplied interface object once
// successfully validated by the supplied validator. An empty or 'new(obj)' variant of the requested data will
// work for this function.The supplied writer value allows for controlling the output when an error occurs.

// MarshalFunc will create a handler that will attempt to unmarshal a copy of the supplied interface object once
// successfully validated by the supplied validator. An empty or 'new(obj)' variant of the requested data will
// work for this function. This function allows for passing a function instead on an interface.
func MarshalFunc(v Validator, i interface{}, h marshalerFunc) Handler {
	return &marshaler{h: h, v: v, o: reflect.TypeOf(i)}
}

// MarshalEx will create a handler that will attempt to unmarshal a copy of the supplied interface object once
// successfully validated by the supplied validator. An empty or 'new(obj)' variant of the requested data will
// work for this function. The supplied writer value allows for controlling the output when an error occurs.
func MarshalEx(v Validator, i interface{}, w writer, h Marshaler) Handler {
	return &marshaler{h: h, v: v, o: reflect.TypeOf(i), w: w}
}
func (h wrapper) Handle(x context.Context, w http.ResponseWriter, r *Request) {
	if r.Body == nil {
		h.h.Handle(x, w, r, nil)
		return
	}
	c, err := r.ContentValidate(h.v)
	if err != nil {
		if h.w != nil {
			w.WriteHeader(http.StatusBadRequest)
			h.w(w, err)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	h.h.Handle(x, w, r, c)
}
func (m marshaler) Handle(x context.Context, w http.ResponseWriter, r *Request) {
	o := reflect.New(m.o)
	if !o.IsValid() {
		if m.w != nil {
			w.WriteHeader(http.StatusInternalServerError)
			m.w(w, errInvalid)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if r.Body == nil {
		m.h.Handle(x, w, r, nil)
		return
	}
	v := o.Interface()
	if err := r.MarshalValidate(m.v, v); err != nil {
		if m.w != nil {
			w.WriteHeader(http.StatusBadRequest)
			m.w(w, err)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	m.h.Handle(x, w, r, v)
}

// MarshalFuncEx will create a handler that will attempt to unmarshal a copy of the supplied interface object once
// successfully validated by the supplied validator. An empty or 'new(obj)' variant of the requested data will
// work for this function. This function allows for passing a function instead on an interface. The supplied writer
// value allows for controlling the output when an error occurs.
func MarshalFuncEx(v Validator, i interface{}, w writer, h marshalerFunc) Handler {
	return &marshaler{h: h, v: v, o: reflect.TypeOf(i), w: w}
}
func (f wrapperFunc) Handle(x context.Context, w http.ResponseWriter, r *Request, c Content) {
	f(x, w, r, c)
}
func (f marshalerFunc) Handle(x context.Context, w http.ResponseWriter, r *Request, i interface{}) {
	f(x, w, r, i)
}
