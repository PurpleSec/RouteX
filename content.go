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

// Content is an alias of a JSON data payload sent to the server.
type Content map[string]interface{}

const (
	// ErrNoBody is an error returned when there is no content passed to a HTTP request when it's required.
	ErrNoBody = strErr("missing HTTP body")
	// ErrNotExists is an error returned from any of the Content getter functions when the value by the supplied name
	// does not exist in the Content map.
	ErrNotExists = strErr("value does not exist")
	// ErrInvalidType is a error returned from any of the Content getter functions when the value by the supplied name
	// is not the requested value.
	ErrInvalidType = strErr("incorrect value type")
)

// Raw returns the raw interface value with the supplied value name. This function returns nil if the name does not
// exist. This is similar to directory calling the name in the map.
func (c Content) Raw(s string) interface{} {
	return c[s]
}

// Int64 attempts to return the value with the provided name as an integer value. This function will return an
// 'ErrNotExists' error if the value by the specified name does not exist or 'ErrInvalidType' if the value does
// not represent an integer type.
func (c Content) Int64(s string) (int64, error) {
	r, err := c.Float64(s)
	if err != nil {
		return 0, err
	}
	return int64(r), nil
}

// Uint64 attempts to return the value with the provided name as an unsigned integer value. This function will
// return an 'ErrNotExists' error if the value by the specified name does not exist or 'ErrInvalidType' if the
// value does not represent an integer type.
func (c Content) Uint64(s string) (uint64, error) {
	r, err := c.Float64(s)
	if err != nil {
		return 0, err
	}
	return uint64(r), nil
}

// String attempts to return the value with the provided name as a string value. This function will
// return an 'ErrNotExists' error if the value by the specified name does not exist.
func (c Content) String(s string) (string, error) {
	v, ok := c[s]
	if !ok {
		return "", wrap(s, ErrNotExists)
	}
	r, ok := v.(string)
	if !ok {
		return "", wrap(s, ErrInvalidType)
	}
	return r, nil
}

// Object attempts to return the value with the provided name as a complex object value (wrapped as a Content alias).
// This function will return an 'ErrNotExists' error if the value by the specified name does not exist or
// 'ErrInvalidType' if the value does not represent an object type.
func (c Content) Object(s string) (Content, error) {
	v, ok := c[s]
	if !ok {
		return nil, wrap(s, ErrNotExists)
	}
	r, ok := v.(map[string]interface{})
	if !ok {
		return nil, wrap(s, ErrInvalidType)
	}
	return r, nil
}

// Float64 attempts to return the value with the provided name as an floating point value. This function will
// return an 'ErrNotExists' error if the value by the specified name does not exist or 'ErrInvalidType' if the
// value does not represent a float type.
func (c Content) Float64(s string) (float64, error) {
	v, ok := c[s]
	if !ok {
		return 0, wrap(s, ErrNotExists)
	}
	r, ok := v.(float64)
	if !ok {
		return 0, wrap(s, ErrInvalidType)
	}
	return r, nil
}
