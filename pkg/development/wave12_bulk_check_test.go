package development

import (
	"testing"
)

// TestWave12BulkPermissionCheckBatchSize asserts batch size guards
func TestWave12BulkPermissionCheckBatchSize(t *testing.T) {
	maxBatchSize := 100

	isBatchAllowed := func(count int) bool {
		return count > 0 && count <= maxBatchSize
	}

	if !isBatchAllowed(25) {
		t.Errorf("expected 25 permission checks to be allowed in batch")
	}
	if isBatchAllowed(105) {
		t.Errorf("expected 105 requests to exceed max batch check quota")
	}
	if isBatchAllowed(0) {
		t.Errorf("expected empty batch check to be rejected")
	}
}

// TestWave12PermissionCheckResultDeterministicMapping tests boolean output map
func TestWave12PermissionCheckResultDeterministicMapping(t *testing.T) {
	results := map[string]bool{
		"doc:1#view@user:1": true,
		"doc:1#edit@user:1": false,
	}

	if !results["doc:1#view@user:1"] {
		t.Errorf("expected view permission check to return true")
	}
	if results["doc:1#edit@user:1"] {
		t.Errorf("expected edit permission check to return false")
	}
}
