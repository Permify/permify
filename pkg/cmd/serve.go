package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	slogmulti "github.com/samber/slog-multi"
	"github.com/sony/gobreaker"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel/sdk/metric"

	"github.com/Permify/permify/internal/engines/balancer"
	"github.com/Permify/permify/internal/engines/cache"
	"github.com/Permify/permify/internal/invoke"
	cacheDecorator "github.com/Permify/permify/internal/storage/decorators/cache"
	cbDecorator "github.com/Permify/permify/internal/storage/decorators/circuitBreaker"
	sfDecorator "github.com/Permify/permify/internal/storage/decorators/singleflight"
	"github.com/Permify/permify/internal/storage/postgres/gc"
	"github.com/Permify/permify/pkg/cmd/flags"
	PQDatabase "github.com/Permify/permify/pkg/database/postgres"

	"go.opentelemetry.io/otel/sdk/trace"
	"golang.org/x/sync/errgroup"

	"github.com/Permify/permify/internal"
	"github.com/Permify/permify/internal/config"
	"github.com/Permify/permify/internal/engines"
	"github.com/Permify/permify/internal/factories"
	"github.com/Permify/permify/internal/servers"
	"github.com/Permify/permify/internal/storage"
	pkgcache "github.com/Permify/permify/pkg/cache"
	"github.com/Permify/permify/pkg/cache/ristretto"
	"github.com/Permify/permify/pkg/telemetry"
	"github.com/Permify/permify/pkg/telemetry/meterexporters"
	"github.com/Permify/permify/pkg/telemetry/tracerexporters"
)

// NewServeCommand returns a new Cobra command that can be used to run the "permify serve" command.
// The command takes no arguments and runs the serve() function to start the Permify service.
// The command has a short description of what it does.
func NewServeCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "serve",
		Short: "serve the Permify server",
		RunE:  serve(),
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
	f.String("log-level", conf.Log.Level, "set log verbosity ('info', 'debug', 'error', 'warning')")
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

	// SilenceUsage is set to true to suppress usage when an error occurs
	command.SilenceUsage = true

	command.PreRun = func(cmd *cobra.Command, args []string) {
		flags.RegisterServeFlags(f)
	}

	return command
}

