package pretty

import (
	"encoding/json"
)

func NewStringer(a any) *prettierString {
	return &prettierString{
		ToString: a,
	}
}

type prettierString struct {
	ToString any
}

func (p *prettierString) String() string {
	b, err := json.MarshalIndent(p.ToString, "", "\t")
	if err != nil {
		return ""
	}
	return string(b)
}
