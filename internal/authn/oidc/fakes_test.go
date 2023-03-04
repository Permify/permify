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
}

func newfakeOidcProvider(issuerURL string) (*fakeOidcProvider, error) {
	rsaPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	rsaPrivateKeyForPS, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	ecdsaPrivateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	hmacKey := []byte("hmackeysecret")

	keyIds := map[jwt.SigningMethod]string{
		jwt.SigningMethodRS256: "rs256keyid",
		jwt.SigningMethodES256: "es256keyid",
		jwt.SigningMethodHS256: "hs256keyid",
		jwt.SigningMethodPS256: "ps256keyid",
	}
	jwks := []jose.JSONWebKey{
		{
			Key:       rsaPrivateKey.Public(),
			KeyID:     keyIds[jwt.SigningMethodRS256],
			Algorithm: "RS256",
			Use:       "sig",
		},
		{
			Key:       ecdsaPrivateKey.Public(),
			KeyID:     keyIds[jwt.SigningMethodES256],
			Algorithm: "ES256",
			Use:       "sig",
		},
		{
			Key:       hmacKey,
			KeyID:     keyIds[jwt.SigningMethodHS256],
			Algorithm: "HS256",
			Use:       "sig",
		},
		{
			Key:       rsaPrivateKeyForPS.Public(),
			KeyID:     keyIds[jwt.SigningMethodPS256],
			Algorithm: "PS256",
			Use:       "sig",
		},
	}

	return &fakeOidcProvider{
		issuerURL:          issuerURL,
		authPath:           "/auth",
		tokenPath:          "/token",
		userInfoPath:       "/userInfo",
		JWKSPath:           "/jwks",
		algorithms:         []string{"RS256", "HS256", "ES256", "PS256"},
		rsaPrivateKey:      rsaPrivateKey,
		rsaPrivateKeyForPS: rsaPrivateKeyForPS,
		hmacKey:            hmacKey,
		jwks:               jwks,
		ecdsaPrivateKey:    ecdsaPrivateKey,
		keyIds:             keyIds,
	}, nil
}

func (s *fakeOidcProvider) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/.well-known/openid-configuration":
		s.ResponseWellKnown(w, r)
	case s.JWKSPath:
		s.ResponseJWKS(w, r)
	case s.authPath, s.tokenPath, s.userInfoPath:
		httpError(w, 404)
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

func (s *fakeOidcProvider) ResponseWellKnown(w http.ResponseWriter, r *http.Request) {
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

func (s *fakeOidcProvider) ResponseJWKS(w http.ResponseWriter, r *http.Request) {
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
		return
	}
}

func httpError(w http.ResponseWriter, code int) {
	http.Error(w, http.StatusText(code), code)
}

func (s *fakeOidcProvider) SignIDToken(unsignedToken *jwt.Token) (string, error) {
	signedToken := ""
	var err error

	switch unsignedToken.Method {
	case jwt.SigningMethodHS256:
		signedToken, err = unsignedToken.SignedString(s.hmacKey)
	case jwt.SigningMethodRS256:
		signedToken, err = unsignedToken.SignedString(s.rsaPrivateKey)
	case jwt.SigningMethodES256:
		signedToken, err = unsignedToken.SignedString(s.ecdsaPrivateKey)
	case jwt.SigningMethodPS256:
		signedToken, err = unsignedToken.SignedString(s.rsaPrivateKeyForPS)

	default:
		return "", fmt.Errorf("Incorrect signing method type, supported algorithms: HS256, RS256, ES256, PS256")
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
	token := jwt.NewWithClaims(method, claims)
	return token
}

func fakeHttpServer(url string, handler http.HandlerFunc) (*httptest.Server, error) {
	l, err := net.Listen("tcp", url)
	if err != nil {
		return nil, err
	}
	ts := httptest.NewUnstartedServer(handler)
	_ = ts.Listener.Close()
	ts.Listener = l
	ts.Start()
	return ts, nil
}
