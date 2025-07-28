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
			name:     "clean release with v prefix",
			version:  "v1.0.0-0-1234567",
			expected: "1.0.0",
		},
		{
			name:     "clean release without v prefix",
			version:  "1.0.0-0-1234567",
			expected: "1.0.0",
		},
		{
			name:     "dev build with commits ahead",
			version:  "v1.0.0-5-abcdef0",
			expected: "1.0.0-dev.abcdef0",
		},
		{
			name:     "dev build with dirty flag",
			version:  "v1.0.0-0-1234567-dirty",
			expected: "1.0.0-dev.1234567",
		},
		{
			name:     "dev build with broken flag",
			version:  "v1.0.0-0-1234567-broken",
			expected: "1.0.0-dev.1234567",
		},
		{
			name:     "dev build with commits ahead and dirty flag",
			version:  "v1.0.0-5-abcdef0-dirty",
			expected: "1.0.0-dev.abcdef0",
		},
		{
			name:     "dev build with commits ahead and broken flag",
			version:  "v1.0.0-5-abcdef0-broken",
			expected: "1.0.0-dev.abcdef0",
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
