package middleware

import (
	"context"

	grpcAuth "github.com/grpc-ecosystem/go-grpc-middleware/auth"

	"github.com/Permify/permify/internal/authn/preshared"
)

// KeyAuthFunc - Middleware that responsible for key authentication
func KeyAuthFunc(authenticator preshared.KeyAuthenticator) grpcAuth.AuthFunc {
	return func(ctx context.Context) (context.Context, error) {
		err := authenticator.Authenticate(ctx)
		if err != nil {
			return nil, err
		}
		return ctx, nil
	}
}
