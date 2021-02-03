package val

import (
	"strconv"
	"unsafe"

	"github.com/PurpleSec/routex"
)

const (
	// Float adds a number constraint to ensure a number does contain a decimal or a decimal of greater than zero.
	Float = number(false)
	// Integer adds a number constraint to ensure a number does not contain a decimal or a decimal of zero.
	Integer = number(true)

	// Positive adds a number constraint to ensure a number is zero or greater.
	Positive = polarity(true)
	// Negative adds a number constraint to ensure a number is less than zero.
	Negative = polarity(false)

	// GreaterThanZero adds a number constraint to ensure a number greater than zero.
	GreaterThanZero = Min(1)
)

var errNotNumber = routex.NewError("value is not a number")

// Min is a wrapper that will add a minimum number value constraint.
type Min float64
type number bool

// Max is a wrapper that will add a maximum number value constraint.
type Max float64
type polarity bool

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

// Validate fulfills the Rule interface.
func (m Max) Validate(i interface{}) error {
	x, ok := i.(float64)
	if !ok {
		return errNotNumber
	}
	if x > float64(m) {
		return routex.NewError("value " + strconv.FormatFloat(x, 'f', 0, 64) + " cannot be more than " + strconv.FormatFloat(float64(m), 'f', 0, 64))
	}
	return nil
}

// Validate fulfills the Rule interface.
func (m Min) Validate(i interface{}) error {
	x, ok := i.(float64)
	if !ok {
		return errNotNumber
	}
	if x < float64(m) {
		return routex.NewError("value " + strconv.FormatFloat(x, 'f', 0, 64) + " cannot be less than " + strconv.FormatFloat(float64(m), 'f', 0, 64))
	}
	return nil
}
func (n number) Validate(i interface{}) error {
	x, ok := i.(float64)
	if !ok {
		return errNotNumber
	}
	v, r := modf(x)
	if bool(n) && (v != x || r) {
		return routex.NewError("value " + strconv.FormatFloat(x, 'f', 2, 64) + " must an integer")
	}
	if !bool(n) && v == x && !r {
		return routex.NewError("value " + strconv.FormatFloat(x, 'f', 0, 64) + " must a float")
	}
	return nil
}
func (p polarity) Validate(i interface{}) error {
	x, ok := i.(float64)
	if !ok {
		return errNotNumber
	}
	if p && x < 0 {
		return routex.NewError("value " + strconv.FormatFloat(x, 'f', 0, 64) + " must be positive")
	}
	if !p && x >= 0 {
		return routex.NewError("value " + strconv.FormatFloat(x, 'f', 0, 64) + " must be negative")
	}
	return nil
}
