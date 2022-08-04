// Package permify configures and runs application.
package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"

	"github.com/Permify/permify/internal/config"
	"github.com/Permify/permify/internal/consumers"
	PQConsumer "github.com/Permify/permify/internal/consumers/postgres"
	v1 "github.com/Permify/permify/internal/controllers/http/v1"
	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/internal/services"
	"github.com/Permify/permify/pkg/database"
	"github.com/Permify/permify/pkg/httpserver"
	"github.com/Permify/permify/pkg/logger"
	PQPublisher "github.com/Permify/permify/pkg/publisher/postgres"
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

	// Services
	schemaService := services.NewSchemaService(entityConfigRepository)
	relationshipService := services.NewRelationshipService(relationTupleRepository)
	permissionService := services.NewPermissionService(relationTupleRepository, schemaService)

	// CDC
	if cfg.Listen != nil && len(cfg.Listen.Tables) > 0 {
		var publisher *PQPublisher.Publisher
		publisher, err = PQPublisher.NewPublisher(context.Background(), cfg.Listen.URL, cfg.Listen.SlotName, cfg.Listen.OutputPlugin, cfg.Listen.Tables, l)
		go publisher.Start()

		notification := make(chan *PQPublisher.Notification)
		publisher.Subscribe(notification)
		defer publisher.Unsubscribe(notification)

		// consumer
		consumer := consumers.New(relationshipService, schemaService)
		pqConsumer := PQConsumer.New(consumer)
		go pqConsumer.Consume(context.Background(), notification)
	}

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
