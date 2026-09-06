package validation

import (
	"testing"
)

func TestWave8OmniPermissionDepthValidation(t *testing.T) {
	isPermissionCycleDetected := func(visited []string, nextNode string) bool {
		for _, node := range visited {
			if node == nextNode {
				return true
			}
		}
		return false
	}

	path := []string{"user:1", "group:eng", "role:admin"}
	if !isPermissionCycleDetected(path, "group:eng") {
		t.Error("expected cycle detection for recurring node")
	}
	if isPermissionCycleDetected(path, "role:superadmin") {
		t.Error("expected no cycle for new node")
	}
}
