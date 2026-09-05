package validation

import (
	"testing"
)

func TestWave4TenantIDSanitization(t *testing.T) {
	isValidTenantID := func(tenant string) bool {
		if len(tenant) == 0 || len(tenant) > 64 {
			return false
		}
		for _, r := range tenant {
			if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_') {
				return false
			}
		}
		return true
	}

	if !isValidTenantID("t_org_prod_99") {
		t.Error("expected valid tenant ID")
	}
	if isValidTenantID("invalid tenant with spaces") {
		t.Error("expected invalid tenant ID with spaces")
	}
	if isValidTenantID("") {
		t.Error("expected invalid empty tenant ID")
	}
}

func TestWave4DepthLimitBoundary(t *testing.T) {
	checkGraphDepth := func(currentDepth int, maxDepth int) bool {
		return currentDepth <= maxDepth && currentDepth >= 0
	}

	if !checkGraphDepth(15, 30) {
		t.Error("depth 15 of 30 should be allowed")
	}
	if checkGraphDepth(31, 30) {
		t.Error("depth 31 of 30 should exceed maximum recursion depth")
	}
}
