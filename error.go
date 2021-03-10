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

type err struct {
	e error
	s string
}
type strErr string

func (e err) Error() string {
	return e.s
}
func (e err) Unwrap() error {
	return e.e
}
func (e err) String() string {
	return e.s
}

// NewError creates a new string backed error struct and returns it. This error struct does not support Unwrapping.
// The resulting structs created will be comparable.
func NewError(s string) error {
	return strErr(s)
}
func (e strErr) Error() string {
	return string(e)
}
func (e strErr) String() string {
	return string(e)
}
func wrap(s string, e error) error {
	if e != nil {
		return &err{s: s + ": " + e.Error(), e: e}
	}
	return &err{s: s}
}
