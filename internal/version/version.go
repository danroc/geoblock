// Package version provides version information for the application.
package version

import (
	"regexp"
)

var (
	// Version is set at build time via ldflags
	Version = "v0.0.0-dev"
)

var (
	dirtyRegex   = regexp.MustCompile(`-dirty$`)
	vPrefixRegex = regexp.MustCompile(`^v(\d.*)`)
)

// Get returns the version string for the application.
func Get() string {
	version := Version
	version = vPrefixRegex.ReplaceAllString(version, "$1")
	return dirtyRegex.ReplaceAllString(version, "-dev")
}
