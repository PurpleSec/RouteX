package routex

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
)

// Request is an extension of the 'http.Request' struct. This struct contains extra data, including the caller
// route name and any parsed values from the calling URL. This struct is to be used in any Handler instances.
type Request struct {
	Route  string
	Values values

	ctx context.Context
	*http.Request
}
type requestValue string
type values map[string]requestValue

// Validator is an interface that allows for validation of Content data. By design, returning nil indicates
// that the supplied Content has passed.
type Validator interface {
	Validate(Content) error
}

// ErrEmptyValue is a error returned from number conversion functions when the string value is empty and cannot
// be converted to a number.
const ErrEmptyValue = strErr("value is empty")

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

// IsDelete returns true if this is a http DELETE request.
func (r *Request) IsDelete() bool {
	return r.Method == http.MethodDelete
}

// IsOptions returns true if this is a http OPTIONS request.
func (r *Request) IsOptions() bool {
	return r.Method == http.MethodOptions
}
func (r requestValue) String() string {
	return string(r)
}
func (v values) Raw(s string) interface{} {
	return v[s]
}

// Context returns the request's context. The returned context is always non-nil. This is a child of the base Handler
// context and cab be cancled if the Handler is closed or any timeout is passed.
func (r *Request) Context() context.Context {
	return r.ctx
}

// Content returns a content map based on the JSO body data passed in this request. The resulting Content may be
// nil if the body is empty. Any parsing errors will also be returned.
func (r *Request) Content() (Content, error) {
	if r.Body == nil {
		return nil, nil
	}
	var (
		c   Content
		err = json.NewDecoder(r.Body).Decode(&c)
	)
	if err == io.EOF {
		return nil, nil
	}
	return c, err
}
func (r requestValue) Int64() (int64, error) {
	if len(r) == 0 {
		return 0, ErrEmptyValue
	}
	return strconv.ParseInt(string(r), 10, 64)
}
func (r requestValue) Uint64() (uint64, error) {
	if len(r) == 0 {
		return 0, ErrEmptyValue
	}
	return strconv.ParseUint(string(r), 10, 64)
}
func (v values) Int64(s string) (int64, error) {
	o, ok := v[s]
	if !ok {
		return 0, WrapError(s, ErrNotExists)
	}
	return o.Int64()
}
func (v values) Uint64(s string) (uint64, error) {
	o, ok := v[s]
	if !ok {
		return 0, WrapError(s, ErrNotExists)
	}
	return o.Uint64()
}
func (v values) String(s string) (string, error) {
	o, ok := v[s]
	if !ok {
		return "", WrapError(s, ErrNotExists)
	}
	return o.String(), nil
}
func (r requestValue) Float64() (float64, error) {
	if len(r) == 0 {
		return 0, ErrEmptyValue
	}
	return strconv.ParseFloat(string(r), 64)
}
func (v values) Float64(s string) (float64, error) {
	o, ok := v[s]
	if !ok {
		return 0, WrapError(s, ErrNotExists)
	}
	return o.Float64()
}

// ContentValidate returns a content map based on the JSO body data passed in this request. The resulting Content may be
// nil if the body is empty. Any parsing errors will also be returned. This function allows for passing a Set that can
// also validate the content before returning. This will only validate if no errors are returned beforehand.
// This function will return 'ErrNoBody' if no content was found.
func (r *Request) ContentValidate(v Validator) (Content, error) {
	c, err := r.Content()
	if v == nil || err != nil {
		return c, err
	}
	if c == nil {
		return nil, ErrNoBody
	}
	return c, v.Validate(c)
}
