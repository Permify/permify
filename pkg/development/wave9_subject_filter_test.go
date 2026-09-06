package development

import (
	"testing"
)

// TestWave9SubjectIDValidation asserts subject ID formatting
func TestWave9SubjectIDValidation(t *testing.T) {
	isValidSubjectID := func(id string) bool {
		return len(id) > 0 && len(id) <= 128
	}

	if !isValidSubjectID("user:12345") {
		t.Errorf("expected valid subject ID assertion to pass")
	}
	if isValidSubjectID("") {
		t.Errorf("expected empty subject ID to fail validation")
	}
}

// TestWave9ActionPermissionSetContains checks action set inclusion
func TestWave9ActionPermissionSetContains(t *testing.T) {
	permissions := map[string]bool{"read": true, "write": true, "admin": true}

	hasPermission := func(action string) bool {
		return permissions[action]
	}

	if !hasPermission("write") {
		t.Errorf("expected write permission to be granted")
	}
	if hasPermission("delete") {
		t.Errorf("expected delete permission to be absent")
	}
}
