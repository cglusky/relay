package pretty

import (
	"encoding/json"
)

func NewStringer(a any) *stringer {
	return &stringer{
		toString: a,
	}
}

type stringer struct {
	toString any
}

func (p *stringer) String() string {
	b, err := json.MarshalIndent(p.toString, "", "\t")
	if err != nil {
		return ""
	}
	return string(b)
}
