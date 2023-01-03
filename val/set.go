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

package val

import (
	"errors"
	"reflect"

	"github.com/PurpleSec/routex"
)

// SubSet is a type if Set that can be used to validate a complex object that is
// a child of an object being validated.
//
// SubSets have the same options as a Set.
type SubSet Set

// Set is an alias for a list of Validators that can be used to validate a request
// data body.
type Set []Validator

// Validator is a struct used to assist with data body validation. This struct can
// be used inside a Set to add rules for incoming data.
//
// The Rules attribute can be used to add more constraints on the Validator.
type Validator struct {
	Name     string `json:"name"`
	Rules    Rules  `json:"rules"`
	Type     kind   `json:"type"`
	Optional bool   `json:"optional,omitempty"`
}

// ErrInvalidName is a validation error returned when a Validator rule has an empty
// name.
var ErrInvalidName = errors.New("invalid name in set")

// Validate fulfills the Rule interface.
func (s SubSet) Validate(i any) error {
	m, ok := i.(map[string]any)
	if !ok {
		return errors.New("type '" + reflect.TypeOf(i).String() + "' is not valid for SubSets")
	}
	return validate(s, m)
}

// Validate will attempt to validate a single validation rule and return an error
// if the supplied interface does not match the Validator's constraints.
func (v Validator) Validate(i any) error {
	if i == nil && v.Type > None {
		return errors.New("'" + v.Name + "': expected '" + v.Type.String() + "' but got 'null'")
	}
	if v.Type > None {
		switch t := i.(type) {
		case bool:
			if v.Type == Bool {
				break
			}
			return errors.New("'" + v.Name + "': expected '" + v.Type.String() + "' but got 'boolean'")
		case string:
			if v.Type == String {
				break
			}
			return errors.New("'" + v.Name + "': expected '" + v.Type.String() + "' but got 'string'")
		case float64:
			if v.Type == Number {
				break
			}
			if v.Type == Int {
				if n, r := modf(t); t != n || r {
					return errors.New("'" + v.Name + "': expected 'integer' but got 'float'")
				}
				break
			}
			return errors.New("'" + v.Name + "': expected '" + v.Type.String() + "' but got 'number'")
		default:
			k := reflect.ValueOf(i).Kind()
			if k == reflect.Map && v.Type != Object {
				return errors.New("'" + v.Name + "': expected 'object' but got '" + reflect.TypeOf(i).String() + "'")
			}
			if k == reflect.Slice && v.Type < List {
				return errors.New("'" + v.Name + "': expected '[]object' but got '" + reflect.TypeOf(i).String() + "'")
			}
			if v.Type > List {
				w, ok := i.([]any)
				if !ok {
					return errors.New("'" + v.Name + "': '[]object' value could not be parsed")
				}
				for e := range w {
					if v.Type == ListNumber {
						if _, ok := w[e].(float64); ok {
							continue
						}
						return errors.New("'" + v.Name + "': '[]number' contains invalid entry")
					}
					if _, ok := w[e].(string); ok {
						continue
					}
					return errors.New("'" + v.Name + "': '[]string' contains invalid entry")
				}
			}
		}
	}
	for x := range v.Rules {
		if err := v.Rules[x].Validate(i); err != nil {
			return errors.New("'" + v.Name + "': " + err.Error())
		}
	}
	return nil
}

// Validate will check the rules of this Set against the supplied content object.
// This function will return nil if the Content is considered valid.
func (s Set) Validate(c routex.Content) error {
	return validate(s, c)
}

// ValidateEmpty will check the rules of this Set against the supplied content
// object.
//
// This function will return nil if the Content is considered valid or if the
// Content is empty. This function allows for specifying and validating optional
// data.
func (s Set) ValidateEmpty(c routex.Content) error {
	if len(c) == 0 {
		return nil
	}
	return validate(s, c)
}
func validate(s []Validator, m routex.Content) error {
	if len(s) == 0 {
		return nil
	}
	for x := range s {
		if len(s[x].Name) == 0 {
			return ErrInvalidName
		}
		i, ok := m[s[x].Name]
		if !ok {
			if s[x].Type == None || s[x].Optional {
				continue
			}
			return errors.New("'" + s[x].Name + "': required")
		}
		if err := s[x].Validate(i); err != nil {
			return err
		}
	}
	return nil
}
