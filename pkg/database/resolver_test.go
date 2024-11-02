package database_test

import (
	"testing"

	"github.com/danroc/geoblock/pkg/database"
)

func TestStrIndex(t *testing.T) {
	tests := []struct {
		data     []string
		index    int
		expected string
	}{
		{[]string{"a", "b", "c"}, 0, "a"},
		{[]string{"a", "b", "c"}, 1, "b"},
		{[]string{"a", "b", "c"}, 2, "c"},
		{[]string{"a", "b", "c"}, 3, ""},
		{[]string{"a", "b", "c"}, -1, ""},
		{[]string{}, 0, ""},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := database.StrIndex(tt.data, tt.index)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestStrToASN(t *testing.T) {
	tests := []struct {
		input    string
		expected uint32
	}{
		{"12345", 12345},
		{"0", 0},
		{"4294967295", 4294967295},
		{"invalid", database.ReservedAS0},
		{"", database.ReservedAS0},
		{"-1", database.ReservedAS0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := database.StrToASN(tt.input)
			if result != tt.expected {
				t.Errorf("got %d, want %d", result, tt.expected)
			}
		})
	}
}
