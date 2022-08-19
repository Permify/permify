package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/dgraph-io/ristretto"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"

	"github.com/Permify/permify/internal/commands"
	"github.com/Permify/permify/internal/config"
	v1 "github.com/Permify/permify/internal/controllers/http/v1"
	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/internal/repositories/decorators"
	"github.com/Permify/permify/internal/repositories/proxies"
	"github.com/Permify/permify/internal/services"
	"github.com/Permify/permify/pkg/database"
	"github.com/Permify/permify/pkg/httpserver"
	"github.com/Permify/permify/pkg/logger"
	"github.com/Permify/permify/pkg/telemetry"
	"github.com/Permify/permify/pkg/telemetry/exporters"
)

// Run creates objects via constructors.
func Run(cfg *config.Config) {
	var err error

	l := logger.New(cfg.Log.Level)

	var DB database.Database
	DB, err = database.DBFactory(cfg.Write)
	if err != nil {
		l.Fatal(fmt.Errorf("permify - Run - DBFactory: %w", err))
	}
	defer DB.Close()

	// Tracing
	if cfg.Tracer != nil && !cfg.Tracer.Disabled {
		exporter, err := exporters.ExporterFactory(cfg.Tracer.Exporter, cfg.Tracer.Endpoint)
		if err != nil {
			l.Fatal(fmt.Errorf("permify - Run - ExporterFactory: %w", err))
		}

		shutdown, err := telemetry.NewTracer(exporter)
		if err != nil {
			l.Fatal(fmt.Errorf("permify - Run - NewTracer: %w", err))
		}

		defer func() {
			if err = shutdown(context.Background()); err != nil {
				l.Fatal("failed to shutdown TracerProvider: %w", err)
			}
		}()
	}

	// cache
	var cache *ristretto.Cache
	cache, err = ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     1 << 30,
		BufferItems: 64,
	})
	if err != nil {
		l.Fatal(fmt.Errorf("permify - Run - ristretto.NewCache: %w", err))
	}

	// Repositories
	relationTupleRepository := repositories.RelationTupleFactory(DB)
	err = relationTupleRepository.Migrate()
	if err != nil {
		l.Fatal(fmt.Errorf("permify - Run - relationTupleRepository.Migrate: %w", err))
	}

	entityConfigRepository := repositories.EntityConfigFactory(DB)
	err = entityConfigRepository.Migrate()
	if err != nil {
		l.Fatal(fmt.Errorf("permify - Run - entityConfigRepository.Migrate: %w", err))
	}

	// decorators
	relationTupleWithCircuitBreaker := decorators.NewRelationTupleWithCircuitBreaker(relationTupleRepository)
	entityConfigWithCircuitBreaker := decorators.NewEntityConfigWithCircuitBreaker(entityConfigRepository)

	// proxies
	entityConfigCacheProxy := proxies.NewEntityConfigCacheProxy(entityConfigWithCircuitBreaker, cache)

	// commands
	checkCommand := commands.NewCheckCommand(relationTupleWithCircuitBreaker, l)
	expandCommand := commands.NewExpandCommand(relationTupleWithCircuitBreaker, l)

	// Services
	schemaService := services.NewSchemaService(entityConfigCacheProxy)
	relationshipService := services.NewRelationshipService(relationTupleWithCircuitBreaker)
	permissionService := services.NewPermissionService(checkCommand, expandCommand, entityConfigCacheProxy)

	// HTTP Server
	handler := echo.New()
	handler.Use(otelecho.Middleware("http.server"))

	v1.NewRouter(handler, l, relationshipService, permissionService, schemaService)
	httpServer := httpserver.New(handler, httpserver.Port(cfg.HTTP.Port))
	l.Info(fmt.Sprintf("http server successfully started: %s", cfg.HTTP.Port))

	// Waiting signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		l.Info("permify - Run - signal: " + s.String())
	case err = <-httpServer.Notify():
		l.Error(fmt.Errorf("permify - Run - httpServer.Notify: %w", err))
	}

	// Shutdown
	err = httpServer.Shutdown()
	if err != nil {
		l.Error(fmt.Errorf("permify - Run - httpServer.Shutdown: %w", err))
	}
}
