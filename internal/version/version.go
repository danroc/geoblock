// Package version provides version information for the application.
package version //nolint:revive // Package name acceptable for internal package

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Version is set at build time via ldflags
var Version = "v0.0.0-0-g0000000-dirty"

var prefixRegex = regexp.MustCompile(`^v(\d.*)`)

// GitDescribe represents the parsed components of a git describe string.
type GitDescribe struct {
	Tag    string
	Ahead  int
	Commit string
	Dirty  bool
	Broken bool
}

// atoiOrZero converts a string to an integer, returning 0 if conversion fails.
func atoiOrZero(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

// removePrefix removes the "v" prefix from version strings.
func removePrefix(version string) string {
	return prefixRegex.ReplaceAllString(version, "$1")
}

// parseGitDescribe parses a git describe string into its components. It
// expects the format: tag-ahead-commit[-dirty|-broken]
func parseGitDescribe(version string) GitDescribe {
	var (
		parts    = strings.Split(version, "-")
		isDirty  = len(parts) > 3 && parts[3] == "dirty"
		isBroken = len(parts) > 3 && parts[3] == "broken"
	)

	return GitDescribe{
		Tag:    removePrefix(parts[0]),
		Ahead:  atoiOrZero(parts[1]),
		Commit: strings.TrimPrefix(parts[2], "g"),
		Dirty:  isDirty,
		Broken: isBroken,
	}
}

// Get returns the version string for the application.
func Get() string {
	var (
		describe = parseGitDescribe(Version)
		isDev    = describe.Broken || describe.Dirty || describe.Ahead > 0
	)

	if isDev {
		return fmt.Sprintf("%s-dev.%s", describe.Tag, describe.Commit)
	}
	return describe.Tag
}
