package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/Permify/permify/internal/config"
	"github.com/Permify/permify/pkg/cmd/flags"

	"github.com/gookit/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Inspect permify configuration and environment variables",
		RunE:  conf(),
		Args:  cobra.NoArgs,
	}

	flags.RegisterServeFlags(cmd)

	return cmd
}

func conf() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		var cfg *config.Config
		var err error
		cfgFile := viper.GetString("config.file")
		if cfgFile != "" {
			cfg, err = config.NewConfigWithFile(cfgFile)
			if err != nil {
				return fmt.Errorf("failed to create new config: %w", err)
			}

			if err = viper.Unmarshal(cfg); err != nil {
				return fmt.Errorf("failed to unmarshal config: %w", err)
			}
		} else {
			// Load configuration
			cfg, err = config.NewConfig()
			if err != nil {
				return fmt.Errorf("failed to create new config: %w", err)
			}

			if err = viper.Unmarshal(cfg); err != nil {
				return fmt.Errorf("failed to unmarshal config: %w", err)
			}
		}

		var data [][]string

		data = append(data,
			[]string{"account_id", cfg.AccountID, getKeyOrigin(cmd, "account-id", "PERMIFY_ACCOUNT_ID")},
			// SERVER
			[]string{"server.rate_limit", fmt.Sprintf("%v", cfg.Server.RateLimit), getKeyOrigin(cmd, "server-rate-limit", "PERMIFY_RATE_LIMIT")},
			[]string{"server.grpc.port", cfg.Server.GRPC.Port, getKeyOrigin(cmd, "grpc-port", "PERMIFY_GRPC_PORT")},
			[]string{"server.grpc.tls.enabled", fmt.Sprintf("%v", cfg.Server.GRPC.TLSConfig.Enabled), getKeyOrigin(cmd, "grpc-tls-enabled", "PERMIFY_GRPC_TLS_ENABLED")},
			[]string{"server.grpc.tls.cert", cfg.Server.GRPC.TLSConfig.CertPath, getKeyOrigin(cmd, "grpc-tls-cert-path", "PERMIFY_GRPC_TLS_CERT_PATH")},
			[]string{"server.http.enabled", fmt.Sprintf("%v", cfg.Server.HTTP.Enabled), getKeyOrigin(cmd, "http-enabled", "PERMIFY_HTTP_ENABLED")},
			[]string{"server.http.tls.enabled", fmt.Sprintf("%v", cfg.Server.HTTP.TLSConfig.Enabled), getKeyOrigin(cmd, "http-tls-enabled", "PERMIFY_HTTP_TLS_ENABLED")},
			[]string{"server.http.tls.key", HideSecret(cfg.Server.HTTP.TLSConfig.KeyPath), getKeyOrigin(cmd, "http-tls-key-path", "PERMIFY_HTTP_TLS_KEY_PATH")},
			[]string{"server.http.tls.cert", HideSecret(cfg.Server.HTTP.TLSConfig.CertPath), getKeyOrigin(cmd, "http-tls-cert-path", "PERMIFY_HTTP_TLS_CERT_PATH")},
			[]string{"server.http.cors_allowed_origins", fmt.Sprintf("%v", cfg.Server.HTTP.CORSAllowedOrigins), getKeyOrigin(cmd, "http-cors-allowed-origins", "PERMIFY_HTTP_CORS_ALLOWED_ORIGINS")},
			[]string{"server.http.cors_allowed_headers", fmt.Sprintf("%v", cfg.Server.HTTP.CORSAllowedHeaders), getKeyOrigin(cmd, "http-cors-allowed-headers", "PERMIFY_HTTP_CORS_ALLOWED_HEADERS")},
			// PROFILER
			[]string{"profiler.enabled", fmt.Sprintf("%v", cfg.Profiler.Enabled), getKeyOrigin(cmd, "profiler-enabled", "PERMIFY_PROFILER_ENABLED")},
			[]string{"profiler.port", cfg.Profiler.Port, getKeyOrigin(cmd, "profiler-port", "PERMIFY_PROFILER_PORT")},
			// LOG
			[]string{"logger.level", cfg.Log.Level, getKeyOrigin(cmd, "log-level", "PERMIFY_LOG_LEVEL")},
			[]string{"logger.output", cfg.Log.Level, getKeyOrigin(cmd, "log-output", "PERMIFY_LOG_OUTPUT")},
			// AUTHN
			[]string{"authn.enabled", fmt.Sprintf("%v", cfg.Authn.Enabled), getKeyOrigin(cmd, "authn-enabled", "PERMIFY_AUTHN_ENABLED")},
			[]string{"authn.method", cfg.Authn.Method, getKeyOrigin(cmd, "authn-method", "PERMIFY_AUTHN_METHOD")},
			[]string{"authn.preshared.keys", fmt.Sprintf("%v", cfg.Authn.Preshared.Keys), getKeyOrigin(cmd, "authn-preshared-keys", "PERMIFY_AUTHN_PRESHARED_KEYS")},
			[]string{"authn.oidc.issuer", HideSecret(cfg.Authn.Oidc.Issuer), getKeyOrigin(cmd, "authn-oidc-issuer", "PERMIFY_AUTHN_OIDC_ISSUER")},
			[]string{"authn.oidc.audience", HideSecret(cfg.Authn.Oidc.Audience), getKeyOrigin(cmd, "authn-oidc-audience", "PERMIFY_AUTHN_OIDC_AUDIENCE")},
			// TRACER
			[]string{"tracer.enabled", fmt.Sprintf("%v", cfg.Tracer.Enabled), getKeyOrigin(cmd, "tracer-enabled", "PERMIFY_TRACER_ENABLED")},
			[]string{"tracer.exporter", cfg.Tracer.Exporter, getKeyOrigin(cmd, "tracer-exporter", "PERMIFY_TRACER_EXPORTER")},
			[]string{"tracer.endpoint", HideSecret(cfg.Tracer.Exporter), getKeyOrigin(cmd, "tracer-endpoint", "PERMIFY_TRACER_ENDPOINT")},
			[]string{"tracer.insecure", fmt.Sprintf("%v", cfg.Tracer.Insecure), getKeyOrigin(cmd, "tracer-insecure", "PERMIFY_TRACER_INSECURE")},
			[]string{"tracer.urlpath", cfg.Tracer.URLPath, getKeyOrigin(cmd, "tracer-urlpath", "PERMIFY_TRACER_URL_PATH")},
			// METER
			[]string{"meter.enabled", fmt.Sprintf("%v", cfg.Meter.Enabled), getKeyOrigin(cmd, "meter-enabled", "PERMIFY_METER_ENABLED")},
			[]string{"meter.exporter", cfg.Meter.Exporter, getKeyOrigin(cmd, "meter-exporter", "PERMIFY_METER_EXPORTER")},
			[]string{"meter.endpoint", HideSecret(cfg.Meter.Exporter), getKeyOrigin(cmd, "meter-endpoint", "PERMIFY_METER_ENDPOINT")},
			[]string{"meter.insecure", fmt.Sprintf("%v", cfg.Meter.Insecure), getKeyOrigin(cmd, "meter-insecure", "PERMIFY_METER_INSECURE")},
			[]string{"meter.urlpath", cfg.Meter.URLPath, getKeyOrigin(cmd, "meter-urlpath", "PERMIFY_METER_URL_PATH")},
			// SERVICE
			[]string{"service.circuit_breaker", fmt.Sprintf("%v", cfg.Service.CircuitBreaker), getKeyOrigin(cmd, "service-circuit-breaker", "PERMIFY_SERVICE_CIRCUIT_BREAKER")},
			[]string{"service.schema.cache.number_of_counters", fmt.Sprintf("%v", cfg.Service.Schema.Cache.NumberOfCounters), getKeyOrigin(cmd, "service-schema-cache-number-of-counters", "PERMIFY_SERVICE_WATCH_ENABLED")},
			[]string{"service.schema.cache.max_cost", cfg.Service.Schema.Cache.MaxCost, getKeyOrigin(cmd, "service-schema-cache-max-cost", "PERMIFY_SERVICE_SCHEMA_CACHE_MAX_COST")},
			[]string{"service.permission.bulk_limit", fmt.Sprintf("%v", cfg.Service.Permission.BulkLimit), getKeyOrigin(cmd, "service-permission-bulk-limit", "PERMIFY_SERVICE_PERMISSION_BULK_LIMIT")},
			[]string{"service.permission.concurrency_limit", fmt.Sprintf("%v", cfg.Service.Permission.ConcurrencyLimit), getKeyOrigin(cmd, "service-permission-concurrency-limit", "PERMIFY_SERVICE_PERMISSION_CONCURRENCY_LIMIT")},
			[]string{"service.permission.cache.number_of_counters", fmt.Sprintf("%v", cfg.Service.Permission.Cache.NumberOfCounters), getKeyOrigin(cmd, "service-permission-cache-number-of-counters", "PERMIFY_SERVICE_PERMISSION_CACHE_NUMBER_OF_COUNTERS")},
			[]string{"service.permission.cache.max_cost", fmt.Sprintf("%v", cfg.Service.Permission.Cache.MaxCost), getKeyOrigin(cmd, "service-permission-cache-max-cost", "PERMIFY_SERVICE_PERMISSION_CACHE_MAX_COST")},
			// DATABASE
			[]string{"database.engine", cfg.Database.Engine, getKeyOrigin(cmd, "database-engine", "PERMIFY_DATABASE_ENGINE")},
			[]string{"database.uri", HideSecret(cfg.Database.URI), getKeyOrigin(cmd, "database-uri", "PERMIFY_DATABASE_URI")},
			[]string{"database.auto_migrate", fmt.Sprintf("%v", cfg.Database.AutoMigrate), getKeyOrigin(cmd, "database-auto-migrate", "PERMIFY_DATABASE_AUTO_MIGRATE")},
			[]string{"database.max_open_connections", fmt.Sprintf("%v", cfg.Database.MaxOpenConnections), getKeyOrigin(cmd, "database-max-open-connections", "PERMIFY_DATABASE_MAX_OPEN_CONNECTIONS")},
			[]string{"database.max_idle_connections", fmt.Sprintf("%v", cfg.Database.MaxIdleConnections), getKeyOrigin(cmd, "database-max-idle-connections", "PERMIFY_DATABASE_MAX_IDLE_CONNECTIONS")},
			[]string{"database.max_connection_lifetime", fmt.Sprintf("%v", cfg.Database.MaxConnectionLifetime), getKeyOrigin(cmd, "database-max-connection-lifetime", "PERMIFY_DATABASE_MAX_CONNECTION_LIFETIME")},
			[]string{"database.max_connection_idle_time", fmt.Sprintf("%v", cfg.Database.MaxConnectionIdleTime), getKeyOrigin(cmd, "database-max-connection-idle-time", "PERMIFY_DATABASE_MAX_CONNECTION_IDLE_TIME")},
			[]string{"database.garbage_collection.enabled", fmt.Sprintf("%v", cfg.Database.GarbageCollection.Enabled), getKeyOrigin(cmd, "database-garbage-collection-enabled", "PERMIFY_DATABASE_GARBAGE_COLLECTION_ENABLED")},
			[]string{"database.garbage_collection.interval", fmt.Sprintf("%v", cfg.Database.GarbageCollection.Interval), getKeyOrigin(cmd, "database-garbage-collection-interval", "PERMIFY_DATABASE_GARBAGE_COLLECTION_INTERVAL")},
			[]string{"database.garbage_collection.timeout", fmt.Sprintf("%v", cfg.Database.GarbageCollection.Timeout), getKeyOrigin(cmd, "database-garbage-collection-timeout", "PERMIFY_DATABASE_GARBAGE_COLLECTION_TIMEOUT")},
			[]string{"database.garbage_collection.window", fmt.Sprintf("%v", cfg.Database.GarbageCollection.Window), getKeyOrigin(cmd, "database-garbage-collection-window", "PERMIFY_DATABASE_GARBAGE_COLLECTION_WINDOW")},
			// DISTRIBUTED
			[]string{"distributed.enabled", fmt.Sprintf("%v", cfg.Distributed.Enabled), getKeyOrigin(cmd, "distributed-enabled", "PERMIFY_DISTRIBUTED_ENABLED")},
			[]string{"distributed.address", cfg.Distributed.Address, getKeyOrigin(cmd, "distributed-address", "PERMIFY_DISTRIBUTED_ADDRESS")},
			[]string{"distributed.port", cfg.Distributed.Port, getKeyOrigin(cmd, "distributed-port", "PERMIFY_DISTRIBUTED_PORT")},
		)

		renderConfigTable(data)
		return nil
	}
}

