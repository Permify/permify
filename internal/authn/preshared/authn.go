package preshared

import (
	"context"
	"fmt"

	grpcAuth "github.com/grpc-ecosystem/go-grpc-middleware/auth"

	"github.com/Permify/permify/internal/config"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// KeyAuthn - Authentication Keys Structure
type KeyAuthn struct {
	keys map[string]struct{}
}

// NewKeyAuthn - Create New Authenticated Keys
func NewKeyAuthn(_ context.Context, cfg config.Preshared) (*KeyAuthn, error) {
	if len(cfg.Keys) < 1 {
		return nil, fmt.Errorf("%s: pre shared key authn must have at least one key", base.ErrorCode_ERROR_CODE_INVALID_ARGUMENT.String())
	}
	mapKeys := make(map[string]struct{})
	for _, k := range cfg.Keys {
		mapKeys[k] = struct{}{}
	}
	return &KeyAuthn{
		keys: mapKeys,
	}, nil
}

// Authenticate - Checking whether any API request contain keys
func (a *KeyAuthn) Authenticate(ctx context.Context) error {
	key, err := grpcAuth.AuthFromMD(ctx, "Bearer")
	if err != nil {
		return fmt.Errorf("%s: failed to extract authorization header from gRPC request: %w", base.ErrorCode_ERROR_CODE_MISSING_BEARER_TOKEN.String(), err)
	}
	if _, found := a.keys[key]; found {
		return nil
	}
	return fmt.Errorf("%s: invalid preshared key provided", base.ErrorCode_ERROR_CODE_UNAUTHENTICATED.String())
}
