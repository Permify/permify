package oidc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	grpcauth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/lestrrat-go/jwx/jwk"

	"github.com/Permify/permify/internal/config"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// Authenticator - Interface for oidc authenticator
type Authenticator interface {
	Authenticate(ctx context.Context) error
}

type Authn struct {
	// URL of the issuer. This is typically the base URL of the identity provider.
	IssuerURL string
	// Audience for which the token is intended. It must match the audience in the JWT.
	Audience string
	// URL of the JSON Web Key Set (JWKS). This URL hosts public keys used to verify JWT signatures.
	JwksURI string
	// Pointer to an AutoRefresh object from the JWKS library. It helps in automatically refreshing the JWKS at predefined intervals.
	jwksSet *jwk.AutoRefresh
	// List of valid signing methods. Specifies which signing algorithms are considered valid for the JWTs.
	validMethods []string
	// Pointer to a JWT parser object. This is used to parse and validate the JWT tokens.
	jwtParser *jwt.Parser
}

// NewOidcAuthn initializes a new instance of the Authn struct with OpenID Connect (OIDC) configuration.
// It takes in a context for managing cancellation and a configuration object. It returns a pointer to an Authn instance or an error.
func NewOidcAuthn(ctx context.Context, conf config.Oidc) (*Authn, error) {
	// Create a new HTTP client with retry capabilities. This client is used for making HTTP requests, particularly for fetching OIDC configuration.
	client := retryablehttp.NewClient()
	client.Logger = nil // Disable logging for the HTTP client to avoid noisy logs.

	// Fetch the OIDC configuration from the issuer's well-known configuration endpoint.
	oidcConf, err := fetchOIDCConfiguration(client.StandardClient(), strings.TrimSuffix(conf.Issuer, "/")+"/.well-known/openid-configuration")
	if err != nil {
		// If there is an error fetching the OIDC configuration, return nil and the error.
		return nil, fmt.Errorf("failed to fetch OIDC configuration: %w", err)
	}

	// Set up automatic refresh of the JSON Web Key Set (JWKS) to ensure the public keys are always up-to-date.
	ar := jwk.NewAutoRefresh(ctx)                                                                                              // Create a new AutoRefresh instance for the JWKS.
	ar.Configure(oidcConf.JWKsURI, jwk.WithHTTPClient(client.StandardClient()), jwk.WithRefreshInterval(conf.RefreshInterval)) // Configure the auto-refresh parameters.

	// Initialize the Authn struct with the OIDC configuration details and other relevant settings.
	oidc := &Authn{
		IssuerURL:    conf.Issuer,                                            // URL of the token issuer.
		Audience:     conf.Audience,                                          // Intended audience of the token.
		JwksURI:      oidcConf.JWKsURI,                                       // URL of the JWKS endpoint.
		validMethods: conf.ValidMethods,                                      // List of acceptable signing methods for the tokens.
		jwtParser:    jwt.NewParser(jwt.WithValidMethods(conf.ValidMethods)), // JWT parser configured with the valid signing methods.
		jwksSet:      ar,                                                     // Set the JWKS auto-refresh instance.
	}

	// Attempt to fetch the JWKS immediately to ensure it's available and valid.
	_, err = oidc.jwksSet.Fetch(ctx, oidc.JwksURI)
	if err != nil {
		// If there is an error fetching the JWKS, return nil and the error.
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}

	// Return the initialized OIDC authentication object and no error.
	return oidc, nil
}

