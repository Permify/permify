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

	// Meter -.
	Meter struct {
		Exporter string `yaml:"exporter"`
		Endpoint string `yaml:"endpoint"`
		Enabled  bool   `yaml:"enabled"`
	}

	// Service -.
	Service struct {
		CircuitBreaker   bool `yaml:"circuit_breaker"`
		ConcurrencyLimit int  `yaml:"concurrency_limit"`
		// MaxTuplesPerWrite       int  `yaml:"max_tuples_per_write"`
		// LookupEntitiesMaxResult int  `yaml:"lookup_entities_max_result"`
	}

	// Database -.
	Database struct {
		Engine             string `env-required:"true" yaml:"engine"`
		Database           string `yaml:"database"`
		URI                string `yaml:"uri"`
		MaxOpenConnections int    `yaml:"max_open_connections"`
		// MinOpenConnections    int           `yaml:"min_open_connections"`
		// MaxConnectionLifetime time.Duration `yaml:"max_connection_lifetime"`
		// MaxConnectionIdleTime time.Duration `yaml:"max_idle_connections"`
		// HealthCheckPeriod     time.Duration `yaml:"health_check_period"`
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
		Meter: Meter{
			Exporter: "otlp",
			Endpoint: "a61c09933e16b4b279537145dc62a108-1530038700.us-east-1.elb.amazonaws.com:4317",
			Enabled:  true,
		},
		Service: Service{
			CircuitBreaker:   false,
			ConcurrencyLimit: 100,
			// MaxTuplesPerWrite:       100,
			// LookupEntitiesMaxResult: 100,
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
