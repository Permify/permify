package flags

import (
	"github.com/Permify/permify/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// RegisterServeFlags - Define and registers permify CLI flags
func RegisterServeFlags(cmd *cobra.Command) {
	conf := config.DefaultConfig()

	var err error

	flags := cmd.Flags()

	// GRPC Server
	flags.String("grpc-port", conf.Server.GRPC.Port, "port that GRPC server run on")
	if err = viper.BindPFlag("server.grpc.port", flags.Lookup("grpc-port")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("server.grpc.port", "PERMIFY_GRPC_PORT"); err != nil {
		panic(err)
	}

	flags.Bool("grpc-tls-enabled", conf.Server.GRPC.TLSConfig.Enabled, "switch option for GRPC tls server")
	if err = viper.BindPFlag("server.grpc.tls.enabled", flags.Lookup("grpc-tls-enabled")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("server.grpc.tls.enabled", "PERMIFY_GRPC_TLS_ENABLED"); err != nil {
		panic(err)
	}

	flags.String("grpc-tls-key-path", conf.Server.GRPC.TLSConfig.KeyPath, "GRPC tls key path")
	if err = viper.BindPFlag("server.grpc.tls.key", flags.Lookup("grpc-tls-key-path")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("server.grpc.tls.key", "PERMIFY_GRPC_TLS_KEY_PATH"); err != nil {
		panic(err)
	}

	flags.String("grpc-tls-cert-path", conf.Server.GRPC.TLSConfig.CertPath, "GRPC tls certificate path")
	if err = viper.BindPFlag("server.grpc.tls.cert", flags.Lookup("grpc-tls-cert-path")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("server.grpc.tls.cert", "PERMIFY_GRPC_TLS_CERT_PATH"); err != nil {
		panic(err)
	}

	// HTTP Server
	flags.Bool("http-enabled", conf.Server.HTTP.Enabled, "switch option for HTTP server")
	if err = viper.BindPFlag("server.http.enabled", flags.Lookup("http-enabled")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("server.http.enabled", "PERMIFY_HTTP_ENABLED"); err != nil {
		panic(err)
	}

	flags.String("http-port", conf.Server.HTTP.Port, "HTTP port address")
	if err = viper.BindPFlag("server.http.port", flags.Lookup("http-port")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("server.http.port", "PERMIFY_HTTP_PORT"); err != nil {
		panic(err)
	}

	flags.String("http-tls-key-path", conf.Server.HTTP.TLSConfig.KeyPath, "HTTP tls key path")
	if err = viper.BindPFlag("server.http.tls.key", flags.Lookup("http-tls-key-path")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("server.http.tls.key", "PERMIFY_HTTP_TLS_KEY_PATH"); err != nil {
		panic(err)
	}

	flags.String("http-tls-cert-path", conf.Server.HTTP.TLSConfig.CertPath, "HTTP tls certificate path")
	if err = viper.BindPFlag("server.http.tls.cert", flags.Lookup("http-tls-cert-path")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("server.http.tls.cert", "PERMIFY_HTTP_TLS_CERT_PATH"); err != nil {
		panic(err)
	}

	flags.StringSlice("http-cors-allowed-origins", conf.Server.HTTP.CORSAllowedOrigins, "CORS allowed origins for http gateway")
	if err = viper.BindPFlag("server.http.cors_allowed_origins", flags.Lookup("http-cors-allowed-origins")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("server.http.cors_allowed_origins", "PERMIFY_HTTP_CORS_ALLOWED_ORIGINS"); err != nil {
		panic(err)
	}

	flags.StringSlice("http-cors-allowed-headers", conf.Server.HTTP.CORSAllowedHeaders, "CORS allowed headers for http gateway")
	if err = viper.BindPFlag("server.http.cors_allowed_headers", flags.Lookup("http-cors-allowed-headers")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("server.http.cors_allowed_headers", "PERMIFY_HTTP_CORS_ALLOWED_HEADERS"); err != nil {
		panic(err)
	}

	// PROFILER
	flags.Bool("profiler-enabled", conf.Profiler.Enabled, "switch option for profiler")
	if err = viper.BindPFlag("profiler.enabled", flags.Lookup("profiler-enabled")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("profiler.enabled", "PERMIFY_PROFILER_ENABLED"); err != nil {
		panic(err)
	}

	flags.String("profiler-port", conf.Profiler.Port, "profiler port address")
	if err = viper.BindPFlag("profiler.port", flags.Lookup("profiler-port")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("profiler.port", "PERMIFY_PROFILER_PORT"); err != nil {
		panic(err)
	}

	// LOG
	flags.String("log-level", conf.Log.Level, "real time logs of authorization. Permify uses zerolog as a logger")
	if err = viper.BindPFlag("logger.level", flags.Lookup("log-level")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("logger.level", "PERMIFY_LOG_LEVEL"); err != nil {
		panic(err)
	}

	// AUTHN
	flags.Bool("authn-enabled", conf.Authn.Enabled, "enable server authentication")
	if err = viper.BindPFlag("authn.enabled", flags.Lookup("authn-enabled")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("authn.enabled", "PERMIFY_AUTHN_ENABLED"); err != nil {
		panic(err)
	}

	flags.StringSlice("authn-keys", conf.Authn.Keys, "preshared key/keys for server authentication")
	if err = viper.BindPFlag("authn.keys", flags.Lookup("authn-keys")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("authn.keys", "PERMIFY_AUTHN_KEYS"); err != nil {
		panic(err)
	}

	flags.String("authn-oidc-issuer", conf.Authn.Oidc.Issuer, "issuer identifier of the OpenID Connect Provider")
	if err = viper.BindPFlag("authn.oidc.issuer", flags.Lookup("authn-oidc-issuer")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("authn.oidc.issuer", "PERMIFY_AUTHN_OIDC_ISSUER"); err != nil {
		panic(err)
	}

	flags.String("authn-oidc-client-id", conf.Authn.Oidc.ClientId, "client ID which requested the token from OIDC issuer")
	if err = viper.BindPFlag("authn.oidc.client_id", flags.Lookup("authn-oidc-client-id")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("authn.oidc.client_id", "PERMIFY_AUTHN_OIDC_CLIENT_ID"); err != nil {
		panic(err)
	}

	// TRACER
	flags.Bool("tracer-enabled", conf.Tracer.Enabled, "switch option for tracing")
	if err = viper.BindPFlag("tracer.enabled", flags.Lookup("tracer-enabled")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("tracer.enabled", "PERMIFY_TRACER_ENABLED"); err != nil {
		panic(err)
	}

	flags.String("tracer-exporter", conf.Tracer.Exporter, "export uri for tracing data")
	if err = viper.BindPFlag("tracer.exporter", flags.Lookup("tracer-exporter")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("tracer.exporter", "PERMIFY_TRACER_EXPORTER"); err != nil {
		panic(err)
	}

	flags.String("tracer-endpoint", conf.Tracer.Endpoint, "can be; jaeger, signoz, zipkin or otlp. (integrated tracing tools)")
	if err = viper.BindPFlag("tracer.endpoint", flags.Lookup("tracer-endpoint")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("tracer.endpoint", "PERMIFY_TRACER_ENDPOINT"); err != nil {
		panic(err)
	}

	// METER
	flags.Bool("meter-enabled", conf.Meter.Enabled, "switch option for metric")
	if err = viper.BindPFlag("meter.enabled", flags.Lookup("meter-enabled")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("meter.enabled", "PERMIFY_METRIC_ENABLED"); err != nil {
		panic(err)
	}

	flags.String("meter-exporter", conf.Meter.Exporter, "export uri for metric data")
	if err = viper.BindPFlag("meter.exporter", flags.Lookup("meter-exporter")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("meter.exporter", "PERMIFY_METER_EXPORTER"); err != nil {
		panic(err)
	}

	flags.String("meter-endpoint", conf.Meter.Endpoint, "can be; otlp. (integrated metric tools)")
	if err = viper.BindPFlag("meter.endpoint", flags.Lookup("meter-endpoint")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("meter.endpoint", "PERMIFY_METER_ENDPOINT"); err != nil {
		panic(err)
	}

	// SERVICE
	flags.Bool("service-circuit-breaker", conf.Service.CircuitBreaker, "switch option for service circuit breaker")
	if err = viper.BindPFlag("service.circuit_breaker", flags.Lookup("service-circuit-breaker")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("service.circuit_breaker", "PERMIFY_SERVICE_CIRCUIT_BREAKER"); err != nil {
		panic(err)
	}

	flags.Int64("service-schema-cache-number-of-counters", conf.Service.Schema.Cache.NumberOfCounters, "schema service cache number of counters")
	if err = viper.BindPFlag("service.schema.cache.number_of_counters", flags.Lookup("service-schema-cache-number-of-counters")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("service.schema.cache.number_of_counters", "PERMIFY_SERVICE_SCHEMA_CACHE_NUMBER_OF_COUNTERS"); err != nil {
		panic(err)
	}

	flags.String("service-schema-cache-max-cost", conf.Service.Schema.Cache.MaxCost, "schema service cache max cost")
	if err = viper.BindPFlag("service.schema.cache.max_cost", flags.Lookup("service-schema-cache-max-cost")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("service.schema.cache.max_cost", "PERMIFY_SERVICE_SCHEMA_CACHE_MAX_COST"); err != nil {
		panic(err)
	}

	flags.Int("service-permission-concurrency-limit", conf.Service.Permission.ConcurrencyLimit, "concurrency limit")
	if err = viper.BindPFlag("service.permission.concurrency_limit", flags.Lookup("service-permission-concurrency-limit")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("service.permission.concurrency_limit", "PERMIFY_SERVICE_PERMISSION_CONCURRENCY_LIMIT"); err != nil {
		panic(err)
	}

	flags.Int64("service-permission-cache-number-of-counters", conf.Service.Permission.Cache.NumberOfCounters, "permission service cache number of counters")
	if err = viper.BindPFlag("service.permission.cache.number_of_counters", flags.Lookup("service-permission-cache-number-of-counters")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("service.permission.cache.number_of_counters", "PERMIFY_SERVICE_PERMISSION_CACHE_NUMBER_OF_COUNTERS"); err != nil {
		panic(err)
	}

	flags.String("service-permission-cache-max-cost", conf.Service.Permission.Cache.MaxCost, "permission service cache max cost")
	if err = viper.BindPFlag("service.permission.cache.max_cost", flags.Lookup("service-permission-cache-max-cost")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("service.permission.cache.max_cost", "PERMIFY_SERVICE_PERMISSION_CACHE_MAX_COST"); err != nil {
		panic(err)
	}

	// DATABASE
	flags.String("database-engine", conf.Database.Engine, "data source. e.g. postgres, memory")
	if err = viper.BindPFlag("database.engine", flags.Lookup("database-engine")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("database.engine", "PERMIFY_DATABASE_ENGINE"); err != nil {
		panic(err)
	}

	flags.String("database-uri", conf.Database.URI, "uri of your data source to store relation tuples and schema")
	if err = viper.BindPFlag("database.uri", flags.Lookup("database-uri")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("database.uri", "PERMIFY_DATABASE_URI"); err != nil {
		panic(err)
	}

	flags.Bool("database-auto-migrate", conf.Database.AutoMigrate, "auto migrate database tables")
	if err = viper.BindPFlag("database.auto_migrate", flags.Lookup("database-auto-migrate")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("database.auto_migrate", "PERMIFY_DATABASE_AUTO_MIGRATE"); err != nil {
		panic(err)
	}

	flags.Int("database-max-open-connections", conf.Database.MaxOpenConnections, "maximum number of parallel connections that can be made to the database at any time")
	if err = viper.BindPFlag("database.max_open_connections", flags.Lookup("database-max-open-connections")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("database.max_open_connections", "PERMIFY_DATABASE_MAX_OPEN_CONNECTIONS"); err != nil {
		panic(err)
	}

	flags.Int("database-max-idle-connections", conf.Database.MaxIdleConnections, "maximum number of idle connections that can be made to the database at any time")
	if err = viper.BindPFlag("database.max_idle_connections", flags.Lookup("database-max-idle-connections")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("database.max_idle_connections", "PERMIFY_DATABASE_MAX_IDLE_CONNECTIONS"); err != nil {
		panic(err)
	}

	flags.Duration("database-max-connection-lifetime", conf.Database.MaxConnectionLifetime, "maximum amount of time a connection may be reused")
	if err = viper.BindPFlag("database.max_connection_lifetime", flags.Lookup("database-max-connection-lifetime")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("database.max_connection_lifetime", "PERMIFY_DATABASE_MAX_CONNECTION_LIFETIME"); err != nil {
		panic(err)
	}

	flags.Duration("database-max-connection-idle-time", conf.Database.MaxConnectionIdleTime, "maximum amount of time a connection may be idle")
	if err = viper.BindPFlag("database.max_connection_idle_time", flags.Lookup("database-max-connection-idle-time")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("database.max_connection_idle_time", "PERMIFY_DATABASE_MAX_CONNECTION_IDLE_TIME"); err != nil {
		panic(err)
	}
}
