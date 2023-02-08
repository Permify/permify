package multitenant

import (
	"github.com/golang-jwt/jwt/v4"
	grpcAuth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"golang.org/x/net/context"

	"github.com/Permify/permify/internal/authn"
)

// TenantAuthenticator - Interface for key authenticator
type TenantAuthenticator interface {
	Authenticate(ctx context.Context, reqTenantID string) error
}

// TenantAuthn - Authentication Keys Structure
type TenantAuthn struct {
	secret     string
	algorithms []string
}

// NewTenantAuthn - Create New NewTenantAuthn Keys
func NewTenantAuthn(secret string, algorithms []string) (*TenantAuthn, error) {
	return &TenantAuthn{
		secret:     secret,
		algorithms: algorithms,
	}, nil
}

// Authenticate - Checking whether any API request contain keys
func (t *TenantAuthn) Authenticate(ctx context.Context, reqTenantID string) error {
	header, err := grpcAuth.AuthFromMD(ctx, "Bearer")
	if err != nil {
		return authn.MissingBearerTokenError
	}
	jwtParser := jwt.NewParser(jwt.WithValidMethods(t.algorithms))
	var token *jwt.Token
	token, err = jwtParser.Parse(header, func(token *jwt.Token) (any, error) {
		return []byte(t.secret), nil
	})
	if err != nil {
		return authn.Unauthenticated
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		var tenantID string
		tenantID, ok = claims["tenant_id"].(string)
		if !ok {
			return authn.MissingTenantIDError
		}
		if reqTenantID != tenantID {
			return authn.Unauthenticated
		}
		return nil
	}
	return authn.Unauthenticated
}
