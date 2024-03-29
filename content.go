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

import "encoding/base64"

// Content is an alias of a JSON data payload sent to the server.
type Content map[string]any

const (
	// ErrNoBody is an error returned when there is no content passed to an HTTP
	// request when it's required.
	ErrNoBody = errStr("missing HTTP body")
	// ErrNotExists is an error returned from any of the Content getter functions
	// when the value by the supplied name does not exist in the Content map.
	ErrNotExists = errStr("value does not exist")
	// ErrInvalidType is an error returned from any of the Content getter functions
	// when the value by the supplied name is not the requested value.
	ErrInvalidType = errStr("incorrect value type")
)

// Raw returns the raw interface value with the supplied value name.
//
// This function returns nil if the name does not exist. This is similar to
// directly calling the name in a map.
func (c Content) Raw(s string) any {
	return c[s]
}

// Bool attempts to return the value with the provided name as a boolean value.
//
// This function will return an 'ErrNotExists' error if the value by the specified
// name does not exist or 'ErrInvalidType' if the value does not represent a boolean
// type.
func (c Content) Bool(s string) (bool, error) {
	v, ok := c[s]
	if !ok {
		return false, &errValue{s: s, e: ErrNotExists}
	}
	r, ok := v.(bool)
	if !ok {
		return false, &errValue{s: s, e: ErrInvalidType}
	}
	return r, nil
}

// Int attempts to return the value with the provided name as an integer value.
//
// This function will return an 'ErrNotExists' error if the value by the specified
// name does not exist or 'ErrInvalidType' if the value does not represent an integer
// type.
func (c Content) Int(s string) (int64, error) {
	r, err := c.Float(s)
	if err != nil {
		return 0, err
	}
	return int64(r), nil
}

// Uint attempts to return the value with the provided name as an unsigned integer
// value.
//
// This function will return an 'ErrNotExists' error if the value by the specified
// name does not exist or 'ErrInvalidType' if the value does not represent an integer
// type.
func (c Content) Uint(s string) (uint64, error) {
	r, err := c.Float(s)
	if err != nil {
		return 0, err
	}
	return uint64(r), nil
}

// Bytes attempts to return the value with the provided name as a byte slice value
// that is represented by a Base64-encoded string.
//
// This function will return an 'ErrNotExists' error if the value by the specified
// name does not exist or 'ErrInvalidType' if the value does not represent a bytes
// type.
//
// This will attempt to decode the Base64 string and will return the encoding
// errors if they occur.
func (c Content) Bytes(s string) ([]byte, error) {
	v, ok := c[s]
	if !ok {
		return nil, &errValue{s: s, e: ErrNotExists}
	}
	r, ok := v.(string)
	if !ok {
		return nil, &errValue{s: s, e: ErrInvalidType}
	}
	return base64.StdEncoding.DecodeString(r)
}

// String attempts to return the value with the provided name as a string value.
//
// This function will return an 'ErrNotExists' error if the value by the specified
// name does not exist or 'ErrInvalidType' if the value does not represent a string
// type.
func (c Content) String(s string) (string, error) {
	v, ok := c[s]
	if !ok {
		return "", &errValue{s: s, e: ErrNotExists}
	}
	r, ok := v.(string)
	if !ok {
		return "", &errValue{s: s, e: ErrInvalidType}
	}
	return r, nil
}

// Float attempts to return the value with the provided name as a floating point
// value.
//
// This function will return an 'ErrNotExists' error if the value by the specified
// name does not exist or 'ErrInvalidType' if the value does not represent a float
// type.
func (c Content) Float(s string) (float64, error) {
	v, ok := c[s]
	if !ok {
		return 0, &errValue{s: s, e: ErrNotExists}
	}
	r, ok := v.(float64)
	if !ok {
		return 0, &errValue{s: s, e: ErrInvalidType}
	}
	return r, nil
}

