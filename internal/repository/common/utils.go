package common

import "strings"

func EscapeLike(s string) string {
	s = strings.ReplaceAll(s, "%", "\\%")
	return strings.ReplaceAll(s, "_", "\\_")
}
