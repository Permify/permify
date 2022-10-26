package authn

import (
	"context"

	grpcAuth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	base "github.com/Permify/permify/pkg/pb/base/v1"
)

var (
	Unauthenticated         = status.Error(codes.Code(base.ErrorCode_ERROR_CODE_UNAUTHENTICATED), "unauthenticated")
	MissingBearerTokenError = status.Error(codes.Code(base.ErrorCode_ERROR_CODE_MISSING_BEARER_TOKEN), "missing bearer token")
)

// KeyAuthenticator -
type KeyAuthenticator interface {
	Authenticate(ctx context.Context) error
}

// KeyAuthn -
type KeyAuthn struct {
	Keys map[string]struct{}
}

// NewKeyAuthn -
func NewKeyAuthn(keys ...string) (*KeyAuthn, error) {
	if len(keys) < 1 {
		return nil, errors.New("pre shared key authn must have at least one key")
	}
	mapKeys := make(map[string]struct{})
	for _, k := range keys {
		mapKeys[k] = struct{}{}
	}
	return &KeyAuthn{
		Keys: mapKeys,
	}, nil
}

// Authenticate -
func (a *KeyAuthn) Authenticate(ctx context.Context) error {
	key, err := grpcAuth.AuthFromMD(ctx, "Bearer")
	if err != nil {
		return MissingBearerTokenError
	}
	if _, found := a.Keys[key]; found {
		return nil
	}
	return Unauthenticated
}