// StringDefault attempts to return the value with the provided name as a string
// value.
//
// This function will return the default value specified if the value does not exist
// or is not a string type.
func (c Content) StringDefault(s, d string) string {
	v, ok := c[s]
	if !ok {
		return d
	}
	r, ok := v.(string)
	if !ok {
		return d
	}
	return r
}

// Object attempts to return the value with the provided name as a complex object
// value (wrapped as a Content alias).
//
// This function will return an 'ErrNotExists' error if the value by the specified
// name does not exist or 'ErrInvalidType' if the value does not represent an object
// type.
func (c Content) Object(s string) (Content, error) {
	v, ok := c[s]
	if !ok {
		return nil, &errValue{s: s, e: ErrNotExists}
	}
	r, ok := v.(map[string]any)
	if !ok {
		return nil, &errValue{s: s, e: ErrInvalidType}
	}
	return r, nil
}

// BoolDefault attempts to return the value with the provided name as a boolean
// value.
//
// This function will return the default value specified if the value does not exist
// or is not a boolean type.
func (c Content) BoolDefault(s string, d bool) bool {
	v, ok := c[s]
	if !ok {
		return d
	}
	r, ok := v.(bool)
	if !ok {
		return d
	}
	return r
}

// IntDefault attempts to return the value with the provided name as an integer
// value.
//
// This function will return the default value specified if the value does not exist
// or is not an integer type.
func (c Content) IntDefault(s string, d int64) int64 {
	r, err := c.Float(s)
	if err != nil {
		return d
	}
	return int64(r)
}

// BytesEmpty attempts to return the value with the provided name as a byte slice
// value that is represented by a Base64-encoded string.
//
// This function will return an 'ErrInvalidType' if the value does not represent
// a bytes type. Empty or missing values will simply return none.
//
// This function is different than the other 'Bytes' function as it allows for
// empty/missing byte slices but not invalid or improperly formatted ones.
//
// This will attempt to decode the Base64 string and will return the encoding
// errors if they occur.
func (c Content) BytesEmpty(s string) ([]byte, error) {
	v, ok := c[s]
	if !ok {
		return nil, nil
	}
	r, ok := v.(string)
	if !ok {
		return nil, &errValue{s: s, e: ErrInvalidType}
	}
	if len(r) == 0 {
		return nil, nil
	}
	return base64.StdEncoding.DecodeString(r)
}

// UintDefault attempts to return the value with the provided name as an unsigned
// integer value.
//
// This function will return the default value specified if the value does not exist
// or is not an unsigned integer type.
func (c Content) UintDefault(s string, d uint64) uint64 {
	r, err := c.Float(s)
	if err != nil {
		return d
	}
	return uint64(r)
}

// BytesDefault to return the value with the provided name as a byte slice value
// that is represented by a Base64-encoded string.
//
// This function will return the default value specified if the value does not exist
// or is not a bytes type.
//
// This will attempt to decode the Base64 string and will return the default value
// if errors occur.
func (c Content) BytesDefault(s string, d []byte) []byte {
	v, ok := c[s]
	if !ok {
		return d
	}
	r, ok := v.(string)
	if !ok {
		return d
	}
	if b, err := base64.StdEncoding.DecodeString(r); err == nil {
		return b
	}
	return d
}

// FloatDefault attempts to return the value with the provided name as a floating
// point value.
//
// This function will return the default value specified if the value does not exist
// or is not a float type.
func (c Content) FloatDefault(s string, d float64) float64 {
	v, ok := c[s]
	if !ok {
		return d
	}
	r, ok := v.(float64)
	if !ok {
		return d
	}
	return r
}

// ObjectDefault attempts to return the value with the provided name as an object
// value (wrapped as a Content alias).
//
// This function will return the default value specified if the value does not exist
// or is not an object type.
func (c Content) ObjectDefault(s string, d Content) Content {
	v, ok := c[s]
	if !ok {
		return d
	}
	r, ok := v.(map[string]any)
	if !ok {
		return d
	}
	return r
}
