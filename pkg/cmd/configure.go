package cmd

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"gopkg.in/yaml.v3"
)

type Credentials struct {
	Endpoint string `yaml:"endpoint"`
	APIToken string `yaml:"api_token"`
	CertPath string `yaml:"cert_path"`
	CertKey  string `yaml:"cert_key"`
}

func getCredentialsPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting home directory:", err)
		os.Exit(1)
	}
	return filepath.Join(home, ".permify", "credentials")
}

func NewConfigureCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "configure",
		Short: "Configure Permify CLI credentials",
		Run: func(cmd *cobra.Command, args []string) {
			reader := bufio.NewReader(os.Stdin)

			fmt.Print("Endpoint (e.g., localhost:3478): ")
			endpoint, _ := reader.ReadString('\n')

			fmt.Print("API Token: ")
			apiToken, _ := reader.ReadString('\n')

			fmt.Print("Cert Path (optional): ")
			certPath, _ := reader.ReadString('\n')

			fmt.Print("Cert Key (optional): ")
			certKey, _ := reader.ReadString('\n')

			creds := Credentials{
				Endpoint: strings.TrimSpace(endpoint),
				APIToken: strings.TrimSpace(apiToken),
				CertPath: strings.TrimSpace(certPath),
				CertKey:  strings.TrimSpace(certKey),
			}

			data, err := yaml.Marshal(&creds)
			if err != nil {
				fmt.Println("Error marshaling credentials:", err)
				os.Exit(1)
			}

			credsPath := getCredentialsPath()
			credsDir := filepath.Dir(credsPath)

			if err := os.MkdirAll(credsDir, 0700); err != nil {
				fmt.Println("Error creating credentials directory:", err)
				os.Exit(1)
			}

			if err := os.WriteFile(credsPath, data, 0600); err != nil {
				fmt.Println("Error writing credentials file:", err)
				os.Exit(1)
			}

			fmt.Println("Credentials saved successfully in ~/.permify/credentials")
		},
	}
}

type TokenAuth struct {
	Token string
}

func (t TokenAuth) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": "Bearer " + t.Token,
	}, nil
}

// FIX: Blindamos el envío del token (Security Risk)
func (TokenAuth) RequireTransportSecurity() bool {
	return true
}

func ConfiguredGRPCClient() ([]grpc.DialOption, string, error) {
	credsPath := getCredentialsPath()
	data, err := os.ReadFile(credsPath)
	if err != nil {
		return nil, "", fmt.Errorf("could not read credentials (run 'permify configure' first): %w", err)
	}

	var creds Credentials
	if err := yaml.Unmarshal(data, &creds); err != nil {
		return nil, "", fmt.Errorf("could not parse credentials: %w", err)
	}

	var opts []grpc.DialOption

	if creds.APIToken != "" {
		opts = append(opts, grpc.WithPerRPCCredentials(TokenAuth{Token: creds.APIToken}))
	}

	if creds.CertPath != "" && creds.CertKey != "" {
		cert, err := tls.LoadX509KeyPair(creds.CertPath, creds.CertKey)
		if err != nil {
			return nil, "", fmt.Errorf("could not load client key pair: %w", err)
		}
		// FIX: Forzamos la versión mínima de TLS
		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS12,
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	endpoint := creds.Endpoint
	if endpoint == "" {
		endpoint = "localhost:3478"
	}

	return opts, endpoint, nil
}
