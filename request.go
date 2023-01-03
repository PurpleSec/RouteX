// Copyright 2021 - 2023 PurpleSec Team
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
	"io"
	"net/http"
)

// Request is an extension of the 'http.Request' struct.
//
// This struct includes parsed values from the calling URL and offers some convenience
// functions for parsing the resulting data.
type Request struct {
	Mux    *Mux
	ctx    context.Context
	Values values
	*http.Request
}

// Validator is an interface that allows for validation of Content data. By design,
// returning nil indicates that the supplied Content has passed all checks.
type Validator interface {
	Validate(Content) error
}

// IsGet returns true if this is a http GET request.
func (r *Request) IsGet() bool {
	return r.Method == http.MethodGet
}

// IsPut returns true if this is a http PUT request.
func (r *Request) IsPut() bool {
	return r.Method == http.MethodPut
}

// IsPost returns true if this is a http POST request.
func (r *Request) IsPost() bool {
	return r.Method == http.MethodPost
}

// IsHead returns true if this is a http HEAD request.
func (r *Request) IsHead() bool {
	return r.Method == http.MethodHead
}

// IsPatch returns true if this is a http PATCH request.
func (r *Request) IsPatch() bool {
	return r.Method == http.MethodPatch
}

// IsDelete returns true if this is a http DELETE request.
func (r *Request) IsDelete() bool {
	return r.Method == http.MethodDelete
}

// IsOptions returns true if this is a http OPTIONS request.
func (r *Request) IsOptions() bool {
	return r.Method == http.MethodOptions
}

// Marshal will attempt to unmarshal the JSON body in the Request into the supplied
// interface.
//
// This function returns 'ErrNoBody' if the Body is nil or empty.
//
// Any JSON parsing errors will also be returned if they occur.
func (r *Request) Marshal(i any) error {
	if r.Body == nil {
		return ErrNoBody
	}
	return json.NewDecoder(r.Body).Decode(&i)
}

// Context returns the request's context. The returned context is always non-nil.
//
// This is a child of the base Handler context if supplied on Mux creation
// and can be canceled if the Handler is closed or any timeout is passed.
func (r *Request) Context() context.Context {
	return r.ctx
}

// Content returns a content map based on the JSON body data passed in this request.
// This function returns 'ErrNoBody' if the Body is nil or empty.
//
// Any JSON parsing errors will also be returned if they occur.
func (r *Request) Content() (Content, error) {
	if r.Body == nil {
		return nil, ErrNoBody
	}
	var (
		c   Content
		err = json.NewDecoder(r.Body).Decode(&c)
	)
	if err == io.EOF {
		return c, nil
	}
	return c, err
}

// ValidateMarshal is similar to the Marshal function but will validate the Request
// content with the specified Validator before returning.
//
// This function returns 'ErrNoBody' if the Body is nil or empty.
//
// Any JSON parsing errors will also be returned if they occur.
func (r *Request) ValidateMarshal(v Validator, i any) error {
	if r.Body == nil {
		return ErrNoBody
	}
	var (
		c      Content
		b, err = io.ReadAll(r.Body)
	)
	if err != nil {
		return err
	}
	if len(b) == 0 {
		return ErrNoBody
	}
	if err = json.Unmarshal(b, &c); err != nil {
		return err
	}
	if v != nil {
		if err = v.Validate(c); err != nil {
			return err
		}
	}
	return json.Unmarshal(b, &i)
}

// ValidateContent returns a content map based on the JSON body data passed in this
// request.
//
// This function allows for passing a Validator that can also validate the content
// before returning.
//
// This will only validate if no JSON parsing errors are returned beforehand.
//
// This function will return 'ErrNoBody' if no content was found or the request
// body is empty.
func (r *Request) ValidateContent(v Validator) (Content, error) {
	c, err := r.Content()
	if err != nil {
		return c, err
	}
	if v == nil {
		return c, nil
	}
	return c, v.Validate(c)
}
