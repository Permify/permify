package tests

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Masterminds/squirrel"
	"github.com/stretchr/testify/require"

	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/internal/storage/postgres"
	PQRepository "github.com/Permify/permify/pkg/database/postgres"
	"github.com/Permify/permify/pkg/logger"
)

func TestWriteSchema_Test(t *testing.T) {
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

	ctx := context.Background()

	schemas := []storage.SchemaDefinition{
		{TenantID: "1", Name: "entity1", SerializedDefinition: []byte("def1"), Version: "v1"},
		{TenantID: "2", Name: "entity2", SerializedDefinition: []byte("def2"), Version: "v2"},
	}

	query := "INSERT INTO schema_definitions \\(entity_type, serialized_definition, version, tenant_id\\) VALUES \\(\\$1,\\$2,\\$3,\\$4\\),\\(\\$5,\\$6,\\$7,\\$8\\)$"
	mock.ExpectExec(query).
		WithArgs(schemas[0].Name, schemas[0].SerializedDefinition, schemas[0].Version, schemas[0].TenantID,
			schemas[1].Name, schemas[1].SerializedDefinition, schemas[1].Version, schemas[1].TenantID).
		WillReturnResult(sqlmock.NewResult(0, 2))

	err = writer.WriteSchema(ctx, schemas)
	require.NoError(t, err)
}
