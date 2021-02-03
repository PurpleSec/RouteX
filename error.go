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

// WrapError creates a new error that wraps the specified error. If not nil, this function will append ": " + 'Error()'
// to the resulting string message.
func WrapError(s string, e error) error {
	if e != nil {
		return &err{s: s + ": " + e.Error(), e: e}
	}
	return &err{s: s}
}
