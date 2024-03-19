package oidc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc"
	"github.com/golang-jwt/jwt/v4"
	grpcauth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/hashicorp/go-retryablehttp"

	"github.com/Permify/permify/internal/config"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// Authenticator - Interface for oidc authenticator
type Authenticator interface {
	Authenticate(ctx context.Context) error
}

// Authn holds configuration for OIDC authentication, including issuer, audience, and key details.
type Authn struct {
	// IssuerURL is the URL of the OIDC issuer.
	IssuerURL string

	// Audience is the intended audience of the tokens, typically the client ID.
	Audience string

	// JwksURI is the URL to fetch the JSON Web Key Set (JWKS) from.
	JwksURI string

	// JWKs holds the JWKS fetched from JwksURI for validating tokens.
	JWKs *keyfunc.JWKS

	// httpClient is used to make HTTP requests, e.g., to fetch the JWKS.
	httpClient *http.Client

	// Last time the JWKS was fetched
	lastKeyFetch time.Time

	// Last time the OIDC configuration was fetched
	lastOIDCConfigFetch time.Time

	// KeyRefreshInterval is the interval to refresh the keys
	keyRefreshInterval time.Duration

	// ConfigRefreshInterval is the interval to refresh the OIDC configuration
	configRefreshInterval time.Duration

	// refreshUnknownKID is a flag to refresh the JWKS when the KID is unknown
	refreshUnknownKID bool
}

// NewOidcAuthn creates a new instance of Authn configured for OIDC authentication.
// It initializes the HTTP client with retry capabilities, sets up the OIDC issuer and audience,
// and attempts to fetch the JWKS keys from the issuer's JWKsURI.
func NewOidcAuthn(_ context.Context, conf config.Oidc) (*Authn, error) {
	// Initialize a new retryable HTTP client to handle transient network errors
	// by retrying failed HTTP requests. The logger is disabled for cleaner output.
	client := retryablehttp.NewClient()
	client.Logger = nil // Disabling logging for the HTTP client

	// Create a new instance of Authn with the provided issuer URL and audience.
	// The httpClient is set to the standard net/http client wrapped with retry logic.
	oidc := &Authn{
		IssuerURL:             conf.Issuer,
		Audience:              conf.Audience,
		httpClient:            client.StandardClient(), // Wrap retryable client as a standard http.Client
		keyRefreshInterval:    conf.KeyRefreshInterval,
		configRefreshInterval: conf.ConfigRefreshInterval,
		refreshUnknownKID:     conf.RefreshUnknownKID,
	}

	err := oidc.fetchKeys()
	if err != nil {
		// If fetching keys fails, return an error to prevent initialization of a non-functional Authn instance.
		return nil, err
	}

	// Return the initialized Authn instance, ready for use in OIDC authentication.
	return oidc, nil
}

// Authenticate validates the authentication token from the request context.
func (oidc *Authn) Authenticate(requestContext context.Context) error {
	// Extract the authentication header from the metadata in the request context.
	authHeader, err := grpcauth.AuthFromMD(requestContext, "Bearer")
	if err != nil {
		// Return an error if the bearer token is missing from the authentication header.
		return errors.New(base.ErrorCode_ERROR_CODE_MISSING_BEARER_TOKEN.String())
	}

	// Initialize a new JWT parser with the RS256 signing method.
	jwtParser := jwt.NewParser(jwt.WithValidMethods([]string{"RS256"}))

	// Parse and validate the JWT from the authentication header.
	token, err := jwtParser.Parse(authHeader, func(token *jwt.Token) (any, error) {
		// If a presented token's KID is not found in the existing headers, initiate a JWKs fetch and validate the token.
		if _, ok := token.Header["kid"].(string); !ok {
			// Whem KID is absent in the header and it has been less than the interval since the last JWKs retrieval attempt, reject the token.
			if !oidc.refreshUnknownKID && time.Since(oidc.lastKeyFetch) < oidc.keyRefreshInterval {
				return nil, errors.New(base.ErrorCode_ERROR_CODE_INVALID_BEARER_TOKEN.String())
			}
		}

		// Use the JWKS from oidc to validate the JWT's signature.
		return oidc.JWKs.Keyfunc(token)
	})
	if err != nil {
		// Return an error if the token is invalid (e.g., expired, wrong signature).
		return errors.New(base.ErrorCode_ERROR_CODE_INVALID_BEARER_TOKEN.String())
	}

	// Check if the parsed token is valid.
	if !token.Valid {
		// Return an error if the token is not valid.
		return errors.New(base.ErrorCode_ERROR_CODE_INVALID_BEARER_TOKEN.String())
	}

	// Extract the claims from the token.
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return errors.New(base.ErrorCode_ERROR_CODE_INVALID_CLAIMS.String())
	}

	if ok := claims.VerifyIssuer(oidc.IssuerURL, true); !ok {
		return errors.New(base.ErrorCode_ERROR_CODE_INVALID_ISSUER.String())
	}

	if ok := claims.VerifyAudience(oidc.Audience, true); !ok {
		return errors.New(base.ErrorCode_ERROR_CODE_INVALID_AUDIENCE.String())
	}

	// If all checks pass, the token is considered valid, and the function returns nil.
	return nil
}

