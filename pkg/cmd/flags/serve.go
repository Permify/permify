package flags

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// RegisterServeFlags - Define and registers permify CLI flags
func RegisterServeFlags(flags *pflag.FlagSet) {
	var err error

	// Config File
	if err = viper.BindPFlag("config.file", flags.Lookup("config")); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("account_id", flags.Lookup("account-id")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("account_id", "PERMIFY_ACCOUNT_ID"); err != nil {
		panic(err)
	}

	// Server
	if err = viper.BindPFlag("server.rate_limit", flags.Lookup("server-rate-limit")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("server.rate_limit", "PERMIFY_RATE_LIMIT"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("server.name_override", flags.Lookup("server-name-override")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("server.name_override", "PERMIFY_NAME_OVERRIDE"); err != nil {
		panic(err)
	}

	// GRPC Server
	if err = viper.BindPFlag("server.grpc.port", flags.Lookup("grpc-port")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("server.grpc.port", "PERMIFY_GRPC_PORT"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("server.grpc.tls.enabled", flags.Lookup("grpc-tls-enabled")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("server.grpc.tls.enabled", "PERMIFY_GRPC_TLS_ENABLED"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("server.grpc.tls.key", flags.Lookup("grpc-tls-key-path")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("server.grpc.tls.key", "PERMIFY_GRPC_TLS_KEY_PATH"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("server.grpc.tls.cert", flags.Lookup("grpc-tls-cert-path")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("server.grpc.tls.cert", "PERMIFY_GRPC_TLS_CERT_PATH"); err != nil {
		panic(err)
	}

	// HTTP Server
	if err = viper.BindPFlag("server.http.enabled", flags.Lookup("http-enabled")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("server.http.enabled", "PERMIFY_HTTP_ENABLED"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("server.http.port", flags.Lookup("http-port")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("server.http.port", "PERMIFY_HTTP_PORT"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("server.http.tls.enabled", flags.Lookup("http-tls-enabled")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("server.http.tls.enabled", "PERMIFY_HTTP_TLS_ENABLED"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("server.http.tls.key", flags.Lookup("http-tls-key-path")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("server.http.tls.key", "PERMIFY_HTTP_TLS_KEY_PATH"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("server.http.tls.cert", flags.Lookup("http-tls-cert-path")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("server.http.tls.cert", "PERMIFY_HTTP_TLS_CERT_PATH"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("server.http.cors_allowed_origins", flags.Lookup("http-cors-allowed-origins")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("server.http.cors_allowed_origins", "PERMIFY_HTTP_CORS_ALLOWED_ORIGINS"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("server.http.cors_allowed_headers", flags.Lookup("http-cors-allowed-headers")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("server.http.cors_allowed_headers", "PERMIFY_HTTP_CORS_ALLOWED_HEADERS"); err != nil {
		panic(err)
	}

	// PROFILER
	if err = viper.BindPFlag("profiler.enabled", flags.Lookup("profiler-enabled")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("profiler.enabled", "PERMIFY_PROFILER_ENABLED"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("profiler.port", flags.Lookup("profiler-port")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("profiler.port", "PERMIFY_PROFILER_PORT"); err != nil {
		panic(err)
	}

	// LOG
	if err = viper.BindPFlag("logger.level", flags.Lookup("log-level")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("logger.level", "PERMIFY_LOG_LEVEL"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("logger.output", flags.Lookup("log-output")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("logger.output", "PERMIFY_LOG_OUTPUT"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("logger.enabled", flags.Lookup("log-enabled")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("logger.enabled", "PERMIFY_LOG_ENABLED"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("logger.exporter", flags.Lookup("log-exporter")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("logger.exporter", "PERMIFY_LOG_EXPORTER"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("logger.endpoint", flags.Lookup("log-endpoint")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("logger.endpoint", "PERMIFY_LOG_ENDPOINT"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("logger.insecure", flags.Lookup("log-insecure")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("logger.insecure", "PERMIFY_LOG_INSECURE"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("logger.urlpath", flags.Lookup("log-urlpath")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("logger.urlpath", "PERMIFY_LOG_URL_PATH"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("logger.headers", flags.Lookup("log-headers")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("logger.headers", "PERMIFY_LOG_HEADERS"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("logger.protocol", flags.Lookup("log-protocol")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("logger.protocol", "PERMIFY_LOG_PROTOCOL"); err != nil {
		panic(err)
	}

	// AUTHN
	if err = viper.BindPFlag("authn.enabled", flags.Lookup("authn-enabled")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("authn.enabled", "PERMIFY_AUTHN_ENABLED"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("authn.method", flags.Lookup("authn-method")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("authn.method", "PERMIFY_AUTHN_METHOD"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("authn.preshared.keys", flags.Lookup("authn-preshared-keys")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("authn.preshared.keys", "PERMIFY_AUTHN_PRESHARED_KEYS"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("authn.oidc.issuer", flags.Lookup("authn-oidc-issuer")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("authn.oidc.issuer", "PERMIFY_AUTHN_OIDC_ISSUER"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("authn.oidc.audience", flags.Lookup("authn-oidc-audience")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("authn.oidc.audience", "PERMIFY_AUTHN_OIDC_AUDIENCE"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("authn.oidc.refresh_interval", flags.Lookup("authn-oidc-refresh-interval")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("authn.oidc.refresh_interval", "PERMIFY_AUTHN_OIDC_REFRESH_INTERVAL"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("authn.oidc.backoff_interval", flags.Lookup("authn-oidc-backoff-interval")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("authn.oidc.backoff_interval", "PERMIFY_AUTHN_OIDC_BACKOFF_INTERVAL"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("authn.oidc.backoff_max_retries", flags.Lookup("authn-oidc-backoff-max-retries")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("authn.oidc.backoff_max_retries", "PERMIFY_AUTHN_OIDC_BACKOFF_RETRIES"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("authn.oidc.backoff_frequency", flags.Lookup("authn-oidc-backoff-frequency")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("authn.oidc.backoff_frequency", "PERMIFY_AUTHN_OIDC_BACKOFF_FREQUENCY"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("authn.oidc.valid_methods", flags.Lookup("authn-oidc-valid-methods")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("authn.oidc.valid_methods", "PERMIFY_AUTHN_OIDC_VALID_METHODS"); err != nil {
		panic(err)
	}

	// TRACER
	if err = viper.BindPFlag("tracer.enabled", flags.Lookup("tracer-enabled")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("tracer.enabled", "PERMIFY_TRACER_ENABLED"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("tracer.exporter", flags.Lookup("tracer-exporter")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("tracer.exporter", "PERMIFY_TRACER_EXPORTER"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("tracer.endpoint", flags.Lookup("tracer-endpoint")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("tracer.endpoint", "PERMIFY_TRACER_ENDPOINT"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("tracer.insecure", flags.Lookup("tracer-insecure")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("tracer.insecure", "PERMIFY_TRACER_INSECURE"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("tracer.urlpath", flags.Lookup("tracer-urlpath")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("tracer.urlpath", "PERMIFY_TRACER_URL_PATH"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("tracer.headers", flags.Lookup("tracer-headers")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("tracer.headers", "PERMIFY_TRACER_HEADERS"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("tracer.protocol", flags.Lookup("tracer-protocol")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("tracer.protocol", "PERMIFY_TRACER_PROTOCOL"); err != nil {
		panic(err)
	}

	// METER
	if err = viper.BindPFlag("meter.enabled", flags.Lookup("meter-enabled")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("meter.enabled", "PERMIFY_METER_ENABLED"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("meter.exporter", flags.Lookup("meter-exporter")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("meter.exporter", "PERMIFY_METER_EXPORTER"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("meter.endpoint", flags.Lookup("meter-endpoint")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("meter.endpoint", "PERMIFY_METER_ENDPOINT"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("meter.insecure", flags.Lookup("meter-insecure")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("meter.insecure", "PERMIFY_METER_INSECURE"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("meter.urlpath", flags.Lookup("meter-urlpath")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("meter.urlpath", "PERMIFY_METER_URL_PATH"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("meter.headers", flags.Lookup("meter-headers")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("meter.headers", "PERMIFY_METER_HEADERS"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("meter.interval", flags.Lookup("meter-interval")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("meter.interval", "PERMIFY_METER_INTERVAL"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("meter.protocol", flags.Lookup("meter-protocol")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("meter.protocol", "PERMIFY_METER_PROTOCOL"); err != nil {
		panic(err)
	}

	// SERVICE
	if err = viper.BindPFlag("service.circuit_breaker", flags.Lookup("service-circuit-breaker")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("service.circuit_breaker", "PERMIFY_SERVICE_CIRCUIT_BREAKER"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("service.watch.enabled", flags.Lookup("service-watch-enabled")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("service.watch.enabled", "PERMIFY_SERVICE_WATCH_ENABLED"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("service.schema.cache.number_of_counters", flags.Lookup("service-schema-cache-number-of-counters")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("service.schema.cache.number_of_counters", "PERMIFY_SERVICE_SCHEMA_CACHE_NUMBER_OF_COUNTERS"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("service.schema.cache.max_cost", flags.Lookup("service-schema-cache-max-cost")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("service.schema.cache.max_cost", "PERMIFY_SERVICE_SCHEMA_CACHE_MAX_COST"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("service.permission.bulk_limit", flags.Lookup("service-permission-bulk-limit")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("service.permission.bulk_limit", "PERMIFY_SERVICE_PERMISSION_BULK_LIMIT"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("service.permission.concurrency_limit", flags.Lookup("service-permission-concurrency-limit")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("service.permission.concurrency_limit", "PERMIFY_SERVICE_PERMISSION_CONCURRENCY_LIMIT"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("service.permission.cache.number_of_counters", flags.Lookup("service-permission-cache-number-of-counters")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("service.permission.cache.number_of_counters", "PERMIFY_SERVICE_PERMISSION_CACHE_NUMBER_OF_COUNTERS"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("service.permission.cache.max_cost", flags.Lookup("service-permission-cache-max-cost")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("service.permission.cache.max_cost", "PERMIFY_SERVICE_PERMISSION_CACHE_MAX_COST"); err != nil {
		panic(err)
	}

	// DATABASE
	if err = viper.BindPFlag("database.engine", flags.Lookup("database-engine")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("database.engine", "PERMIFY_DATABASE_ENGINE"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("database.uri", flags.Lookup("database-uri")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("database.uri", "PERMIFY_DATABASE_URI"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("database.writer.uri", flags.Lookup("database-writer-uri")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("database.writer.uri", "PERMIFY_DATABASE_WRITER_URI"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("database.reader.uri", flags.Lookup("database-reader-uri")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("database.reader.uri", "PERMIFY_DATABASE_READER_URI"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("database.auto_migrate", flags.Lookup("database-auto-migrate")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("database.auto_migrate", "PERMIFY_DATABASE_AUTO_MIGRATE"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("database.max_open_connections", flags.Lookup("database-max-open-connections")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("database.max_open_connections", "PERMIFY_DATABASE_MAX_OPEN_CONNECTIONS"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("database.max_idle_connections", flags.Lookup("database-max-idle-connections")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("database.max_idle_connections", "PERMIFY_DATABASE_MAX_IDLE_CONNECTIONS"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("database.max_connection_lifetime", flags.Lookup("database-max-connection-lifetime")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("database.max_connection_lifetime", "PERMIFY_DATABASE_MAX_CONNECTION_LIFETIME"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("database.max_connection_idle_time", flags.Lookup("database-max-connection-idle-time")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("database.max_connection_idle_time", "PERMIFY_DATABASE_MAX_CONNECTION_IDLE_TIME"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("database.max_data_per_write", flags.Lookup("database-max-data-per-write")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("database.max_data_per_write", "PERMIFY_DATABASE_MAX_DATA_PER_WRITE"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("database.max_retries", flags.Lookup("database-max-retries")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("database.max_retries", "PERMIFY_DATABASE_MAX_RETRIES"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("database.watch_buffer_size", flags.Lookup("database-watch-buffer-size")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("database.watch_buffer_size", "PERMIFY_DATABASE_WATCH_BUFFER_SIZE"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("database.garbage_collection.enabled", flags.Lookup("database-garbage-collection-enabled")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("database.garbage_collection.enabled", "PERMIFY_DATABASE_GARBAGE_COLLECTION_ENABLED"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("database.garbage_collection.interval", flags.Lookup("database-garbage-collection-interval")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("database.garbage_collection.interval", "PERMIFY_DATABASE_GARBAGE_COLLECTION_INTERVAL"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("database.garbage_collection.timeout", flags.Lookup("database-garbage-collection-timeout")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("database.garbage_collection.timeout", "PERMIFY_DATABASE_GARBAGE_COLLECTION_TIMEOUT"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("database.garbage_collection.window", flags.Lookup("database-garbage-collection-window")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("database.garbage_collection.window", "PERMIFY_DATABASE_GARBAGE_COLLECTION_WINDOW"); err != nil {
		panic(err)
	}

	// DISTRIBUTED
	if err = viper.BindPFlag("distributed.enabled", flags.Lookup("distributed-enabled")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("distributed.enabled", "PERMIFY_DISTRIBUTED_ENABLED"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("distributed.address", flags.Lookup("distributed-address")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("distributed.address", "PERMIFY_DISTRIBUTED_ADDRESS"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("distributed.port", flags.Lookup("distributed-port")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("distributed.port", "PERMIFY_DISTRIBUTED_PORT"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("distributed.partition_count", flags.Lookup("distributed-partition-count")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("distributed.partition_count", "PERMIFY_DISTRIBUTED_PARTITION_COUNT"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("distributed.replication_factor", flags.Lookup("distributed-replication-factor")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("distributed.replication_factor", "PERMIFY_DISTRIBUTED_REPLICATION_FACTOR"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("distributed.load", flags.Lookup("distributed-load")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("distributed.load", "PERMIFY_DISTRIBUTED_LOAD"); err != nil {
		panic(err)
	}

	if err = viper.BindPFlag("distributed.picker_width", flags.Lookup("distributed-picker-width")); err != nil {
		panic(err)
	}
	if err = viper.BindEnv("distributed.picker_width", "PERMIFY_DISTRIBUTED_PICKER_WIDTH"); err != nil {
		panic(err)
	}
}
