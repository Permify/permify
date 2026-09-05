package validation

import (
	"strings"
	"testing"
)

func TestWave3RelationTupleFormatting(t *testing.T) {
	validateEntity := func(entity string) bool {
		parts := strings.Split(entity, ":")
		if len(parts) != 2 {
			return false
		}
		return len(parts[0]) > 0 && len(parts[1]) > 0
	}

	if !validateEntity("user:12345") {
		t.Error("expected valid entity user:12345")
	}
	if validateEntity("invalid_entity_no_colon") {
		t.Error("expected invalid entity without colon")
	}
	if validateEntity(":only_id") {
		t.Error("expected invalid entity with empty namespace")
	}
}

func TestWave3PermissionNameValidation(t *testing.T) {
	isValidPermissionName := func(perm string) bool {
		if len(perm) == 0 || len(perm) > 64 {
			return false
		}
		for _, r := range perm {
			if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-') {
				return false
			}
		}
		return true
	}

	if !isValidPermissionName("view_document") {
		t.Error("view_document should be valid permission name")
	}
	if isValidPermissionName("invalid perm with spaces") {
		t.Error("permission with spaces should be rejected")
	}
}
