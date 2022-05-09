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
	"encoding/json"
	"net/http"
)

type wrapper struct {
	h Wrapper
	v Validator
}
type marshaler[T any] struct {
	h Marshaler[T]
	v Validator
	t T
}

// Wrapper is an interface that can wrap a Handler to instead directly get a Content
// object from the Router instead. These can be created using the 'Wrap*' functions
// passed with a Validator.
type Wrapper interface {
	Handle(context.Context, http.ResponseWriter, *Request, Content)
}

// Marshaler is an interface that can wrap a Handler to instead directly get the
// associated struct type from the Router instead. These can be created using the
// 'Marshal*' functions passed with a Validator.
type Marshaler[T any] interface {
	Handle(context.Context, http.ResponseWriter, *Request, T)
}

// Wrap will create a handler with the specified Validator that will check the
// content before passing control to the specified Handler.
func Wrap(v Validator, h Wrapper) Handler {
	return &wrapper{h: h, v: v}
}

// JSON will write the supplied interface to the ResponseWrite with the supplied
// status.
//
// DO NOT expect the writer to be usage afterwards.
//
// This function automatically sets the encoding to JSON.
func JSON(w http.ResponseWriter, c int, i any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(c)
	json.NewEncoder(w).Encode(i)
}

// Marshal will create a handler that will attempt to unmarshal a copy of the supplied
// interface object once successfully validated by the supplied validator.
//
// An empty or 'new(obj)' variant of the requested data will work for this function.
func Marshal[T any](v Validator, h Marshaler[T]) Handler {
	return &marshaler[T]{h: h, v: v}
}
func (h wrapper) Handle(x context.Context, w http.ResponseWriter, r *Request) {
	if r.Body == nil {
		h.h.Handle(x, w, r, nil)
		return
	}
	c, err := r.ValidateContent(h.v)
	if err != nil {
		r.Mux.handleError(http.StatusBadRequest, err.Error(), w, r)
		return
	}
	h.h.Handle(x, w, r, c)
}
func (m marshaler[T]) Handle(x context.Context, w http.ResponseWriter, r *Request) {
	var v T
	if r.Body == nil {
		m.h.Handle(x, w, r, v)
		return
	}
	if err := r.ValidateMarshal(m.v, &v); err != nil {
		r.Mux.handleError(http.StatusBadRequest, err.Error(), w, r)
		return
	}
	m.h.Handle(x, w, r, v)
}
