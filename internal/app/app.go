// Package app configures and runs application.
package app

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	"github.com/lib/pq"

	"github.com/labstack/echo/v4"

	"github.com/Permify/permify/internal/config"
	"github.com/Permify/permify/internal/consumers"
	PQConsumer "github.com/Permify/permify/internal/consumers/postgres"
	v1 "github.com/Permify/permify/internal/controllers/http/v1"
	"github.com/Permify/permify/internal/migrations"
	PQRepository "github.com/Permify/permify/internal/repositories/postgres"
	"github.com/Permify/permify/internal/services"
	PQDatabase "github.com/Permify/permify/pkg/database/postgres"
	"github.com/Permify/permify/pkg/dsl/parser"
	"github.com/Permify/permify/pkg/dsl/schema"
	"github.com/Permify/permify/pkg/httpserver"
	"github.com/Permify/permify/pkg/logger"
	"github.com/Permify/permify/pkg/migration"
	"github.com/Permify/permify/pkg/notifier/postgres"
)

// Run creates objects via constructors.
func Run(cfg *config.Config) {
	var err error

	l := logger.New(cfg.Log.Level)

	// config statement
	var s []byte
	s, err = ioutil.ReadFile(schema.Path)
	if err != nil {
		s, err = ioutil.ReadFile(schema.DefaultPath)
		if err != nil {
			l.Fatal(fmt.Errorf("%w", err))
		}
	}

	statement := parser.TranslateToSchema(string(s))

	// migrations register
	var notifierMigrations *migration.Migration
	notifierMigrations, err = migrations.RegisterNotifierMigrations(statement.GetTableNames())
	if err != nil {
		l.Fatal(fmt.Errorf("%w", err))
	}

	var subscriberMigrations *migration.Migration
	subscriberMigrations, err = migrations.RegisterSubscriberMigrations()
	if err != nil {
		l.Fatal(fmt.Errorf("%w", err))
	}
	l.Info("migrations successfully registered")

	// Listen DB
	var listen *PQDatabase.Postgres
	listen, err = PQDatabase.New(cfg.Listen.URL, PQDatabase.MaxPoolSize(cfg.Listen.PoolMax))
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - postgres.New: %w", err))
	}
	defer listen.Close()

	err = listen.Migrate(*notifierMigrations)
	if err != nil {
		l.Fatal(fmt.Errorf("app - Migration - postgres.Migrate: %w", err))
	}
	l.Info("listen db successfully migrated")

	// Write DB
	var write *PQDatabase.Postgres
	write, err = PQDatabase.New(cfg.Write.URL, PQDatabase.MaxPoolSize(cfg.Write.PoolMax))
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - postgres.New: %w", err))
	}
	defer write.Close()

	err = write.Migrate(*subscriberMigrations)
	if err != nil {
		l.Fatal(fmt.Errorf("app - Migration - postgres.Migrate: %w", err))
	}
	l.Info("write db successfully migrated")

	// notifier
	var notifier *postgres.Notifier
	notifier, err = postgres.New(cfg.Listen.URL, l)
	if err != nil {
		l.Fatal(fmt.Errorf("app - notifier.New: %w", err))
	}

	err = notifier.Register(consumers.AuthorizationEvents)
	if err != nil {
		l.Fatal(fmt.Errorf("app - notifier.Register: %w", err))
	}
	err = notifier.Start()
	if err != nil {
		l.Fatal(fmt.Errorf("app - notifier.Start: %w", err))
	}
	l.Info("notifier service successfully started")

	notification := make(chan *pq.Notification)
	notifier.Subscribe(notification)
	defer notifier.Unsubscribe(notification)

	// repositories
	relationTupleRepository := PQRepository.NewRelationTupleRepository(write)
	// entityConfigRepository := PQRepository.NewEntityConfigRepository(write)

	// services
	// schemaService := services.NewSchemaService(entityConfigRepository, cache, write)
	relationshipService := services.NewRelationshipService(relationTupleRepository)
	permissionService := services.NewPermissionService(relationTupleRepository, statement)

	// consumer
	consumer := consumers.New(relationshipService, statement)
	pqConsumer := PQConsumer.New(consumer)
	pqConsumer.Consume(context.Background(), notification)

	// HTTP Server
	handler := echo.New()
	v1.NewRouter(handler, l, relationshipService, permissionService)
	httpServer := httpserver.New(handler, httpserver.Port(cfg.HTTP.Port))
	l.Info("http server successfully started")

	// Waiting signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		l.Info("app - Run - signal: " + s.String())
	case err = <-httpServer.Notify():
		l.Error(fmt.Errorf("app - Run - httpServer.Notify: %w", err))
	}

	// Shutdown
	err = httpServer.Shutdown()
	if err != nil {
		l.Error(fmt.Errorf("app - Run - httpServer.Shutdown: %w", err))
	}
}
