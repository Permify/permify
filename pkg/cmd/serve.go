package cmd

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/Permify/permify/internal/invoke"
	"github.com/Permify/permify/internal/repositories/postgres"
	PQDatabase "github.com/Permify/permify/pkg/database/postgres"

	"go.opentelemetry.io/otel/sdk/metric"

	"github.com/spf13/viper"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel/sdk/trace"
	"golang.org/x/sync/errgroup"

	"github.com/Permify/permify/internal"
	"github.com/Permify/permify/internal/config"
	"github.com/Permify/permify/internal/engines"
	"github.com/Permify/permify/internal/factories"
	"github.com/Permify/permify/internal/keys"
	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/internal/repositories/decorators"
	"github.com/Permify/permify/internal/servers"
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
		// Load configuration
		cfg, err := config.NewConfig()
		if err != nil {
			return fmt.Errorf("failed to create new config: %w", err)
		}

		if err = viper.Unmarshal(cfg); err != nil {
			return fmt.Errorf("failed to unmarshal config: %w", err)
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
			err = repositories.Migrate(cfg.Database, l)
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
			exporter, err = tracerexporters.ExporterFactory(cfg.Tracer.Exporter, cfg.Tracer.Endpoint)
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
		if cfg.DatabaseGarbageCollection.Enable && cfg.Database.Engine != "memory" {
			l.Info("🗑️ starting database garbage collection...")
			gc := postgres.NewGarbageCollector(ctx, db.(*PQDatabase.Postgres), l, cfg.DatabaseGarbageCollection)

			err := gc.Start()
			if err != nil {
				l.Fatal(err)
			}

			defer func() {
				gc.Stop()
			}()
		}

		// Meter
		// meter := telemetry.NewNoopMeter()
		if cfg.Meter.Enabled {
			var exporter metric.Exporter
			exporter, err = meterexporters.ExporterFactory(cfg.Meter.Exporter, cfg.Meter.Endpoint)
			if err != nil {
				l.Fatal(err)
			}

			_, err = telemetry.NewMeter(exporter)
			if err != nil {
				l.Fatal(err)
			}
		}

		// schema cache
		var schemaCache cache.Cache
		schemaCache, err = ristretto.New(ristretto.NumberOfCounters(cfg.Schema.Cache.NumberOfCounters), ristretto.MaxCost(cfg.Schema.Cache.MaxCost))
		if err != nil {
			l.Fatal(err)
		}

		// engines cache keys
		var engineKeyCache cache.Cache
		engineKeyCache, err = ristretto.New(ristretto.NumberOfCounters(cfg.Permission.Cache.NumberOfCounters), ristretto.MaxCost(cfg.Permission.Cache.MaxCost))
		if err != nil {
			l.Fatal(err)
		}

		// Initialize the repositories with factory methods
		relationshipReader := factories.RelationshipReaderFactory(db, l)
		relationshipWriter := factories.RelationshipWriterFactory(db, l)
		schemaReader := factories.SchemaReaderFactory(db, l)
		schemaWriter := factories.SchemaWriterFactory(db, l)
		tenantReader := factories.TenantReaderFactory(db, l)
		tenantWriter := factories.TenantWriterFactory(db, l)

		// Add caching to the schema reader using a decorator
		schemaReader = decorators.NewSchemaReaderWithCache(schemaReader, schemaCache)

		// Check if circuit breaker should be enabled for services
		if cfg.Service.CircuitBreaker {
			// Add circuit breaker to the relationship reader and writer using decorators
			relationshipWriter = decorators.NewRelationshipWriterWithCircuitBreaker(relationshipWriter)
			relationshipReader = decorators.NewRelationshipReaderWithCircuitBreaker(relationshipReader)

			// Add circuit breaker to the schema reader and writer using decorators
			schemaWriter = decorators.NewSchemaWriterWithCircuitBreaker(schemaWriter)
			schemaReader = decorators.NewSchemaReaderWithCircuitBreaker(schemaReader)
		}

		// Initialize the key manager for the check engine
		checkKeyManager := keys.NewCheckEngineKeys(engineKeyCache)

		// Initialize the engines using the key manager, schema reader, and relationship reader
		checkEngine := engines.NewCheckEngine(checkKeyManager, schemaReader, relationshipReader, engines.CheckConcurrencyLimit(cfg.Permission.ConcurrencyLimit))
		linkedEntityEngine := engines.NewLinkedEntityEngine(schemaReader, relationshipReader)
		lookupEntityEngine := engines.NewLookupEntityEngine(checkEngine, linkedEntityEngine, engines.LookupEntityConcurrencyLimit(cfg.Permission.BulkLimit))
		expandEngine := engines.NewExpandEngine(schemaReader, relationshipReader)
		schemaLookupEngine := engines.NewLookupSchemaEngine(schemaReader)

		// Create the container with engines, repositories, and other dependencies
		container := servers.NewContainer(
			invoke.NewDirectInvoker(
				checkEngine,
				expandEngine,
				schemaLookupEngine,
				lookupEntityEngine,
			),
			relationshipReader,
			relationshipWriter,
			schemaReader,
			schemaWriter,
			tenantReader,
			tenantWriter,
		)

		// Create an error group with the provided context
		var g *errgroup.Group
		g, ctx = errgroup.WithContext(ctx)

		// Add the container.Run function to the error group
		g.Go(func() error {
			return container.Run(ctx, &cfg.Server, &cfg.Authn, &cfg.Profiler, l)
		})

		// Wait for the error group to finish and log any errors
		if err = g.Wait(); err != nil {
			l.Error(err)
		}

		return nil
	}
}
