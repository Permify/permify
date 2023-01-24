package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

const (
	Path = "config/config.yaml"
)

type (
	// Config -
	Config struct {
		Server   `mapstructure:"server"`
		Log      `mapstructure:"logger"`
		Profiler `mapstructure:"profiler"`
		Authn    `mapstructure:"authn"`
		Tracer   `mapstructure:"tracer"`
		Meter    `mapstructure:"meter"`
		Service  `mapstructure:"service"`
		Database `mapstructure:"database"`
	}

	Server struct {
		HTTP `mapstructure:"http"`
		GRPC `mapstructure:"grpc"`
	}

	// HTTP -.
	HTTP struct {
		Enabled            bool      `mapstructure:"enabled"`
		Port               string    `mapstructure:"port"`
		TLSConfig          TLSConfig `mapstructure:"tls"`
		CORSAllowedOrigins []string  `mapstructure:"cors_allowed_origins"`
		CORSAllowedHeaders []string  `mapstructure:"cors_allowed_headers"`
	}

	GRPC struct {
		Port      string    `mapstructure:"port"`
		TLSConfig TLSConfig `mapstructure:"tls"`
	}

	TLSConfig struct {
		Enabled  bool   `mapstructure:"enabled"`
		CertPath string `mapstructure:"cert_path"`
		KeyPath  string `mapstructure:"key_path"`
	}

	// Authn -.
	Authn struct {
		Enabled bool     `mapstructure:"enabled"`
		Keys    []string `mapstructure:"keys"`
	}

	// Profiler -.
	Profiler struct {
		Enabled bool   `mapstructure:"enabled"`
		Port    string `mapstructure:"port"`
	}

	// Log -.
	Log struct {
		Level string `mapstructure:"level"`
	}

	// Tracer -.
	Tracer struct {
		Enabled  bool   `mapstructure:"enabled"`
		Exporter string `mapstructure:"exporter"`
		Endpoint string `mapstructure:"endpoint"`
	}

	// Meter -.
	Meter struct {
		Enabled  bool   `mapstructure:"enabled"`
		Exporter string `mapstructure:"exporter"`
		Endpoint string `mapstructure:"endpoint"`
	}

	// Service -.
	Service struct {
		CircuitBreaker   bool `mapstructure:"circuit_breaker"`
		ConcurrencyLimit int  `mapstructure:"concurrency_limit"`
	}

	// Database -.
	Database struct {
		Engine                string        `mapstructure:"engine"`
		URI                   string        `mapstructure:"uri"`
		AutoMigrate           bool          `mapstructure:"auto_migrate"`
		MaxOpenConnections    int           `mapstructure:"max_open_connections"`
		MaxIdleConnections    int           `mapstructure:"max_idle_connections"`
		MaxConnectionLifetime time.Duration `mapstructure:"max_connection_lifetime"`
		MaxConnectionIdleTime time.Duration `mapstructure:"max_connection_idle_time"`
	}
)

// NewConfig - Creates new config
func NewConfig() (*Config, error) {
	cfg := DefaultConfig()

	if _, err := os.Stat("./" + Path); !os.IsNotExist(err) {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath("./config")

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
		Profiler: Profiler{
			Enabled: false,
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
