package openid // OpenID test fakes and mocks

import ( // Test helper imports
	"crypto/ecdsa"      // ECDSA key generation
	"crypto/elliptic"   // Elliptic curve operations
	"crypto/rand"       // Random number generation
	"crypto/rsa"        // RSA key generation
	"encoding/json"     // JSON encoding
	"fmt"               // Formatting utilities
	"net"               // Network utilities
	"net/http"          // HTTP server
	"net/http/httptest" // HTTP testing utilities
	"sync"              // Synchronization primitives

	"github.com/go-jose/go-jose/v3" // JOSE library for JWKS
	"github.com/golang-jwt/jwt/v4"  // JWT library
) // End of imports

type fakeOidcProvider struct { // Mock OIDC provider for testing
	issuerURL    string // Provider issuer URL
	authPath     string // Authorization endpoint path
	tokenPath    string // Token endpoint path
	userInfoPath string // User info endpoint path
	JWKSPath     string // JWKS endpoint path

	algorithms         []string                     // Supported algorithms
	signingKeyMap      map[jwt.SigningMethod]string // Mapping of signing methods to key IDs
	jwks               []jose.JSONWebKey            // JSON Web Key Set
	rsaPrivateKey      *rsa.PrivateKey              // RSA private key for RS256
	rsaPrivateKeyForPS *rsa.PrivateKey              // RSA private key for PS256
	ecdsaPrivateKey    *ecdsa.PrivateKey            // ECDSA private key
	hmacKey            []byte                       // HMAC secret key

	mu sync.RWMutex // Mutex for concurrent access
} // End of fakeOidcProvider struct

type ProviderConfig struct { // Configuration for fake provider
	IssuerURL    string   // Provider issuer URL
	AuthPath     string   // Authorization endpoint path
	TokenPath    string   // Token endpoint path
	UserInfoPath string   // User info endpoint path
	JWKSPath     string   // JWKS endpoint path
	Algorithms   []string // Supported signing algorithms
} // End of ProviderConfig struct

func newFakeOidcProvider(config ProviderConfig) (*fakeOidcProvider, error) { // Create new fake OIDC provider
	rsaPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048) // Generate RSA key
	if err != nil {                                          // Check for errors
		return nil, fmt.Errorf("failed to generate RSA key: %w", err)
	} // End of RSA key generation
	rsaPrivateKeyForPS, err := rsa.GenerateKey(rand.Reader, 2048) // Generate RSA key for PS
	if err != nil {                                               // Check for errors
		return nil, fmt.Errorf("failed to generate RSA key for PS: %w", err)
	} // End of PS RSA key generation
	ecdsaPrivateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader) // Generate ECDSA key
	if err != nil {                                                         // Check for errors
		return nil, fmt.Errorf("failed to generate ECDSA key: %w", err)
	} // End of ECDSA key generation
	hmacKey := []byte("hmackeysecret") // HMAC secret key
	// Initialize signing key mapping
	signingKeyMap := map[jwt.SigningMethod]string{ // Map signing methods to key IDs
		jwt.SigningMethodRS256: "rs256keyid", // RS256 key ID
	} // End of signing key map
	jwks := []jose.JSONWebKey{ // Create JWKS array
		{ // RS256 public key
			Key:       rsaPrivateKey.Public(),                // Public key
			KeyID:     signingKeyMap[jwt.SigningMethodRS256], // Key ID
			Algorithm: "RS256",                               // Algorithm
			Use:       "sig",                                 // Usage: signature
		}, // End of RS256 key
	} // End of JWKS array
	// Return configured provider
	return &fakeOidcProvider{ // Create provider instance
		issuerURL:          config.IssuerURL,
		authPath:           config.AuthPath,
		tokenPath:          config.TokenPath,
		userInfoPath:       config.UserInfoPath,
		JWKSPath:           config.JWKSPath,
		algorithms:         config.Algorithms,  // Supported algorithms
		rsaPrivateKey:      rsaPrivateKey,      // RS256 private key
		rsaPrivateKeyForPS: rsaPrivateKeyForPS, // PS256 private key
		hmacKey:            hmacKey,            // HMAC key
		jwks:               jwks,               // Public key set
		ecdsaPrivateKey:    ecdsaPrivateKey,    // ECDSA private key
		signingKeyMap:      signingKeyMap,      // Key ID mapping
		mu:                 sync.RWMutex{},     // Read-write mutex
	}, nil // Return provider without errors
} // End of newFakeOidcProvider

func (s *fakeOidcProvider) ServeHTTP(w http.ResponseWriter, r *http.Request) { // HTTP handler for fake provider
	s.mu.RLock()         // Acquire read lock
	defer s.mu.RUnlock() // Release lock on return
	// Route requests based on path
	switch r.URL.Path { // Match request path
	case "/.well-known/openid-configuration": // OIDC discovery endpoint
		s.responseWellKnown(w)
	case s.JWKSPath: // JWKS endpoint
		s.responseJWKS(w)
	case s.authPath, s.tokenPath, s.userInfoPath: // Other endpoints
		httpError(w, http.StatusNotFound)
	default: // Unknown paths
		httpError(w, http.StatusNotFound)
	} // End of path routing
} // End of ServeHTTP

