//go:build integration

package tests

import (
	"context"
	"testing"

	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/internal/storage/postgres"

	"github.com/stretchr/testify/require"

	"github.com/Permify/permify/pkg/database"
	PQDatabase "github.com/Permify/permify/pkg/database/postgres"
)

func TestSchemaWriter_Integration(t *testing.T) {
	ctx := context.Background()

	err := storage.Migrate(cfg)
	require.NoError(t, err)

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

	// Create a TenantWriter instance
	schemaWriter := postgres.NewSchemaWriter(db.(*PQDatabase.Postgres))

	schemas := []storage.SchemaDefinition{
		{TenantID: "t1", EntityType: "entity3", SerializedDefinition: []byte("def2"), Version: "v3"},
	}
	// Test the CreateTenant method
	err = schemaWriter.WriteSchema(ctx, schemas)
	require.NoError(t, err)
}
