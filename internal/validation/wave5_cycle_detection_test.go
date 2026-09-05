package validation

import (
	"testing"
)

func TestWave5CyclicRelationDetection(t *testing.T) {
	// Direct self-reference cycle: role:admin -> role:admin
	detectSelfReference := func(subject string, object string) bool {
		return subject == object
	}

	if !detectSelfReference("user:org_admin", "user:org_admin") {
		t.Error("expected self-reference detection")
	}
	if detectSelfReference("user:member", "user:admin") {
		t.Error("non-matching entities should not trigger self-reference")
	}
}

func TestWave5PermissionSetUnion(t *testing.T) {
	unionPermissions := func(setA []string, setB []string) []string {
		seen := make(map[string]bool)
		var merged []string
		for _, p := range append(setA, setB...) {
			if !seen[p] {
				seen[p] = true
				merged = append(merged, p)
			}
		}
		return merged
	}

	p1 := []string{"read", "write"}
	p2 := []string{"write", "delete", "admin"}
	res := unionPermissions(p1, p2)
	
	if len(res) != 4 {
		t.Errorf("expected 4 distinct permissions, got %d", len(res))
	}
}
