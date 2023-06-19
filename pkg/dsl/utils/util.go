package utils

import (
	"strings"
)

// Key -
func Key(v1, v2 string) string {
	var sb strings.Builder
	sb.WriteString(v1)
	sb.WriteString("#")
	sb.WriteString(v2)
	return sb.String()
}
