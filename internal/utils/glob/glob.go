// Package glob provides simple globbing functions.
package glob

// MatchFold matches a string against a pattern that may contain * as a wildcard. The *
// character matches zero or more characters. Matching is case-insensitive (ASCII fold).
func MatchFold(p, s string) bool {
	if p == "" {
		return s == ""
	}

	if p[0] == '*' {
		return MatchFold(p[1:], s) || (s != "" && MatchFold(p, s[1:]))
	}

	return s != "" && toLower(p[0]) == toLower(s[0]) && MatchFold(p[1:], s[1:])
}

// toLower returns the ASCII lowercase version of a byte.
func toLower(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		return c + ('a' - 'A')
	}
	return c
}
