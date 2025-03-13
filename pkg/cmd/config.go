package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/gookit/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/Permify/permify/internal/config"
	"github.com/Permify/permify/pkg/cmd/flags"
)

func NewConfigCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "config",
		Short: "inspect permify configuration and environment variables",
		RunE:  conf(),
		Args:  cobra.NoArgs,
	}

	conf := config.DefaultConfig()
	f := command.Flags()
	f.StringP("config", "c", "", "config file (default is $HOME/.permify.yaml)")
	f.Bool("http-enabled", conf.Server.HTTP.Enabled, "switch option for HTTP server")
	f.String("account-id", conf.AccountID, "account id")
	f.Int64("server-rate-limit", conf.Server.RateLimit, "the maximum number of requests the server should handle per second")
	f.String("server-name-override", conf.Server.NameOverride, "server name override")
	f.String("grpc-port", conf.Server.GRPC.Port, "port that GRPC server run on")
	f.Bool("grpc-tls-enabled", conf.Server.GRPC.TLSConfig.Enabled, "switch option for GRPC tls server")
	f.String("grpc-tls-key-path", conf.Server.GRPC.TLSConfig.KeyPath, "GRPC tls key path")
	f.String("grpc-tls-cert-path", conf.Server.GRPC.TLSConfig.CertPath, "GRPC tls certificate path")
	f.String("http-port", conf.Server.HTTP.Port, "HTTP port address")
	f.Bool("http-tls-enabled", conf.Server.HTTP.TLSConfig.Enabled, "switch option for HTTP tls server")
	f.String("http-tls-key-path", conf.Server.HTTP.TLSConfig.KeyPath, "HTTP tls key path")
	f.String("http-tls-cert-path", conf.Server.HTTP.TLSConfig.CertPath, "HTTP tls certificate path")
	f.StringSlice("http-cors-allowed-origins", conf.Server.HTTP.CORSAllowedOrigins, "CORS allowed origins for http gateway")
	f.StringSlice("http-cors-allowed-headers", conf.Server.HTTP.CORSAllowedHeaders, "CORS allowed headers for http gateway")
	f.Bool("profiler-enabled", conf.Profiler.Enabled, "switch option for profiler")
	f.String("profiler-port", conf.Profiler.Port, "profiler port address")
	f.String("log-level", conf.Log.Level, "real time logs of authorization. Permify uses slog as a logger")
	f.String("log-output", conf.Log.Output, "logger output valid values json, text")
	f.Bool("log-enabled", conf.Log.Enabled, "logger exporter enabled")
	f.String("log-exporter", conf.Log.Exporter, "can be; otlp. (integrated metric tools)")
	f.String("log-endpoint", conf.Log.Endpoint, "export uri for logs")
	f.Bool("log-insecure", conf.Log.Insecure, "use https or http for logs")
	f.String("log-urlpath", conf.Log.Urlpath, "allow to set url path for otlp exporter")
	f.StringSlice("log-headers", conf.Log.Headers, "allows setting custom headers for the log exporter in key-value pairs")
	f.String("log-protocol", conf.Log.Protocol, "allows setting the communication protocol for the log exporter, with options http or grpc")
	f.Bool("authn-enabled", conf.Authn.Enabled, "enable server authentication")
	f.String("authn-method", conf.Authn.Method, "server authentication method")
	f.StringSlice("authn-preshared-keys", conf.Authn.Preshared.Keys, "preshared key/keys for server authentication")
	f.String("authn-oidc-issuer", conf.Authn.Oidc.Issuer, "issuer identifier of the OpenID Connect Provider")
	f.String("authn-oidc-audience", conf.Authn.Oidc.Audience, "intended audience of the OpenID Connect token")
	f.Duration("authn-oidc-refresh-interval", conf.Authn.Oidc.RefreshInterval, "refresh interval for the OpenID Connect configuration")
	f.Duration("authn-oidc-backoff-interval", conf.Authn.Oidc.BackoffInterval, "backoff interval for the OpenID Connect configuration")
	f.Duration("authn-oidc-backoff-frequency", conf.Authn.Oidc.BackoffFrequency, "backoff frequency for the OpenID Connect configuration")
	f.Int("authn-oidc-backoff-max-retries", conf.Authn.Oidc.BackoffMaxRetries, "defines the maximum number of retries for the OpenID Connect configuration")
	f.StringSlice("authn-oidc-valid-methods", conf.Authn.Oidc.ValidMethods, "list of valid JWT signing methods for OpenID Connect")
	f.Bool("tracer-enabled", conf.Tracer.Enabled, "switch option for tracing")
	f.String("tracer-exporter", conf.Tracer.Exporter, "can be; jaeger, signoz, zipkin or otlp. (integrated tracing tools)")
	f.String("tracer-endpoint", conf.Tracer.Endpoint, "export uri for tracing data")
	f.Bool("tracer-insecure", conf.Tracer.Insecure, "use https or http for tracer data, only used for otlp exporter or signoz")
	f.String("tracer-urlpath", conf.Tracer.Urlpath, "allow to set url path for otlp exporter")
	f.StringSlice("tracer-headers", conf.Tracer.Headers, "allows setting custom headers for the tracer exporter in key-value pairs")
	f.String("tracer-protocol", conf.Tracer.Protocol, "allows setting the communication protocol for the tracer exporter, with options http or grpc")
	f.Bool("meter-enabled", conf.Meter.Enabled, "switch option for metric")
	f.String("meter-exporter", conf.Meter.Exporter, "can be; otlp. (integrated metric tools)")
	f.String("meter-endpoint", conf.Meter.Endpoint, "export uri for metric data")
	f.Bool("meter-insecure", conf.Meter.Insecure, "use https or http for metric data")
	f.String("meter-urlpath", conf.Meter.Urlpath, "allow to set url path for otlp exporter")
	f.StringSlice("meter-headers", conf.Meter.Headers, "allows setting custom headers for the metric exporter in key-value pairs")
	f.Int("meter-interval", conf.Meter.Interval, "allows to set metrics to be pushed in certain time interval")
	f.String("meter-protocol", conf.Meter.Protocol, "allows setting the communication protocol for the meter exporter, with options http or grpc")
	f.Bool("service-circuit-breaker", conf.Service.CircuitBreaker, "switch option for service circuit breaker")
	f.Bool("service-watch-enabled", conf.Service.Watch.Enabled, "switch option for watch service")
	f.Int64("service-schema-cache-number-of-counters", conf.Service.Schema.Cache.NumberOfCounters, "schema service cache number of counters")
	f.String("service-schema-cache-max-cost", conf.Service.Schema.Cache.MaxCost, "schema service cache max cost")
	f.Int("service-permission-bulk-limit", conf.Service.Permission.BulkLimit, "bulk operations limit")
	f.Int("service-permission-concurrency-limit", conf.Service.Permission.ConcurrencyLimit, "concurrency limit")
	f.Int64("service-permission-cache-number-of-counters", conf.Service.Permission.Cache.NumberOfCounters, "permission service cache number of counters")
	f.String("service-permission-cache-max-cost", conf.Service.Permission.Cache.MaxCost, "permission service cache max cost")
	f.String("database-engine", conf.Database.Engine, "data source. e.g. postgres, memory")
	f.String("database-uri", conf.Database.URI, "uri of your data source to store relation tuples and schema")
	f.String("database-writer-uri", conf.Database.Writer.URI, "writer uri of your data source to store relation tuples and schema")
	f.String("database-reader-uri", conf.Database.Reader.URI, "reader uri of your data source to store relation tuples and schema")
	f.Bool("database-auto-migrate", conf.Database.AutoMigrate, "auto migrate database tables")
	f.Int("database-max-open-connections", conf.Database.MaxOpenConnections, "maximum number of parallel connections that can be made to the database at any time")
	f.Int("database-max-idle-connections", conf.Database.MaxIdleConnections, "maximum number of idle connections that can be made to the database at any time")
	f.Duration("database-max-connection-lifetime", conf.Database.MaxConnectionLifetime, "maximum amount of time a connection may be reused")
	f.Duration("database-max-connection-idle-time", conf.Database.MaxConnectionIdleTime, "maximum amount of time a connection may be idle")
	f.Int("database-max-data-per-write", conf.Database.MaxDataPerWrite, "sets the maximum amount of data per write operation to the database")
	f.Int("database-max-retries", conf.Database.MaxRetries, "defines the maximum number of retries for database operations in case of failure")
	f.Int("database-watch-buffer-size", conf.Database.WatchBufferSize, "specifies the buffer size for database watch operations, impacting how many changes can be queued")
	f.Bool("database-garbage-collection-enabled", conf.Database.GarbageCollection.Enabled, "use database garbage collection for expired relationships and attributes")
	f.Duration("database-garbage-collection-interval", conf.Database.GarbageCollection.Interval, "interval for database garbage collection")
	f.Duration("database-garbage-collection-timeout", conf.Database.GarbageCollection.Timeout, "timeout for database garbage collection")
	f.Duration("database-garbage-collection-window", conf.Database.GarbageCollection.Window, "window for database garbage collection")
	f.Bool("distributed-enabled", conf.Distributed.Enabled, "enable distributed")
	f.String("distributed-address", conf.Distributed.Address, "distributed address")
	f.String("distributed-port", conf.Distributed.Port, "distributed port")
	f.Int("distributed-partition-count", conf.Distributed.PartitionCount, "number of partitions for distributed hashing")
	f.Int("distributed-replication-factor", conf.Distributed.ReplicationFactor, "number of replicas for distributed hashing")
	f.Float64("distributed-load", conf.Distributed.Load, "load factor for distributed hashing")
	f.Int("distributed-picker-width", conf.Distributed.PickerWidth, "picker width for distributed hashing")

	command.PreRun = func(cmd *cobra.Command, args []string) {
		flags.RegisterServeFlags(f)
	}

	return command
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
			[]string{"server.name_override", fmt.Sprintf("%v", cfg.Server.NameOverride), getKeyOrigin(cmd, "server-name-override", "PERMIFY_NAME_OVERRIDE")},
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
			[]string{"logger.output", cfg.Log.Output, getKeyOrigin(cmd, "log-output", "PERMIFY_LOG_OUTPUT")},
			[]string{"logger.enabled", fmt.Sprintf("%v", cfg.Log.Enabled), getKeyOrigin(cmd, "log-enabled", "PERMIFY_LOG_ENABLED")},
			[]string{"logger.exporter", cfg.Log.Exporter, getKeyOrigin(cmd, "log-exporter", "PERMIFY_LOG_EXPORTER")},
			[]string{"logger.endpoint", HideSecret(cfg.Log.Exporter), getKeyOrigin(cmd, "log-endpoint", "PERMIFY_LOG_ENDPOINT")},
			[]string{"logger.insecure", fmt.Sprintf("%v", cfg.Log.Insecure), getKeyOrigin(cmd, "log-insecure", "PERMIFY_LOG_INSECURE")},
			[]string{"logger.urlpath", cfg.Log.Urlpath, getKeyOrigin(cmd, "log-urlpath", "PERMIFY_LOG_URL_PATH")},
			[]string{"logger.headers", fmt.Sprintf("%v", cfg.Log.Headers), getKeyOrigin(cmd, "log-headers", "PERMIFY_LOG_HEADERS")},
			[]string{"logger.protocol", cfg.Log.Protocol, getKeyOrigin(cmd, "log-protocol", "PERMIFY_LOG_PROTOCOL")},
			// AUTHN
			[]string{"authn.enabled", fmt.Sprintf("%v", cfg.Authn.Enabled), getKeyOrigin(cmd, "authn-enabled", "PERMIFY_AUTHN_ENABLED")},
			[]string{"authn.method", cfg.Authn.Method, getKeyOrigin(cmd, "authn-method", "PERMIFY_AUTHN_METHOD")},
			[]string{"authn.preshared.keys", fmt.Sprintf("%v", HideSecrets(cfg.Authn.Preshared.Keys...)), getKeyOrigin(cmd, "authn-preshared-keys", "PERMIFY_AUTHN_PRESHARED_KEYS")},
			[]string{"authn.oidc.issuer", HideSecret(cfg.Authn.Oidc.Issuer), getKeyOrigin(cmd, "authn-oidc-issuer", "PERMIFY_AUTHN_OIDC_ISSUER")},
			[]string{"authn.oidc.audience", HideSecret(cfg.Authn.Oidc.Audience), getKeyOrigin(cmd, "authn-oidc-audience", "PERMIFY_AUTHN_OIDC_AUDIENCE")},
			[]string{"authn.oidc.refresh_interval", fmt.Sprintf("%v", cfg.Authn.Oidc.RefreshInterval), getKeyOrigin(cmd, "authn-oidc-refresh-interval", "PERMIFY_AUTHN_OIDC_REFRESH_INTERVAL")},
			[]string{"authn.oidc.backoff_interval", fmt.Sprintf("%v", cfg.Authn.Oidc.BackoffInterval), getKeyOrigin(cmd, "authn-oidc-backoff-interval", "PERMIFY_AUTHN_OIDC_BACKOFF_INTERVAL")},
			[]string{"authn.oidc.backoff_max_retries", fmt.Sprintf("%v", cfg.Authn.Oidc.BackoffMaxRetries), getKeyOrigin(cmd, "authn-oidc-backoff-max-retries", "PERMIFY_AUTHN_OIDC_BACKOFF_RETRIES")},
			[]string{"authn.oidc.backoff_frequency", fmt.Sprintf("%v", cfg.Authn.Oidc.BackoffFrequency), getKeyOrigin(cmd, "authn-oidc-backoff-frequency", "PERMIFY_AUTHN_OIDC_BACKOFF_FREQUENCY")},
			[]string{"authn.oidc.valid_methods", fmt.Sprintf("%v", cfg.Authn.Oidc.ValidMethods), getKeyOrigin(cmd, "authn-oidc-valid-methods", "PERMIFY_AUTHN_OIDC_VALID_METHODS")},
			// TRACER
			[]string{"tracer.enabled", fmt.Sprintf("%v", cfg.Tracer.Enabled), getKeyOrigin(cmd, "tracer-enabled", "PERMIFY_TRACER_ENABLED")},
			[]string{"tracer.exporter", cfg.Tracer.Exporter, getKeyOrigin(cmd, "tracer-exporter", "PERMIFY_TRACER_EXPORTER")},
			[]string{"tracer.endpoint", HideSecret(cfg.Tracer.Exporter), getKeyOrigin(cmd, "tracer-endpoint", "PERMIFY_TRACER_ENDPOINT")},
			[]string{"tracer.insecure", fmt.Sprintf("%v", cfg.Tracer.Insecure), getKeyOrigin(cmd, "tracer-insecure", "PERMIFY_TRACER_INSECURE")},
			[]string{"tracer.urlpath", cfg.Tracer.Urlpath, getKeyOrigin(cmd, "tracer-urlpath", "PERMIFY_TRACER_URL_PATH")},
			[]string{"tracer.headers", fmt.Sprintf("%v", cfg.Tracer.Headers), getKeyOrigin(cmd, "tracer-headers", "PERMIFY_TRACER_HEADERS")},
			[]string{"tracer.protocol", cfg.Tracer.Protocol, getKeyOrigin(cmd, "tracer-protocol", "PERMIFY_TRACER_PROTOCOL")},
			// METER
			[]string{"meter.enabled", fmt.Sprintf("%v", cfg.Meter.Enabled), getKeyOrigin(cmd, "meter-enabled", "PERMIFY_METER_ENABLED")},
			[]string{"meter.exporter", cfg.Meter.Exporter, getKeyOrigin(cmd, "meter-exporter", "PERMIFY_METER_EXPORTER")},
			[]string{"meter.endpoint", HideSecret(cfg.Meter.Exporter), getKeyOrigin(cmd, "meter-endpoint", "PERMIFY_METER_ENDPOINT")},
			[]string{"meter.insecure", fmt.Sprintf("%v", cfg.Meter.Insecure), getKeyOrigin(cmd, "meter-insecure", "PERMIFY_METER_INSECURE")},
			[]string{"meter.urlpath", cfg.Meter.Urlpath, getKeyOrigin(cmd, "meter-urlpath", "PERMIFY_METER_URL_PATH")},
			[]string{"meter.headers", fmt.Sprintf("%v", cfg.Meter.Headers), getKeyOrigin(cmd, "meter-headers", "PERMIFY_METER_HEADERS")},
			[]string{"meter.protocol", cfg.Meter.Protocol, getKeyOrigin(cmd, "meter-protocol", "PERMIFY_METER_PROTOCOL")},
			[]string{"meter.interval", fmt.Sprintf("%v", cfg.Meter.Interval), getKeyOrigin(cmd, "meter-interval", "PERMIFY_METER_INTERVAL")},
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
			[]string{"database.writer.uri", HideSecret(cfg.Database.Writer.URI), getKeyOrigin(cmd, "database-writer-uri", "PERMIFY_DATABASE_WRITER_URI")},
			[]string{"database.reader.uri", HideSecret(cfg.Database.Reader.URI), getKeyOrigin(cmd, "database-reader-uri", "PERMIFY_DATABASE_READER_URI")},
			[]string{"database.auto_migrate", fmt.Sprintf("%v", cfg.Database.AutoMigrate), getKeyOrigin(cmd, "database-auto-migrate", "PERMIFY_DATABASE_AUTO_MIGRATE")},
			[]string{"database.max_open_connections", fmt.Sprintf("%v", cfg.Database.MaxOpenConnections), getKeyOrigin(cmd, "database-max-open-connections", "PERMIFY_DATABASE_MAX_OPEN_CONNECTIONS")},
			[]string{"database.max_idle_connections", fmt.Sprintf("%v", cfg.Database.MaxIdleConnections), getKeyOrigin(cmd, "database-max-idle-connections", "PERMIFY_DATABASE_MAX_IDLE_CONNECTIONS")},
			[]string{"database.max_connection_lifetime", fmt.Sprintf("%v", cfg.Database.MaxConnectionLifetime), getKeyOrigin(cmd, "database-max-connection-lifetime", "PERMIFY_DATABASE_MAX_CONNECTION_LIFETIME")},
			[]string{"database.max_connection_idle_time", fmt.Sprintf("%v", cfg.Database.MaxConnectionIdleTime), getKeyOrigin(cmd, "database-max-connection-idle-time", "PERMIFY_DATABASE_MAX_CONNECTION_IDLE_TIME")},
			[]string{"database.max_data_per_write", fmt.Sprintf("%v", cfg.Database.MaxDataPerWrite), getKeyOrigin(cmd, "database-max-data-per-write", "PERMIFY_DATABASE_MAX_DATA_PER_WRITE")},
			[]string{"database.max_retries", fmt.Sprintf("%v", cfg.Database.MaxRetries), getKeyOrigin(cmd, "database-max-retries", "PERMIFY_DATABASE_MAX_RETRIES")},
			[]string{"database.watch_buffer_size", fmt.Sprintf("%v", cfg.Database.WatchBufferSize), getKeyOrigin(cmd, "database-watch-buffer-size", "PERMIFY_DATABASE_WATCH_BUFFER_SIZE")},
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

// HideSecrets obscures each string in a given list.
func HideSecrets(secrets ...string) (rv []string) {
	// Convert each secret to its hidden version and collect them.
	for _, secret := range secrets {
		rv = append(rv, HideSecret(secret)) // Hide each secret.
	}
	return
}
