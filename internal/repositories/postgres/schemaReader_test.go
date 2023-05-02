package postgres

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Masterminds/squirrel"
	"github.com/Permify/permify/internal/repositories"

	PQRepository "github.com/Permify/permify/pkg/database/postgres"
	"github.com/Permify/permify/pkg/logger"
	"github.com/rs/xid"
	"github.com/stretchr/testify/require"
)

var schemaExample = "entity user {}\\n\\nentity organization {\\n\\n    // relations\\n    relation admin @user\\n    relation member @user\\n\\n    // actions\\n    action create_repository = (admin or member)\\n    action delete = admin\\n}\\n\\nentity repository {\\n\\n    // relations\\n    relation owner @user @organization#member\\n    relation parent @organization\\n\\n    // actions\\n    action push = owner\\n    action read = (owner and (parent.admin and not parent.member))\\n    \\n    // parent.create_repository means user should be\\n    // organization admin or organization member\\n    action delete = (owner or (parent.create_repository))\\n}"

func TestHeadVersion_Test(t *testing.T) {
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
	writer := NewSchemaWriter(pg, log)
	reader := NewSchemaReader(pg, log)

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
	writer := NewSchemaWriter(pg, log)
	reader := NewSchemaReader(pg, log)

	ctx := context.Background()

	version := xid.New().String()
	schemas := map[string]repositories.SchemaDefinition{
		version: {TenantID: "1", EntityType: "user", SerializedDefinition: []byte(schemaExample), Version: version},
	}

	writeSchemas := []repositories.SchemaDefinition{
		{TenantID: "1", EntityType: "user", SerializedDefinition: []byte(schemaExample), Version: version},
	}

	query := "INSERT INTO schema_definitions \\(entity_type, serialized_definition, version, tenant_id\\) VALUES \\(\\$1,\\$2,\\$3,\\$4\\)$"
	mock.ExpectExec(query).
		WithArgs(schemas[version].EntityType, schemas[version].SerializedDefinition, schemas[version].Version, schemas[version].TenantID).
		WillReturnResult(sqlmock.NewResult(0, 2))

	err = writer.WriteSchema(ctx, writeSchemas)
	require.NoError(t, err)

	expectedQuery := "SELECT entity_type, serialized_definition, version FROM schema_definitions WHERE tenant_id = \\$1 AND version = \\$2"
	expectedRows := sqlmock.NewRows([]string{"entity_type", "serialized_definition", "version"}).
		AddRow("user", []byte(schemaExample), version)

	mock.ExpectQuery(expectedQuery).WithArgs("1", version).WillReturnRows(expectedRows)

	_, err = reader.ReadSchema(ctx, "1", version)
	require.NoError(t, err)
}

func TestReadSchemaDefinition_Test(t *testing.T) {
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
	writer := NewSchemaWriter(pg, log)
	reader := NewSchemaReader(pg, log)

	ctx := context.Background()

	version := xid.New().String()
	schemas := map[string]repositories.SchemaDefinition{
		version: {TenantID: "1", EntityType: "user", SerializedDefinition: []byte(schemaExample), Version: version},
	}

	writeSchemas := []repositories.SchemaDefinition{
		{TenantID: "1", EntityType: "user", SerializedDefinition: []byte(schemaExample), Version: version},
	}

	query := "INSERT INTO schema_definitions \\(entity_type, serialized_definition, version, tenant_id\\) VALUES \\(\\$1,\\$2,\\$3,\\$4\\)$"
	mock.ExpectExec(query).
		WithArgs(schemas[version].EntityType, schemas[version].SerializedDefinition, schemas[version].Version, schemas[version].TenantID).
		WillReturnResult(sqlmock.NewResult(0, 2))

	err = writer.WriteSchema(ctx, writeSchemas)
	require.NoError(t, err)

	expectedQuery := "SELECT entity_type, serialized_definition, version FROM schema_definitions WHERE entity_type = \\$1 AND tenant_id = \\$2 AND version = \\$3 LIMIT 1"
	expectedRows := sqlmock.NewRows([]string{"entity_type", "serialized_definition", "version"}).
		AddRow("user", schemaExample, version)

	mock.ExpectQuery(expectedQuery).WithArgs("user", "1", version).WillReturnRows(expectedRows)

	_, v, err := reader.ReadSchemaDefinition(ctx, "1", "user", version)
	require.NoError(t, err)
	require.Equal(t, version, v)
}
