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

var (
	errEmpty   = &err{s: "empty body"}
	errInvalid = &err{s: "object is not valid"}
)

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

type Wrapper interface {
	Handle(context.Context, http.ResponseWriter, *Request, Content)
}

type Marshaler interface {
	Handle(context.Context, http.ResponseWriter, *Request, interface{})
}
type writer func(http.ResponseWriter, error)
type wrapperFunc func(context.Context, http.ResponseWriter, *Request, Content)
type marshalerFunc func(context.Context, http.ResponseWriter, *Request, interface{})

func Wrap(v Validator, h Wrapper) Handler {
	return &wrapper{h: h, v: v}
}
func WrapFunc(v Validator, h wrapperFunc) Handler {
	return &wrapper{h: h, v: v}
}
func WrapEx(v Validator, w writer, h Wrapper) Handler {
	return &wrapper{h: h, v: v, w: w}
}
func WrapFuncEx(v Validator, w writer, h wrapperFunc) Handler {
	return &wrapper{h: h, v: v, w: w}
}

func Marshal(v Validator, i interface{}, h Marshaler) Handler {
	return &marshaler{h: h, v: v, o: reflect.TypeOf(i)}
}
func MarshalFunc(v Validator, i interface{}, h marshalerFunc) Handler {
	return &marshaler{h: h, v: v, o: reflect.TypeOf(i)}
}

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
		if err == errEmpty {
			m.h.Handle(x, w, r, nil)
			return
		}
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

func MarshalFuncEx(v Validator, i interface{}, w writer, h marshalerFunc) Handler {
	return &marshaler{h: h, v: v, o: reflect.TypeOf(i), w: w}
}
func (f wrapperFunc) Handle(x context.Context, w http.ResponseWriter, r *Request, c Content) {
	f(x, w, r, c)
}
func (f marshalerFunc) Handle(x context.Context, w http.ResponseWriter, r *Request, i interface{}) {
	f(x, w, r, i)
}
