package oidc

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"

	"github.com/go-jose/go-jose/v3"
	"github.com/golang-jwt/jwt/v4"
)

type fakeOidcProvider struct {
	issuerURL    string
	authPath     string
	tokenPath    string
	userInfoPath string
	JWKSPath     string

	algorithms         []string
	keyIds             map[jwt.SigningMethod]string
	jwks               []jose.JSONWebKey
	rsaPrivateKey      *rsa.PrivateKey
	rsaPrivateKeyForPS *rsa.PrivateKey
	ecdsaPrivateKey    *ecdsa.PrivateKey
	hmacKey            []byte

	mu sync.RWMutex
}

type ProviderConfig struct {
	IssuerURL    string
	AuthPath     string
	TokenPath    string
	UserInfoPath string
	JWKSPath     string
	Algorithms   []string
}

func newFakeOidcProvider(config ProviderConfig) (*fakeOidcProvider, error) {
	rsaPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}
	rsaPrivateKeyForPS, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key for PS: %w", err)
	}
	ecdsaPrivateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ECDSA key: %w", err)
	}
	hmacKey := []byte("hmackeysecret")

	keyIds := map[jwt.SigningMethod]string{
		jwt.SigningMethodRS256: "rs256keyid",
	}
	jwks := []jose.JSONWebKey{
		{
			Key:       rsaPrivateKey.Public(),
			KeyID:     keyIds[jwt.SigningMethodRS256],
			Algorithm: "RS256",
			Use:       "sig",
		},
	}

	return &fakeOidcProvider{
		issuerURL:          config.IssuerURL,
		authPath:           config.AuthPath,
		tokenPath:          config.TokenPath,
		userInfoPath:       config.UserInfoPath,
		JWKSPath:           config.JWKSPath,
		algorithms:         config.Algorithms,
		rsaPrivateKey:      rsaPrivateKey,
		rsaPrivateKeyForPS: rsaPrivateKeyForPS,
		hmacKey:            hmacKey,
		jwks:               jwks,
		ecdsaPrivateKey:    ecdsaPrivateKey,
		keyIds:             keyIds,
		mu:                 sync.RWMutex{},
	}, nil
}

func (s *fakeOidcProvider) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	switch r.URL.Path {
	case "/.well-known/openid-configuration":
		s.responseWellKnown(w)
	case s.JWKSPath:
		s.responseJWKS(w)
	case s.authPath, s.tokenPath, s.userInfoPath:
		httpError(w, http.StatusNotFound)
	default:
		httpError(w, http.StatusNotFound)
	}
}

type providerJSON struct {
	Issuer      string   `json:"issuer"`
	AuthURL     string   `json:"authorization_endpoint"`
	TokenURL    string   `json:"token_endpoint"`
	JWKSURL     string   `json:"jwks_uri"`
	UserInfoURL string   `json:"userinfo_endpoint"`
	Algorithms  []string `json:"id_token_signing_alg_values_supported"`
}

func (s *fakeOidcProvider) responseWellKnown(w http.ResponseWriter) {
	jso := providerJSON{
		Issuer:      s.issuerURL,
		AuthURL:     s.issuerURL + s.authPath,
		TokenURL:    s.issuerURL + s.tokenPath,
		JWKSURL:     s.issuerURL + s.JWKSPath,
		UserInfoURL: s.issuerURL + s.userInfoPath,
		Algorithms:  s.algorithms,
	}
	httpJSON(w, jso)
}

func (s *fakeOidcProvider) responseJWKS(w http.ResponseWriter) {
	jwks := &jose.JSONWebKeySet{
		Keys: s.jwks,
	}
	httpJSON(w, jwks)
}

func httpJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	e := json.NewEncoder(w)
	e.SetIndent("", "  ")
	if err := e.Encode(v); err != nil {
		httpError(w, http.StatusInternalServerError)
	}
}

func httpError(w http.ResponseWriter, code int) {
	http.Error(w, http.StatusText(code), code)
}

func (s *fakeOidcProvider) UpdateKeyID(method jwt.SigningMethod, newKeyID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.keyIds[method] = newKeyID
	for i, key := range s.jwks {
		if key.Algorithm == method.Alg() {
			s.jwks[i].KeyID = newKeyID
		}
	}
}

func (s *fakeOidcProvider) SignIDToken(unsignedToken *jwt.Token) (string, error) {
	var signedToken string
	var err error

	switch unsignedToken.Method {
	case jwt.SigningMethodRS256:
		signedToken, err = unsignedToken.SignedString(s.rsaPrivateKey)
	default:
		return "", fmt.Errorf("incorrect signing method type, supported algorithms: HS256, RS256, ES256, PS256")
	}

	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func createUnsignedToken(regClaims jwt.RegisteredClaims, method jwt.SigningMethod) *jwt.Token {
	claims := struct {
		jwt.RegisteredClaims
	}{
		RegisteredClaims: regClaims,
	}
	return jwt.NewWithClaims(method, claims)
}

func fakeHttpServer(url string, handler http.HandlerFunc) (*httptest.Server, error) {
	l, err := net.Listen("tcp", url)
	if err != nil {
		return nil, fmt.Errorf("failed to start listener: %w", err)
	}
	ts := httptest.NewUnstartedServer(handler)
	_ = ts.Listener.Close()
	ts.Listener = l
	ts.Start()
	return ts, nil
}
