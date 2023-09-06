package cmd

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/Permify/permify/internal/engines/consistent"
	"github.com/Permify/permify/internal/engines/keys"
	"github.com/Permify/permify/internal/invoke"
	"github.com/Permify/permify/internal/storage/postgres/gc"
	hash "github.com/Permify/permify/pkg/consistent"
	PQDatabase "github.com/Permify/permify/pkg/database/postgres"

	"github.com/spf13/viper"
	"go.opentelemetry.io/otel/sdk/metric"

	"github.com/Permify/permify/pkg/gossip"

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
	"github.com/Permify/permify/pkg/logger"
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
		l := logger.New(cfg.Log.Level)
		l.Info("🚀 starting permify service...")

		// Set up context and signal handling
		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		// Run database migration if enabled
		if cfg.AutoMigrate {
			err = storage.Migrate(cfg.Database, l)
			if err != nil {
				l.Fatal("failed to migrate database: %w", err)
			}
		}

		// Initialize database
		db, err := factories.DatabaseFactory(cfg.Database)
		if err != nil {
			l.Fatal("failed to initialize database: %w", err)
		}
		defer func() {
			if err = db.Close(); err != nil {
				l.Fatal("failed to close database: %v", err)
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
				l.Fatal(err)
			}

			shutdown := telemetry.NewTracer(exporter)

			defer func() {
				if err = shutdown(context.Background()); err != nil {
					l.Fatal(err)
				}
			}()
		}

		// Garbage collection
		if cfg.Database.GarbageCollection.Timeout > 0 && cfg.Database.GarbageCollection.Enabled && cfg.Database.Engine != "memory" {
			l.Info("🗑️ starting database garbage collection...")

			garbageCollector := gc.NewGC(
				db.(*PQDatabase.Postgres),
				l,
				gc.Interval(cfg.Database.GarbageCollection.Interval),
				gc.Window(cfg.Database.GarbageCollection.Window),
				gc.Timeout(cfg.Database.GarbageCollection.Timeout),
			)

			go func() {
				err = garbageCollector.Start(ctx)
				if err != nil {
					l.Fatal(err)
				}
			}()
		}

		// Meter
		meter := telemetry.NewNoopMeter()
		if cfg.Meter.Enabled {
			var exporter metric.Exporter
			exporter, err = meterexporters.ExporterFactory(cfg.Meter.Exporter, cfg.Meter.Endpoint)
			if err != nil {
				l.Fatal(err)
			}

			meter, err = telemetry.NewMeter(exporter)
			if err != nil {
				l.Fatal(err)
			}
		}

		var gossipEngine gossip.IGossip
		var consistencyChecker *hash.ConsistentHash
		if cfg.Distributed.Enabled {
			l.Info("🔗 starting distributed mode...")

			consistencyChecker, err = hash.NewConsistentHash(100, nil, cfg.GRPC)
			if err != nil {
				l.Fatal(err)
			}

			externalIP, err := gossip.ExternalIP()
			if err != nil {
				l.Fatal(err)
			}
			l.Info("🔗 external IP: " + externalIP)

			gossipEngine, err = gossip.InitMemberList(cfg.Distributed.Node, cfg.Distributed.Protocol)
			if err != nil {
				l.Info("🔗 failed to start distributed mode: %s ", err.Error())
				return err
			}

			go gossipEngine.SyncNodes(ctx, consistencyChecker, cfg.Distributed.NodeName, cfg.Server.GRPC.Port)

			defer func() {
				if err = gossipEngine.Shutdown(); err != nil {
					l.Fatal(err)
				}
			}()
		}

		// schema cache
		var schemaCache cache.Cache
		schemaCache, err = ristretto.New(ristretto.NumberOfCounters(cfg.Service.Schema.Cache.NumberOfCounters), ristretto.MaxCost(cfg.Service.Schema.Cache.MaxCost))
		if err != nil {
			l.Fatal(err)
		}

		// engines cache keys
		var engineKeyCache cache.Cache
		engineKeyCache, err = ristretto.New(ristretto.NumberOfCounters(cfg.Service.Permission.Cache.NumberOfCounters), ristretto.MaxCost(cfg.Service.Permission.Cache.MaxCost))
		if err != nil {
			l.Fatal(err)
		}

		watcher := storage.NewNoopWatcher()
		if cfg.Service.Watch.Enabled {
			watcher = factories.WatcherFactory(db, l)
		}

		// Initialize the storage with factory methods
		dataReader := factories.DataReaderFactory(db, l)
		dataWriter := factories.DataWriterFactory(db, l)
		schemaReader := factories.SchemaReaderFactory(db, l)
		schemaWriter := factories.SchemaWriterFactory(db, l)
		tenantReader := factories.TenantReaderFactory(db, l)
		tenantWriter := factories.TenantWriterFactory(db, l)

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

		// filters
		schemaBaseEntityFilter := engines.NewSchemaBasedEntityFilter(schemaReader, dataReader)
		massEntityFilter := engines.NewMassEntityFilter(dataReader)

		schemaBaseSubjectFilter := engines.NewSchemaBasedSubjectFilter(schemaReader, dataReader, engines.SchemaBaseSubjectFilterConcurrencyLimit(cfg.Service.Permission.ConcurrencyLimit))
		massSubjectFilter := engines.NewMassSubjectFilter(dataReader)

		// Initialize the engines using the key manager, schema reader, and relationship reader
		checkEngine := engines.NewCheckEngine(schemaReader, dataReader, engines.CheckConcurrencyLimit(cfg.Service.Permission.ConcurrencyLimit))
		expandEngine := engines.NewExpandEngine(schemaReader, dataReader)
		lookupEntityEngine := engines.NewLookupEntityEngine(checkEngine, schemaReader, schemaBaseEntityFilter, massEntityFilter, engines.LookupEntityConcurrencyLimit(cfg.Service.Permission.BulkLimit))
		lookupSubjectEngine := engines.NewLookupSubjectEngine(checkEngine, schemaReader, schemaBaseSubjectFilter, massSubjectFilter, engines.LookupSubjectConcurrencyLimit(cfg.Service.Permission.ConcurrencyLimit))
		subjectPermissionEngine := engines.NewSubjectPermission(checkEngine, schemaReader, engines.SubjectPermissionConcurrencyLimit(cfg.Service.Permission.ConcurrencyLimit))

		var check invoke.Check
		if cfg.Distributed.Enabled {
			check, err = consistent.NewCheckEngineWithHashring(
				keys.NewCheckEngineWithKeys(
					checkEngine,
					engineKeyCache,
					l,
				),
				consistencyChecker,
				gossipEngine,
				cfg.Server.GRPC.Port,
				l,
			)
			if err != nil {
				return err
			}
		} else {
			check = keys.NewCheckEngineWithKeys(
				checkEngine,
				engineKeyCache,
				l,
			)
		}

		invoker := invoke.NewDirectInvoker(
			schemaReader,
			dataReader,
			check,
			expandEngine,
			lookupEntityEngine,
			lookupSubjectEngine,
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
				l,
			)
		})

		// Wait for the error group to finish and log any errors
		if err = g.Wait(); err != nil {
			l.Error(err)
		}

		return nil
	}
}