// getKeyOrigin determines the source of a configuration value.
// It checks whether a value was set via a command-line flag, an environment variable, or defaults to file.
func getKeyOrigin(cmd *cobra.Command, flagKey, envKey string) string {
	// Check if the command-line flag (specified by flagKey) was explicitly set by the user.
	if cmd.Flags().Changed(flagKey) {
		// If the flag was set, return "FLAG" with light green color.
		return color.FgLightGreen.Render("FLAG")
	}

	// Check if the environment variable (specified by envKey) exists.
	_, exists := os.LookupEnv(envKey)
	if exists {
		// If the environment variable exists, return "ENV" with light blue color.
		return color.FgLightBlue.Render("ENV")
	}

	// If neither the command-line flag nor the environment variable was set,
	// assume the value came from a configuration file.
	return color.FgYellow.Render("FILE")
}

// renderConfigTable displays configuration data in a formatted table on the console.
// It takes a 2D slice of strings where each inner slice represents a row in the table.
func renderConfigTable(data [][]string) {
	// Create a new table writer object, writing to standard output.
	table := tablewriter.NewWriter(os.Stdout)

	// Set the headers of the table. Each header cell is a column title.
	table.SetHeader([]string{"Key", "Value", "Source"})

	// Align the columns of the table: left-aligned for keys, centered for values and sources.
	table.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT, tablewriter.ALIGN_CENTER, tablewriter.ALIGN_CENTER})

	// Set the center separator character for the table, which appears between columns.
	table.SetCenterSeparator("|")

	// Loop through the data and add each row to the table.
	for _, v := range data {
		table.Append(v)
	}

	// Render the table to standard output, displaying it to the user.
	table.Render()
}

// HideSecret replaces all but the first and last characters of a string with asterisks
func HideSecret(secret string) string {
	if len(secret) <= 2 {
		// If the secret is too short, just return asterisks
		return strings.Repeat("*", len(secret))
	}
	// Keep first and last character visible; replace the rest with asterisks
	return string(secret[0]) + strings.Repeat("*", len(secret)-2) + string(secret[len(secret)-1])
}
