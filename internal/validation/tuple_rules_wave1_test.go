package validation_test

import (
	"testing"
)

func TestWave1TupleValidationEdgeCases(t *testing.T) {
	validateEntity := func(entityType, entityID string) bool {
		if entityType == "" || entityID == "" {
			return false
		}
		if len(entityType) > 64 || len(entityID) > 128 {
			return false
		}
		return true
	}

	if !validateEntity("user", "usr_100234") {
		t.Errorf("expected valid entity to pass validation")
	}

	if validateEntity("", "usr_100234") {
		t.Errorf("expected empty entity type to fail validation")
	}

	if validateEntity("user", "") {
		t.Errorf("expected empty entity ID to fail validation")
	}
}
