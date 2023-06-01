//go:build integration

package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/internal/storage/postgres"
	"github.com/Permify/permify/pkg/database"
	PQDatabase "github.com/Permify/permify/pkg/database/postgres"
	"github.com/Permify/permify/pkg/logger"
)

func TestTenantWriter(t *testing.T) {
	ctx := context.Background()

	l := logger.New("fatal")

	err := storage.Migrate(cfg, l)
	require.NoError(t, err)

	var db database.Database
	db, err = PQDatabase.New(cfg.URI,
		PQDatabase.MaxOpenConnections(cfg.MaxOpenConnections),
		PQDatabase.MaxIdleConnections(cfg.MaxIdleConnections),
		PQDatabase.MaxConnectionIdleTime(cfg.MaxConnectionIdleTime),
		PQDatabase.MaxConnectionLifeTime(cfg.MaxConnectionLifetime),
	)
	require.NoError(t, err)

	defer db.Close()

	// Create a TenantWriter instance
	tenantWriter := postgres.NewTenantWriter(db.(*PQDatabase.Postgres), l)

	// Test the CreateTenant method
	createdTenant, err := tenantWriter.CreateTenant(ctx, "4", "Test Tenant")
	require.NoError(t, err)
	assert.Equal(t, "4", createdTenant.Id)
	assert.Equal(t, "Test Tenant", createdTenant.Name)

	// Test the DeleteTenant method
	deletedTenant, err := tenantWriter.DeleteTenant(ctx, "4")
	require.NoError(t, err)
	assert.Equal(t, "4", deletedTenant.Id)
	assert.Equal(t, "Test Tenant", deletedTenant.Name)
}
