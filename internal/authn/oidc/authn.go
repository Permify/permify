package oidc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
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
	client.Logger = OIDCSlogAdapter{Logger: slog.Default()}

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
		// Log the error if the authorization header is missing or does not start with "Bearer"
		slog.Error("failed to extract authorization header from gRPC request", "error", err)
		// Return an error indicating the missing or incorrect bearer token
		return errors.New(base.ErrorCode_ERROR_CODE_MISSING_BEARER_TOKEN.String())
	}

	// Log the successful extraction of the authorization header for debugging purposes.
	// Remember, do not log the actual content of authHeader as it might contain sensitive information.
	slog.Debug("Successfully extracted authorization header from gRPC request")

	// Parse and validate the JWT token extracted from the authorization header.
	parsedToken, err := oidc.jwtParser.Parse(authHeader, func(token *jwt.Token) (interface{}, error) {
		slog.Info("starting JWT parsing and validation.")

		// Fetch the public keys from the JWKS endpoint configured for the OIDC.
		jwks, err := oidc.jwksSet.Fetch(requestContext, oidc.JwksURI)
		if err != nil {
			slog.Error("failed to fetch JWKS", "uri", oidc.JwksURI, "error", err)
			// If fetching the JWKS fails, return an error.
			return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
		}

		slog.Info("successfully fetched JWKS")

		// Retrieve the key ID from the JWT header and find the corresponding key in the JWKS.
		if keyID, ok := token.Header["kid"].(string); ok {
			slog.Debug("attempting to find key in JWKS", "kid", keyID)

			if key, found := jwks.LookupKeyID(keyID); found {
				// If the key is found, convert it to a usable format.
				var k interface{}
				if err := key.Raw(&k); err != nil {
					slog.Error("failed to get raw public key", "kid", keyID, "error", err)
					return nil, fmt.Errorf("failed to get raw public key: %w", err)
				}
				slog.Debug("successfully obtained raw public key", "key", k)
				return k, nil // Return the public key for JWT signature verification.
			}
			slog.Error("key ID not found in JWKS", "kid", keyID)
			// If the specified key ID is not found in the JWKS, return an error.
			return nil, fmt.Errorf("kid %s not found", keyID)
		}
		slog.Error("jwt does not contain a key ID")
		// If the JWT does not contain a key ID, return an error.
		return nil, errors.New("kid must be specified in the token header")
	})
	if err != nil {
		// Log that the token parsing or validation failed
		slog.Error("token parsing or validation failed", "error", err)
		// If token parsing or validation fails, return an error indicating the token is invalid.
		return errors.New(base.ErrorCode_ERROR_CODE_INVALID_BEARER_TOKEN.String())
	}

	// Ensure the token is valid.
	if !parsedToken.Valid {
		// Log that the parsed token was not valid
		slog.Warn("parsed token is invalid")
		// Return an error indicating the invalid token
		return errors.New(base.ErrorCode_ERROR_CODE_INVALID_BEARER_TOKEN.String())
	}

	// Extract the claims from the token.
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		// Log that the claims were in an incorrect format
		slog.Warn("token claims are in an incorrect format")
		// Return an error
		return errors.New(base.ErrorCode_ERROR_CODE_INVALID_CLAIMS.String())
	}

	slog.Debug("extracted token claims", "claims", claims)

	// Verify the issuer of the token matches the expected issuer.
	if ok := claims.VerifyIssuer(oidc.IssuerURL, true); !ok {
		// Log that the issuer did not match the expected issuer
		slog.Warn("token issuer is invalid", "expected", oidc.IssuerURL, "actual", claims["iss"])
		// Return an error
		return errors.New(base.ErrorCode_ERROR_CODE_INVALID_ISSUER.String())
	}

	// Verify the audience of the token matches the expected audience.
	if ok := claims.VerifyAudience(oidc.Audience, true); !ok {
		// Log that the audience did not match the expected audience
		slog.Warn("token audience is invalid", "expected", oidc.Audience, "actual", claims["aud"])
		// Return an error
		return errors.New(base.ErrorCode_ERROR_CODE_INVALID_AUDIENCE.String())
	}

	// Log that the token's issuer and audience were successfully validated
	slog.Info("token validation succeeded")

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
	// Log the attempt to create a new HTTP GET request
	slog.Debug("creating new HTTP GET request", "url", url)

	// Create a new HTTP GET request.
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		// Log the error if creating the HTTP request fails
		slog.Error("failed to create HTTP request", "url", url, "error", err)
		return nil, fmt.Errorf("failed to create HTTP request for OIDC configuration: %s", err)
	}

	// Log the execution of the HTTP request
	slog.Debug("executing HTTP request", "url", url)

	// Send the request using the configured HTTP client.
	res, err := client.Do(req)
	if err != nil {
		// Log the error if executing the HTTP request fails
		slog.Error("failed to execute HTTP request", "url", url, "error", err)
		return nil, fmt.Errorf("failed to execute HTTP request for OIDC configuration: %s", err)
	}

	// Log the HTTP status code of the response
	slog.Debug("received HTTP response", "status_code", res.StatusCode, "url", url)

	// Ensure the response body is closed after reading.
	defer res.Body.Close()

	// Check if the HTTP status code indicates success.
	if res.StatusCode != http.StatusOK {
		// Log the unexpected status code
		slog.Warn("received unexpected status code", "status_code", res.StatusCode, "url", url)
		return nil, fmt.Errorf("received unexpected status code (%d) while fetching OIDC configuration", res.StatusCode)
	}

	// Log the attempt to read the response body
	slog.Debug("reading response body", "url", url)

	// Read the response body.
	body, err := io.ReadAll(res.Body)
	if err != nil {
		// Log the error if reading the response body fails
		slog.Error("failed to read response body", "url", url, "error", err)
		return nil, fmt.Errorf("failed to read response body from OIDC configuration request: %s", err)
	}

	// Log the successful retrieval of the response body
	slog.Debug("successfully read response body", "url", url, "response_length", len(body))

	// Return the response body.
	return body, nil
}

