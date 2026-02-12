// Package version provides build-time version information.
package version

// Set via ldflags. Defaults are used for builds without the Makefile (e.g. go install).
var (
	Version = "dev"
	Commit  = "unknown" // e.g. "1234567" or "1234567-dirty"
)
