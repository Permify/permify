package cmd

import (
	"context"
	"fmt"
	"os/signal"
	"strconv"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/Permify/permify/internal/engines/consistent"
	"github.com/Permify/permify/internal/engines/keys"
	"github.com/Permify/permify/internal/invoke"
	"github.com/Permify/permify/internal/storage/postgres"
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
		if cfg.DatabaseGarbageCollection.Enabled && cfg.Database.Engine != "memory" {
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

		var gossipEngine *gossip.Gossip
		var consistencyChecker *hash.ConsistentHash
		if cfg.Distributed.Enabled {
			l.Info("🔗 starting distributed mode...")

			consistencyChecker = hash.NewConsistentHash(100, cfg.Distributed.Nodes, nil)

			externalIP, err := gossip.ExternalIP()
			if err != nil {
				l.Fatal(err)
			}
			l.Info("🔗 external IP: " + externalIP)

			consistencyChecker.Add(externalIP + ":" + cfg.HTTP.Port)

			grpcPort, err := strconv.Atoi(cfg.Server.GRPC.Port)
			if err != nil {
				return err
			}

			gossipEngine, err = gossip.InitMemberList(cfg.Distributed.Nodes, grpcPort)
			if err != nil {
				l.Info("🔗 failed to start distributed mode: %s ", err.Error())
				return err
			}

			go func() {
				for {
					consistencyChecker.SyncNodes(gossipEngine)
				}
			}()

			defer func() {
				if err = gossipEngine.Shutdown(); err != nil {
					l.Fatal(err)
				}
			}()
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

		// Initialize the engines using the key manager, schema reader, and relationship reader
		checkEngine := engines.NewCheckEngine(schemaReader, dataReader, engines.CheckConcurrencyLimit(cfg.Permission.ConcurrencyLimit))
		expandEngine := engines.NewExpandEngine(schemaReader, dataReader)
		entityFilterEngine := engines.NewEntityFilterEngine(schemaReader, dataReader)
		lookupEntityEngine := engines.NewLookupEntityEngine(checkEngine, entityFilterEngine, engines.LookupEntityConcurrencyLimit(cfg.Permission.BulkLimit))
		lookupSubjectEngine := engines.NewLookupSubjectEngine(schemaReader, dataReader, engines.LookupSubjectConcurrencyLimit(cfg.Permission.ConcurrencyLimit))
		subjectPermissionEngine := engines.NewSubjectPermission(checkEngine, schemaReader, engines.SubjectPermissionConcurrencyLimit(cfg.Permission.ConcurrencyLimit))

		var check invoke.Check
		if cfg.Distributed.Enabled {
			options := []grpc.DialOption{
				grpc.WithBlock(),
			}
			if cfg.GRPC.TLSConfig.Enabled {
				c, err := credentials.NewClientTLSFromFile(cfg.GRPC.TLSConfig.CertPath, "")
				if err != nil {
					return err
				}
				options = append(options, grpc.WithTransportCredentials(c))
			} else {
				options = append(options, grpc.WithTransportCredentials(insecure.NewCredentials()))
			}

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
				options...,
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
			return container.Run(ctx, &cfg.Server, &cfg.Authn, &cfg.Profiler, l)
		})

		// Wait for the error group to finish and log any errors
		if err = g.Wait(); err != nil {
			l.Error(err)
		}

		return nil
	}
}