// parseOIDCConfiguration decodes the OIDC configuration from the given JSON body.
func parseOIDCConfiguration(body []byte) (*Config, error) {
	var oidcConfig Config
	// Attempt to unmarshal the JSON body into the oidcConfig struct.
	if err := json.Unmarshal(body, &oidcConfig); err != nil {
		// Log the error if unmarshalling
		slog.Error("failed to unmarshal OIDC configuration", "error", err)
		return nil, fmt.Errorf("failed to decode OIDC configuration: %s", err)
	}
	// Log the successful decoding of OIDC configuration
	slog.Debug("successfully decoded OIDC configuration")

	if oidcConfig.Issuer == "" {
		// Log missing issuer value
		slog.Warn("missing issuer value in OIDC configuration")
		return nil, errors.New("issuer value is required but missing in OIDC configuration")
	}

	if oidcConfig.JWKsURI == "" {
		// Log missing JWKsURI value
		slog.Warn("missing JWKsURI value in OIDC configuration")
		return nil, errors.New("JWKsURI value is required but missing in OIDC configuration")
	}

	// Log the successful parsing of the OIDC configuration
	slog.Info("successfully parsed OIDC configuration", "issuer", oidcConfig.Issuer, "jwks_uri", oidcConfig.JWKsURI)

	// Return the successfully parsed configuration.
	return &oidcConfig, nil
}

// OIDCSlogAdapter adapts the slog.Logger to be compatible with retryablehttp.LeveledLogger.
type OIDCSlogAdapter struct {
	Logger *slog.Logger
}

// Error logs messages at error level.
func (a OIDCSlogAdapter) Error(msg string, keysAndValues ...interface{}) {
	a.Logger.Error(msg, keysAndValues...)
}

// Info logs messages at info level.
func (a OIDCSlogAdapter) Info(msg string, keysAndValues ...interface{}) {
	a.Logger.Info(msg, keysAndValues...)
}

// Debug logs messages at debug level.
func (a OIDCSlogAdapter) Debug(msg string, keysAndValues ...interface{}) {
	a.Logger.Info(msg, keysAndValues...)
}

// Warn logs messages at warn level.
func (a OIDCSlogAdapter) Warn(msg string, keysAndValues ...interface{}) {
	a.Logger.Warn(msg, keysAndValues...)
}