type providerJSON struct { // OIDC provider metadata JSON structure
	Issuer      string   `json:"issuer"`                                // Provider issuer URL
	AuthURL     string   `json:"authorization_endpoint"`                // Authorization endpoint
	TokenURL    string   `json:"token_endpoint"`                        // Token endpoint
	JWKSURL     string   `json:"jwks_uri"`                              // JWKS endpoint
	UserInfoURL string   `json:"userinfo_endpoint"`                     // User info endpoint
	Algorithms  []string `json:"id_token_signing_alg_values_supported"` // Supported algorithms
} // End of providerJSON struct

func (s *fakeOidcProvider) responseWellKnown(w http.ResponseWriter) { // Return OIDC configuration
	jso := providerJSON{ // Create provider metadata
		Issuer:      s.issuerURL,                  // Set issuer
		AuthURL:     s.issuerURL + s.authPath,     // Build auth URL
		TokenURL:    s.issuerURL + s.tokenPath,    // Build token URL
		JWKSURL:     s.issuerURL + s.JWKSPath,     // Build JWKS URL
		UserInfoURL: s.issuerURL + s.userInfoPath, // Build user info URL
		Algorithms:  s.algorithms,                 // Set algorithms
	} // End of provider metadata
	httpJSON(w, jso) // Write JSON response
} // End of responseWellKnown

func (s *fakeOidcProvider) responseJWKS(w http.ResponseWriter) { // Return JWKS
	jwks := &jose.JSONWebKeySet{ // Create key set
		Keys: s.jwks, // Set keys
	} // End of key set
	httpJSON(w, jwks) // Write JSON response
} // End of responseJWKS

func httpJSON(w http.ResponseWriter, v interface{}) { // Write JSON response
	w.Header().Set("Content-Type", "application/json") // Set content type
	encoder := json.NewEncoder(w)                      // Create JSON encoder
	encoder.SetIndent("", "  ")                        // Set indentation for readability
	if err := encoder.Encode(v); err != nil {          // Encode and write
		httpError(w, http.StatusInternalServerError) // Handle encoding errors
	} // End of encoding
} // End of httpJSON

func httpError(w http.ResponseWriter, code int) { // Write HTTP error
	http.Error(w, http.StatusText(code), code) // Send error response
} // End of httpError

func (s *fakeOidcProvider) UpdateKeyID(method jwt.SigningMethod, newKeyID string) { // Update key ID for method
	s.mu.Lock()         // Acquire write lock
	defer s.mu.Unlock() // Release lock on return
	// Update key ID mapping
	s.signingKeyMap[method] = newKeyID // Update mapping
	for i, key := range s.jwks {       // Iterate through keys
		if key.Algorithm == method.Alg() { // Find matching algorithm
			s.jwks[i].KeyID = newKeyID // Update key ID
		} // End of algorithm check
	} // End of key iteration
} // End of UpdateKeyID

func (s *fakeOidcProvider) SignIDToken(unsignedToken *jwt.Token) (string, error) { // Sign JWT token
	var signedToken string // Signed token result
	var err error          // Error variable
	// Sign based on method
	switch unsignedToken.Method { // Check signing method
	case jwt.SigningMethodRS256: // RS256 algorithm
		signedToken, err = unsignedToken.SignedString(s.rsaPrivateKey) // Sign with RSA key
	default: // Unsupported method
		return "", fmt.Errorf("incorrect signing method type, supported algorithms: HS256, RS256, ES256, PS256")
	} // End of signing method switch
	// Check for signing errors
	if err != nil { // Handle errors
		return "", err // Return error
	} // End of error check
	// Return signed token
	return signedToken, nil // Success
} // End of SignIDToken

func createUnsignedToken(regClaims jwt.RegisteredClaims, method jwt.SigningMethod) *jwt.Token { // Create unsigned JWT
	claims := struct { // Claims structure
		jwt.RegisteredClaims // Embed registered claims
	}{ // Initialize claims
		RegisteredClaims: regClaims, // Set claims
	} // End of claims
	return jwt.NewWithClaims(method, claims) // Create token with claims
} // End of createUnsignedToken

func fakeHttpServer(url string, handler http.HandlerFunc) (*httptest.Server, error) { // Create fake HTTP server
	listener, err := net.Listen("tcp", url) // Create TCP listener
	if err != nil {                         // Check for errors
		return nil, fmt.Errorf("failed to start listener: %w", err)
	} // End of listener creation
	testServer := httptest.NewUnstartedServer(handler) // Create test server
	_ = testServer.Listener.Close()                    // Close default listener
	testServer.Listener = listener                     // Use custom listener
	testServer.Start()                                 // Start server
	return testServer, nil                             // Return server
} // End of fakeHttpServer
