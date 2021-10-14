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

package val

import (
	"errors"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var (
	// NoEmpty adds a string constraint to ensure a string value cannot be empty.
	NoEmpty = &Length{Min: 1}

	errNotString = errors.New("value is not a string")
)

type regex struct {
	*regexp.Regexp
}

// Length adds a string  or array length constraint to be used. Max value is ignored if empty or less than Min.
type Length struct {
	Min, Max uint64
}
type strPrefix string
type strSuffix string
type strContains string

// Prefix returns a Rule that will verify that the value is a string and starts with the supplied string.
func Prefix(s string) Rule {
	return strPrefix(s)
}

// Suffix returns a Rule that will verify that the value is a string and ends with the supplied string.
func Suffix(s string) Rule {
	return strSuffix(s)
}

// Contains returns a Rule that will verify that the value is a string and contains the supplied string.
func Contains(s string) Rule {
	return strContains(s)
}

// MustRegex will return a regular expression validator. This function will panic if the expression has any
// errors compiling.
func MustRegex(s string) Rule {
	r, err := regexp.Compile(s)
	if err != nil {
		panic(err.Error())
	}
	return &regex{r}
}

// Regex will return a regular expression validator. This function will return an error if the expression has
// any errors compiling.
func Regex(s string) (Rule, error) {
	r, err := regexp.Compile(s)
	if err != nil {
		return nil, err
	}
	return &regex{r}, nil
}

// Validate fulfills the Rule interface.
func (l Length) Validate(i interface{}) error {
	var (
		v, ok = i.(string)
		x     uint64
	)
	if ok {
		x = uint64(len(v))
	} else {
		switch s := reflect.ValueOf(i); s.Kind() {
		case reflect.Map, reflect.Slice, reflect.Array:
			x = uint64(s.Len())
		default:
			return errNotString
		}
	}
	if x < l.Min {
		return errors.New("length " + strconv.FormatUint(x, 10) + " must be at least " + strconv.FormatUint(l.Min, 10))
	}
	if l.Min >= l.Max {
		return nil
	}
	if x > l.Max {
		return errors.New("length " + strconv.FormatUint(x, 10) + " cannot be more than " + strconv.FormatUint(l.Min, 10))
	}
	return nil
}
func (r *regex) Validate(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return errNotString
	}
	if r.MatchString(v) {
		return nil
	}
	return errors.New("string does not match expression '" + r.String() + "'")
}
func (s strPrefix) Validate(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return errNotString
	}
	if strings.HasPrefix(v, string(s)) {
		return nil
	}
	return errors.New("string does not have prefix '" + string(s) + "'")
}
func (s strSuffix) Validate(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return errNotString
	}
	if strings.HasSuffix(v, string(s)) {
		return nil
	}
	return errors.New("string does not have suffix '" + string(s) + "'")
}
func (s strContains) Validate(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return errNotString
	}
	if strings.Contains(v, string(s)) {
		return nil
	}
	return errors.New("string does not contain '" + string(s) + "'")
}
