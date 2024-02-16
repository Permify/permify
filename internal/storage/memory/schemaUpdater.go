package memory

import (
	"context"
	"errors"

	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/internal/storage/memory/constants"
	db "github.com/Permify/permify/pkg/database/memory"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/hashicorp/go-memdb"
)

// SchemaUpdater - Structure for Schema Updater
type SchemaUpdater struct {
	database *db.Memory
}

// NewSchemaUpdater - Creates a new SchemaUpdater
func NewSchemaUpdater(database *db.Memory) *SchemaUpdater {
	return &SchemaUpdater{
		database: database,
	}
}

// UpdateSchema - Update entity config in the database
func (u *SchemaUpdater) UpdateSchema(ctx context.Context, tenantID, version string, definitions map[string]map[string][]string) (schema []string, err error) {
	txn := u.database.DB.Txn(false)
	defer txn.Abort()
	var it memdb.ResultIterator
	it, err = txn.Get(constants.SchemaDefinitionsTable, "version", tenantID, version)
	if err != nil {
		return schema, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}

	var storedDefinitions []string
	for obj := it.Next(); obj != nil; obj = it.Next() {
		storedDefinitions = append(storedDefinitions, obj.(storage.SchemaDefinition).Serialized())
	}
	return schema, nil
}
