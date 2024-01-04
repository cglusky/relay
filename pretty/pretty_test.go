package pretty

import (
	"reflect"
	"testing"
)

func TestNewStringer(t *testing.T) {
	testCases := []struct {
		name  string
		input any
	}{
		{
			name:  "Test with string",
			input: "test",
		},
		{
			name:  "Test with integer",
			input: 123,
		},
		{
			name:  "Test with slice",
			input: []int{1, 2, 3},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewStringer(tc.input)
			if !reflect.DeepEqual(s.toString, tc.input) {
				t.Errorf("Expected %v, but got %v", tc.input, s.toString)
			}
		})
	}
}

func TestStringerString(t *testing.T) {
	testCases := []struct {
		name     string
		input    any
		expected string
	}{
		{
			name:     "Test with string",
			input:    "test",
			expected: "\"test\"",
		},
		{
			name:     "Test with integer",
			input:    123,
			expected: "123",
		},
		{
			name:     "Test with slice",
			input:    []int{1, 2, 3},
			expected: "[\n\t1,\n\t2,\n\t3\n]",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewStringer(tc.input)
			result := s.String()
			if result != tc.expected {
				t.Errorf("Expected %s, but got %s", tc.expected, result)
			}
		})
	}
}
