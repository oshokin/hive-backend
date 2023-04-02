package common

import "strings"

// EscapeLike escapes the '%' and ” characters in the input string s by prefixing them with a backslash.
// This is commonly used for constructing SQL LIKE queries with wildcards.
// The function returns the escaped string.
func EscapeLike(s string) string {
	s = strings.ReplaceAll(s, "%", "\\%")
	return strings.ReplaceAll(s, "_", "\\_")
}
