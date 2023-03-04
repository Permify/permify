package preshared

import (
	"context"

	grpcAuth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/pkg/errors"

	"github.com/Permify/permify/internal/authn"
	"github.com/Permify/permify/internal/config"
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
func NewKeyAuthn(ctx context.Context, cfg config.Preshared) (*KeyAuthn, error) {
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
		return authn.MissingBearerTokenError
	}
	if _, found := a.keys[key]; found {
		return nil
	}
	return authn.Unauthenticated
}
