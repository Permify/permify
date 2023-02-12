package cmd

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/spf13/viper"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
	"golang.org/x/sync/errgroup"

	`github.com/Permify/permify/internal`
	"github.com/Permify/permify/internal/commands"
	"github.com/Permify/permify/internal/config"
	"github.com/Permify/permify/internal/factories"
	"github.com/Permify/permify/internal/keys"
	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/internal/repositories/decorators"
	"github.com/Permify/permify/internal/servers"
	"github.com/Permify/permify/internal/services"
	"github.com/Permify/permify/pkg/cache"
	"github.com/Permify/permify/pkg/cache/ristretto"
	"github.com/Permify/permify/pkg/database"
	"github.com/Permify/permify/pkg/logger"
	"github.com/Permify/permify/pkg/telemetry"
	"github.com/Permify/permify/pkg/telemetry/meterexporters"
	"github.com/Permify/permify/pkg/telemetry/tracerexporters"
)

// NewServeCommand - Creates new server command
func NewServeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "serve the Permify server",
		RunE:  serve(),
		Args:  cobra.NoArgs,
	}
}

// serve - permify serve command
func serve() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		cfg, err := config.NewConfig()
		if err = viper.Unmarshal(cfg); err != nil {
			return err
		}

		red := color.New(color.FgGreen)
		_, _ = red.Printf(internal.Banner, internal.Version)

		l := logger.New(cfg.Log.Level)

		l.Info("🚀 starting permify service...")

		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		if cfg.AutoMigrate {
			err = repositories.Migrate(cfg.Database, l)
			if err != nil {
				l.Fatal(err)
			}
		}

		var db database.Database
		db, err = factories.DatabaseFactory(cfg.Database)
		if err != nil {
			l.Fatal(err)
		}
		defer db.Close()

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

		// schema cache
		var schemaCache cache.Cache
		schemaCache, err = ristretto.New(ristretto.NumberOfCounters(cfg.Schema.Cache.NumberOfCounters), ristretto.MaxCost(cfg.Schema.Cache.MaxCost))
		if err != nil {
			l.Fatal(err)
		}

		// commands cache keys
		var commandsKeyCache cache.Cache
		commandsKeyCache, err = ristretto.New(ristretto.NumberOfCounters(cfg.Permission.Cache.NumberOfCounters), ristretto.MaxCost(cfg.Permission.Cache.MaxCost))
		if err != nil {
			l.Fatal(err)
		}

		// Repositories
		relationshipReader := factories.RelationshipReaderFactory(db, l)
		relationshipWriter := factories.RelationshipWriterFactory(db, l)
		schemaReader := factories.SchemaReaderFactory(db, l)
		schemaWriter := factories.SchemaWriterFactory(db, l)
		tenantReader := factories.TenantReaderFactory(db, l)
		tenantWriter := factories.TenantWriterFactory(db, l)

		// decorators
		schemaReader = decorators.NewSchemaReaderWithCache(schemaReader, schemaCache)

		// Service
		if cfg.Service.CircuitBreaker {
			relationshipWriter = decorators.NewRelationshipWriterWithCircuitBreaker(relationshipWriter)
			relationshipReader = decorators.NewRelationshipReaderWithCircuitBreaker(relationshipReader)

			schemaWriter = decorators.NewSchemaWriterWithCircuitBreaker(schemaWriter)
			schemaReader = decorators.NewSchemaReaderWithCircuitBreaker(schemaReader)
		}

		// key managers
		checkKeyManager := keys.NewCheckCommandKeys(commandsKeyCache)

		// commands
		var checkCommand *commands.CheckCommand
		checkCommand, err = commands.NewCheckCommand(checkKeyManager, schemaReader, relationshipReader, meter, commands.ConcurrencyLimit(cfg.Permission.ConcurrencyLimit))
		if err != nil {
			l.Fatal(err)
		}

		expandCommand := commands.NewExpandCommand(schemaReader, relationshipReader)
		schemaLookupCommand := commands.NewLookupSchemaCommand(schemaReader)
		lookupEntityCommand := commands.NewLookupEntityCommand(checkCommand, schemaReader, relationshipReader)

		// Services
		relationshipService := services.NewRelationshipService(relationshipReader, relationshipWriter, schemaReader)
		permissionService := services.NewPermissionService(checkCommand, expandCommand, schemaLookupCommand, lookupEntityCommand)
		schemaService := services.NewSchemaService(schemaWriter, schemaReader)
		tenancyService := services.NewTenancyService(tenantWriter, tenantReader)

		container := servers.ServiceContainer{
			RelationshipService: relationshipService,
			PermissionService:   permissionService,
			SchemaService:       schemaService,
			TenancyService:      tenancyService,
		}

		var g *errgroup.Group
		g, ctx = errgroup.WithContext(ctx)

		g.Go(func() error {
			return container.Run(ctx, &cfg.Server, &cfg.Authn, &cfg.Profiler, l)
		})

		if err = g.Wait(); err != nil {
			l.Error(err)
		}

		return nil
	}
}
