//go:build !integration

package postgres_test

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Masterminds/squirrel"
	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/internal/repositories/postgres"
	PQRepository "github.com/Permify/permify/pkg/database/postgres"
	"github.com/Permify/permify/pkg/logger"
	"github.com/rs/xid"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestReadSchema_Test(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	pg := &PQRepository.Postgres{
		DB:      db,
		Builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	// Create Logger
	log := logger.New("debug")

	// Create SchemaWriter
	writer := postgres.NewSchemaWriter(pg, log)
	reader := postgres.NewSchemaReader(pg, log)

	ctx := context.Background()

	version := xid.New().String()
	schemas := map[string]repositories.SchemaDefinition{
		version: {TenantID: "1", EntityType: "entity1", SerializedDefinition: []byte("def1"), Version: version},
	}

	writeSchemas := []repositories.SchemaDefinition{
		{TenantID: "1", EntityType: "entity1", SerializedDefinition: []byte("def1"), Version: version},
	}

	query := "INSERT INTO schema_definitions \\(entity_type, serialized_definition, version, tenant_id\\) VALUES \\(\\$1,\\$2,\\$3,\\$4\\)$"
	mock.ExpectExec(query).
		WithArgs(schemas[version].EntityType, schemas[version].SerializedDefinition, schemas[version].Version, schemas[version].TenantID).
		WillReturnResult(sqlmock.NewResult(0, 2))

	err = writer.WriteSchema(ctx, writeSchemas)
	require.NoError(t, err)

	versionQuery := "SELECT version FROM schema_definitions WHERE tenant_id = \\$1 ORDER BY version DESC LIMIT 1$"
	rows := sqlmock.NewRows([]string{"version"}).AddRow(schemas[version].Version)
	mock.ExpectQuery(versionQuery).WithArgs(schemas[version].TenantID).WillReturnRows(rows)

	versionActual, err := reader.HeadVersion(ctx, "1")
	require.NoError(t, err)
	require.Equal(t, version, versionActual)
}
