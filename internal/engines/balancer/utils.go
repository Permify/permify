package balancer

import (
	"context"
	"fmt"

	"github.com/Permify/permify/internal/config"
)

// secureTokenCredentials represents a map used for storing secure tokens.
// These tokens require transport security.
type secureTokenCredentials map[string]string

// RequireTransportSecurity indicates that transport security is required for these credentials.
func (c secureTokenCredentials) RequireTransportSecurity() bool {
	return true // Transport security is required for secure tokens.
}

// GetRequestMetadata retrieves the current metadata (secure tokens) for a request.
func (c secureTokenCredentials) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	return c, nil // Returns the secure tokens as metadata with no error.
}

// nonSecureTokenCredentials represents a map used for storing non-secure tokens.
// These tokens do not require transport security.
type nonSecureTokenCredentials map[string]string

// RequireTransportSecurity indicates that transport security is not required for these credentials.
func (c nonSecureTokenCredentials) RequireTransportSecurity() bool {
	return false // Transport security is not required for non-secure tokens.
}

// GetRequestMetadata retrieves the current metadata (non-secure tokens) for a request.
func (c nonSecureTokenCredentials) GetRequestMetadata(_ context.Context, _ ...string) (map[string]string, error) {
	return c, nil // Returns the non-secure tokens as metadata with no error.
}

// setupAuthn configures the authentication token based on the provided authentication method.
// It returns the token string and an error if any.
func setupAuthn(_ context.Context, authn *config.Authn) (string, error) {
	var token string

	switch authn.Method {
	case "preshared":
		token = authn.Preshared.Keys[0]
	case "oidc":
		return "", fmt.Errorf("unsupported authentication method: '%s'", authn.Method)
	default:
		return "", fmt.Errorf("unknown authentication method: '%s'", authn.Method)
	}

	return token, nil
}
