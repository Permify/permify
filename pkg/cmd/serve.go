package cmd

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
	"golang.org/x/sync/errgroup"

	"github.com/Permify/permify/internal/commands"
	"github.com/Permify/permify/internal/config"
	"github.com/Permify/permify/internal/factories"
	"github.com/Permify/permify/internal/keys"
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

const (
	// Version of Permify
	Version = "v0.2.2"
	banner  = `

██████╗ ███████╗██████╗ ███╗   ███╗██╗███████╗██╗   ██╗
██╔══██╗██╔════╝██╔══██╗████╗ ████║██║██╔════╝╚██╗ ██╔╝
██████╔╝█████╗  ██████╔╝██╔████╔██║██║█████╗   ╚████╔╝ 
██╔═══╝ ██╔══╝  ██╔══██╗██║╚██╔╝██║██║██╔══╝    ╚██╔╝  
██║     ███████╗██║  ██║██║ ╚═╝ ██║██║██║        ██║   
╚═╝     ╚══════╝╚═╝  ╚═╝╚═╝     ╚═╝╚═╝╚═╝        ╚═╝   
_______________________________________________________
Fine-grained Authorization System %s
`
)

// NewServeCommand - Creates new server command
func NewServeCommand(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "serve the Permify server",
		RunE:  serve(cfg),
		Args:  cobra.NoArgs,
	}
}

// serve - permify serve command
func serve(cfg *config.Config) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		var err error

		red := color.New(color.FgGreen)
		_, _ = red.Printf(banner, Version)

		l := logger.New(cfg.Log.Level)

		l.Info("🚀 starting permify service...")

		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		var db database.Database
		db, err = factories.DatabaseFactory(cfg.Database)
		if err != nil {
			l.Fatal(err)
		}
		defer db.Close()

		// Migration
		err = db.Migrate(factories.MigrationFactory(database.Engine(cfg.Database.Engine)))
		if err != nil {
			l.Fatal(err)
		}

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
		schemaCache, err = ristretto.New()
		if err != nil {
			l.Fatal(err)
		}

		// commands cache keys
		var commandsKeyCache cache.Cache
		commandsKeyCache, err = ristretto.New()
		if err != nil {
			l.Fatal(err)
		}

		// Repositories
		relationshipReader := factories.RelationshipReaderFactory(db, l)
		relationshipWriter := factories.RelationshipWriterFactory(db, l)
		schemaReader := factories.SchemaReaderFactory(db, l)
		schemaWriter := factories.SchemaWriterFactory(db, l)

		// decorators
		schemaReaderWithCache := decorators.NewSchemaReaderWithCache(schemaReader, schemaCache)

		// Service
		if cfg.Service.CircuitBreaker {
			relationshipWriter = decorators.NewRelationshipWriterWithCircuitBreaker(relationshipWriter)
			relationshipReader = decorators.NewRelationshipReaderWithCircuitBreaker(relationshipReader)

			schemaWriter = decorators.NewSchemaWriterWithCircuitBreaker(schemaWriter)
			schemaReader = decorators.NewSchemaReaderWithCircuitBreaker(schemaReaderWithCache)
		}

		// key managers
		checkKeyManager := keys.NewCheckCommandKeys(commandsKeyCache)

		// commands
		var checkCommand *commands.CheckCommand
		checkCommand, err = commands.NewCheckCommand(checkKeyManager, schemaReader, relationshipReader, meter, commands.ConcurrencyLimit(cfg.Service.ConcurrencyLimit))
		if err != nil {
			l.Fatal(err)
		}

		expandCommand := commands.NewExpandCommand(schemaReader, relationshipReader)
		schemaLookupCommand := commands.NewLookupSchemaCommand(schemaReader)
		lookupEntityCommand := commands.NewLookupEntityCommand(checkCommand, schemaReader, relationshipReader)

		// Services
		relationshipService := services.NewRelationshipService(relationshipReader, relationshipWriter, schemaReader)
		permissionService := services.NewPermissionService(checkCommand, expandCommand, schemaLookupCommand, lookupEntityCommand)
		schemaService := services.NewSchemaService(schemaWriter, schemaReaderWithCache)

		container := servers.ServiceContainer{
			RelationshipService: relationshipService,
			PermissionService:   permissionService,
			SchemaService:       schemaService,
		}

		var g *errgroup.Group
		g, ctx = errgroup.WithContext(ctx)

		g.Go(func() error {
			return container.Run(ctx, &cfg.Server, &cfg.Authn, l)
		})

		if err = g.Wait(); err != nil {
			l.Error(err)
		}

		return nil
	}
}

