package authn

import (
	"context"
)

// Authenticator - Interface for oidc authenticator
type Authenticator interface {
	Authenticate(ctx context.Context) error
}
