package config

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/pkg/errors"

	"github.com/spf13/viper"
)

type (
	// Config is the main configuration structure containing various sections for different aspects of the application.
	Config struct {
		AccountID   string                       `mapstructure:"account_id"`
		Server      `mapstructure:"server"`      // Server configuration for both HTTP and gRPC
		Log         `mapstructure:"logger"`      // Logging configuration
		Profiler    `mapstructure:"profiler"`    // Profiler configuration
		Authn       `mapstructure:"authn"`       // Authentication configuration
		Tracer      `mapstructure:"tracer"`      // Tracing configuration
		Meter       `mapstructure:"meter"`       // Metrics configuration
		Service     `mapstructure:"service"`     // Service configuration
		Database    `mapstructure:"database"`    // Database configuration
		Distributed `mapstructure:"distributed"` // Distributed configuration
	}

	// Server contains the configurations for both HTTP and gRPC servers.
	Server struct {
		HTTP         `mapstructure:"http"` // HTTP server configuration
		GRPC         `mapstructure:"grpc"` // gRPC server configuration
		NameOverride string                `mapstructure:"name_override"`
		RateLimit    int64                 `mapstructure:"rate_limit"` // Rate limit configuration
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
		Issuer            string        `mapstructure:"issuer"`   // OIDC issuer URL
		Audience          string        `mapstructure:"audience"` // OIDC client ID
		RefreshInterval   time.Duration `mapstructure:"refresh_interval"`
		BackoffInterval   time.Duration `mapstructure:"backoff_interval"`
		BackoffFrequency  time.Duration `mapstructure:"backoff_frequency"`
		BackoffMaxRetries int           `mapstructure:"backoff_max_retries"`
		ValidMethods      []string      `mapstructure:"valid_methods"`
	}

	// Profiler contains configuration for the profiler.
	Profiler struct {
		Enabled bool   `mapstructure:"enabled"` // Whether the profiler is enabled
		Port    string `mapstructure:"port"`    // Port for the profiler
	}

	// Log contains configuration for logging.
	Log struct {
		Level    string   `mapstructure:"level"`    // Logging level
		Output   string   `mapstructure:"output"`   // Logging output format, e.g., text, json
		Enabled  bool     `mapstructure:"enabled"`  // Whether logging collection is enabled
		Exporter string   `mapstructure:"exporter"` // Exporter for log data
		Endpoint string   `mapstructure:"endpoint"` // Endpoint for the log exporter
		Insecure bool     `mapstructure:"insecure"` // Connect to the collector using the HTTP scheme, instead of HTTPS.
		Urlpath  string   `mapstructure:"urlpath"`  // Path for the log exporter, if not defined /v1/logs will be used
		Headers  []string `mapstructure:"headers"`
		Protocol string   `mapstructure:"protocol"` // Protocol for the log exporter, e.g., http, grpc
	}

	// Tracer contains configuration for distributed tracing.
	Tracer struct {
		Enabled  bool     `mapstructure:"enabled"`  // Whether tracing collection is enabled
		Exporter string   `mapstructure:"exporter"` // Exporter for tracing data
		Endpoint string   `mapstructure:"endpoint"` // Endpoint for the tracing exporter
		Insecure bool     `mapstructure:"insecure"` // Connect to the collector using the HTTP scheme, instead of HTTPS.
		Urlpath  string   `mapstructure:"urlpath"`  // Path for the tracing exporter, if not defined /v1/trace will be used
		Headers  []string `mapstructure:"headers"`
		Protocol string   `mapstructure:"protocol"` // Protocol for the tracing exporter, e.g., http, grpc
	}

	// Meter contains configuration for metrics collection and reporting.
	Meter struct {
		Enabled  bool     `mapstructure:"enabled"`  // Whether metrics collection is enabled
		Exporter string   `mapstructure:"exporter"` // Exporter for metrics data
		Endpoint string   `mapstructure:"endpoint"` // Endpoint for the metrics exporter
		Insecure bool     `mapstructure:"insecure"` // Connect to the collector using the HTTP scheme, instead of HTTPS.
		Urlpath  string   `mapstructure:"urlpath"`  // Path for the metrics exporter, if not defined /v1/metrics will be used
		Headers  []string `mapstructure:"headers"`
		Interval int      `mapstructure:"interval"`
		Protocol string   `mapstructure:"protocol"` // Protocol for the metrics exporter, e.g., http, grpc
	}

	// Service contains configuration for various service-level features.
	Service struct {
		CircuitBreaker bool       `mapstructure:"circuit_breaker"` // Whether to enable the circuit breaker pattern
		Watch          Watch      `mapstructure:"watch"`           // Watch service configuration
		Schema         Schema     `mapstructure:"schema"`          // Schema service configuration
		Permission     Permission `mapstructure:"permission"`      // Permission service configuration
		Data           Data       `mapstructure:"data"`            // Data service configuration
	}

	// Watch contains configuration for the watch service.
	Watch struct {
		Enabled bool `mapstructure:"enabled"`
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

	// Data is a placeholder struct for the data service configuration.
	Data struct{}

	// Cache contains configuration for caching.
	Cache struct {
		NumberOfCounters int64  `mapstructure:"number_of_counters"` // Number of counters for the cache
		MaxCost          string `mapstructure:"max_cost"`           // Maximum cost for the cache
	}

	// Database contains configuration for the database.
	Database struct {
		Engine string `mapstructure:"engine"` // Database engine type (e.g., "postgres" or "memory")
		URI    string `mapstructure:"uri"`    // Database connection URI
		Writer struct {
			URI string `mapstructure:"uri"`
		} `mapstructure:"writer"`
		Reader struct {
			URI string `mapstructure:"uri"`
		} `mapstructure:"reader"`
		AutoMigrate           bool              `mapstructure:"auto_migrate"`            // Whether to enable automatic migration
		MaxOpenConnections    int               `mapstructure:"max_open_connections"`    // Maximum number of open connections to the database
		MaxIdleConnections    int               `mapstructure:"max_idle_connections"`    // Maximum number of idle connections to the database
		MaxConnectionLifetime time.Duration     `mapstructure:"max_connection_lifetime"` // Maximum duration a connection can be reused
		MaxConnectionIdleTime time.Duration     `mapstructure:"max_connection_idle_time"`
		MaxDataPerWrite       int               `mapstructure:"max_data_per_write"`
		MaxRetries            int               `mapstructure:"max_retries"`
		WatchBufferSize       int               `mapstructure:"watch_buffer_size"`
		GarbageCollection     GarbageCollection `mapstructure:"garbage_collection"`
	}

	GarbageCollection struct {
		Enabled  bool          `mapstructure:"enabled"`
		Interval time.Duration `mapstructure:"interval"`
		Timeout  time.Duration `mapstructure:"timeout"`
		Window   time.Duration `mapstructure:"window"`
	}

	Distributed struct {
		Enabled           bool    `mapstructure:"enabled"`
		Address           string  `mapstructure:"address"`
		Port              string  `mapstructure:"port"`
		PartitionCount    int     `mapstructure:"partition_count"`
		ReplicationFactor int     `mapstructure:"replication_factor"`
		Load              float64 `mapstructure:"load"`
		PickerWidth       int     `mapstructure:"picker_width"`
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

// NewConfigWithFile initializes and returns a new Config object by reading and unmarshalling
// the configuration file from the given path. It falls back to the DefaultConfig if the
// file is not found. If there's an error during the process, it returns the error.
func NewConfigWithFile(dir string) (*Config, error) {
	// Start with the default configuration values
	cfg := DefaultConfig()

	viper.SetConfigFile(dir)

	err := isYAML(dir)
	if err != nil {
		return nil, err
	}

	// Read the config file
	err = viper.ReadInConfig()
	// If there's an error during reading the config file
	if err != nil {
		// Check if the error is because of the config file not being found
		if ok := errors.As(err, &viper.ConfigFileNotFoundError{}); !ok {
			// If it's not a "file not found" error, return the error with a message
			return nil, fmt.Errorf("failed to load server config: %w", err)
		}
		if ok := errors.As(err, &viper.ConfigMarshalError{}); !ok {
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
		AccountID: "",
		Server: Server{
			NameOverride: "",
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
			RateLimit: 10_000,
		},
		Profiler: Profiler{
			Enabled: false,
		},
		Log: Log{
			Level:    "info",
			Enabled:  false,
			Exporter: "otlp",
			Headers:  []string{},
			Protocol: "http",
		},
		Tracer: Tracer{
			Enabled:  false,
			Headers:  []string{},
			Protocol: "http",
		},
		Meter: Meter{
			Enabled:  false,
			Exporter: "otlp",
			Endpoint: "telemetry.permify.co",
			Headers:  []string{},
			Interval: 300,
			Protocol: "http",
		},
		Service: Service{
			CircuitBreaker: false,
			Watch: Watch{
				Enabled: false,
			},
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
			Data: Data{},
		},
		Authn: Authn{
			Enabled:   false,
			Preshared: Preshared{},
			Oidc: Oidc{
				RefreshInterval:   time.Minute * 15,
				ValidMethods:      []string{"RS256", "HS256"},
				BackoffMaxRetries: 5,
				BackoffInterval:   12 * time.Second,
			},
		},
		Database: Database{
			Engine:                "memory",
			AutoMigrate:           true,
			MaxOpenConnections:    20,
			MaxIdleConnections:    1,
			MaxConnectionLifetime: time.Second * 300,
			MaxConnectionIdleTime: time.Second * 60,
			MaxDataPerWrite:       1000,
			MaxRetries:            10,
			WatchBufferSize:       100,
			GarbageCollection: GarbageCollection{
				Enabled: false,
			},
		},
		Distributed: Distributed{
			Enabled: false,
			Port:    "5000",
		},
	}
}

func isYAML(file string) error {
	ext := filepath.Ext(file)
	if ext != ".yaml" {
		return errors.New("file is not yaml")
	}
	return nil
}