// RegisterServeFlags - Define and registers permify CLI flags
func RegisterServeFlags(cmd *cobra.Command, config *config.Config) {
	// GRPC Server
	cmd.Flags().StringVar(&config.Server.GRPC.Port, "grpc-port", config.Server.GRPC.Port, "port that GRPC server run on")
	cmd.Flags().StringVar(&config.Server.GRPC.TLSConfig.KeyPath, "grpc-tls-config-key-path", config.Server.GRPC.TLSConfig.KeyPath, "GRPC tls key path")
	cmd.Flags().StringVar(&config.Server.GRPC.TLSConfig.CertPath, "grpc-tls-config-cert-path", config.Server.GRPC.TLSConfig.CertPath, "GRPC tls certificate path")

	// HTTP Server
	cmd.Flags().BoolVar(&config.Server.HTTP.Enabled, "http-enabled", config.Server.HTTP.Enabled, "switch option for HTTP server")
	cmd.Flags().StringVar(&config.Server.HTTP.Port, "http-port", config.Server.HTTP.Port, "HTTP port address")
	cmd.Flags().StringVar(&config.Server.HTTP.TLSConfig.KeyPath, "http-tls-config-key-path", config.Server.HTTP.TLSConfig.KeyPath, "HTTP tls key path")
	cmd.Flags().StringVar(&config.Server.HTTP.TLSConfig.CertPath, "http-tls-config-cert-path", config.Server.HTTP.TLSConfig.CertPath, "HTTP tls certificate path")
	cmd.Flags().StringSliceVar(&config.Server.HTTP.CORSAllowedOrigins, "http-cors-allowed-origins", config.Server.HTTP.CORSAllowedOrigins, "CORS allowed origins for http gateway")
	cmd.Flags().StringSliceVar(&config.Server.HTTP.CORSAllowedHeaders, "http-cors-allowed-headers", config.Server.HTTP.CORSAllowedHeaders, "CORS allowed headers for http gateway")

	// LOG
	cmd.Flags().StringVar(&config.Log.Level, "log-level", config.Log.Level, "real time logs of authorization. Permify uses zerolog as a logger")

	// AUTHN
	cmd.Flags().BoolVar(&config.Authn.Enabled, "authn-enabled", config.Authn.Enabled, "enable server authentication")
	cmd.Flags().StringSliceVar(&config.Authn.Keys, "authn-preshared-keys", config.Authn.Keys, "preshared key/keys for server authentication")

	// TRACER
	cmd.Flags().BoolVar(&config.Tracer.Enabled, "tracer-enabled", config.Tracer.Enabled, "switch option for tracing")
	cmd.Flags().StringVar(&config.Tracer.Exporter, "tracer-exporter", config.Tracer.Exporter, "export uri for tracing data")
	cmd.Flags().StringVar(&config.Tracer.Endpoint, "tracer-endpoint", config.Tracer.Endpoint, "can be; jaeger, signoz, zipkin or otlp. (integrated tracing tools)")

	// METER
	cmd.Flags().BoolVar(&config.Meter.Enabled, "meter-enabled", config.Meter.Enabled, "switch option for metric")
	cmd.Flags().StringVar(&config.Meter.Exporter, "meter-exporter", config.Meter.Exporter, "export uri for metric data")
	cmd.Flags().StringVar(&config.Meter.Endpoint, "meter-endpoint", config.Meter.Endpoint, "can be; otlp. (integrated metric tools)")

	// SERVICE
	cmd.Flags().BoolVar(&config.Service.CircuitBreaker, "service-circuit-breaker", config.Service.CircuitBreaker, "switch option for service circuit breaker")
	cmd.Flags().IntVar(&config.Service.ConcurrencyLimit, "service-concurrency-limit", config.Service.ConcurrencyLimit, "concurrency limit")

	// DATABASE
	cmd.Flags().StringVar(&config.Database.Engine, "database-engine", config.Database.Engine, "data source. e.g. postgres, memory")
	cmd.Flags().IntVar(&config.Database.MaxOpenConnections, "database-max-open-connections", config.Database.MaxOpenConnections, "maximum number of parallel connections that can be made to the database at any time")
	cmd.Flags().StringVar(&config.Database.Database, "database-name", config.Database.Database, "custom database name")
	cmd.Flags().StringVar(&config.Database.URI, "database-uri", config.Database.URI, "uri of your data source to store relation tuples and schema")
}
