package pretty

import (
	"encoding/json"
	"fmt"
)

type Prettier interface {
	Stringer(any) (string, error)
	Printer(any) error
}

func Stringer(a any) (string, error) {
	s, err := json.MarshalIndent(a, "", "\t")
	if err != nil {
		return "", err
	}
	return string(s), nil
}

func Printer(a any) error {
	s, err := json.MarshalIndent(a, "", "\t")
	if err != nil {
		return err
	}

	fmt.Println(string(s))
	return nil
}
