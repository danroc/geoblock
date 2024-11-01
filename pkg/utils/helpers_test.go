package utils_test

import (
	"testing"

	"github.com/danroc/geoblock/pkg/utils"
)

func TestAny(t *testing.T) {
	tests := []struct {
		name   string
		values []int
		f      func(int) bool
		want   bool
	}{
		{"empty slice", []int{}, func(v int) bool { return v > 0 }, false},
		{"no match", []int{1, 2, 3}, func(v int) bool { return v > 3 }, false},
		{"one match", []int{1, 2, 3}, func(v int) bool { return v == 2 }, true},
		{"all match", []int{1, 2, 3}, func(v int) bool { return v > 0 }, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.Any(tt.values, tt.f); got != tt.want {
				t.Errorf("Any() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAll(t *testing.T) {
	tests := []struct {
		name   string
		values []int
		f      func(int) bool
		want   bool
	}{
		{"empty slice", []int{}, func(v int) bool { return v > 0 }, true},
		{"no match", []int{1, 2, 3}, func(v int) bool { return v > 3 }, false},
		{"one match", []int{1, 2, 3}, func(v int) bool { return v == 2 }, false},
		{"all match", []int{1, 2, 3}, func(v int) bool { return v > 0 }, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.All(tt.values, tt.f); got != tt.want {
				t.Errorf("All() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNone(t *testing.T) {
	tests := []struct {
		name   string
		values []int
		f      func(int) bool
		want   bool
	}{
		{"empty slice", []int{}, func(v int) bool { return v > 0 }, true},
		{"no match", []int{1, 2, 3}, func(v int) bool { return v > 3 }, true},
		{"one match", []int{1, 2, 3}, func(v int) bool { return v == 2 }, false},
		{"all match", []int{1, 2, 3}, func(v int) bool { return v > 0 }, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.None(tt.values, tt.f); got != tt.want {
				t.Errorf("None() = %v, want %v", got, tt.want)
			}
		})
	}
}
