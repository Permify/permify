//go:build !integration
// +build !integration

package postgres_test

import (
	"context"
	"github.com/Masterminds/squirrel"
	postgres2 "github.com/Permify/permify/internal/repositories/postgres"
	"github.com/Permify/permify/pkg/database/postgres"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Permify/permify/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListTenantWriter_Test(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	pg := &postgres.Postgres{
		DB:      db,
		Builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	// Create Fileds for Create Tenant
	log := logger.New("debug")
	writer := postgres2.NewTenantWriter(pg, log)
	ctx := context.Background()

	id := "2"
	name := "tenant_1"
	createdAt := time.Now()

	mock.ExpectQuery("INSERT INTO tenants \\(id, name\\) VALUES \\(\\$1,\\$2\\) RETURNING created_at").WithArgs(id, name).
		WillReturnRows(sqlmock.NewRows([]string{"created_at"}).AddRow(createdAt))

	tenant, err := writer.CreateTenant(ctx, id, name)
	require.NoError(t, err)
	assert.NotNil(t, tenant)
	assert.Equal(t, id, tenant.Id)
	assert.Equal(t, name, tenant.Name)
}
