package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Permify/permify/internal/engines/consistent"
	"github.com/Permify/permify/internal/engines/keys"
	"github.com/Permify/permify/internal/invoke"
	"github.com/Permify/permify/internal/storage/postgres/gc"
	hash "github.com/Permify/permify/pkg/consistent"
	PQDatabase "github.com/Permify/permify/pkg/database/postgres"
	"github.com/Permify/permify/pkg/gossip"

	"github.com/spf13/viper"
	"go.opentelemetry.io/otel/sdk/metric"

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
	"github.com/Permify/permify/internal/storage/decorators"
	"github.com/Permify/permify/pkg/cache"
	"github.com/Permify/permify/pkg/cache/ristretto"
	"github.com/Permify/permify/pkg/telemetry"
	"github.com/Permify/permify/pkg/telemetry/meterexporters"
	"github.com/Permify/permify/pkg/telemetry/tracerexporters"
)

// NewServeCommand returns a new Cobra command that can be used to run the "permify serve" command.
// The command takes no arguments and runs the serve() function to start the Permify service.
// The command has a short description of what it does.
func NewServeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "serve the Permify server",
		RunE:  serve(),
		Args:  cobra.NoArgs,
	}
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

		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		slog.SetDefault(logger)

		slog.Info("🚀 starting permify service...")

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
			exporter, err = meterexporters.ExporterFactory(cfg.Meter.Exporter, cfg.Meter.Endpoint)
			if err != nil {
				slog.Error(err.Error())
			}

			meter, err = telemetry.NewMeter(exporter)
			if err != nil {
				slog.Error(err.Error())
			}
		}

		var gossipEngine gossip.IGossip
		var consistencyChecker *hash.ConsistentHash

		if cfg.Distributed.Enabled {
			slog.Info("🔗 starting distributed mode...")

			consistencyChecker, err = hash.NewConsistentHash(100, nil, cfg.GRPC)
			if err != nil {
				slog.Error(err.Error())
			}

			externalIP, err := gossip.ExternalIP()
			if err != nil {
				slog.Error(err.Error())
			}

			slog.Info("🔗 external IP: " + externalIP)

			gossipEngine, err = gossip.InitMemberList(cfg.Distributed.Node, cfg.Distributed.Protocol)
			if err != nil {
				slog.Info("🔗 failed to start distributed mode: %s ", err.Error())
				return err
			}

			go gossipEngine.SyncNodes(ctx, consistencyChecker, cfg.Distributed.NodeName, cfg.Server.GRPC.Port)

			defer func() {
				if err = gossipEngine.Shutdown(); err != nil {
					slog.Error(err.Error())
				}
			}()
		}

		// schema cache
		var schemaCache cache.Cache
		schemaCache, err = ristretto.New(ristretto.NumberOfCounters(cfg.Service.Schema.Cache.NumberOfCounters), ristretto.MaxCost(cfg.Service.Schema.Cache.MaxCost))
		if err != nil {
			slog.Error(err.Error())
		}

		// engines cache keys
		var engineKeyCache cache.Cache
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
		schemaReader := factories.SchemaReaderFactory(db)
		schemaWriter := factories.SchemaWriterFactory(db)
		tenantReader := factories.TenantReaderFactory(db)
		tenantWriter := factories.TenantWriterFactory(db)

		// Add caching to the schema reader using a decorator
		schemaReader = decorators.NewSchemaReaderWithCache(schemaReader, schemaCache)

		// Check if circuit breaker should be enabled for services
		if cfg.Service.CircuitBreaker {
			// Add circuit breaker to the relationship reader and writer using decorators
			dataWriter = decorators.NewDataWriterWithCircuitBreaker(dataWriter)
			dataReader = decorators.NewDataReaderWithCircuitBreaker(dataReader)

			// Add circuit breaker to the schema reader and writer using decorators
			schemaWriter = decorators.NewSchemaWriterWithCircuitBreaker(schemaWriter)
			schemaReader = decorators.NewSchemaReaderWithCircuitBreaker(schemaReader)
		}

		// Initialize the engines using the key manager, schema reader, and relationship reader
		checkEngine := engines.NewCheckEngine(schemaReader, dataReader, engines.CheckConcurrencyLimit(cfg.Service.Permission.ConcurrencyLimit))
		expandEngine := engines.NewExpandEngine(schemaReader, dataReader)

		var checker invoke.Check
		if cfg.Distributed.Enabled {
			checker, err = consistent.NewCheckEngineWithHashring(
				keys.NewCheckEngineWithKeys(
					checkEngine,
					schemaReader,
					engineKeyCache,
				),
				schemaReader,
				consistencyChecker,
				gossipEngine,
				cfg.Server.GRPC.Port,
			)
			if err != nil {
				return err
			}
		} else {
			checker = keys.NewCheckEngineWithKeys(
				checkEngine,
				schemaReader,
				engineKeyCache,
			)
		}

		lookupEngine := engines.NewLookupEngine(
			checker,
			schemaReader,
			dataReader,
			engines.LookupConcurrencyLimit(cfg.Service.Permission.BulkLimit),
		)

		subjectPermissionEngine := engines.NewSubjectPermission(
			checker,
			schemaReader,
			engines.SubjectPermissionConcurrencyLimit(cfg.Service.Permission.ConcurrencyLimit),
		)

		invoker := invoke.NewDirectInvoker(
			schemaReader,
			dataReader,
			checker,
			expandEngine,
			lookupEngine,
			subjectPermissionEngine,
			meter,
		)

		checkEngine.SetInvoker(invoker)

		// Create the container with engines, storage, and other dependencies
		container := servers.NewContainer(
			invoker,
			dataReader,
			dataWriter,
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
				&cfg.Authn,
				&cfg.Profiler,
			)
		})

		// Wait for the error group to finish and log any errors
		if err = g.Wait(); err != nil {
			slog.Error(err.Error())
		}

		return nil
	}
}