func (oidc *Authn) fetchKeys() error {
	if oidc.JwksURI == "" || time.Since(oidc.lastOIDCConfigFetch) > oidc.configRefreshInterval {
		oidcConfig, err := oidc.fetchOIDCConfiguration()
		if err != nil {
			return fmt.Errorf("error fetching OIDC configuration: %w", err)
		}

		oidc.JwksURI = oidcConfig.JWKsURI
		oidc.lastOIDCConfigFetch = time.Now()
	}

	jwks, err := oidc.GetKeys()
	if err != nil {
		return fmt.Errorf("error fetching OIDC keys: %w", err)
	}

	oidc.JWKs = jwks
	oidc.lastKeyFetch = time.Now()

	return nil
}

// GetKeys fetches the JSON Web Key Set (JWKS) from the configured JWKS URI.
func (oidc *Authn) GetKeys() (*keyfunc.JWKS, error) {
	// Use the keyfunc package to fetch the JWKS from the JWKS URI.
	// The keyfunc.Options struct is used to configure the HTTP client used for the request
	// and set a refresh interval for the keys.
	jwks, err := keyfunc.Get(oidc.JwksURI, keyfunc.Options{
		Client:            oidc.httpClient,         // Use the HTTP client configured in the Authn struct.
		RefreshInterval:   oidc.keyRefreshInterval, // Set the interval to refresh the keys.
		RefreshUnknownKID: oidc.refreshUnknownKID,  // Set the flag to refresh the JWKS when the KID is unknown.
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch keys from '%s': %s", oidc.JwksURI, err)
	}

	// Return the fetched JWKS and nil for the error if successful.
	return jwks, nil
}

// Config holds OpenID Connect (OIDC) configuration details.
type Config struct {
	// Issuer is the OIDC provider's unique identifier URL.
	Issuer string `json:"issuer"`
	// JWKsURI is the URL to the JSON Web Key Set (JWKS) provided by the OIDC issuer.
	JWKsURI string `json:"jwks_uri"`
}

// Fetches OIDC configuration using the well-known endpoint.
func (oidc *Authn) fetchOIDCConfiguration() (*Config, error) {
	wellKnownURL := oidc.getWellKnownURL()
	body, err := oidc.doHTTPRequest(wellKnownURL)
	if err != nil {
		return nil, err
	}

	oidcConfig, err := parseOIDCConfiguration(body)
	if err != nil {
		return nil, err
	}

	return oidcConfig, nil
}

// Constructs the well-known URL for fetching OIDC configuration.
func (oidc *Authn) getWellKnownURL() string {
	return strings.TrimSuffix(oidc.IssuerURL, "/") + "/.well-known/openid-configuration"
}

// doHTTPRequest makes an HTTP GET request to the specified URL and returns the response body.
func (oidc *Authn) doHTTPRequest(url string) ([]byte, error) {
	// Create a new HTTP GET request.
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request for OIDC configuration: %s", err)
	}

	// Send the request using the configured HTTP client.
	res, err := oidc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request for OIDC configuration: %s", err)
	}
	// Ensure the response body is closed after reading.
	defer res.Body.Close()

	// Check if the HTTP status code indicates success.
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received unexpected status code (%d) while fetching OIDC configuration", res.StatusCode)
	}

	// Read the response body.
	body, err := io.ReadAll(res.Body)
	if err != nil {
		// Return an error if reading the response body fails.
		return nil, fmt.Errorf("failed to read response body from OIDC configuration request: %s", err)
	}

	// Return the response body.
	return body, nil
}

// parseOIDCConfiguration decodes the OIDC configuration from the given JSON body.
func parseOIDCConfiguration(body []byte) (*Config, error) {
	var oidcConfig Config
	// Attempt to unmarshal the JSON body into the oidcConfig struct.
	if err := json.Unmarshal(body, &oidcConfig); err != nil {
		return nil, fmt.Errorf("failed to decode OIDC configuration: %s", err)
	}

	if oidcConfig.Issuer == "" {
		return nil, errors.New("issuer value is required but missing in OIDC configuration")
	}

	if oidcConfig.JWKsURI == "" {
		return nil, errors.New("JWKsURI value is required but missing in OIDC configuration")
	}

	// Return the successfully parsed configuration.
	return &oidcConfig, nil
}

func (oidc *Authn) Close() {
	oidc.JWKs.EndBackground()
}
