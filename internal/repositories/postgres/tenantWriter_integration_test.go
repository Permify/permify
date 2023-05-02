//go:build integration

package postgres

import (
	"context"
	"testing"

	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/pkg/database"
	PQDatabase "github.com/Permify/permify/pkg/database/postgres"
	"github.com/Permify/permify/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTenantWriter(t *testing.T) {
	ctx := context.Background()

	l := logger.New("fatal")

	err := repositories.Migrate(cfg, l)
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
	tenantWriter := NewTenantWriter(db.(*PQDatabase.Postgres), l)

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
