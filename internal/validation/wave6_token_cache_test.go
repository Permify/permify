package validation

import (
	"testing"
)

func TestWave6SnapTokenTTLValidation(t *testing.T) {
	isSnapTokenValid := func(issuedAt int64, ttlSeconds int64, currentTs int64) bool {
		return (currentTs - issuedAt) <= ttlSeconds && currentTs >= issuedAt
	}

	now := int64(1788500000)
	ttl := int64(300) // 5 minutes

	if !isSnapTokenValid(now-100, ttl, now) {
		t.Error("token within TTL should be valid")
	}
	if isSnapTokenValid(now-301, ttl, now) {
		t.Error("token exceeding TTL should be expired")
	}
	if isSnapTokenValid(now+100, ttl, now) {
		t.Error("token with future timestamp should be rejected")
	}
}

func TestWave6PermissionCacheKeyGenerator(t *testing.T) {
	generateCacheKey := func(tenantID string, entityType string, entityID string, permission string) string {
		return tenantID + "#" + entityType + ":" + entityID + "@" + permission
	}

	key := generateCacheKey("t_123", "organization", "456", "admin")
	expected := "t_123#organization:456@admin"
	if key != expected {
		t.Errorf("generateCacheKey = %q; want %q", key, expected)
	}
}
