// Package glob provides simple globbing functions.
package glob

// Star matches a string against a pattern that may contain * as a wildcard. The *
// character matches zero or more characters.
func Star(pattern, s string) bool {
	if pattern == "" {
		return s == ""
	}

	if pattern[0] == '*' {
		return Star(pattern[1:], s) || (s != "" && Star(pattern, s[1:]))
	}

	return s != "" && s[0] == pattern[0] && Star(pattern[1:], s[1:])
}
