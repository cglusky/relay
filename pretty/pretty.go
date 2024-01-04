// Package pretty provides utilities for pretty-printing any value that can be marshalled to JSON.
package pretty

import (
	"encoding/json"
)

// NewStringer creates a new stringer instance with the given value.
// The value will be converted to a string representation using JSON marshaling.
func NewStringer(a any) stringer {
	return stringer{
		toString: a,
	}
}

// stringer is a struct that holds a value to be converted to a string.
type stringer struct {
	toString any
}

// String returns the string representation of the value stored in the stringer.
// The value is marshaled to JSON with indentation for readability.
// If an error occurs during marshaling, an empty string is returned.
func (p stringer) String() string {
	b, err := json.MarshalIndent(p.toString, "", "\t")
	if err != nil {
		return ""
	}
	return string(b)
}
