package keys

import (
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// CommandKeyManager - Key manager interface for commands
type CommandKeyManager interface {
	// SetCheckKey sets the value for the given key.
	SetCheckKey(key *base.PermissionCheckRequest, decision *base.PermissionCheckResponse) bool
	// GetCheckKey gets the value for the given key.
	GetCheckKey(key *base.PermissionCheckRequest) (*base.PermissionCheckResponse, bool)
}
