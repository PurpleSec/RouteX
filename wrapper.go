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

const errInvalid = strErr("object is not valid")

type wrapper struct {
	e ErrorHandler
	h Wrapper
	v Validator
}
type marshaler struct {
	e ErrorHandler
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

// Wrap will create a handler with the specified Validator that will check the content before passing control
// to the specified Handler.
func Wrap(v Validator, h Wrapper) Handler {
	return &wrapper{h: h, v: v}
}

// WrapEx will create a handler with the specified Validator that will check the content before passing control
// to the specified Handler. The supplied writer value allows for controlling the output when an error occurs.
func WrapEx(v Validator, e ErrorHandler, h Wrapper) Handler {
	return &wrapper{h: h, v: v, e: e}
}

// Marshal will create a handler that will attempt to unmarshal a copy of the supplied interface object once
// successfully validated by the supplied validator. An empty or 'new(obj)' variant of the requested data will
// work for this function.The supplied writer value allows for controlling the output when an error occurs.
func Marshal(v Validator, i interface{}, h Marshaler) Handler {
	return &marshaler{h: h, v: v, o: reflect.TypeOf(i)}
}
func (h wrapper) Handle(x context.Context, w http.ResponseWriter, r *Request) {
	if r.Body == nil {
		h.h.Handle(x, w, r, nil)
		return
	}
	c, err := r.ContentValidate(h.v)
	if err != nil {
		if h.e != nil {
			h.e.HandleError(http.StatusBadRequest, err.Error(), w, r)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}
	h.h.Handle(x, w, r, c)
}
func (m marshaler) Handle(x context.Context, w http.ResponseWriter, r *Request) {
	o := reflect.New(m.o)
	if !o.IsValid() {
		if m.e != nil {
			m.e.HandleError(http.StatusInternalServerError, errInvalid.Error(), w, r)
		} else {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}
	if r.Body == nil {
		m.h.Handle(x, w, r, nil)
		return
	}
	v := o.Interface()
	if err := r.MarshalValidate(m.v, v); err != nil {
		if m.e != nil {
			m.e.HandleError(http.StatusBadRequest, err.Error(), w, r)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}
	m.h.Handle(x, w, r, v)
}

// MarshalEx will create a handler that will attempt to unmarshal a copy of the supplied interface object once
// successfully validated by the supplied validator. An empty or 'new(obj)' variant of the requested data will
// work for this function. The supplied writer value allows for controlling the output when an error occurs.
func MarshalEx(v Validator, i interface{}, e ErrorHandler, h Marshaler) Handler {
	return &marshaler{h: h, v: v, o: reflect.TypeOf(i), e: e}
}
