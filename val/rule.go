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

package val

// Rules is an alias for a list of Rules that can be used to validate Content
// constraints.
type Rules []Rule

// Rule is an interface that can be used to assist with verifying the correct data
// passed to the Validator.
//
// The 'Validate' function will be passed the object in question and should return
// nil if it passes.
type Rule interface {
	Validate(any) error
}

// ID is a ruleset that allows for identifying a ID value.
var ID = Rules{Integer, GreaterThanZero}
