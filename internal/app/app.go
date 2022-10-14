package app

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"

	"github.com/Permify/permify/internal/commands"
	"github.com/Permify/permify/internal/config"
	"github.com/Permify/permify/internal/factories"
	"github.com/Permify/permify/internal/managers"
	"github.com/Permify/permify/internal/repositories/decorators"
	"github.com/Permify/permify/internal/servers"
	"github.com/Permify/permify/internal/services"
	"github.com/Permify/permify/pkg/cache"
	"github.com/Permify/permify/pkg/cache/ristretto"
	"github.com/Permify/permify/pkg/database"
	"github.com/Permify/permify/pkg/logger"
	"github.com/Permify/permify/pkg/telemetry"
	"github.com/Permify/permify/pkg/telemetry/exporters"
)

const (
	// Version of Permify
	Version = "v0.0.0-alpha7"
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

// Start creates objects via constructors.
func Start(cfg *config.Config) {
	log.Printf(color, fmt.Sprintf(banner, Version))

	var err error
	l := logger.New(cfg.Log.Level)

	l.Info("🚀 starting permify service...")

	var db database.Database
	db, err = factories.DatabaseFactory(cfg.Database)
	if err != nil {
		l.Fatal(err)
	}
	defer db.Close()

	// Tracing
	if cfg.Tracer != nil && cfg.Tracer.Enabled {
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
		l.Fatal(fmt.Errorf("permify - Run - cache.Factory: %w", err))
	}

	// Repositories
	relationTupleRepository := factories.RelationTupleFactory(db)
	err = relationTupleRepository.Migrate()
	if err != nil {
		l.Fatal(fmt.Errorf("permify - Run - relationTupleRepository.Migrate: %w", err))
	}

	entityConfigRepository := factories.EntityConfigFactory(db)
	err = entityConfigRepository.Migrate()
	if err != nil {
		l.Fatal(fmt.Errorf("permify - Run - entityConfigRepository.Migrate: %w", err))
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

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return container.Run(context.Background(), &cfg.Server, cfg.Authn, l)
	})

	if err = g.Wait(); err != nil {
		l.Error(err)
	}

	l.Info("Server is shutting down")
}
