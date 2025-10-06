package maps

import (
	"reflect"
	"testing"
)

func TestKeys(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]int
		expected []string
	}{
		{
			name:     "empty map",
			input:    map[string]int{},
			expected: []string{},
		},
		{
			name:     "single element",
			input:    map[string]int{"a": 1},
			expected: []string{"a"},
		},
		{
			name:     "multiple elements",
			input:    map[string]int{"a": 1, "b": 2, "c": 3},
			expected: []string{"a", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Keys(tt.input)

			if len(result) != len(tt.expected) {
				t.Fatalf(
					"Keys() returned %d elements, want %d",
					len(result),
					len(tt.expected),
				)
			}

			// Since Keys returns in no particular order, we need to
			// check containment.
			resultMap := make(map[string]bool)
			for _, k := range result {
				resultMap[k] = true
			}

			for _, expected := range tt.expected {
				if !resultMap[expected] {
					t.Errorf("Keys() missing expected key %q", expected)
				}
			}
		})
	}
}

func TestSortedKeys(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]int
		expected []string
	}{
		{
			name:     "empty map",
			input:    map[string]int{},
			expected: []string{},
		},
		{
			name:     "single element",
			input:    map[string]int{"a": 1},
			expected: []string{"a"},
		},
		{
			name:     "multiple elements",
			input:    map[string]int{"c": 3, "a": 1, "b": 2},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "already sorted",
			input:    map[string]int{"a": 1, "b": 2, "c": 3},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "reverse order",
			input:    map[string]int{"z": 3, "y": 2, "x": 1},
			expected: []string{"x", "y", "z"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SortedKeys(tt.input)

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("SortedKeys() = %v, want %v", result, tt.expected)
			}
		})
	}
}
