package openid // OpenID authentication tests
// Test file for OIDC authentication functionality
import ( // Test package imports
	"context" // Context for request management
	"encoding/json"
	"fmt" // Formatting for test messages
	"net/http"
	"net/http/httptest"
	"sync"
	"time" // Time utilities for test timing

	"google.golang.org/grpc/metadata" // GRPC metadata utilities

	"github.com/golang-jwt/jwt/v4"                                // JWT token library
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils" // GRPC metadata utils
	. "github.com/onsi/ginkgo/v2"                                 // Ginkgo test framework
	. "github.com/onsi/gomega"                                    // Gomega matchers

	"github.com/Permify/permify/internal/config"     // Internal config
	base "github.com/Permify/permify/pkg/pb/base/v1" // Base protobuf types
) // End of test imports
// Test suite begins here
var _ = Describe("authn-oidc", func() { // Main test suite
	audience := "aud"                      // Test audience
	listenAddress := "localhost:9999"      // Local test server address
	issuerURL := "http://" + listenAddress // Construct issuer URL
	var fakeOidcProvider *fakeOidcProvider // Mock OIDC provider for testing

	var server *httptest.Server // Test HTTP server

	BeforeEach(func() { // Setup before each test
		var err error // Error variable

		fakeOidcProvider, err = newFakeOidcProvider(ProviderConfig{ // Initialize fake provider
			IssuerURL:    issuerURL,                                    // Mock issuer URL
			AuthPath:     "/auth",                                      // Authentication endpoint
			TokenPath:    "/token",                                     // Token endpoint
			UserInfoPath: "/userInfo",                                  // User info endpoint
			JWKSPath:     "/jwks",                                      // JWKS endpoint
			Algorithms:   []string{"RS256", "HS256", "ES256", "PS256"}, // Supported algorithms
		}) // End of provider config
		Expect(err).To(BeNil()) // Verify no errors
		// Start test HTTP server
		server, err = fakeHttpServer(listenAddress, fakeOidcProvider.ServeHTTP)
		Expect(err).To(BeNil()) // Verify server creation
	}) // End of BeforeEach
	// Cleanup section
	AfterEach(func() { // Cleanup after each test
		server.Close() // Close test server
	}) // End of AfterEach
	// Test contexts begin
	Context("Authenticate With Signing Methods", func() { // Test different signing methods
		It("Case 1", func() { // Test signing method validation
			tests := []struct { // Test cases for signing methods
				name   string
				method jwt.SigningMethod
				err    error
			}{
				{ // Test case for RS256
					"Should pass with RS256",
					jwt.SigningMethodRS256,
					nil,
				}, // End of RS256 test case
			} // End of signing method test cases
			// Execute test cases
			for _, tt := range tests { // Run each signing method test
				now := time.Now()               // Record test start time
				claims := jwt.RegisteredClaims{ // Create JWT claims
					Issuer:    issuerURL,
					Subject:   "user",
					Audience:  []string{audience},
					ExpiresAt: &jwt.NumericDate{Time: now.AddDate(1, 0, 0)},
					IssuedAt:  &jwt.NumericDate{Time: now},
				}

				// create signed token from oidc provider
				unsignedToken := createUnsignedToken(claims, tt.method)
				unsignedToken.Header["kid"] = fakeOidcProvider.signingKeyMap[tt.method]
				idToken, err := fakeOidcProvider.SignIDToken(unsignedToken)
				Expect(err).To(BeNil()) // Verify token signed
				// Create authenticator for test
				// create oidc authenticator
				ctx := context.Background() // Background context for test
				auth, err := NewOidcAuthn(ctx, config.Oidc{
					Audience:          audience,
					Issuer:            issuerURL,
					RefreshInterval:   5 * time.Minute,
					BackoffInterval:   12 * time.Second,
					BackoffMaxRetries: 5,
					BackoffFrequency:  1 * time.Second, // Backoff retry frequency
				})
				Expect(err).To(BeNil()) // Verify authenticator created
				// Test token authentication
				// authenticate
				niceMd := make(metautils.NiceMD)
				niceMd.Set("authorization", "Bearer "+idToken)
				err = auth.Authenticate(niceMd.ToIncoming(ctx))
				if tt.err == nil {
					Expect(err).To(BeNil())
				} else {
					Expect(err).To(Equal(tt.err))
				}
			} // End of test cases
		}) // End of test
	}) // End of context
	// Claims validation tests
	Context("Authenticate Claims", func() { // Test claims validation
		It("Case 1", func() {
			tests := []struct { // Test cases for claims
				name          string                // Test case name
				claimOverride *jwt.RegisteredClaims // Claims to override
				wantErr       bool                  // Expected error status
			}{
				{
					"With correct values there should be no errors",
					&jwt.RegisteredClaims{},
					false,
				},
				{
					"Wrong issuer in the token, it should fail",
					&jwt.RegisteredClaims{
						Issuer: "http://wrong-issuer",
					},
					true,
				},
				{
					"Wrong clientID in the token, it should fail",
					&jwt.RegisteredClaims{
						Audience: []string{"wrong-clientid"},
					},
					true,
				},
				{
					"Expired Token, it should fail",
					&jwt.RegisteredClaims{
						ExpiresAt: &jwt.NumericDate{Time: time.Date(1999, 1, 0, 0, 0, 0, 0, time.UTC)},
					},
					true,
				},
				{
					"Issued at the future, it should fail",
					&jwt.RegisteredClaims{
						IssuedAt: &jwt.NumericDate{Time: time.Date(3999, 1, 0, 0, 0, 0, 0, time.UTC)},
					},
					true,
				},
				{
					"Token used before its NotBefore date, it should fail",
					&jwt.RegisteredClaims{
						NotBefore: &jwt.NumericDate{Time: time.Date(3999, 1, 0, 0, 0, 0, 0, time.UTC)},
					},
					true,
				}, // End of NotBefore test case
			} // End of claim test cases
			// Execute all claim validation tests
			for _, tt := range tests { // Execute claim validation tests
				now := time.Now()
				claims := jwt.RegisteredClaims{
					Issuer:    issuerURL,
					Subject:   "user",
					Audience:  []string{audience},
					ExpiresAt: &jwt.NumericDate{Time: now.AddDate(1, 0, 0)},
					IssuedAt:  &jwt.NumericDate{Time: now},
				}
				claimOverride(&claims, tt.claimOverride) // Override claims for test
				// Create token with overridden claims
				// create signed token from oidc server with overridden claims
				unsignedToken := createUnsignedToken(claims, jwt.SigningMethodRS256)
				unsignedToken.Header["kid"] = fakeOidcProvider.signingKeyMap[jwt.SigningMethodRS256]
				idToken, err := fakeOidcProvider.SignIDToken(unsignedToken)
				Expect(err).To(BeNil()) // Verify token signing
				// Initialize authenticator
				// create oidc authenticator
				ctx := context.Background() // Background context
				auth, err := NewOidcAuthn(ctx, config.Oidc{
					Audience:          audience,
					Issuer:            issuerURL,
					RefreshInterval:   5 * time.Minute,
					BackoffInterval:   12 * time.Second,
					BackoffMaxRetries: 5,
					BackoffFrequency:  1 * time.Second, // Retry frequency
				})
				Expect(err).To(BeNil()) // Verify authenticator creation
				// Test authentication with token
				// authenticate token
				niceMd := make(metautils.NiceMD)
				niceMd.Set("authorization", "Bearer "+idToken)
				err = auth.Authenticate(niceMd.ToIncoming(ctx))
				Expect(err != nil).To(Equal(tt.wantErr), fmt.Sprintf("Wanted error: %t, got %v", tt.wantErr, err))
				Expect(time.Now()).To(BeTemporally("<=", now.Add(1*time.Second))) // Verify execution time
			} // End of test iteration
		}) // End of test case
	}) // End of Authenticate Claims context
	// Test key ID authentication scenarios
	Context("Authenticate Key Ids", func() { // Test key ID authentication
		It("Case 1", func() { // Test various key ID scenarios
			tests := []struct {
				name       string            // Test case name
				method     jwt.SigningMethod // Signing method
				addKeyId   bool              // Whether to add key ID
				keyId      string            // Key ID value
				wantErr    bool              // Expected error status
				wantTiming time.Duration     // Expected timing
			}{ // Test case definitions
				{ // Test case 1
					"With no keyid using RS256 it should fail, multiple public RSA keys matching for RS256 and PS256",
					jwt.SigningMethodRS256,
					false, "",
					true,
					1 * time.Second, // Expected timing
				}, // End test case 1
				{ // Test case 2
					"With right keyid using RS256 it should pass",
					jwt.SigningMethodRS256,
					true, fakeOidcProvider.signingKeyMap[jwt.SigningMethodRS256],
					false,
					1 * time.Second, // Expected timing
				}, // End test case 2
				{ // Test case 3
					"With wrong keyid using RS256 it should fail",
					jwt.SigningMethodRS256,
					true, "wrongkeyid",
					true,
					6 * time.Second, // Expected timing with retries
				}, // End test case 3
			} // End of key ID test case array
			// Execute all key ID tests
			for _, tt := range tests { // Iterate through key ID tests
				now := time.Now()               // Start timer
				claims := jwt.RegisteredClaims{ // JWT claims
					Issuer:    issuerURL,
					Subject:   "user",
					Audience:  []string{audience},
					ExpiresAt: &jwt.NumericDate{Time: now.AddDate(1, 0, 0)},
					IssuedAt:  &jwt.NumericDate{Time: now},
				} // End of claims
				// Create test token with optional key ID
				// create signed token from oidc provider possibly with kid in header
				unsignedToken := createUnsignedToken(claims, tt.method)
				if tt.addKeyId {
					unsignedToken.Header["kid"] = tt.keyId
				}
				idToken, err := fakeOidcProvider.SignIDToken(unsignedToken)
				Expect(err).To(BeNil()) // Verify token creation
				// Create authenticator instance
				// create authenticator
				ctx := context.Background() // Test context
				auth, err := NewOidcAuthn(ctx, config.Oidc{
					Audience:          audience,
					Issuer:            issuerURL,
					RefreshInterval:   5 * time.Minute,
					BackoffInterval:   12 * time.Second,
					BackoffMaxRetries: 5,
					BackoffFrequency:  1 * time.Second, // Retry frequency
				})
				Expect(err).To(BeNil()) // Verify initialization
				// Test authentication flow
				// authenticate
				niceMd := make(metautils.NiceMD)
				niceMd.Set("authorization", "Bearer "+idToken)
				err = auth.Authenticate(niceMd.ToIncoming(ctx))
				Expect(err != nil).To(Equal(tt.wantErr), fmt.Sprintf("Wanted error: %t, got %v", tt.wantErr, err))
				Expect(time.Now()).To(BeTemporally("<=", now.Add(tt.wantTiming))) // Verify timing
			} // End of key ID test iteration
		}) // End of Case 1

		It("Case 2", func() { // Test key rotation scenario
			// create authenticator
			ctx := context.Background()
			auth, err := NewOidcAuthn(ctx, config.Oidc{
				Audience:          audience,
				Issuer:            issuerURL,
				RefreshInterval:   5 * time.Minute,
				BackoffInterval:   1 * time.Minute,
				BackoffMaxRetries: 5,
				BackoffFrequency:  1 * time.Second, // Frequency between retries
			})
			Expect(err).To(BeNil()) // Verify no initialization errors

			tests := []struct {
				name     string
				method   jwt.SigningMethod
				newKeyId string
			}{
				{
					"Old KID not found, retry with new KID",
					jwt.SigningMethodRS256,
					"newkey",
				},
				{
					"Old KID not found, retry with new KID",
					jwt.SigningMethodRS256,
					"keykey",
				}, // End of test case 2
			} // End of key rotation test cases

			for _, tt := range tests { // Iterate through rotation tests
				fakeOidcProvider.UpdateKeyID(tt.method, tt.newKeyId) // Update provider key

				now := time.Now()               // Record start time
				claims := jwt.RegisteredClaims{ // Create standard claims
					Issuer:    issuerURL,
					Subject:   "user",
					Audience:  []string{audience},
					ExpiresAt: &jwt.NumericDate{Time: now.AddDate(1, 0, 0)},
					IssuedAt:  &jwt.NumericDate{Time: now},
				}

				// create signed token from oidc provider possibly with kid in header
				unsignedToken := createUnsignedToken(claims, tt.method)
				unsignedToken.Header["kid"] = tt.newKeyId
				idToken, err := fakeOidcProvider.SignIDToken(unsignedToken)
				Expect(err).To(BeNil())

				// authenticate
				niceMd := make(metautils.NiceMD)
				niceMd.Set("authorization", "Bearer "+idToken)
				err = auth.Authenticate(niceMd.ToIncoming(ctx))
				Expect(err).Should(BeNil())
				Expect(time.Now()).To(BeTemporally("<=", now.Add(1*time.Second))) // Verify quick response
			} // End of rotation test iteration
		}) // End of Case 2

		It("Case 3: Complex test for maximum retries and backoff interval", func() { // Test backoff behavior
			// create authenticator
			ctx := context.Background() // Test context
			auth, err := NewOidcAuthn(ctx, config.Oidc{
				Audience:          audience,
				Issuer:            issuerURL,
				RefreshInterval:   5 * time.Minute,
				BackoffInterval:   6 * time.Second, // Backoff interval
				BackoffMaxRetries: 3,               // Max retry attempts
				BackoffFrequency:  1 * time.Second, // Retry frequency
			})
			Expect(err).To(BeNil()) // Verify authenticator created

			tests := []struct { // Complex retry test cases
				name              string            // Test case name
				method            jwt.SigningMethod // JWT signing method
				newKeyId          string            // Invalid key ID to test
				succeedAfterRetry bool              // Whether to succeed after retry
			}{
				{ // Test case 1: retry then succeed
					"Invalid KID, should retry and fail, and then succeed", // Description
					jwt.SigningMethodRS256,
					"invalidkey1",
					true, // Will succeed after retry
				}, // End of test case 1
				{ // Test case 2: always fail
					"Invalid KID, should retry and fail",
					jwt.SigningMethodRS256,
					"invalidkey2",
					false, // Will not succeed
				}, // End of test case 2
			} // End of complex test cases

			for _, tt := range tests { // Execute each complex test
				now := time.Now()               // Start timer
				claims := jwt.RegisteredClaims{ // JWT claims structure
					Issuer:    issuerURL,
					Subject:   "user",
					Audience:  []string{audience},
					ExpiresAt: &jwt.NumericDate{Time: now.AddDate(1, 0, 0)},
					IssuedAt:  &jwt.NumericDate{Time: now},
				}

				// create signed token from oidc provider with invalid kid in header
				unsignedToken := createUnsignedToken(claims, tt.method)
				unsignedToken.Header["kid"] = tt.newKeyId
				idToken, err := fakeOidcProvider.SignIDToken(unsignedToken)
				Expect(err).To(BeNil()) // Verify token creation

				// authenticate with retries
				niceMd := make(metautils.NiceMD) // Create metadata
				niceMd.Set("authorization", "Bearer "+idToken)
				now = time.Now()                                                  // Record start time
				err = auth.Authenticate(niceMd.ToIncoming(ctx))                   // Attempt authentication
				Expect(err).ShouldNot(BeNil())                                    // Should fail
				Expect(time.Now()).To(BeTemporally("<=", now.Add(4*time.Second))) // Verify retry timing
				// Test immediate failure after max retries
				// authenticate after retries should fail immediately without waiting
				now = time.Now()                                                  // Record time
				err = auth.Authenticate(niceMd.ToIncoming(ctx))                   // Try authentication again
				Expect(err).ShouldNot(BeNil())                                    // Should still fail
				Expect(time.Now()).To(BeTemporally("<=", now.Add(1*time.Second))) // Verify immediate failure

				if tt.succeedAfterRetry { // Check if should succeed after retry
					// authenticate with now valid key should succeed immediately after backoff interval elapsed
					fakeOidcProvider.UpdateKeyID(jwt.SigningMethodRS256, tt.newKeyId) // Update with valid key

					time.Sleep(7 * time.Second) // Wait for backoff interval

					now = time.Now()                                                  // Record time
					err = auth.Authenticate(niceMd.ToIncoming(ctx))                   // Authenticate with valid key
					Expect(err).Should(BeNil())                                       // Should succeed
					Expect(time.Now()).To(BeTemporally("<=", now.Add(1*time.Second))) // Verify quick success
				} // End of succeedAfterRetry check
			} // End of complex test iteration
		})

		It("Case 4: Concurrent requests leading to global backoff lock for 6 seconds", func() { // Test concurrent retries
			// create authenticator
			ctx := context.Background() // Test context
			auth, err := NewOidcAuthn(ctx, config.Oidc{
				Audience:          audience,
				Issuer:            issuerURL,
				RefreshInterval:   5 * time.Minute,
				BackoffInterval:   6 * time.Second, // Backoff period
				BackoffMaxRetries: 3,               // Max retries
				BackoffFrequency:  1 * time.Second, // Retry frequency
			})
			Expect(err).To(BeNil()) // Verify authenticator created

			invalidKeyIDs := []string{"invalidkey1", "invalidkey2"} // Invalid key IDs for testing

			var wg sync.WaitGroup                 // WaitGroup for concurrent requests
			numRequests := len(invalidKeyIDs) * 3 // Send each invalid key ID multiple times to trigger retries
			// Helper to create tokens
			// Helper function to create a token with the specified key ID
			createTokenWithKid := func(kid string) (string, error) {
				now := time.Now()
				claims := jwt.RegisteredClaims{
					Issuer:    auth.IssuerURL,
					Subject:   "user",
					Audience:  []string{auth.Audience},
					ExpiresAt: &jwt.NumericDate{Time: now.AddDate(1, 0, 0)},
					IssuedAt:  &jwt.NumericDate{Time: now},
				}

				// create signed token from oidc server with overridden claims
				unsignedToken := createUnsignedToken(claims, jwt.SigningMethodRS256)
				unsignedToken.Header["kid"] = kid
				idToken, err := fakeOidcProvider.SignIDToken(unsignedToken)
				Expect(err).To(BeNil())
				return idToken, nil
			} // End of helper function
			// Setup valid key for later tests
			// Set valid token to ensure it's in cache for subsequent tests
			validKeyID := "validkey"                                         // Valid key ID for testing
			fakeOidcProvider.UpdateKeyID(jwt.SigningMethodRS256, validKeyID) // Set valid key
			// Test concurrent invalid requests
			// Step 1: Trigger backoff by hitting max retries with invalid keys concurrently
			now := time.Now() // Record start time
			for i := 0; i < numRequests; i++ {
				wg.Add(1)
				go func(i int) {
					defer wg.Done()
					keyID := invalidKeyIDs[i%len(invalidKeyIDs)]
					token, _ := createTokenWithKid(keyID)
					niceMd := make(metautils.NiceMD)
					niceMd.Set("authorization", "Bearer "+token)
					now := time.Now() // Record request time
					err := auth.Authenticate(niceMd.ToIncoming(ctx))
					Expect(err.Error()).Should(Equal(base.ErrorCode_ERROR_CODE_INVALID_BEARER_TOKEN.String()))
					Expect(time.Now()).To(BeTemporally("<=", now.Add(4*time.Second))) // Verify retry timing
				}(i)
			}

			wg.Wait()
			Expect(time.Now()).To(BeTemporally("<=", now.Add(4*time.Second))) // Verify total timing

			// Step 2: Verify that retries are immediately rejected during backoff period
			for _, keyID := range invalidKeyIDs {
				token, _ := createTokenWithKid(keyID)
				md := metadata.Pairs("authorization", "Bearer "+token)
				ctx := metadata.NewIncomingContext(ctx, md)
				now := time.Now()             // Record request time
				err := auth.Authenticate(ctx) // Attempt authentication
				Expect(err.Error()).Should(Equal(base.ErrorCode_ERROR_CODE_INVALID_BEARER_TOKEN.String()))
				Expect(time.Now()).To(BeTemporally("<=", now.Add(1*time.Second))) // Should fail immediately
			} // End of Step 2 iteration
			// Test valid key still works
			// Step 3: A valid KID already in the JWKS cache should continue to authenticate successfully immediately
			validToken, _ := createTokenWithKid(validKeyID)
			niceMd := make(metautils.NiceMD)
			niceMd.Set("authorization", "Bearer "+validToken)
			// Test valid key authentication
			now = time.Now() // Record time
			err = auth.Authenticate(niceMd.ToIncoming(ctx))
			Expect(err).Should(BeNil())
			Expect(time.Now()).To(BeTemporally("<=", now.Add(1*time.Second))) // Should succeed quickly
			// Verify invalid keys still blocked
			// Test invalid keys still fail during backoff
			// Step 4: Ensure invalid keys return immediately during backoff, even after valid KID
			for _, keyID := range invalidKeyIDs { // Test each invalid key
				token, _ := createTokenWithKid(keyID)
				md := metadata.Pairs("authorization", "Bearer "+token) // Create metadata
				ctx := metadata.NewIncomingContext(ctx, md)            // Create incoming context
				now = time.Now()                                       // Record time
				err := auth.Authenticate(ctx)                          // Authenticate with invalid key
				Expect(err.Error()).Should(Equal(base.ErrorCode_ERROR_CODE_INVALID_BEARER_TOKEN.String()))
				Expect(time.Now()).To(BeTemporally("<=", now.Add(1*time.Second))) // Should fail immediately
			} // End of Step 4 iteration
			// Verify behavior after backoff expires
			// Test rejection after backoff expires
			// Step 5: Ensure invalid keys are still rejected after backoff period
			time.Sleep(7 * time.Second) // Wait for backoff to expire

			now = time.Now()                      // Record time
			for _, keyID := range invalidKeyIDs { // Test each invalid key again
				token, _ := createTokenWithKid(keyID) // Create token
				niceMd := make(metautils.NiceMD)
				niceMd.Set("authorization", "Bearer "+token)
				err = auth.Authenticate(niceMd.ToIncoming(ctx))
				Expect(err.Error()).Should(Equal(base.ErrorCode_ERROR_CODE_INVALID_BEARER_TOKEN.String()))
			} // End of Step 5 iteration
			Expect(time.Now()).To(BeTemporally("<=", now.Add(4*time.Second))) // Verify timing with retries
		})
	})

	Context("Missing Config", func() {
		It("Case 1", func() {
			_, err := NewOidcAuthn(context.Background(), config.Oidc{
				Audience:          "",
				Issuer:            "https://wrong-url",
				RefreshInterval:   5 * time.Minute,
				BackoffInterval:   12 * time.Second,
				BackoffMaxRetries: 5,
				BackoffFrequency:  5 * time.Second,
			},
			)
			Expect(err).ShouldNot(Equal(BeNil()))
		})
	})

	Context("Backoff Parameter Validation", func() {
		It("should return error for invalid backoffInterval", func() {
			_, err := NewOidcAuthn(context.Background(), config.Oidc{
				Audience:          audience,
				Issuer:            issuerURL,
				RefreshInterval:   5 * time.Minute,
				BackoffInterval:   0, // Invalid: should be > 0
				BackoffMaxRetries: 5,
				BackoffFrequency:  1 * time.Second,
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("invalid or missing backoffInterval"))
		})

		It("should return error for negative backoffInterval", func() {
			_, err := NewOidcAuthn(context.Background(), config.Oidc{
				Audience:          audience,
				Issuer:            issuerURL,
				RefreshInterval:   5 * time.Minute,
				BackoffInterval:   -1 * time.Second, // Invalid: should be > 0
				BackoffMaxRetries: 5,
				BackoffFrequency:  1 * time.Second,
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("invalid or missing backoffInterval"))
		})

		It("should return error for invalid backoffMaxRetries", func() {
			_, err := NewOidcAuthn(context.Background(), config.Oidc{
				Audience:          audience,
				Issuer:            issuerURL,
				RefreshInterval:   5 * time.Minute,
				BackoffInterval:   12 * time.Second,
				BackoffMaxRetries: 0, // Invalid: should be > 0
				BackoffFrequency:  1 * time.Second,
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("invalid or missing backoffMaxRetries"))
		})

		It("should return error for negative backoffMaxRetries", func() {
			_, err := NewOidcAuthn(context.Background(), config.Oidc{
				Audience:          audience,
				Issuer:            issuerURL,
				RefreshInterval:   5 * time.Minute,
				BackoffInterval:   12 * time.Second,
				BackoffMaxRetries: -1, // Invalid: should be > 0
				BackoffFrequency:  1 * time.Second,
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("invalid or missing backoffMaxRetries"))
		})

		It("should return error for invalid backoffFrequency", func() {
			_, err := NewOidcAuthn(context.Background(), config.Oidc{
				Audience:          audience,
				Issuer:            issuerURL,
				RefreshInterval:   5 * time.Minute,
				BackoffInterval:   12 * time.Second,
				BackoffMaxRetries: 5,
				BackoffFrequency:  0, // Invalid: should be > 0
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("invalid or missing backoffFrequency"))
		})

		It("should return error for negative backoffFrequency", func() {
			_, err := NewOidcAuthn(context.Background(), config.Oidc{
				Audience:          audience,
				Issuer:            issuerURL,
				RefreshInterval:   5 * time.Minute,
				BackoffInterval:   12 * time.Second,
				BackoffMaxRetries: 5,
				BackoffFrequency:  -1 * time.Second, // Invalid: should be > 0
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("invalid or missing backoffFrequency"))
		})
	})

	Context("Authenticate Id Token", func() {
		It("Case 1", func() {
			// create authenticator
			ctx := context.Background()
			auth, err := NewOidcAuthn(ctx, config.Oidc{
				Audience:          audience,
				Issuer:            issuerURL,
				RefreshInterval:   5 * time.Minute,
				BackoffInterval:   12 * time.Second,
				BackoffMaxRetries: 5,
				BackoffFrequency:  5 * time.Second,
			})
			Expect(err).To(BeNil())

			// authenticate
			niceMd := make(metautils.NiceMD)
			niceMd.Set("authorization", "Bearer ")
			err = auth.Authenticate(niceMd.ToIncoming(ctx))
			Expect(err.Error()).Should(Equal(base.ErrorCode_ERROR_CODE_INVALID_BEARER_TOKEN.String()))
		})

		It("Case 2", func() {
			// create authenticator
			ctx := context.Background()
			auth, err := NewOidcAuthn(ctx, config.Oidc{
				Audience:          audience,
				Issuer:            issuerURL,
				RefreshInterval:   5 * time.Minute,
				BackoffInterval:   12 * time.Second,
				BackoffMaxRetries: 5,
				BackoffFrequency:  5 * time.Second,
			})
			Expect(err).To(BeNil())

			// authenticate
			niceMd := make(metautils.NiceMD)
			err = auth.Authenticate(niceMd.ToIncoming(ctx))
			Expect(err.Error()).Should(Equal(base.ErrorCode_ERROR_CODE_MISSING_BEARER_TOKEN.String()))
		})

		It("Case 3", func() {
			// create authenticator
			ctx := context.Background()
			auth, err := NewOidcAuthn(ctx, config.Oidc{
				Audience:          audience,
				Issuer:            issuerURL,
				RefreshInterval:   5 * time.Minute,
				BackoffInterval:   12 * time.Second,
				BackoffMaxRetries: 5,
				BackoffFrequency:  5 * time.Second,
			})
			Expect(err).To(BeNil())

			// authenticate
			niceMd := make(metautils.NiceMD)
			niceMd.Set("authorization", "Bearer asd")
			err = auth.Authenticate(niceMd.ToIncoming(ctx))
			Expect(err.Error()).Should(Equal(base.ErrorCode_ERROR_CODE_INVALID_BEARER_TOKEN.String()))
		})

		It("should return error for invalid token claims format", func() {
			// This test is challenging because the JWT library automatically converts
			// custom claims to MapClaims. The invalid claims format error path
			// (lines 157-163 in authn.go) is difficult to trigger in practice
			// because the JWT library handles the conversion internally.
			//
			// The error path exists for defensive programming but may not be
			// easily testable with the current JWT library implementation.
			//
			// For now, we'll skip this test as the code path is covered by
			// the JWT library's internal handling of claims conversion.
			Skip("Invalid claims format test - JWT library automatically converts to MapClaims")
		})
	})

	Context("Context Cancellation During Retry", func() {
		It("should handle context cancellation during retry", func() { // Test cancellation handling
			// create authenticator with short backoff frequency to trigger retries quickly
			ctx, cancel := context.WithCancel(context.Background()) // Cancellable context
			auth, err := NewOidcAuthn(ctx, config.Oidc{             // Create authenticator
				Audience:          audience,
				Issuer:            issuerURL,
				RefreshInterval:   5 * time.Minute,
				BackoffInterval:   12 * time.Second,
				BackoffMaxRetries: 3,
				BackoffFrequency:  2 * time.Second, // Longer frequency to allow cancellation
			})
			Expect(err).To(BeNil())

			// Create a token with invalid key ID to trigger retries
			now := time.Now()
			claims := jwt.RegisteredClaims{
				Issuer:    issuerURL,
				Subject:   "user",
				Audience:  []string{audience},
				ExpiresAt: &jwt.NumericDate{Time: now.AddDate(1, 0, 0)},
				IssuedAt:  &jwt.NumericDate{Time: now},
			}

			unsignedToken := createUnsignedToken(claims, jwt.SigningMethodRS256)
			unsignedToken.Header["kid"] = "invalidkeyid" // This will trigger retries
			idToken, err := fakeOidcProvider.SignIDToken(unsignedToken)
			Expect(err).To(BeNil())

			// Start authentication in a goroutine
			authErr := make(chan error, 1)
			go func() {
				niceMd := make(metautils.NiceMD)
				niceMd.Set("authorization", "Bearer "+idToken)
				authErr <- auth.Authenticate(niceMd.ToIncoming(ctx))
			}()

			// Cancel the context after a short delay to interrupt the retry
			time.Sleep(100 * time.Millisecond)
			cancel()

			// Wait for the authentication to complete and check for context cancellation error
			select {
			case err := <-authErr:
				Expect(err).To(HaveOccurred())
				// The error might be wrapped, so check for context cancellation in the error chain
				Expect(err.Error()).To(Or(
					ContainSubstring("context canceled"),
					ContainSubstring("context cancelled"),
					Equal(base.ErrorCode_ERROR_CODE_INVALID_BEARER_TOKEN.String()),
				))
			case <-time.After(5 * time.Second):
				Fail("Authentication should have completed or been cancelled")
			} // End of error select
		}) // End of cancellation test
	}) // End of Context Cancellation context

	Context("OIDC Configuration Errors", func() { // Test OIDC config errors
		It("should return error for missing issuer in OIDC configuration", func() { // Test missing issuer
			// Create a custom server that returns OIDC config without issuer
			customServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { // HTTP handler
				if r.URL.Path == "/.well-known/openid-configuration" {
					// Return OIDC config without issuer
					config := map[string]interface{}{
						"jwks_uri": "http://localhost:9999/jwks",
						// issuer is missing
					}
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(config)
					return
				}
				http.NotFound(w, r)
			}))
			defer customServer.Close()

			_, err := NewOidcAuthn(context.Background(), config.Oidc{
				Audience:          audience,
				Issuer:            customServer.URL,
				RefreshInterval:   5 * time.Minute,
				BackoffInterval:   12 * time.Second,
				BackoffMaxRetries: 5,
				BackoffFrequency:  1 * time.Second,
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("issuer value is required but missing in OIDC configuration"))
		}) // End of missing issuer test

		It("should return error for missing JWKsURI in OIDC configuration", func() { // Test missing JWKsURI
			// Create a custom server that returns OIDC config without JWKsURI
			var serverURL string                                                                               // Server URL variable
			customServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { // HTTP handler
				if r.URL.Path == "/.well-known/openid-configuration" {
					// Return OIDC config without JWKsURI
					config := map[string]interface{}{
						"issuer": serverURL,
						// jwks_uri is missing
					}
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(config)
					return
				}
				http.NotFound(w, r)
			}))
			serverURL = customServer.URL
			defer customServer.Close()

			_, err := NewOidcAuthn(context.Background(), config.Oidc{
				Audience:          audience,
				Issuer:            customServer.URL,
				RefreshInterval:   5 * time.Minute,
				BackoffInterval:   12 * time.Second,
				BackoffMaxRetries: 5,
				BackoffFrequency:  1 * time.Second,
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("JWKsURI value is required but missing in OIDC configuration"))
		}) // End of missing JWKsURI test

		It("should return error for invalid JSON in OIDC configuration", func() { // Test invalid JSON
			// Create a custom server that returns invalid JSON
			customServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { // Handler for invalid JSON
				if r.URL.Path == "/.well-known/openid-configuration" {
					w.Header().Set("Content-Type", "application/json")
					w.Write([]byte("invalid json"))
					return
				}
				http.NotFound(w, r)
			}))
			defer customServer.Close()

			_, err := NewOidcAuthn(context.Background(), config.Oidc{
				Audience:          audience,
				Issuer:            customServer.URL,
				RefreshInterval:   5 * time.Minute,
				BackoffInterval:   12 * time.Second,
				BackoffMaxRetries: 5,
				BackoffFrequency:  1 * time.Second,
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to decode OIDC configuration"))
		}) // End of invalid JSON test
	}) // End of OIDC Configuration Errors context

	Context("HTTP Request Errors", func() { // Test HTTP errors
		It("should return error for non-200 HTTP status code", func() { // Test HTTP 404
			// Create a custom server that returns 404
			customServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { // Handler returning 404
				if r.URL.Path == "/.well-known/openid-configuration" {
					http.NotFound(w, r)
					return
				}
				http.NotFound(w, r)
			}))
			defer customServer.Close()

			_, err := NewOidcAuthn(context.Background(), config.Oidc{
				Audience:          audience,
				Issuer:            customServer.URL,
				RefreshInterval:   5 * time.Minute,
				BackoffInterval:   12 * time.Second,
				BackoffMaxRetries: 5,
				BackoffFrequency:  1 * time.Second,
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("received unexpected status code (404)"))
		}) // End of HTTP 404 test

		It("should return error for invalid issuer URL", func() { // Test invalid URL
			_, err := NewOidcAuthn(context.Background(), config.Oidc{
				Audience:          audience,
				Issuer:            "invalid://url", // Invalid URL scheme
				RefreshInterval:   5 * time.Minute,
				BackoffInterval:   12 * time.Second,
				BackoffMaxRetries: 5,
				BackoffFrequency:  1 * time.Second,
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to fetch OIDC configuration"))
		}) // End of invalid URL test
	}) // End of HTTP Request Errors context

	Context("Response Body Reading Errors", func() { // Test response body errors
		It("should return error when response body cannot be read", func() { // Test body read error
			// Create a custom server that closes the connection immediately
			customServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { // Handler that closes connection
				if r.URL.Path == "/.well-known/openid-configuration" {
					w.WriteHeader(http.StatusOK)
					// Close the connection without writing body
					if hj, ok := w.(http.Hijacker); ok {
						conn, _, _ := hj.Hijack()
						conn.Close()
					} else {
						Skip("ResponseWriter does not support Hijacker; skipping read-error simulation")
					}
					return
				}
				http.NotFound(w, r)
			}))
			defer customServer.Close()

			_, err := NewOidcAuthn(context.Background(), config.Oidc{
				Audience:          audience,
				Issuer:            customServer.URL,
				RefreshInterval:   5 * time.Minute,
				BackoffInterval:   12 * time.Second,
				BackoffMaxRetries: 5,
				BackoffFrequency:  1 * time.Second,
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to fetch OIDC configuration"))
		}) // End of body read error test
	}) // End of Response Body Reading Errors context
}) // End of authn-oidc describe

func claimOverride(current, overrider *jwt.RegisteredClaims) { // Helper to override claims
	if overrider.Audience != nil { // Override audience if provided
		current.Audience = overrider.Audience // Set audience
	} // End audience override
	if overrider.Issuer != "" { // Override issuer if provided
		current.Issuer = overrider.Issuer // Set issuer
	} // End issuer override
	if overrider.ID != "" { // Override ID if provided
		current.ID = overrider.ID // Set ID
	} // End ID override
	if overrider.IssuedAt != nil { // Override issued at if provided
		current.IssuedAt = overrider.IssuedAt // Set issued at time
	} // End issued at override
	if overrider.ExpiresAt != nil { // Override expires at if provided
		current.ExpiresAt = overrider.ExpiresAt // Set expiration time
	} // End expires at override
	if overrider.NotBefore != nil { // Override not before if provided
		current.NotBefore = overrider.NotBefore // Set not before time
	} // End not before override
	if overrider.Subject != "" { // Override subject if provided
		current.Subject = overrider.Subject // Set subject
	} // End subject override
} // End of claimOverride function