// Authenticate validates the JWT token found in the authorization header of the incoming request.
// It uses the OIDC configuration to validate the token against the issuer's public keys.
func (oidc *Authn) Authenticate(requestContext context.Context) error {
	// Extract the authorization header from the metadata of the incoming gRPC request.
	authHeader, err := grpcauth.AuthFromMD(requestContext, "Bearer")
	if err != nil {
		// If the authorization header is missing or does not start with "Bearer", return an error.
		return errors.New(base.ErrorCode_ERROR_CODE_MISSING_BEARER_TOKEN.String())
	}

	// Parse and validate the JWT token extracted from the authorization header.
	parsedToken, err := oidc.jwtParser.Parse(authHeader, func(token *jwt.Token) (interface{}, error) {
		// Fetch the public keys from the JWKS endpoint configured for the OIDC.
		jwks, err := oidc.jwksSet.Fetch(requestContext, oidc.JwksURI)
		if err != nil {
			// If fetching the JWKS fails, return an error.
			return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
		}

		// Retrieve the key ID from the JWT header and find the corresponding key in the JWKS.
		if keyID, ok := token.Header["kid"].(string); ok {
			if key, found := jwks.LookupKeyID(keyID); found {
				// If the key is found, convert it to a usable format.
				var k interface{}
				if err := key.Raw(&k); err != nil {
					return nil, fmt.Errorf("failed to get raw public key: %w", err)
				}
				return k, nil // Return the public key for JWT signature verification.
			}
			// If the specified key ID is not found in the JWKS, return an error.
			return nil, fmt.Errorf("kid %s not found", keyID)
		}
		// If the JWT does not contain a key ID, return an error.
		return nil, errors.New("kid must be specified in the token header")
	})
	if err != nil {
		// If token parsing or validation fails, return an error indicating the token is invalid.
		return errors.New(base.ErrorCode_ERROR_CODE_INVALID_BEARER_TOKEN.String())
	}

	// Ensure the token is valid.
	if !parsedToken.Valid {
		return errors.New(base.ErrorCode_ERROR_CODE_INVALID_BEARER_TOKEN.String())
	}

	// Extract the claims from the token.
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		// If the claims are in an incorrect format, return an error.
		return errors.New(base.ErrorCode_ERROR_CODE_INVALID_CLAIMS.String())
	}

	// Verify the issuer of the token matches the expected issuer.
	if ok := claims.VerifyIssuer(oidc.IssuerURL, true); !ok {
		return errors.New(base.ErrorCode_ERROR_CODE_INVALID_ISSUER.String())
	}

	// Verify the audience of the token matches the expected audience.
	if ok := claims.VerifyAudience(oidc.Audience, true); !ok {
		return errors.New(base.ErrorCode_ERROR_CODE_INVALID_AUDIENCE.String())
	}

	// If all validations pass, return nil indicating the token is valid.
	return nil
}

// Config holds OpenID Connect (OIDC) configuration details.
type Config struct {
	// Issuer is the OIDC provider's unique identifier URL.
	Issuer string `json:"issuer"`
	// JWKsURI is the URL to the JSON Web Key Set (JWKS) provided by the OIDC issuer.
	JWKsURI string `json:"jwks_uri"`
}

// fetchOIDCConfiguration sends an HTTP request to the given URL to fetch the OpenID Connect (OIDC) configuration.
// It requires an HTTP client and the URL from which to fetch the configuration.
func fetchOIDCConfiguration(client *http.Client, url string) (*Config, error) {
	// Send an HTTP GET request to the provided URL to fetch the OIDC configuration.
	// This typically points to the well-known configuration endpoint of the OIDC provider.
	body, err := doHTTPRequest(client, url)
	if err != nil {
		// If there is an error in fetching the configuration (network error, bad response, etc.), return nil and the error.
		return nil, err
	}

	// Parse the JSON response body into an OIDC Config struct.
	// This involves unmarshalling the JSON into a struct that matches the expected fields of the OIDC configuration.
	oidcConfig, err := parseOIDCConfiguration(body)
	if err != nil {
		// If there is an error in parsing the JSON response (missing fields, incorrect format, etc.), return nil and the error.
		return nil, err
	}

	// Return the parsed OIDC configuration and nil as the error (indicating success).
	return oidcConfig, nil
}

// doHTTPRequest makes an HTTP GET request to the specified URL and returns the response body.
func doHTTPRequest(client *http.Client, url string) ([]byte, error) {
	// Create a new HTTP GET request.
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request for OIDC configuration: %s", err)
	}

	// Send the request using the configured HTTP client.
	res, err := client.Do(req)
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
