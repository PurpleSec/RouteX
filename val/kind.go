package val

const (
	// Any is the default type value. This means the Validator will accept any object type and just confirms it exists.
	Any kind = iota
	// None is the opposite of the Any type. This will instruct the validator to ensure the value is null
	// or non-existant.
	None
	// Number represents a type of an integer or float value. These are stored as float64 values.
	Number
	// Int represents a type of integer value. Similar to having the Rule 'Integer'.
	Int
	// String represents a string type value.
	String
	// Bool is a type that is either true or false.
	Bool
	// Object is a type that can be used to ensure the result is a complex map or non-list type.
	Object
	// List is a type that will represent a generic list/array/slice input.
	List
	// ListNumber is a type that goes further than List and ensures that all the list entries are valid numbers.
	ListNumber
	// ListString is a type that goes further than List and ensures that all the list entries are valid strings.
	ListString
)

type kind uint8

func (k kind) String() string {
	switch k {
	case Any:
		return "any"
	case None:
		return "null"
	case Number:
		return "number"
	case Int:
		return "integer"
	case String:
		return "string"
	case Bool:
		return "boolean"
	case Object:
		return "object"
	case List:
		return "[]object"
	case ListNumber:
		return "[]number"
	case ListString:
		return "[]string"
	}
	return "invalid"
}
