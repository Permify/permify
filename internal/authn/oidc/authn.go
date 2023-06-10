package oidc

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	grpcAuth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/zitadel/oidc/pkg/client"
	"github.com/zitadel/oidc/pkg/client/rp"
	"github.com/zitadel/oidc/pkg/oidc"

	"github.com/Permify/permify/internal/config"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// OidcAuthenticator - Interface for oidc authenticator
type OidcAuthenticator interface {
	Authenticate(ctx context.Context) error
}

// OidcAuthn - Oidc verifier structure
type OidcAuthn struct {
	verifier rp.IDTokenVerifier
}

// NewOidcAuthn - Create new Oidc verifier
func NewOidcAuthn(_ context.Context, cfg config.Oidc) (*OidcAuthn, error) {
	dis, err := client.Discover(cfg.Issuer, http.DefaultClient)
	if err != nil {
		return nil, err
	}
	remoteKeySet := rp.NewRemoteKeySet(http.DefaultClient, dis.JwksURI)
	verifier := rp.NewIDTokenVerifier(dis.Issuer, cfg.ClientId, remoteKeySet,
		rp.WithSupportedSigningAlgorithms(dis.IDTokenSigningAlgValuesSupported...))

	return &OidcAuthn{verifier: verifier}, nil
}

// Authenticate - Checking whether JWT token is signed by the provider and is valid
func (t *OidcAuthn) Authenticate(ctx context.Context) error {
	rawToken, err := grpcAuth.AuthFromMD(ctx, "Bearer")
	if err != nil {
		return errors.New(base.ErrorCode_ERROR_CODE_MISSING_BEARER_TOKEN.String())
	}

	claims, err := rp.VerifyIDToken(ctx, rawToken, t.verifier)
	if err != nil {
		return status.Error(codes.Unauthenticated, err.Error())
	}

	if err := t.validateOtherClaims(claims); err != nil {
		return status.Error(codes.Unauthenticated, err.Error())
	}
	return nil
}

// validateOtherClaims - Validate claims that are not validated by the oidc client library
func (t *OidcAuthn) validateOtherClaims(claims oidc.IDTokenClaims) error {
	return checkNotBefore(claims, t.verifier.Offset())
}

// checkNotBefore - Validate current time to be not before the notBefore claim value
// Use the same logic for time comparisons as in the zitadel oidc verifier
func checkNotBefore(claims oidc.IDTokenClaims, offset time.Duration) error {
	notBefore := claims.GetNotBefore().Round(time.Second)
	if notBefore.IsZero() {
		return nil
	}

	nowWithOffset := time.Now().UTC().Add(offset).Round(time.Second)
	if nowWithOffset.Before(notBefore) {
		return fmt.Errorf("token's notBefore date is %s, should be used after that time", notBefore)
	}
	return nil
}
