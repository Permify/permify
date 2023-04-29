//go:build integration

package postgres_test

import (
	"context"
	"fmt"
	"github.com/Permify/permify/internal/config"
	"github.com/Permify/permify/internal/repositories"
	pg "github.com/Permify/permify/internal/repositories/postgres"
	"github.com/Permify/permify/pkg/database"
	PQDatabase "github.com/Permify/permify/pkg/database/postgres"
	"github.com/Permify/permify/pkg/logger"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"io"
	"log"
	"testing"
)

func TestSchemaWriter(t *testing.T) {
	// Start a PostgreSQL container
	ctx := context.Background()

	testcontainers.Logger = log.New(io.Discard, "", 0)

	req := testcontainers.ContainerRequest{
		Image:        "postgres:14-alpine",
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor:   wait.ForLog("database system is ready to accept connections"),
		Env:          map[string]string{"POSTGRES_USER": "postgres", "POSTGRES_PASSWORD": "postgres", "POSTGRES_DB": "permify"},
	}

	postgres, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	defer postgres.Terminate(ctx)

	// Get the host and port
	host, err := postgres.Host(ctx)
	require.NoError(t, err)

	port, err := postgres.MappedPort(ctx, "5432")
	require.NoError(t, err)

	l := logger.New("fatal")

	dbAddr := fmt.Sprintf("%s:%s", host, port.Port())
	databaseURL := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", "postgres", "postgres", dbAddr, "permify")

	cfg := config.Database{
		Engine:                "postgres",
		AutoMigrate:           true,
		URI:                   databaseURL,
		MaxOpenConnections:    20,
		MaxIdleConnections:    1,
		MaxConnectionLifetime: 300,
		MaxConnectionIdleTime: 60,
	}
	var db database.Database

	db, err = PQDatabase.New(cfg.URI,
		PQDatabase.MaxOpenConnections(cfg.MaxOpenConnections),
		PQDatabase.MaxIdleConnections(cfg.MaxIdleConnections),
		PQDatabase.MaxConnectionIdleTime(cfg.MaxConnectionIdleTime),
		PQDatabase.MaxConnectionLifeTime(cfg.MaxConnectionLifetime),
	)
	require.NoError(t, err)

	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	err = repositories.Migrate(cfg, l)
	require.NoError(t, err)

	// Create a TenantWriter instance
	schemaWriter := pg.NewSchemaWriter(db.(*PQDatabase.Postgres), l)

	schemas := []repositories.SchemaDefinition{
		{TenantID: "t1", EntityType: "entity1", SerializedDefinition: []byte("def1"), Version: "v1"},
		{TenantID: "t1", EntityType: "entity2", SerializedDefinition: []byte("def2"), Version: "v2"},
	}
	// Test the CreateTenant method
	err = schemaWriter.WriteSchema(ctx, schemas)
	require.NoError(t, err)
}
