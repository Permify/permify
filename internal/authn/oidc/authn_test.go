package oidc

import (
	"context"
	"fmt"
	"net/http/httptest"
	"time"

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
	fakeOidcProvider, _ := newFakeOidcProvider(issuerURL)

	var server *httptest.Server

	BeforeEach(func() {
		var err error
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
					BackoffInterval:   12 * time.Second,
					BackoffMaxRetries: 5,
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
					BackoffInterval:   12 * time.Second,
					BackoffMaxRetries: 5,
				})
				Expect(err).To(BeNil())

				// authenticate token
				niceMd := make(metautils.NiceMD)
				niceMd.Set("authorization", "Bearer "+idToken)
				err = auth.Authenticate(niceMd.ToIncoming(ctx))
				Expect(err != nil).To(Equal(tt.wantErr), fmt.Sprintf("Wanted error: %t, got %v", tt.wantErr, err))
			}
		})
	})

	Context("Authenticate Key Ids", func() {
		It("Case 1", func() {
			tests := []struct {
				name     string
				method   jwt.SigningMethod
				addKeyId bool
				keyId    string
				wantErr  bool
			}{
				{
					"With no keyid using RS256 it should fail, multiple public RSA keys matching for RS256 and PS256",
					jwt.SigningMethodRS256,
					false, "",
					true,
				},
				{
					"With right keyid using RS256 it should pass",
					jwt.SigningMethodRS256,
					true, fakeOidcProvider.keyIds[jwt.SigningMethodRS256],
					false,
				},
				{
					"With wrong keyid using RS256 it should fail",
					jwt.SigningMethodRS256,
					true, "wrongkeyid",
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
					BackoffInterval:   12 * time.Second,
					BackoffMaxRetries: 5,
				})
				Expect(err).To(BeNil())

				// authenticate
				niceMd := make(metautils.NiceMD)
				niceMd.Set("authorization", "Bearer "+idToken)
				err = auth.Authenticate(niceMd.ToIncoming(ctx))
				Expect(err != nil).To(Equal(tt.wantErr), fmt.Sprintf("Wanted error: %t, got %v", tt.wantErr, err))
			}
		})
	})

	Context("Missing Config", func() {
		It("Case 1", func() {
			_, err := NewOidcAuthn(context.Background(), config.Oidc{
				Audience:          "",
				Issuer:            "https://wrong-url",
				BackoffInterval:   12 * time.Second,
				BackoffMaxRetries: 5,
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
				BackoffInterval:   12 * time.Second,
				BackoffMaxRetries: 5,
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
				BackoffInterval:   12 * time.Second,
				BackoffMaxRetries: 5,
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
				BackoffInterval:   12 * time.Second,
				BackoffMaxRetries: 5,
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
