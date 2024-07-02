package oidc

import (
	"context"
	"fmt"
	"net/http/httptest"
	"sync"
	"time"

	"google.golang.org/grpc/metadata"

	"github.com/golang-jwt/jwt/v4"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/internal/config"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

var _ = Describe("authn-oidc", func() {
	audience := "aud"
	listenAddress := "localhost:9999"
	issuerURL := "http://" + listenAddress
	var fakeOidcProvider *fakeOidcProvider

	var server *httptest.Server

	BeforeEach(func() {
		var err error

		fakeOidcProvider, err = newFakeOidcProvider(ProviderConfig{
			IssuerURL:    issuerURL,
			AuthPath:     "/auth",
			TokenPath:    "/token",
			UserInfoPath: "/userInfo",
			JWKSPath:     "/jwks",
			Algorithms:   []string{"RS256", "HS256", "ES256", "PS256"},
		})
		Expect(err).To(BeNil())

		server, err = fakeHttpServer(listenAddress, fakeOidcProvider.ServeHTTP)
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		server.Close()
	})

	Context("Authenticate With Signing Methods", func() {
		It("Case 1", func() {
			tests := []struct {
				name   string
				method jwt.SigningMethod
				err    error
			}{
				{
					"Should pass with RS256",
					jwt.SigningMethodRS256,
					nil,
				},
			}

			for _, tt := range tests {
				now := time.Now()
				claims := jwt.RegisteredClaims{
					Issuer:    issuerURL,
					Subject:   "user",
					Audience:  []string{audience},
					ExpiresAt: &jwt.NumericDate{Time: now.AddDate(1, 0, 0)},
					IssuedAt:  &jwt.NumericDate{Time: now},
				}

				// create signed token from oidc provider
				unsignedToken := createUnsignedToken(claims, tt.method)
				unsignedToken.Header["kid"] = fakeOidcProvider.keyIds[tt.method]
				idToken, err := fakeOidcProvider.SignIDToken(unsignedToken)
				Expect(err).To(BeNil())

				// create oidc authenticator
				ctx := context.Background()
				auth, err := NewOidcAuthn(ctx, config.Oidc{
					Audience:          audience,
					Issuer:            issuerURL,
					RefreshInterval:   5 * time.Minute,
					BackoffInterval:   12 * time.Second,
					BackoffMaxRetries: 5,
					BackoffFrequency:  1 * time.Second,
				})
				Expect(err).To(BeNil())

				// authenticate
				niceMd := make(metautils.NiceMD)
				niceMd.Set("authorization", "Bearer "+idToken)
				err = auth.Authenticate(niceMd.ToIncoming(ctx))
				if tt.err == nil {
					Expect(err).To(BeNil())
				} else {
					Expect(err).To(Equal(tt.err))
				}
			}
		})
	})

	Context("Authenticate Claims", func() {
		It("Case 1", func() {
			tests := []struct {
				name          string
				claimOverride *jwt.RegisteredClaims
				wantErr       bool
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
				},
			}

			for _, tt := range tests {
				now := time.Now()
				claims := jwt.RegisteredClaims{
					Issuer:    issuerURL,
					Subject:   "user",
					Audience:  []string{audience},
					ExpiresAt: &jwt.NumericDate{Time: now.AddDate(1, 0, 0)},
					IssuedAt:  &jwt.NumericDate{Time: now},
				}
				claimOverride(&claims, tt.claimOverride)

				// create signed token from oidc server with overridden claims
				unsignedToken := createUnsignedToken(claims, jwt.SigningMethodRS256)
				unsignedToken.Header["kid"] = fakeOidcProvider.keyIds[jwt.SigningMethodRS256]
				idToken, err := fakeOidcProvider.SignIDToken(unsignedToken)
				Expect(err).To(BeNil())

				// create oidc authenticator
				ctx := context.Background()
				auth, err := NewOidcAuthn(ctx, config.Oidc{
					Audience:          audience,
					Issuer:            issuerURL,
					RefreshInterval:   5 * time.Minute,
					BackoffInterval:   12 * time.Second,
					BackoffMaxRetries: 5,
					BackoffFrequency:  1 * time.Second,
				})
				Expect(err).To(BeNil())

				// authenticate token
				niceMd := make(metautils.NiceMD)
				niceMd.Set("authorization", "Bearer "+idToken)
				err = auth.Authenticate(niceMd.ToIncoming(ctx))
				Expect(err != nil).To(Equal(tt.wantErr), fmt.Sprintf("Wanted error: %t, got %v", tt.wantErr, err))
				Expect(time.Now()).To(BeTemporally("<=", now.Add(1*time.Second)))
			}
		})
	})

	Context("Authenticate Key Ids", func() {
		It("Case 1", func() {
			tests := []struct {
				name       string
				method     jwt.SigningMethod
				addKeyId   bool
				keyId      string
				wantErr    bool
				wantTiming time.Duration
			}{
				{
					"With no keyid using RS256 it should fail, multiple public RSA keys matching for RS256 and PS256",
					jwt.SigningMethodRS256,
					false, "",
					true,
					1 * time.Second,
				},
				{
					"With right keyid using RS256 it should pass",
					jwt.SigningMethodRS256,
					true, fakeOidcProvider.keyIds[jwt.SigningMethodRS256],
					false,
					1 * time.Second,
				},
				{
					"With wrong keyid using RS256 it should fail",
					jwt.SigningMethodRS256,
					true, "wrongkeyid",
					true,
					6 * time.Second,
				},
			}

			for _, tt := range tests {
				now := time.Now()
				claims := jwt.RegisteredClaims{
					Issuer:    issuerURL,
					Subject:   "user",
					Audience:  []string{audience},
					ExpiresAt: &jwt.NumericDate{Time: now.AddDate(1, 0, 0)},
					IssuedAt:  &jwt.NumericDate{Time: now},
				}

				// create signed token from oidc provider possibly with kid in header
				unsignedToken := createUnsignedToken(claims, tt.method)
				if tt.addKeyId {
					unsignedToken.Header["kid"] = tt.keyId
				}
				idToken, err := fakeOidcProvider.SignIDToken(unsignedToken)
				Expect(err).To(BeNil())

				// create authenticator
				ctx := context.Background()
				auth, err := NewOidcAuthn(ctx, config.Oidc{
					Audience:          audience,
					Issuer:            issuerURL,
					RefreshInterval:   5 * time.Minute,
					BackoffInterval:   12 * time.Second,
					BackoffMaxRetries: 5,
					BackoffFrequency:  1 * time.Second,
				})
				Expect(err).To(BeNil())

				// authenticate
				niceMd := make(metautils.NiceMD)
				niceMd.Set("authorization", "Bearer "+idToken)
				err = auth.Authenticate(niceMd.ToIncoming(ctx))
				Expect(err != nil).To(Equal(tt.wantErr), fmt.Sprintf("Wanted error: %t, got %v", tt.wantErr, err))
				Expect(time.Now()).To(BeTemporally("<=", now.Add(tt.wantTiming)))
			}
		})

		It("Case 2", func() {
			// create authenticator
			ctx := context.Background()
			auth, err := NewOidcAuthn(ctx, config.Oidc{
				Audience:          audience,
				Issuer:            issuerURL,
				RefreshInterval:   5 * time.Minute,
				BackoffInterval:   1 * time.Minute,
				BackoffMaxRetries: 5,
				BackoffFrequency:  1 * time.Second,
			})
			Expect(err).To(BeNil())

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
				},
			}

			for _, tt := range tests {
				fakeOidcProvider.UpdateKeyID(tt.method, tt.newKeyId)

				now := time.Now()
				claims := jwt.RegisteredClaims{
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
				Expect(time.Now()).To(BeTemporally("<=", now.Add(1*time.Second)))
			}
		})

		It("Case 3: Complex test for maximum retries and backoff interval", func() {
			// create authenticator
			ctx := context.Background()
			auth, err := NewOidcAuthn(ctx, config.Oidc{
				Audience:          audience,
				Issuer:            issuerURL,
				RefreshInterval:   5 * time.Minute,
				BackoffInterval:   6 * time.Second,
				BackoffMaxRetries: 3,
				BackoffFrequency:  1 * time.Second,
			})
			Expect(err).To(BeNil())

			tests := []struct {
				name              string
				method            jwt.SigningMethod
				newKeyId          string
				succeedAfterRetry bool
			}{
				{
					"Invalid KID, should retry and fail, and then succeed",
					jwt.SigningMethodRS256,
					"invalidkey1",
					true,
				},
				{
					"Invalid KID, should retry and fail",
					jwt.SigningMethodRS256,
					"invalidkey2",
					false,
				},
			}

			for _, tt := range tests {
				now := time.Now()
				claims := jwt.RegisteredClaims{
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
				Expect(err).To(BeNil())

				// authenticate with retries
				niceMd := make(metautils.NiceMD)
				niceMd.Set("authorization", "Bearer "+idToken)
				now = time.Now()
				err = auth.Authenticate(niceMd.ToIncoming(ctx))
				Expect(err).ShouldNot(BeNil())
				Expect(time.Now()).To(BeTemporally("<=", now.Add(4*time.Second)))

				// authenticate after retries should fail immediately
				now = time.Now()
				err = auth.Authenticate(niceMd.ToIncoming(ctx))
				Expect(err).ShouldNot(BeNil())
				Expect(time.Now()).To(BeTemporally("<=", now.Add(1*time.Second)))

				if tt.succeedAfterRetry {
					// authenticate with now valid key should succeed immediately after backoff interval elasped
					fakeOidcProvider.UpdateKeyID(jwt.SigningMethodRS256, tt.newKeyId)

					time.Sleep(7 * time.Second)

					now = time.Now()
					err = auth.Authenticate(niceMd.ToIncoming(ctx))
					Expect(err).Should(BeNil())
					Expect(time.Now()).To(BeTemporally("<=", now.Add(1*time.Second)))
				}
			}
		})

		It("Case 4: Concurrent requests leading to global backoff lock for 6 seconds", func() {
			// create authenticator
			ctx := context.Background()
			auth, err := NewOidcAuthn(ctx, config.Oidc{
				Audience:          audience,
				Issuer:            issuerURL,
				RefreshInterval:   5 * time.Minute,
				BackoffInterval:   6 * time.Second,
				BackoffMaxRetries: 3,
				BackoffFrequency:  1 * time.Second,
			})
			Expect(err).To(BeNil())

			invalidKeyIDs := []string{"invalidkey1", "invalidkey2"}

			var wg sync.WaitGroup
			numRequests := len(invalidKeyIDs) * 3 // Send each invalid key multiple times to ensure retries are hit

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
			}

			// Set valid token to ensure it's already in cache for later tests
			validKeyID := "validkey"
			fakeOidcProvider.UpdateKeyID(jwt.SigningMethodRS256, validKeyID)

			// Step 1: Trigger backoff by hitting max retries with invalid keys concurrently
			now := time.Now()
			for i := 0; i < numRequests; i++ {
				wg.Add(1)
				go func(i int) {
					defer wg.Done()
					keyID := invalidKeyIDs[i%len(invalidKeyIDs)]
					token, _ := createTokenWithKid(keyID)
					niceMd := make(metautils.NiceMD)
					niceMd.Set("authorization", "Bearer "+token)
					now := time.Now()
					err := auth.Authenticate(niceMd.ToIncoming(ctx))
					Expect(err.Error()).Should(Equal(base.ErrorCode_ERROR_CODE_INVALID_BEARER_TOKEN.String()))
					Expect(time.Now()).To(BeTemporally("<=", now.Add(4*time.Second)))
				}(i)
			}

			wg.Wait()
			Expect(time.Now()).To(BeTemporally("<=", now.Add(4*time.Second)))

			// Step 2: Verify that retries are immediately rejected during backoff period
			for _, keyID := range invalidKeyIDs {
				token, _ := createTokenWithKid(keyID)
				md := metadata.Pairs("authorization", "Bearer "+token)
				ctx := metadata.NewIncomingContext(ctx, md)
				now := time.Now()
				err := auth.Authenticate(ctx)
				Expect(err.Error()).Should(Equal(base.ErrorCode_ERROR_CODE_INVALID_BEARER_TOKEN.String()))
				Expect(time.Now()).To(BeTemporally("<=", now.Add(1*time.Second)))
			}

			// Step 3: A valid KID already in the JWKS cache should continue to authenticate succesfully immediately
			validToken, _ := createTokenWithKid(validKeyID)
			niceMd := make(metautils.NiceMD)
			niceMd.Set("authorization", "Bearer "+validToken)

			now = time.Now()
			err = auth.Authenticate(niceMd.ToIncoming(ctx))
			Expect(err).Should(BeNil())
			Expect(time.Now()).To(BeTemporally("<=", now.Add(1*time.Second)))

			// Step 4: Ensure that invalid keys return immediately during backoff interval, even after valid KID
			for _, keyID := range invalidKeyIDs {
				token, _ := createTokenWithKid(keyID)
				md := metadata.Pairs("authorization", "Bearer "+token)
				ctx := metadata.NewIncomingContext(ctx, md)
				now = time.Now()
				err := auth.Authenticate(ctx)
				Expect(err.Error()).Should(Equal(base.ErrorCode_ERROR_CODE_INVALID_BEARER_TOKEN.String()))
				Expect(time.Now()).To(BeTemporally("<=", now.Add(1*time.Second)))
			}

			// Step 5: Ensure that invalid keys are still rejected after backoff period
			time.Sleep(7 * time.Second)

			now = time.Now()
			for _, keyID := range invalidKeyIDs {
				token, _ := createTokenWithKid(keyID)
				niceMd := make(metautils.NiceMD)
				niceMd.Set("authorization", "Bearer "+token)
				err = auth.Authenticate(niceMd.ToIncoming(ctx))
				Expect(err.Error()).Should(Equal(base.ErrorCode_ERROR_CODE_INVALID_BEARER_TOKEN.String()))
			}
			Expect(time.Now()).To(BeTemporally("<=", now.Add(4*time.Second)))
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
	})
})

func claimOverride(current, overrider *jwt.RegisteredClaims) {
	if overrider.Audience != nil {
		current.Audience = overrider.Audience
	}
	if overrider.Issuer != "" {
		current.Issuer = overrider.Issuer
	}
	if overrider.ID != "" {
		current.ID = overrider.ID
	}
	if overrider.IssuedAt != nil {
		current.IssuedAt = overrider.IssuedAt
	}
	if overrider.ExpiresAt != nil {
		current.ExpiresAt = overrider.ExpiresAt
	}
	if overrider.NotBefore != nil {
		current.NotBefore = overrider.NotBefore
	}
	if overrider.Subject != "" {
		current.Subject = overrider.Subject
	}
}
