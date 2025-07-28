package version

import (
	"testing"
)

func TestGet(t *testing.T) {
	// Save original version for cleanup
	originalVersion := Version
	t.Cleanup(func() {
		Version = originalVersion
	})

	tests := []struct {
		name     string
		version  string
		expected string
	}{
		{
			name:     "semantic version with v prefix",
			version:  "v1.0.0",
			expected: "1.0.0",
		},
		{
			name:     "dev version",
			version:  "dev",
			expected: "dev",
		},
		{
			name:     "git describe style version with v prefix",
			version:  "v1.0.0-g1234567",
			expected: "1.0.0-g1234567",
		},
		{
			name:     "dirty version gets dev suffix and v prefix removed",
			version:  "v1.0.0-g1234567-dirty",
			expected: "1.0.0-g1234567-dev",
		},
		{
			name:     "simple dirty version with v prefix removed",
			version:  "v1.0.0-dirty",
			expected: "1.0.0-dev",
		},
		{
			name:     "version without v prefix remains unchanged",
			version:  "1.2.3",
			expected: "1.2.3",
		},
		{
			name:     "v followed by non-digit is preserved",
			version:  "version1.0.0",
			expected: "version1.0.0",
		},
		{
			name:     "v prefix with major version only",
			version:  "v2",
			expected: "2",
		},
		{
			name:     "v prefix with beta version",
			version:  "v1.0.0-beta1",
			expected: "1.0.0-beta1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Version = tt.version
			got := Get()
			if got != tt.expected {
				t.Errorf("Get() = %q, want %q", got, tt.expected)
			}
		})
	}
}
