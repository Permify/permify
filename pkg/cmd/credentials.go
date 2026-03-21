package cmd

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"gopkg.in/yaml.v3"
)

// CredentialsFile is the YAML format stored at ~/.permify/credentials (endpoint, optional api_token, tls_ca_path).
type CredentialsFile struct {
	Endpoint  string `yaml:"endpoint"`
	APIToken  string `yaml:"api_token"`
	TLSCAPath string `yaml:"tls_ca_path"`
}

// DefaultCredentialsPath returns $HOME/.permify/credentials.
func DefaultCredentialsPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, ".permify", "credentials"), nil
}

// ResolveCredentialsPath returns flagPath if set, otherwise DefaultCredentialsPath.
func ResolveCredentialsPath(flagPath string) (string, error) {
	if strings.TrimSpace(flagPath) != "" {
		abs, err := filepath.Abs(flagPath)
		if err != nil {
			return "", fmt.Errorf("resolve credentials path: %w", err)
		}
		return abs, nil
	}
	return DefaultCredentialsPath()
}

// LoadCredentials reads and parses a credentials YAML file.
func LoadCredentials(path string) (*CredentialsFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("credentials file not found at %q; create it with endpoint (and optional api_token, tls_ca_path): %w", path, err)
		}
		return nil, fmt.Errorf("read credentials file: %w", err)
	}

	var c CredentialsFile
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("parse credentials YAML: %w", err)
	}

	c.Endpoint = strings.TrimSpace(c.Endpoint)
	c.APIToken = strings.TrimSpace(c.APIToken)
	c.TLSCAPath = strings.TrimSpace(c.TLSCAPath)

	if c.Endpoint == "" {
		return nil, fmt.Errorf("credentials file %q: endpoint is required", path)
	}

	return &c, nil
}

type bearerTokenCreds struct {
	token string
}

func (b bearerTokenCreds) GetRequestMetadata(_ context.Context, _ ...string) (map[string]string, error) {
	return map[string]string{"authorization": "Bearer " + b.token}, nil
}

func (b bearerTokenCreds) RequireTransportSecurity() bool {
	return true
}

// GRPCDialOptions builds dial options: TLS + bearer token when api_token is set, otherwise insecure.
func GRPCDialOptions(c *CredentialsFile) ([]grpc.DialOption, error) {
	token := strings.TrimSpace(c.APIToken)
	if token == "" {
		return []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}, nil
	}

	var tlsCfg *tls.Config
	if c.TLSCAPath != "" {
		pemData, err := os.ReadFile(c.TLSCAPath)
		if err != nil {
			return nil, fmt.Errorf("read tls_ca_path: %w", err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(pemData) {
			return nil, fmt.Errorf("tls_ca_path: no PEM certificates found")
		}
		tlsCfg = &tls.Config{
			RootCAs:    pool,
			MinVersion: tls.VersionTLS12,
		}
	} else {
		tlsCfg = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	tc := credentials.NewTLS(tlsCfg)
	return []grpc.DialOption{
		grpc.WithTransportCredentials(tc),
		grpc.WithPerRPCCredentials(bearerTokenCreds{token: token}),
	}, nil
}

// DialGRPC opens a client connection using a credentials file path.
func DialGRPC(credPath string) (*grpc.ClientConn, error) {
	path, err := ResolveCredentialsPath(credPath)
	if err != nil {
		return nil, fmt.Errorf("resolve credentials path: %w", err)
	}
	creds, err := LoadCredentials(path)
	if err != nil {
		return nil, fmt.Errorf("load credentials: %w", err)
	}
	opts, err := GRPCDialOptions(creds)
	if err != nil {
		return nil, fmt.Errorf("gRPC dial options: %w", err)
	}
	conn, err := grpc.NewClient(creds.Endpoint, opts...)
	if err != nil {
		return nil, fmt.Errorf("dial gRPC %q: %w", creds.Endpoint, err)
	}
	return conn, nil
}
