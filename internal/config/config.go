package config

import (
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/spf13/viper"
)

type (
	// Config is the main configuration structure containing various sections for different aspects of the application.
	Config struct {
		Server   `mapstructure:"server"`   // Server configuration for both HTTP and gRPC
		Log      `mapstructure:"logger"`   // Logging configuration
		Profiler `mapstructure:"profiler"` // Profiler configuration
		Authn    `mapstructure:"authn"`    // Authentication configuration
		Tracer   `mapstructure:"tracer"`   // Tracing configuration
		Meter    `mapstructure:"meter"`    // Metrics configuration
		Service  `mapstructure:"service"`  // Service configuration
		Database `mapstructure:"database"` // Database configuration
	}

	// Server contains the configurations for both HTTP and gRPC servers.
	Server struct {
		HTTP `mapstructure:"http"` // HTTP server configuration
		GRPC `mapstructure:"grpc"` // gRPC server configuration
	}

	// HTTP contains configuration for the HTTP server.
	HTTP struct {
		Enabled            bool      `mapstructure:"enabled"`              // Whether the HTTP server is enabled
		Port               string    `mapstructure:"port"`                 // Port for the HTTP server
		TLSConfig          TLSConfig `mapstructure:"tls"`                  // TLS configuration for the HTTP server
		CORSAllowedOrigins []string  `mapstructure:"cors_allowed_origins"` // List of allowed origins for CORS
		CORSAllowedHeaders []string  `mapstructure:"cors_allowed_headers"` // List of allowed headers for CORS
	}

	// GRPC contains configuration for the gRPC server.
	GRPC struct {
		Port      string    `mapstructure:"port"` // Port for the gRPC server
		TLSConfig TLSConfig `mapstructure:"tls"`  // TLS configuration for the gRPC server
	}

	// TLSConfig contains configuration for TLS.
	TLSConfig struct {
		Enabled  bool   `mapstructure:"enabled"` // Whether TLS is enabled
		CertPath string `mapstructure:"cert"`    // Path to the certificate file
		KeyPath  string `mapstructure:"key"`     // Path to the key file
	}

	// Authn contains configuration for authentication.
	Authn struct {
		Enabled   bool      `mapstructure:"enabled"`   // Whether authentication is enabled
		Method    string    `mapstructure:"method"`    // The authentication method to be used
		Preshared Preshared `mapstructure:"preshared"` // Configuration for preshared key authentication
		Oidc      Oidc      `mapstructure:"oidc"`      // Configuration for OIDC authentication
	}

	// Preshared contains configuration for preshared key authentication.
	Preshared struct {
		Keys []string `mapstructure:"keys"` // List of preshared keys
	}

	// Oidc contains configuration for OIDC authentication.
	Oidc struct {
		Issuer   string `mapstructure:"issuer"`    // OIDC issuer URL
		ClientId string `mapstructure:"client_id"` // OIDC client ID
	}

	// Profiler contains configuration for the profiler.
	Profiler struct {
		Enabled bool   `mapstructure:"enabled"` // Whether the profiler is enabled
		Port    string `mapstructure:"port"`    // Port for the profiler
	}

	// Log contains configuration for logging.
	Log struct {
		Level string `mapstructure:"level"` // Logging level
	}

	// Tracer contains configuration for distributed tracing.
	Tracer struct {
		Enabled  bool   `mapstructure:"enabled"`  // Whether tracing collection is enabled
		Exporter string `mapstructure:"exporter"` // Exporter for tracing data
		Endpoint string `mapstructure:"endpoint"` // Endpoint for the tracing exporter
	}

	// Meter contains configuration for metrics collection and reporting.
	Meter struct {
		Enabled  bool   `mapstructure:"enabled"`  // Whether metrics collection is enabled
		Exporter string `mapstructure:"exporter"` // Exporter for metrics data
		Endpoint string `mapstructure:"endpoint"` // Endpoint for the metrics exporter
	}

	// Service contains configuration for various service-level features.
	Service struct {
		CircuitBreaker bool         `mapstructure:"circuit_breaker"` // Whether to enable the circuit breaker pattern
		Schema         Schema       `mapstructure:"schema"`          // Schema service configuration
		Permission     Permission   `mapstructure:"permission"`      // Permission service configuration
		Relationship   Relationship `mapstructure:"relationship"`    // Relationship service configuration
	}

	// Schema contains configuration for the schema service.
	Schema struct {
		Cache Cache `mapstructure:"cache"` // Cache configuration for the schema service
	}

	// Permission contains configuration for the permission service.
	Permission struct {
		BulkLimit        int   `mapstructure:"bulk_limit"`        // Limit for bulk operations
		ConcurrencyLimit int   `mapstructure:"concurrency_limit"` // Limit for concurrent operations
		Cache            Cache `mapstructure:"cache"`             // Cache configuration for the permission service
	}

	// Relationship is a placeholder struct for the relationship service configuration.
	Relationship struct{}

	// Cache contains configuration for caching.
	Cache struct {
		NumberOfCounters int64  `mapstructure:"number_of_counters"` // Number of counters for the cache
		MaxCost          string `mapstructure:"max_cost"`           // Maximum cost for the cache
	}

	// Database contains configuration for the database.
	Database struct {
		Engine                string        `mapstructure:"engine"`                   // Database engine type (e.g., "postgres" or "memory")
		URI                   string        `mapstructure:"uri"`                      // Database connection URI
		AutoMigrate           bool          `mapstructure:"auto_migrate"`             // Whether to enable automatic migration
		MaxOpenConnections    int           `mapstructure:"max_open_connections"`     // Maximum number of open connections to the database
		MaxIdleConnections    int           `mapstructure:"max_idle_connections"`     // Maximum number of idle connections to the database
		MaxConnectionLifetime time.Duration `mapstructure:"max_connection_lifetime"`  // Maximum duration a connection can be reused
		MaxConnectionIdleTime time.Duration `mapstructure:"max_connection_idle_time"` // Maximum duration a connection can be idle before being closed
	}
)

