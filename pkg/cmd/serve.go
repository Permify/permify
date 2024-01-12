package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sony/gobreaker"
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

	"github.com/fatih/color"
	"github.com/spf13/cobra"
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

	// register flags for serve
	flags.RegisterServeFlags(command)

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
		red := color.New(color.FgGreen)
		_, _ = red.Printf(internal.Banner, internal.Version)

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

		logger := slog.New(handler)

		slog.SetDefault(logger)

		slog.Info("🚀 starting permify service...")

		internal.Identifier = cfg.AccountID
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

		// Set up context and signal handling
		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		// Run database migration if enabled
		if cfg.Database.AutoMigrate {
			err = storage.Migrate(cfg.Database)
			if err != nil {
				slog.Error("failed to migrate database: %w", err)
			}
		}

		// Initialize database
		db, err := factories.DatabaseFactory(cfg.Database)
		if err != nil {
			slog.Error("failed to initialize database: %w", err)
		}
		defer func() {
			if err = db.Close(); err != nil {
				slog.Error("failed to close database: %v", err)
			}
		}()

		// Tracing
		if cfg.Tracer.Enabled {
			var exporter trace.SpanExporter
			exporter, err = tracerexporters.ExporterFactory(
				cfg.Tracer.Exporter,
				cfg.Tracer.Endpoint,
				cfg.Tracer.Insecure,
				cfg.Tracer.URLPath,
			)
			if err != nil {
				slog.Error(err.Error())
			}

			shutdown := telemetry.NewTracer(exporter)

			defer func() {
				if err = shutdown(context.Background()); err != nil {
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
		meter := telemetry.NewNoopMeter()
		if cfg.Meter.Enabled {
			var exporter metric.Exporter
			exporter, err = meterexporters.ExporterFactory(
				cfg.Meter.Exporter,
				cfg.Meter.Endpoint,
				cfg.Meter.Insecure,
				cfg.Meter.URLPath,
			)
			if err != nil {
				slog.Error(err.Error())
			}

			meter, err = telemetry.NewMeter(exporter)
			if err != nil {
				slog.Error(err.Error())
			}
		}

		// schema cache
		var schemaCache pkgcache.Cache
		schemaCache, err = ristretto.New(ristretto.NumberOfCounters(cfg.Service.Schema.Cache.NumberOfCounters), ristretto.MaxCost(cfg.Service.Schema.Cache.MaxCost))
		if err != nil {
			slog.Error(err.Error())
		}

		// engines cache cache
		var engineKeyCache pkgcache.Cache
		engineKeyCache, err = ristretto.New(ristretto.NumberOfCounters(cfg.Service.Permission.Cache.NumberOfCounters), ristretto.MaxCost(cfg.Service.Permission.Cache.MaxCost))
		if err != nil {
			slog.Error(err.Error())
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

		// Create the checker either with load balancing or caching capabilities.
		if cfg.Distributed.Enabled {
			checker, err = balancer.NewCheckEngineWithBalancer(
				context.Background(),
				checkEngine,
				schemaReader,
				&cfg.Distributed,
				&cfg.Server.GRPC,
				&cfg.Authn,
			)
			// Handle potential error during checker creation.
			if err != nil {
				return err
			}
			checker = cache.NewCheckEngineWithCache(checker, schemaReader, engineKeyCache)
		} else {
			checker = cache.NewCheckEngineWithCache(checkEngine, schemaReader, engineKeyCache)
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
			meter,
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
			meter,
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
