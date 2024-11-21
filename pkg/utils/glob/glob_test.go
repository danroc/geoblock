package glob_test

import (
	"testing"

	"github.com/danroc/geoblock/pkg/utils/glob"
)

func TestStar(t *testing.T) {
	tests := []struct {
		pattern string
		s       string
		want    bool
	}{
		{"", "", true},
		{"*", "", true},
		{"", "abc", false},
		{"*", "abc", true},
		{"a*", "abc", true},
		{"*c", "abc", true},
		{"*b*", "abc", true},
		{"a*c", "abc", true},
		{"a*b*c", "abc", true},
		{"a*d", "abc", false},
		{"a*b*d", "abc", false},
		{"abc", "abc", true},
		{"abc*", "abc", true},
		{"*abc", "abc", true},
		{"*a*b*c*", "abc", true},
		{"*a*b*d*", "abc", false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.s, func(t *testing.T) {
			if got := glob.Star(tt.pattern, tt.s); got != tt.want {
				t.Errorf(
					"Star(%q, %q) = %v, want %v",
					tt.pattern,
					tt.s,
					got,
					tt.want,
				)
			}
		})
	}
}
