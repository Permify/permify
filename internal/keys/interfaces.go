package keys

import (
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// EngineKeyManager is an interface that defines the methods required for managing
// engine keys, specifically for caching permission check requests and responses.
type EngineKeyManager interface {
	// SetCheckKey sets the value for the given key in the cache. It takes a PermissionCheckRequest
	// as the key and a PermissionCheckResponse as the value to be stored. It returns a boolean
	// value indicating whether the operation was successful or not.
	SetCheckKey(key *base.PermissionCheckRequest, decision *base.PermissionCheckResponse) bool

	// GetCheckKey retrieves the value for the given key from the cache. It takes a
	// PermissionCheckRequest as the key and returns the corresponding PermissionCheckResponse
	// if the key is found, along with a boolean value indicating whether the key was found or not.
	GetCheckKey(key *base.PermissionCheckRequest) (*base.PermissionCheckResponse, bool)

	SetKey(key *base.PermissionCheckRequest, decision *base.PermissionCheckResponse) bool

	GetKey(key *base.PermissionCheckRequest) (*base.PermissionCheckResponse, bool)
}
