package config

import (
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

const (
	Path = "config/config.yaml"
)

type (
	// Config -
	Config struct {
		Server   `yaml:"server"`
		Log      `yaml:"logger"`
		Authn    `yaml:"authn"`
		Tracer   `yaml:"tracer"`
		Service  `yaml:"service"`
		Database `yaml:"database"`
	}

	Server struct {
		HTTP `yaml:"http"`
		GRPC `yaml:"grpc"`
	}

	// HTTP -.
	HTTP struct {
		Enabled            bool      `yaml:"enabled"`
		Port               string    `env-required:"true" yaml:"port"`
		TLSConfig          TLSConfig `yaml:"tls"`
		CORSAllowedOrigins []string  `yaml:"cors_allowed_origins"`
		CORSAllowedHeaders []string  `yaml:"cors_allowed_headers"`
	}

	GRPC struct {
		Port      string    `env-required:"true" yaml:"port"`
		TLSConfig TLSConfig `yaml:"tls"`
	}

	TLSConfig struct {
		Enabled  bool   `yaml:"enabled"`
		CertPath string `yaml:"cert_path"`
		KeyPath  string `yaml:"key_path"`
	}

	// Authn -.
	Authn struct {
		Enabled bool     `yaml:"enabled"`
		Keys    []string `yaml:"keys"`
	}

	// Log -.
	Log struct {
		Level string `env-required:"true" yaml:"level"`
	}

	// Tracer -.
	Tracer struct {
		Exporter string `yaml:"exporter"`
		Endpoint string `yaml:"endpoint"`
		Enabled  bool   `yaml:"enabled"`
	}

	// Service -.
	Service struct {
		CircuitBreaker bool `yaml:"circuit_breaker"`
	}

	// Database -.
	Database struct {
		Engine   string `env-required:"true" yaml:"engine"`
		PoolMax  int    `yaml:"pool_max"`
		Database string `yaml:"database"`
		URI      string `yaml:"uri"`
	}
)

// NewConfig - Creates new config
func NewConfig() (*Config, error) {
	cfg := &Config{}

	if _, err := os.Stat("./" + Path); !os.IsNotExist(err) {
		err = cleanenv.ReadConfig("./"+Path, cfg)
		if err != nil {
			return nil, err
		}
	} else {
		cfg = DefaultConfig()
	}

	return cfg, nil
}

// DefaultConfig - Creates default config.
func DefaultConfig() *Config {
	return &Config{
		Server: Server{
			HTTP: HTTP{
				Enabled: true,
				Port:    "3476",
				TLSConfig: TLSConfig{
					Enabled: false,
				},
				CORSAllowedOrigins: []string{"*"},
				CORSAllowedHeaders: []string{"*"},
			},
			GRPC: GRPC{
				Port: "3478",
				TLSConfig: TLSConfig{
					Enabled: false,
				},
			},
		},
		Log: Log{
			Level: "debug",
		},
		Tracer: Tracer{
			Enabled: false,
		},
		Service: Service{
			CircuitBreaker: false,
		},
		Authn: Authn{
			Enabled: false,
			Keys:    []string{},
		},
		Database: Database{
			Engine: "memory",
		},
	}
}
