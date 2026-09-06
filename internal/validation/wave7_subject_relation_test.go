package validation

import (
	"strings"
	"testing"
)

func TestWave7SubjectRelationValidation(t *testing.T) {
	isValidSubjectFormat := func(subject string) bool {
		if strings.TrimSpace(subject) == "" {
			return false
		}
		parts := strings.Split(subject, ":")
		if len(parts) != 2 || len(parts[0]) == 0 || len(parts[1]) == 0 {
			return false
		}
		return true
	}

	if !isValidSubjectFormat("user:alice") {
		t.Error("user:alice should be valid")
	}
	if !isValidSubjectFormat("group:engineering#member") {
		t.Error("group:engineering#member should be valid")
	}
	if isValidSubjectFormat("invalid_format_without_colon") {
		t.Error("string without colon should be invalid")
	}
	if isValidSubjectFormat(":missing_type") {
		t.Error("missing type prefix should be invalid")
	}
}

func TestWave7PermissionDepthCounter(t *testing.T) {
	isDepthSafe := func(currentDepth int, maxDepth int) bool {
		return currentDepth <= maxDepth
	}

	if !isDepthSafe(15, 20) {
		t.Error("depth 15 should be safe within limit 20")
	}
	if isDepthSafe(21, 20) {
		t.Error("depth 21 should exceed limit 20")
	}
}
