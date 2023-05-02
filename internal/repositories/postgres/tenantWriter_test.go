package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
	PQRepository "github.com/Permify/permify/pkg/database/postgres"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Permify/permify/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTenantWriter_CreateTenant(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	pg := &PQRepository.Postgres{
		DB:      db,
		Builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	// Create Logger
	log := logger.New("debug")

	// Create TenantWriter
	writer := NewTenantWriter(pg, log)

	ctx := context.Background()

	// Create Fields for CreateTenant
	id := "2"
	name := "tenant_1"
	createdAt := time.Now()

	// SQL sorgusunu bekleyen mock.ExpectQuery oluştur
	mock.ExpectQuery("INSERT INTO tenants \\(id, name\\) VALUES \\(\\$1,\\$2\\) RETURNING created_at").WithArgs(id, name).
		WillReturnRows(sqlmock.NewRows([]string{"created_at"}).AddRow(createdAt))

	tenant, err := writer.CreateTenant(ctx, id, name)
	require.NoError(t, err)
	assert.NotNil(t, tenant)
	assert.Equal(t, id, tenant.Id)
	assert.Equal(t, name, tenant.Name)
}

func TestTenantWriter_DeleteTenant(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	pg := &PQRepository.Postgres{
		DB:      db,
		Builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	// Create Logger
	log := logger.New("debug")

	// Create TenantWriter
	writer := NewTenantWriter(pg, log)

	ctx := context.Background()

	// Create Fields for CreateTenant
	id := "2"
	name := "tenant_1"
	createdAt := time.Now()

	// SQL sorgusunu bekleyen mock.ExpectQuery oluştur
	mock.ExpectQuery("INSERT INTO tenants \\(id, name\\) VALUES \\(\\$1,\\$2\\) RETURNING created_at").WithArgs(id, name).
		WillReturnRows(sqlmock.NewRows([]string{"created_at"}).AddRow(createdAt))

	tenant, err := writer.CreateTenant(ctx, id, name)

	mock.ExpectQuery("DELETE FROM tenants WHERE id = \\$1 RETURNING name, created_at").WithArgs(tenant.Id).
		WillReturnRows(sqlmock.NewRows([]string{"name", "created_at"}).AddRow(tenant.Name, tenant.CreatedAt.AsTime()))

	deletedTenant, err := writer.DeleteTenant(ctx, tenant.Id)
	require.NoError(t, err)
	assert.NotNil(t, deletedTenant)
	assert.Equal(t, id, deletedTenant.Id)
}
