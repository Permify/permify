package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/labstack/echo/v4/middleware"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"

	"github.com/Permify/permify/internal/authn"
	"github.com/Permify/permify/internal/commands"
	"github.com/Permify/permify/internal/config"
	"github.com/Permify/permify/internal/factories"
	"github.com/Permify/permify/internal/managers"
	"github.com/Permify/permify/internal/repositories/decorators"
	grpcServerV1 "github.com/Permify/permify/internal/servers/grpc/v1"
	httpV1 "github.com/Permify/permify/internal/servers/http/v1"
	"github.com/Permify/permify/internal/services"
	"github.com/Permify/permify/pkg/cache"
	"github.com/Permify/permify/pkg/cache/ristretto"
	"github.com/Permify/permify/pkg/database"
	"github.com/Permify/permify/pkg/grpcserver"
	"github.com/Permify/permify/pkg/httpserver"
	"github.com/Permify/permify/pkg/logger"
	grpcV1 "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/telemetry"
	"github.com/Permify/permify/pkg/telemetry/exporters"
)

// Run creates objects via constructors.
func Run(cfg *config.Config) {
	var err error

	l := logger.New(cfg.Log.Level)

	var db database.Database
	db, err = factories.DatabaseFactory(cfg.Write)
	if err != nil {
		l.Fatal(fmt.Errorf("permify - Run - DBFactory: %w", err))
	}
	defer db.Close()

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

	// HTTP Server
	handler := echo.New()
	handler.Use(otelecho.Middleware("http.server"))

	if cfg.Authn != nil && !cfg.Authn.Disabled {
		if len(cfg.Authn.Keys) > 0 {
			authenticator := authn.NewKeyAuthn(cfg.Authn.Keys...)
			handler.Use(middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
				Validator: authenticator.Validator(),
			}))
		}
	}

	// HTTP SERVER
	httpV1.NewServer(handler, l, relationshipService, permissionService, schemaService, schemaManager)
	httpServer := httpserver.New(handler, httpserver.Port(cfg.HTTP.Port))
	httpServer.Run()
	l.Info(fmt.Sprintf("🚀 http server successfully started: %s", cfg.HTTP.Port))

	// GRPC SERVER
	grpcServer := grpcserver.New(grpcserver.Port(cfg.GRPC.Port))
	grpcV1.RegisterPermissionAPIServer(grpcServer.Server, grpcServerV1.NewPermissionServer(permissionService, l))
	grpcV1.RegisterSchemaAPIServer(grpcServer.Server, grpcServerV1.NewSchemaServer(schemaManager, schemaService, l))
	grpcV1.RegisterRelationshipAPIServer(grpcServer.Server, grpcServerV1.NewRelationshipServer(relationshipService, l))
	grpcServer.Run()
	l.Info(fmt.Sprintf("🚀 grpc server successfully started: %s", cfg.GRPC.Port))

	// Waiting signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		l.Info(s.String())
	case err = <-httpServer.Notify():
		l.Error(err.Error())
	case err = <-grpcServer.Notify():
		l.Error(err.Error())
	}

	// Shutdown
	err = httpServer.Shutdown()
	if err != nil {
		l.Error(err.Error())
	}

	err = grpcServer.Shutdown()
	if err != nil {
		l.Error(err.Error())
	}
}