// serve is the main function for the "permify serve" command. It starts the Permify service by configuring and starting the necessary components.
// It initializes the configuration, logger, database, tracing and metering components, and creates instances of the necessary engines, services, and decorators.
// It then creates a ServiceContainer and runs it with the given configuration.
// The function uses errgroup to manage the goroutines and gracefully shuts down the service upon receiving a termination signal.
// It returns an error if there is an issue with any of the components or if any goroutine fails.
func serve() func(cmd *cobra.Command, args []string) error {
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

		// Print banner and initialize logger
		internal.PrintBanner()

		// Set up context and signal handling
		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		internal.Identifier = cfg.AccountID

		var logger *slog.Logger
		var handler slog.Handler

		switch cfg.Log.Output {
		case "json":
			handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				Level: getLogLevel(cfg.Log.Level),
			})
		case "text":
			handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
				Level: getLogLevel(cfg.Log.Level),
			})
		default:
			handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
				Level: getLogLevel(cfg.Log.Level),
			})
		}

		if cfg.Log.Enabled {
			headers := map[string]string{}
			for _, header := range cfg.Log.Headers {
				h := strings.Split(header, ":")
				if len(h) != 2 {
					return errors.New("invalid header format; expected 'key:value'")
				}
				headers[h[0]] = h[1]
			}

			customHandler, err := telemetry.HandlerFactory(
				cfg.Log.Exporter,
				cfg.Log.Endpoint,
				cfg.Log.Insecure,
				cfg.Log.Urlpath,
				headers,
				cfg.Log.Protocol,
				getLogLevel(cfg.Log.Level),
			)

			if err != nil {
				slog.Error("invalid logger exporter", slog.Any("error", err))
				logger = slog.New(handler)
			} else {
				logger = slog.New(
					slogmulti.Fanout(
						customHandler,
						handler,
					),
				)
			}
		} else {
			logger = slog.New(handler)
		}

		slog.SetDefault(logger)

		if internal.Identifier == "" {
			message := "Account ID is not set. Please fill in the Account ID for better support. Get your Account ID from https://permify.co/account"
			slog.Error(message)

			ticker := time.NewTicker(24 * time.Hour)
			defer ticker.Stop()

			go func() {
				for range ticker.C {
					slog.Error(message)
				}
			}()
		}

		slog.Info("🚀 starting permify service...")

		// Run database migration if enabled
		if cfg.Database.AutoMigrate {
			err = storage.Migrate(cfg.Database)
			if err != nil {
				slog.Error("failed to migrate database", slog.Any("error", err))
				return err
			}
		}

		// Initialize database
		db, err := factories.DatabaseFactory(cfg.Database)
		if err != nil {
			slog.Error("failed to initialize database", slog.Any("error", err))
			return err
		}
		defer func() {
			if err = db.Close(); err != nil {
				slog.Error("failed to close database", slog.Any("error", err))
			}
		}()

		// Tracing
		if cfg.Tracer.Enabled {
			headers := map[string]string{}
			for _, header := range cfg.Tracer.Headers {
				h := strings.Split(header, ":")
				if len(h) != 2 {
					return errors.New("invalid header format; expected 'key:value'")
				}
				headers[h[0]] = h[1]
			}

			var exporter trace.SpanExporter
			exporter, err = tracerexporters.ExporterFactory(
				cfg.Tracer.Exporter,
				cfg.Tracer.Endpoint,
				cfg.Tracer.Insecure,
				cfg.Tracer.Urlpath,
				headers,
				cfg.Tracer.Protocol,
			)
			if err != nil {
				slog.Error(err.Error())
			}

			shutdown := telemetry.NewTracer(exporter)

			defer func() {
				if err = shutdown(ctx); err != nil {
					slog.Error(err.Error())
				}
			}()
		}

		// Garbage collection
		if cfg.Database.GarbageCollection.Timeout > 0 && cfg.Database.GarbageCollection.Enabled && cfg.Database.Engine != "memory" {
			slog.Info("🗑️ starting database garbage collection...")

			garbageCollector := gc.NewGC(
				db.(*PQDatabase.Postgres),
				gc.Interval(cfg.Database.GarbageCollection.Interval),
				gc.Window(cfg.Database.GarbageCollection.Window),
				gc.Timeout(cfg.Database.GarbageCollection.Timeout),
			)

			go func() {
				err = garbageCollector.Start(ctx)
				if err != nil {
					slog.Error(err.Error())
				}
			}()
		}

		// Meter
		if cfg.Meter.Enabled {
			headers := map[string]string{}
			for _, header := range cfg.Meter.Headers {
				h := strings.Split(header, ":")
				if len(h) != 2 {
					return errors.New("invalid header format; expected 'key:value'")
				}
				headers[h[0]] = h[1]
			}

			var exporter metric.Exporter
			exporter, err = meterexporters.ExporterFactory(
				cfg.Meter.Exporter,
				cfg.Meter.Endpoint,
				cfg.Meter.Insecure,
				cfg.Meter.Urlpath,
				headers,
				cfg.Meter.Protocol,
			)
			if err != nil {
				slog.Error(err.Error())
			}

			shutdown := telemetry.NewMeter(exporter, time.Duration(cfg.Meter.Interval)*time.Second)

			defer func() {
				if err = shutdown(ctx); err != nil {
					slog.Error(err.Error())
				}
			}()
		}

		// schema cache
		var schemaCache pkgcache.Cache
		schemaCache, err = ristretto.New(ristretto.NumberOfCounters(cfg.Service.Schema.Cache.NumberOfCounters), ristretto.MaxCost(cfg.Service.Schema.Cache.MaxCost))
		if err != nil {
			slog.Error(err.Error())
			return err
		}

		// engines cache cache
		var engineKeyCache pkgcache.Cache
		engineKeyCache, err = ristretto.New(ristretto.NumberOfCounters(cfg.Service.Permission.Cache.NumberOfCounters), ristretto.MaxCost(cfg.Service.Permission.Cache.MaxCost))
		if err != nil {
			slog.Error(err.Error())
			return err
		}

		watcher := storage.NewNoopWatcher()
		if cfg.Service.Watch.Enabled {
			watcher = factories.WatcherFactory(db)
		}

		// Initialize the storage with factory methods
		dataReader := factories.DataReaderFactory(db)
		dataWriter := factories.DataWriterFactory(db)
		bundleReader := factories.BundleReaderFactory(db)
		bundleWriter := factories.BundleWriterFactory(db)
		schemaReader := factories.SchemaReaderFactory(db)
		schemaWriter := factories.SchemaWriterFactory(db)
		tenantReader := factories.TenantReaderFactory(db)
		tenantWriter := factories.TenantWriterFactory(db)

		// Add caching to the schema reader using a decorator
		schemaReader = cacheDecorator.NewSchemaReader(schemaReader, schemaCache)

		dataReader = sfDecorator.NewDataReader(dataReader)
		schemaReader = sfDecorator.NewSchemaReader(schemaReader)

		// Check if circuit breaker should be enabled for services
		if cfg.Service.CircuitBreaker {
			var cb *gobreaker.CircuitBreaker
			var st gobreaker.Settings
			st.Name = "storage"
			st.ReadyToTrip = func(counts gobreaker.Counts) bool {
				failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
				return counts.Requests >= 10 && failureRatio >= 0.6
			}

			cb = gobreaker.NewCircuitBreaker(st)

			// Add circuit breaker to the relationship reader using decorator
			dataReader = cbDecorator.NewDataReader(dataReader, cb)

			// Add circuit breaker to the bundle reader using decorators
			bundleReader = cbDecorator.NewBundleReader(bundleReader, cb)

			// Add circuit breaker to the schema reader using decorator
			schemaReader = cbDecorator.NewSchemaReader(schemaReader, cb)

			// Add circuit breaker to the tenant reader using decorator
			tenantReader = cbDecorator.NewTenantReader(tenantReader, cb)
		}

		// Initialize the engines using the key manager, schema reader, and relationship reader
		checkEngine := engines.NewCheckEngine(schemaReader, dataReader, engines.CheckConcurrencyLimit(cfg.Service.Permission.ConcurrencyLimit))
		expandEngine := engines.NewExpandEngine(schemaReader, dataReader)

		// Declare a variable `checker` of type `invoke.Check`.
		var checker invoke.Check

		checker = cache.NewCheckEngineWithCache(
			checkEngine,
			schemaReader,
			engineKeyCache,
		)

		// Create the checker either with load balancing or caching capabilities.
		if cfg.Distributed.Enabled {
			if cfg.Authn.Enabled && cfg.Authn.Method == "oidc" {
				return errors.New("OIDC authentication method cannot be used in distributed mode. Please check your configuration")
			}

			checker, err = balancer.NewCheckEngineWithBalancer(
				ctx,
				checker,
				schemaReader,
				cfg.NameOverride,
				&cfg.Distributed,
				&cfg.Server.GRPC,
				&cfg.Authn,
			)
			// Handle potential error during checker creation.
			if err != nil {
				return err
			}
		}

		// Create a localChecker which directly checks without considering distributed setup.
		// This also includes caching capabilities.
		localChecker := cache.NewCheckEngineWithCache(
			checkEngine,
			schemaReader,
			engineKeyCache,
		)

		// Initialize the lookupEngine, which is responsible for looking up certain entities or values.
		lookupEngine := engines.NewLookupEngine(
			checker,
			schemaReader,
			dataReader,
			// Set concurrency limit based on the configuration.
			engines.LookupConcurrencyLimit(cfg.Service.Permission.BulkLimit),
		)

		// Initialize the subjectPermissionEngine, responsible for handling subject permissions.
		subjectPermissionEngine := engines.NewSubjectPermission(
			checker,
			schemaReader,
			// Set concurrency limit for the subject permission checks.
			engines.SubjectPermissionConcurrencyLimit(cfg.Service.Permission.ConcurrencyLimit),
		)

		// Create a new invoker that is used to directly call various functions or engines.
		// It encompasses the schema, data, checker, and other engines.
		invoker := invoke.NewDirectInvoker(
			schemaReader,
			dataReader,
			checker,
			expandEngine,
			lookupEngine,
			subjectPermissionEngine,
		)

		// Associate the invoker with the checkEngine.
		checkEngine.SetInvoker(invoker)

		// Create a local invoker for local operations.
		localInvoker := invoke.NewDirectInvoker(
			schemaReader,
			dataReader,
			localChecker,
			expandEngine,
			lookupEngine,
			subjectPermissionEngine,
		)

		// Initialize the container which brings together multiple components such as the invoker, data readers/writers, and schema handlers.
		container := servers.NewContainer(
			invoker,
			dataReader,
			dataWriter,
			bundleReader,
			bundleWriter,
			schemaReader,
			schemaWriter,
			tenantReader,
			tenantWriter,
			watcher,
		)

		// Create an error group with the provided context
		var g *errgroup.Group
		g, ctx = errgroup.WithContext(ctx)

		// Add the container.Run function to the error group
		g.Go(func() error {
			return container.Run(
				ctx,
				&cfg.Server,
				logger,
				&cfg.Distributed,
				&cfg.Authn,
				&cfg.Profiler,
				localInvoker,
			)
		})

		// Wait for the error group to finish and log any errors
		if err = g.Wait(); err != nil {
			slog.Error(err.Error())
		}

		return nil
	}
}

// getLogLevel converts a string representation of log level to its corresponding slog.Level value.
func getLogLevel(level string) slog.Level {
	switch level {
	case "info":
		return slog.LevelInfo // Return Info level
	case "warn":
		return slog.LevelWarn // Return Warning level
	case "error":
		return slog.LevelError // Return Error level
	case "debug":
		return slog.LevelDebug // Return Debug level
	default:
		return slog.LevelInfo // Default to Info level if unrecognized
	}
}
