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
	"strconv"
	"unsafe"
)

const (
	// Float adds a number constraint to ensure a number does contain a decimal
	// or a decimal of greater than zero.
	Float = number(false)
	// Integer adds a number constraint to ensure a number does not contain a
	// decimal or a decimal of zero.
	Integer = number(true)

	// Positive adds a number constraint to ensure a number is zero or greater.
	Positive = polarity(true)
	// Negative adds a number constraint to ensure a number is less than zero.
	Negative = polarity(false)

	// GreaterThanZero adds a number constraint to ensure a number greater than
	//zero.
	GreaterThanZero = Min(1)
)

var errNotNumber = errors.New("value is not a number")

// Min is a wrapper that will add a minimum number value constraint.
type Min float64

// Max is a wrapper that will add a maximum number value constraint.
type Max float64
type number bool
type polarity bool

// Validate fulfills the Rule interface.
func (m Max) Validate(i any) error {
	x, ok := i.(float64)
	if !ok {
		return errNotNumber
	}
	if x > float64(m) {
		return errors.New("value " + strconv.FormatFloat(x, 'f', 0, 64) + " cannot be more than " + strconv.FormatFloat(float64(m), 'f', 0, 64))
	}
	return nil
}

// Validate fulfills the Rule interface.
func (m Min) Validate(i any) error {
	x, ok := i.(float64)
	if !ok {
		return errNotNumber
	}
	if x < float64(m) {
		return errors.New("value " + strconv.FormatFloat(x, 'f', 0, 64) + " cannot be less than " + strconv.FormatFloat(float64(m), 'f', 0, 64))
	}
	return nil
}
func modf(f float64) (float64, bool) {
	var (
		i = *(*uint64)(unsafe.Pointer(&f))
		e = uint64(i>>52)&0x7FF - 1023
	)
	if e < 52 {
		i &^= 1<<(52-e) - 1
	}
	r := *(*float64)(unsafe.Pointer(&i))
	return r, (f - r) > 0
}
func (n number) Validate(i any) error {
	x, ok := i.(float64)
	if !ok {
		return errNotNumber
	}
	v, r := modf(x)
	if bool(n) && (v != x || r) {
		return errors.New("value " + strconv.FormatFloat(x, 'f', 2, 64) + " must an integer")
	}
	if !bool(n) && v == x && !r {
		return errors.New("value " + strconv.FormatFloat(x, 'f', 0, 64) + " must a float")
	}
	return nil
}
func (p polarity) Validate(i any) error {
	x, ok := i.(float64)
	if !ok {
		return errNotNumber
	}
	if p && x < 0 {
		return errors.New("value " + strconv.FormatFloat(x, 'f', 0, 64) + " must be positive")
	}
	if !p && x >= 0 {
		return errors.New("value " + strconv.FormatFloat(x, 'f', 0, 64) + " must be negative")
	}
	return nil
}
