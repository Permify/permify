package oidc

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/internal/config"
)

func Test_AuthenticateWithSigningMethods(t *testing.T) {
	RegisterFailHandler(fail(t))

	clientId := "test-client"
	listenAddress := "localhost:9999"
	issuerURL := "http://" + listenAddress

	// Start oidc provider server
	fakeOidcProvider, err := newfakeOidcProvider(issuerURL)
	Expect(err).To(BeNil())
	server, err := fakeHttpServer(listenAddress, fakeOidcProvider.ServeHTTP)
	Expect(err).To(BeNil())
	defer server.Close()

	tests := []struct {
		name    string
		method  jwt.SigningMethod
		wantErr bool
	}{
		{"Should pass with RS256",
			jwt.SigningMethodRS256,
			false,
		},
		{"Should fail with HS256, zitadel/oidc does not support HSXXX algorithms by default",
			// see https://github.com/zitadel/oidc/blob/v1.13.0/pkg/oidc/keyset.go#L94
			jwt.SigningMethodHS256,
			true,
		},
		{"Should pass with ES256",
			jwt.SigningMethodES256,
			false,
		},
		{"Should pass with PS256",
			jwt.SigningMethodPS256,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now := time.Now()
			claims := jwt.RegisteredClaims{
				Issuer:    issuerURL,
				Subject:   "user",
				Audience:  []string{clientId},
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
				ClientId: clientId,
				Issuer:   issuerURL,
			})
			Expect(err).To(BeNil())

			// authenticate
			niceMd := make(metautils.NiceMD)
			niceMd.Set("authorization", "Bearer "+idToken)
			err = auth.Authenticate(niceMd.ToIncoming(ctx))
			Expect(err != nil).To(Equal(tt.wantErr), fmt.Sprintf("Wanted error: %t, got %v", tt.wantErr, err))
		})
	}
}

func Test_AuthenticateClaims(t *testing.T) {
	RegisterFailHandler(fail(t))

	clientId := "test-client"
	listenAddress := "localhost:9999"
	issuerURL := "http://" + listenAddress

	// Start oidc provider server
	fakeOidcProvider, err := newfakeOidcProvider(issuerURL)
	Expect(err).To(BeNil())
	server, err := fakeHttpServer(listenAddress, fakeOidcProvider.ServeHTTP)
	Expect(err).To(BeNil())
	defer server.Close()

	tests := []struct {
		name          string
		claimOverride *jwt.RegisteredClaims
		wantErr       bool
	}{
		{"With correct values there should be no errors",
			&jwt.RegisteredClaims{},
			false,
		},
		{"Wrong issuer in the token, it should fail",
			&jwt.RegisteredClaims{
				Issuer: "http://wrong-issuer",
			},
			true,
		},
		{"Wrong clientID in the token, it should fail",
			&jwt.RegisteredClaims{
				Audience: []string{"wrong-clientid"},
			},
			true,
		},
		{"Expired Token, it should fail",
			&jwt.RegisteredClaims{
				ExpiresAt: &jwt.NumericDate{Time: time.Date(1999, 1, 0, 0, 0, 0, 0, time.UTC)},
			},
			true,
		},
		{"Issued at the future, it should fail",
			&jwt.RegisteredClaims{
				IssuedAt: &jwt.NumericDate{Time: time.Date(3999, 1, 0, 0, 0, 0, 0, time.UTC)},
			},
			true,
		},
		{"Token used before its NotBefore date, it should fail",
			&jwt.RegisteredClaims{
				NotBefore: &jwt.NumericDate{Time: time.Date(3999, 1, 0, 0, 0, 0, 0, time.UTC)},
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now := time.Now()
			claims := jwt.RegisteredClaims{
				Issuer:    issuerURL,
				Subject:   "user",
				Audience:  []string{clientId},
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
				ClientId: clientId,
				Issuer:   issuerURL,
			})
			Expect(err).To(BeNil())

			// authenticate token
			niceMd := make(metautils.NiceMD)
			niceMd.Set("authorization", "Bearer "+idToken)
			err = auth.Authenticate(niceMd.ToIncoming(ctx))
			Expect(err != nil).To(Equal(tt.wantErr), fmt.Sprintf("Wanted error: %t, got %v", tt.wantErr, err))
		})
	}
}

func Test_AuthenticateKeyIds(t *testing.T) {
	RegisterFailHandler(fail(t))

	clientId := "test-client"
	listenAddress := "localhost:9999"
	issuerURL := "http://" + listenAddress

	// Start oidc provider server
	fakeOidcProvider, err := newfakeOidcProvider(issuerURL)
	Expect(err).To(BeNil())
	server, err := fakeHttpServer(listenAddress, fakeOidcProvider.ServeHTTP)
	Expect(err).To(BeNil())
	defer server.Close()

	tests := []struct {
		name     string
		method   jwt.SigningMethod
		addKeyId bool
		keyId    string
		wantErr  bool
	}{
		{"With no keyid using RS256 it should fail, multiple public RSA keys matching for RS256 and PS256",
			jwt.SigningMethodRS256,
			false, "",
			true,
		},
		{"With no keyid using ES256 it should pass, unique public ecdsa key in keyset",
			jwt.SigningMethodES256,
			true, "",
			false,
		},
		{"With right keyid using RS256 it should pass",
			jwt.SigningMethodRS256,
			true, fakeOidcProvider.keyIds[jwt.SigningMethodRS256],
			false,
		},
		{"With wrong keyid using RS256 it should fail",
			jwt.SigningMethodRS256,
			true, "wrongkeyid",
			true,
		},
		{"With keyid belonging to a different key it should fail",
			jwt.SigningMethodES256,
			true, fakeOidcProvider.keyIds[jwt.SigningMethodRS256],
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now := time.Now()
			claims := jwt.RegisteredClaims{
				Issuer:    issuerURL,
				Subject:   "user",
				Audience:  []string{clientId},
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
				ClientId: clientId,
				Issuer:   issuerURL,
			})
			Expect(err).To(BeNil())

			// authenticate
			niceMd := make(metautils.NiceMD)
			niceMd.Set("authorization", "Bearer "+idToken)
			err = auth.Authenticate(niceMd.ToIncoming(ctx))
			Expect(err != nil).To(Equal(tt.wantErr), fmt.Sprintf("Wanted error: %t, got %v", tt.wantErr, err))
		})
	}
}

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

func fail(t *testing.T) func(message string, callerSkip ...int) {
	return func(message string, callerSkip ...int) {
		t.Errorf(message)
	}
}
