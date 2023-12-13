package pretty

import (
	"encoding/json"
)

func NewStringer(a any) *stringer {
	return &stringer{
		ToString: a,
	}
}

type stringer struct {
	ToString any
}

func (p *stringer) String() string {
	b, err := json.MarshalIndent(p.ToString, "", "\t")
	if err != nil {
		return ""
	}
	return string(b)
}
