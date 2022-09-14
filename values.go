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

import "strconv"

type value string
type values map[string]value

// ErrEmptyValue is an error returned from number conversion functions when the
// string value is empty and does not represent a number.
const ErrEmptyValue = errStr("value is empty")

func (v value) String() string {
	return string(v)
}
func (v value) Bool() (bool, error) {
	if len(v) == 0 {
		return false, ErrEmptyValue
	}
	return strconv.ParseBool(string(v))
}
func (v value) Int() (int64, error) {
	if len(v) == 0 {
		return 0, ErrEmptyValue
	}
	return strconv.ParseInt(string(v), 10, 64)
}
func (v value) Uint() (uint64, error) {
	if len(v) == 0 {
		return 0, ErrEmptyValue
	}
	return strconv.ParseUint(string(v), 10, 64)
}
func (v value) Float() (float64, error) {
	if len(v) == 0 {
		return 0, ErrEmptyValue
	}
	return strconv.ParseFloat(string(v), 64)
}
func (v values) Bool(s string) (bool, error) {
	o, ok := v[s]
	if !ok {
		return false, &errValue{s: s, e: ErrNotExists}
	}
	return o.Bool()
}
func (v values) Int(s string) (int64, error) {
	o, ok := v[s]
	if !ok {
		return 0, &errValue{s: s, e: ErrNotExists}
	}
	return o.Int()
}
func (v values) Uint(s string) (uint64, error) {
	o, ok := v[s]
	if !ok {
		return 0, &errValue{s: s, e: ErrNotExists}
	}
	return o.Uint()
}
func (v values) String(s string) (string, error) {
	o, ok := v[s]
	if !ok {
		return "", &errValue{s: s, e: ErrNotExists}
	}
	return o.String(), nil
}
func (v values) Float(s string) (float64, error) {
	o, ok := v[s]
	if !ok {
		return 0, &errValue{s: s, e: ErrNotExists}
	}
	return o.Float()
}
func (v values) StringDefault(s, d string) string {
	o, ok := v[s]
	if !ok {
		return d
	}
	return o.String()
}
func (v values) BoolDefault(s string, d bool) bool {
	o, ok := v[s]
	if !ok {
		return d
	}
	if r, err := o.Bool(); err == nil {
		return r
	}
	return d
}
func (v values) IntDefault(s string, d int64) int64 {
	o, ok := v[s]
	if !ok {
		return d
	}
	if r, err := o.Int(); err == nil {
		return r
	}
	return d
}
func (v values) UintDefault(s string, d uint64) uint64 {
	o, ok := v[s]
	if !ok {
		return d
	}
	if r, err := o.Uint(); err == nil {
		return r
	}
	return d
}
func (v values) FloatDefault(s string, d float64) float64 {
	o, ok := v[s]
	if !ok {
		return d
	}
	if r, err := o.Float(); err == nil {
		return r
	}
	return d
}
