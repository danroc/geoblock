package glob_test

import (
	"testing"

	"github.com/danroc/geoblock/internal/utils/glob"
)

func TestMatchFold(t *testing.T) {
	tests := []struct {
		pattern string
		s       string
		want    bool
	}{
		{"", "", true},
		{"*", "", true},
		{"a", "", false},
		{"", "abc", false},
		{"*", "abc", true},
		{"abc", "abc", true},
		{"a*", "abc", true},
		{"*c", "abc", true},
		{"a*c", "abc", true},
		{"a*b*c", "abc", true},
		{"a*d", "abc", false},
		{"ABC", "abc", true},
		{"A*C", "abc", true},
		{"*A*B*C*", "XaYbZc", true},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.s, func(t *testing.T) {
			if got := glob.MatchFold(tt.pattern, tt.s); got != tt.want {
				t.Errorf(
					"MatchFold(%q, %q) = %v, want %v", tt.pattern, tt.s, got, tt.want,
				)
			}
		})
	}
}
