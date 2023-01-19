package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
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
		Meter    `yaml:"meter"`
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
		Port               string    `yaml:"port"`
		TLSConfig          TLSConfig `yaml:"tls"`
		CORSAllowedOrigins []string  `yaml:"cors_allowed_origins"`
		CORSAllowedHeaders []string  `yaml:"cors_allowed_headers"`
	}

	GRPC struct {
		Port      string    `yaml:"port"`
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
		Level string `yaml:"level"`
	}

	// Tracer -.
	Tracer struct {
		Enabled  bool   `yaml:"enabled"`
		Exporter string `yaml:"exporter"`
		Endpoint string `yaml:"endpoint"`
	}

	// Meter -.
	Meter struct {
		Enabled  bool   `yaml:"enabled"`
		Exporter string `yaml:"exporter"`
		Endpoint string `yaml:"endpoint"`
	}

	// Service -.
	Service struct {
		CircuitBreaker   bool `yaml:"circuit_breaker"`
		ConcurrencyLimit int  `yaml:"concurrency_limit"`
	}

	// Database -.
	Database struct {
		Engine                string        `yaml:"engine"`
		URI                   string        `yaml:"uri"`
		AutoMigrate           bool          `yaml:"auto_migrate"`
		MaxOpenConnections    int           `yaml:"max_open_connections"`
		MaxIdleConnections    int           `yaml:"max_idle_connections"`
		MaxConnectionLifetime time.Duration `yaml:"max_connection_lifetime"`
		MaxConnectionIdleTime time.Duration `yaml:"max_connection_idle_time"`
	}
)

// NewConfig - Creates new config
func NewConfig() (*Config, error) {
	cfg := DefaultConfig()

	if _, err := os.Stat("./" + Path); !os.IsNotExist(err) {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath("./config")

		viper.SetEnvPrefix("PERMIFY")
		viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		viper.AutomaticEnv()

		err = viper.ReadInConfig()
		if err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return nil, fmt.Errorf("failed to load server config: %w", err)
			}
		}

		if err = viper.Unmarshal(cfg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal server config: %w", err)
		}
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
			Level: "info",
		},
		Tracer: Tracer{
			Enabled: false,
		},
		Meter: Meter{
			Enabled:  true,
			Exporter: "otlp",
			Endpoint: "telemetry.permify.co",
		},
		Service: Service{
			CircuitBreaker:   false,
			ConcurrencyLimit: 100,
		},
		Authn: Authn{
			Enabled: false,
			Keys:    []string{},
		},
		Database: Database{
			Engine:      "memory",
			AutoMigrate: true,
		},
	}
}
