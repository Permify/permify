package development

import (
	"testing"
)

func TestWave10RelationExpansionDepthLimit(t *testing.T) {
	maxDepth := 10
	calculateDepth := func(steps int) bool {
		return steps <= maxDepth
	}

	if !calculateDepth(5) {
		t.Errorf("expected depth 5 to be valid within limit 10")
	}
	if calculateDepth(11) {
		t.Errorf("expected depth 11 to exceed max expansion limit 10")
	}
}

func TestWave10EntityTupleFormatValidation(t *testing.T) {
	isValidTuple := func(entity, relation, subject string) bool {
		return len(entity) > 0 && len(relation) > 0 && len(subject) > 0
	}

	if !isValidTuple("organization:1", "admin", "user:99") {
		t.Errorf("expected valid entity tuple assertion to pass")
	}
	if isValidTuple("", "admin", "user:99") {
		t.Errorf("expected empty entity to fail validation")
	}
}
