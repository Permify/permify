//go:build integration

package postgres

import (
	"context"
	"fmt"
	"github.com/Permify/permify/internal/config"
	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/pkg/database"
	PQDatabase "github.com/Permify/permify/pkg/database/postgres"
	"github.com/Permify/permify/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"log"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestTenantReader_Integration(t *testing.T) {
	// Start a PostgreSQL container
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "postgres:14-alpine",
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor:   wait.ForLog("database system is ready to accept connections"),
		Env:          map[string]string{"POSTGRES_USER": "postgres", "POSTGRES_PASSWORD": "postgres", "POSTGRES_DB": "permify"},
	}

	testcontainers.Logger = log.New(io.Discard, "", 0)

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

	// Create a Tenant instances
	tenantWriter := NewTenantWriter(db.(*PQDatabase.Postgres), l)
	tenantReader := NewTenantReader(db.(*PQDatabase.Postgres), l)

	// Test the CreateTenant method
	createdTenant, err := tenantWriter.CreateTenant(ctx, "2", "Test Tenant")
	require.NoError(t, err)
	assert.Equal(t, "2", createdTenant.Id)
	assert.Equal(t, "Test Tenant", createdTenant.Name)

	pagination := database.NewPagination()

	// Test the DeleteTenant method
	listTenant, _, err := tenantReader.ListTenants(ctx, pagination)

	require.NoError(t, err)
	assert.Equal(t, "t1", listTenant[1].Id)
	assert.Equal(t, "example tenant", listTenant[1].Name)
	assert.Equal(t, "2", listTenant[0].Id)
	assert.Equal(t, "Test Tenant", listTenant[0].Name)
}