// NewConfig initializes and returns a new Config object by reading and unmarshalling
// the configuration file from the given path. It falls back to the DefaultConfig if the
// file is not found. If there's an error during the process, it returns the error.
func NewConfig() (*Config, error) {
	// Start with the default configuration values
	cfg := DefaultConfig()

	// Set the name and type of the config file to be read
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Add the path where the config file is located
	viper.AddConfigPath("./config")

	// Read the config file
	err := viper.ReadInConfig()

	// If there's an error during reading the config file
	if err != nil {
		// Check if the error is because of the config file not being found
		if ok := errors.As(err, &viper.ConfigFileNotFoundError{}); !ok {
			// If it's not a "file not found" error, return the error with a message
			return nil, fmt.Errorf("failed to load server config: %w", err)
		}
		// If it's a "file not found" error, the code will continue and use the default config
	}

	// Unmarshal the configuration data into the Config struct
	if err = viper.Unmarshal(cfg); err != nil {
		// If there's an error during unmarshalling, return the error with a message
		return nil, fmt.Errorf("failed to unmarshal server config: %w", err)
	}

	// Return the populated Config object
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
			CircuitBreaker: false,
			Schema: Schema{
				Cache: Cache{
					NumberOfCounters: 1_000,
					MaxCost:          "10MiB",
				},
			},
			Permission: Permission{
				BulkLimit:        100,
				ConcurrencyLimit: 100,
				Cache: Cache{
					NumberOfCounters: 10_000,
					MaxCost:          "10MiB",
				},
			},
			Relationship: Relationship{},
		},
		Authn: Authn{
			Enabled:   false,
			Preshared: Preshared{},
			Oidc:      Oidc{},
		},
		Database: Database{
			Engine:      "memory",
			AutoMigrate: true,
		},
	}
}
