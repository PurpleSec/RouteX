package val

// Rules is an alias for a list of Rules that can be used to validate Content constraints.
type Rules []Rule

// Rule is an interface that can be used to assist with verifying the correct data passed to the Validator. The
// 'Validate' function will be passed the object in question and should return nil if it passes.
type Rule interface {
	Validate(interface{}) error
}

// ID is a ruleset that allows for identifying a ID value.
var ID = Rules{Integer, GreaterThanZero}
