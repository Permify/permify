package middleware

import (
	"context"

	grpcAuth "github.com/grpc-ecosystem/go-grpc-middleware/auth"

	"github.com/Permify/permify/internal/authn"
)

// AuthFunc - Middleware that responsible for key authentication
func AuthFunc(authenticator authn.Authenticator) grpcAuth.AuthFunc {
	return func(ctx context.Context) (context.Context, error) {
		err := authenticator.Authenticate(ctx)
		if err != nil {
			return nil, err
		}
		return ctx, nil
	}
}
