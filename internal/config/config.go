package config

import (
	"os"
	"time"

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
		Enabled            bool      `yaml:"enabled" env-default:"true"`
		Port               string    `yaml:"port" env-default:"3476"`
		TLSConfig          TLSConfig `yaml:"tls"`
		CORSAllowedOrigins []string  `yaml:"cors_allowed_origins"`
		CORSAllowedHeaders []string  `yaml:"cors_allowed_headers"`
	}

	GRPC struct {
		Port      string    `yaml:"port" env-default:"3478"`
		TLSConfig TLSConfig `yaml:"tls"`
	}

	TLSConfig struct {
		Enabled  bool   `yaml:"enabled"`
		CertPath string `yaml:"cert_path"`
		KeyPath  string `yaml:"key_path"`
	}

	// Authn -.
	Authn struct {
		Enabled bool     `yaml:"enabled" env-default:"false"`
		Keys    []string `yaml:"keys"`
	}

	// Log -.
	Log struct {
		Level string `yaml:"level" env-default:"debug"`
	}

	// Tracer -.
	Tracer struct {
		Enabled  bool   `yaml:"enabled" env-default:"false"`
		Exporter string `yaml:"exporter"`
		Endpoint string `yaml:"endpoint"`
	}

	// Meter -.
	Meter struct {
		Enabled  bool   `yaml:"enabled" env-default:"true"`
		Exporter string `yaml:"exporter" env-default:"otlp"`
		Endpoint string `yaml:"endpoint" env-default:"telemetry.permify.co"`
	}

	// Service -.
	Service struct {
		CircuitBreaker   bool `yaml:"circuit_breaker" env-default:"false"`
		ConcurrencyLimit int  `yaml:"concurrency_limit" env-default:"100"`
	}

	// Database -.
	Database struct {
		Engine                string        `yaml:"engine" env-default:"memory"`
		URI                   string        `yaml:"uri"`
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
		err = cleanenv.ReadConfig("./"+Path, cfg)
		if err != nil {
			return nil, err
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
			Level: "debug",
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
			Engine: "memory",
		},
	}
}
