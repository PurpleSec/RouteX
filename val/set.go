package val

import (
	"reflect"

	"github.com/PurpleSec/routex"
)

// SubSet is a type if Set that can be used to validate a complex object that is a child of an object being validated.
// SubSets have the same options as a Set.
type SubSet Set

// Set is an alias for a list of Validators that can be used to validate a request data body.
type Set []Validator

// Validator is a struct used to assist with data body validation. This struct can be used inside a Set to add rules
// for incoming data. The Rules attribute can be used to add more constraints on the Validator.
type Validator struct {
	Name     string `json:"name"`
	Rules    Rules  `json:"rules"`
	Type     kind   `json:"type"`
	Optional bool   `json:"optional,omitempty"`
}

// ErrInvalidName is a validation error returned when a Validator rule has an empty name.
var ErrInvalidName = routex.NewError("invalid name in set")

// Validate fulfills the Rule interface.
func (s SubSet) Validate(i interface{}) error {
	m, ok := i.(map[string]interface{})
	if !ok {
		return routex.NewError("type '" + reflect.TypeOf(i).String() + "' is not valid for SubSets")
	}
	return validate(s, m)
}

// Validate will check the rules of this Set against the supplied content object. This function will return nil if
// the Content is considered valid.
func (s Set) Validate(c routex.Content) error {
	return validate(s, c)
}

// Validate will attempt to validate a single validation rule and return an error if the supplied interface does not
// match the Validator's constraints.
func (v Validator) Validate(i interface{}) error {
	if i == nil && v.Type > None {
		return routex.NewError("'" + v.Name + "': expected '" + v.Type.String() + "' but got 'null'")
	}
	if v.Type > None {
		switch t := i.(type) {
		case bool:
			if v.Type == Bool {
				break
			}
			return routex.NewError("'" + v.Name + "': expected '" + v.Type.String() + "' but got 'boolean'")
		case string:
			if v.Type == String {
				break
			}
			return routex.NewError("'" + v.Name + "': expected '" + v.Type.String() + "' but got 'string'")
		case float64:
			if v.Type == Number {
				break
			}
			if v.Type == Int {
				if n, r := modf(t); t != n || r {
					return routex.NewError("'" + v.Name + "': expected 'integer' but got 'float'")
				}
				break
			}
			return routex.NewError("'" + v.Name + "': expected '" + v.Type.String() + "' but got 'number'")
		default:
			k := reflect.ValueOf(i).Kind()
			if k == reflect.Map && v.Type != Object {
				return routex.NewError("'" + v.Name + "': expected 'object' but got '" + reflect.TypeOf(i).String() + "'")
			}
			if k == reflect.Slice && v.Type < List {
				return routex.NewError("'" + v.Name + "': expected '[]object' but got '" + reflect.TypeOf(i).String() + "'")
			}
			if v.Type > List {
				w, ok := i.([]interface{})
				if !ok {
					return routex.NewError("'" + v.Name + "': '[]object' value could not be parsed")
				}
				for e := range w {
					if v.Type == ListNumber {
						if _, ok := w[e].(float64); ok {
							continue
						}
						return routex.NewError("'" + v.Name + "': '[]number' contains invalid entry")
					}
					if _, ok := w[e].(string); ok {
						continue
					}
					return routex.NewError("'" + v.Name + "': '[]string' contains invalid entry")
				}
			}
		}
	}
	for x := range v.Rules {
		if err := v.Rules[x].Validate(i); err != nil {
			return routex.NewError("'" + v.Name + "': " + err.Error())
		}
	}
	return nil
}

// ValidateEmpty will check the rules of this Set against the supplied content object. This function will return nil if
// the Content is considered valid or if the Content is empty. This function allows for specifying and validating
// optional data.
func (s Set) ValidateEmpty(c routex.Content) error {
	if len(c) == 0 {
		return nil
	}
	return validate(s, c)
}
func validate(s []Validator, m map[string]interface{}) error {
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
			return routex.NewError("'" + s[x].Name + "': required")
		}
		if err := s[x].Validate(i); err != nil {
			return err
		}
	}
	return nil
}
