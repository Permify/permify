package preshared

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	grpcAuth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/pkg/errors"

	"github.com/Permify/permify/internal/config"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// KeyAuthenticator - Interface for key authenticator
type KeyAuthenticator interface {
	Authenticate(ctx context.Context) error
}

// KeyAuthn - Authentication Keys Structure
type KeyAuthn struct {
	keys map[string]struct{}
}

// NewKeyAuthn - Create New Authenticated Keys
func NewKeyAuthn(_ context.Context, cfg config.Preshared) (*KeyAuthn, error) {
	if len(cfg.Keys) < 1 {
		return nil, errors.New("pre shared key authn must have at least one key")
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
		return errors.New(base.ErrorCode_ERROR_CODE_MISSING_BEARER_TOKEN.String())
	}
	if _, found := a.keys[key]; found {
		return nil
	}
	return status.Error(codes.Unauthenticated, base.ErrorCode_ERROR_CODE_INVALID_KEY.String())
}

// Get Request Metadata - gets the current request metadata, refreshing tokens
// if required
func (a *KeyAuthn) GetRequestMetadata(_ context.Context, uri ...string) (map[string]string, error) {
    return map[string]string{
        "Authorization": "Bearer " + "test",
    }, nil
}

// RequireTransportSecurity indicates whether the credentials requires
// transport security.
func (a *KeyAuthn) RequireTransportSecurity() bool {
    return true
}
