package balancer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

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

// OIDCTokenResponse represents the response from the OIDC token endpoint
type OIDCTokenResponse struct {
	AccessToken string `json:"access_token"`
}

func getOIDCToken(ctx context.Context, issuer, clientID string) (string, error) {
	// Prepare the request data
	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("grant_type", "client_credentials")

	// Create the request
	req, err := http.NewRequestWithContext(ctx, "POST", issuer+"/token", strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("error creating token request: %v", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending token request: %v", err)
	}
	defer resp.Body.Close()

	// Decode the response
	var tokenResponse OIDCTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return "", fmt.Errorf("error decoding token response: %v", err)
	}

	return tokenResponse.AccessToken, nil
}

// setupAuthn configures the authentication token based on the provided authentication method.
// It returns the token string and an error if any.
func setupAuthn(ctx context.Context, authn *config.Authn) (string, error) {
	var token string
	var err error

	switch authn.Method {
	case "preshared":
		token = authn.Preshared.Keys[0]
	case "oidc":
		token, err = getOIDCToken(ctx, authn.Oidc.Issuer, authn.Oidc.ClientID)
		if err != nil {
			return "", fmt.Errorf("failed to get OIDC token: %s", err)
		}
	default:
		return "", fmt.Errorf("unknown authentication method: '%s'", authn.Method)
	}

	return token, nil
}
