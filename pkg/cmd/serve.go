package cmd

import (
	`context`
	`fmt`
	`github.com/spf13/cobra`
	`golang.org/x/sync/errgroup`
	`log`
	`os/signal`
	`syscall`

	`github.com/Permify/permify/internal/commands`
	`github.com/Permify/permify/internal/config`
	`github.com/Permify/permify/internal/factories`
	`github.com/Permify/permify/internal/managers`
	`github.com/Permify/permify/internal/repositories/decorators`
	`github.com/Permify/permify/internal/servers`
	`github.com/Permify/permify/internal/services`
	`github.com/Permify/permify/pkg/cache`
	`github.com/Permify/permify/pkg/cache/ristretto`
	`github.com/Permify/permify/pkg/database`
	`github.com/Permify/permify/pkg/logger`
	`github.com/Permify/permify/pkg/telemetry`
	`github.com/Permify/permify/pkg/telemetry/exporters`
)

const (
	// Version of Permify
	Version = "v0.0.0-alpha8"
	color   = "\033[0;37m%s\033[0m"
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

// NewServeCommand -
func NewServeCommand(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "serve the Permify server",
		RunE:  serve(cfg),
		Args:  cobra.NoArgs,
	}
}

// serve -
func serve(cfg *config.Config) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		var err error

		log.Printf(color, fmt.Sprintf(banner, Version))

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

		// Tracing
		if cfg.Tracer.Enabled {
			exporter, err := exporters.ExporterFactory(cfg.Tracer.Exporter, cfg.Tracer.Endpoint)
			if err != nil {
				l.Fatal(err)
			}

			shutdown, err := telemetry.NewTracer(exporter)
			if err != nil {
				l.Fatal(err)
			}

			defer func() {
				if err = shutdown(context.Background()); err != nil {
					l.Fatal(err)
				}
			}()
		}

		// cache
		var ch cache.Cache
		ch, err = ristretto.New()
		if err != nil {
			l.Fatal(err)
		}

		// Repositories
		relationTupleRepository := factories.RelationTupleFactory(db)
		err = relationTupleRepository.Migrate()
		if err != nil {
			l.Fatal(err)
		}

		entityConfigRepository := factories.EntityConfigFactory(db)
		err = entityConfigRepository.Migrate()
		if err != nil {
			l.Fatal(err)
		}

		// decorators
		relationTupleWithCircuitBreaker := decorators.NewRelationTupleWithCircuitBreaker(relationTupleRepository)
		entityConfigWithCircuitBreaker := decorators.NewEntityConfigWithCircuitBreaker(entityConfigRepository)

		// manager
		schemaManager := managers.NewEntityConfigManager(entityConfigWithCircuitBreaker, ch)

		// commands
		checkCommand := commands.NewCheckCommand(relationTupleWithCircuitBreaker, l)
		expandCommand := commands.NewExpandCommand(relationTupleWithCircuitBreaker, l)
		lookupQueryCommand := commands.NewLookupQueryCommand(relationTupleWithCircuitBreaker, l)
		schemaLookupCommand := commands.NewSchemaLookupCommand(l)

		// Services
		relationshipService := services.NewRelationshipService(relationTupleWithCircuitBreaker, schemaManager)
		permissionService := services.NewPermissionService(checkCommand, expandCommand, lookupQueryCommand, schemaManager)
		schemaService := services.NewSchemaService(schemaLookupCommand, schemaManager)

		container := servers.ServiceContainer{
			RelationshipService: relationshipService,
			PermissionService:   permissionService,
			SchemaService:       schemaService,
			SchemaManager:       schemaManager,
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

// RegisterServeFlags -
func RegisterServeFlags(cmd *cobra.Command, config *config.Config) {

	// GRPC Server
	cmd.Flags().StringVar(&config.Server.GRPC.Port, "grpc-port", config.Server.GRPC.Port, "set grpc port address")
	cmd.Flags().StringVar(&config.Server.GRPC.TLSConfig.KeyPath, "grpc-tls-config-key-path", config.Server.GRPC.TLSConfig.KeyPath, "set grpc tls config key path")
	cmd.Flags().StringVar(&config.Server.GRPC.TLSConfig.CertPath, "grpc-tls-config-cert-path", config.Server.GRPC.TLSConfig.CertPath, "set grpc tls config cert path")

	// HTTP Server
	cmd.Flags().BoolVar(&config.Server.HTTP.Enabled, "http-enabled", config.Server.HTTP.Enabled, "enable http server")
	cmd.Flags().StringVar(&config.Server.HTTP.Port, "http-port", config.Server.HTTP.Port, "set http port address")
	cmd.Flags().StringVar(&config.Server.HTTP.TLSConfig.KeyPath, "http-tls-config-key-path", config.Server.HTTP.TLSConfig.KeyPath, "set http tls config key path")
	cmd.Flags().StringVar(&config.Server.HTTP.TLSConfig.CertPath, "http-tls-config-cert-path", config.Server.HTTP.TLSConfig.CertPath, "set http tls config cert path")
	cmd.Flags().StringSliceVar(&config.Server.HTTP.CORSAllowedOrigins, "http-cors-allowed-origins", config.Server.HTTP.CORSAllowedOrigins, "set CORS allowed origins for http gateway")
	cmd.Flags().StringSliceVar(&config.Server.HTTP.CORSAllowedHeaders, "http-cors-allowed-headers", config.Server.HTTP.CORSAllowedHeaders, "set CORS allowed headers for http gateway")

	// LOG
	cmd.Flags().StringVar(&config.Log.Level, "log-level", config.Log.Level, "set log level")

	// AUTHN
	cmd.Flags().BoolVar(&config.Authn.Enabled, "authn-enabled", config.Authn.Enabled, "enable authentication")
	cmd.Flags().StringSliceVar(&config.Authn.Keys, "authn-preshared-keys", config.Authn.Keys, "set preshared keys")

	// TRACER
	cmd.Flags().BoolVar(&config.Tracer.Enabled, "tracer-enabled", config.Tracer.Enabled, "enable tracer")
	cmd.Flags().StringVar(&config.Tracer.Exporter, "tracer-exporter", config.Tracer.Exporter, "set tracer exporter")
	cmd.Flags().StringVar(&config.Tracer.Endpoint, "tracer-endpoint", config.Tracer.Endpoint, "set tracer endpoint")

	// DATABASE
	cmd.Flags().StringVar(&config.Database.Engine, "database-engine", config.Database.Engine, "set database engine")
	cmd.Flags().IntVar(&config.Database.PoolMax, "database-pool-max", config.Database.PoolMax, "set database pool max")
	cmd.Flags().StringVar(&config.Database.Database, "database-name", config.Database.Database, "set database name")
	cmd.Flags().StringVar(&config.Database.URI, "database-uri", config.Database.URI, "set database uri")
}
