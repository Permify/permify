package memory

import (
	"context"
	"errors"

	"github.com/hashicorp/go-memdb"

	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/internal/schema"
	db "github.com/Permify/permify/pkg/database/memory"
	"github.com/Permify/permify/pkg/logger"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// SchemaReader - Structure for Schema Reader
type SchemaReader struct {
	database *db.Memory
	// logger
	logger logger.Interface
}

// NewSchemaReader - Creates a new SchemaReader
func NewSchemaReader(database *db.Memory, logger logger.Interface) *SchemaReader {
	return &SchemaReader{
		database: database,
		logger:   logger,
	}
}

// ReadSchema - Reads a new schema from repository
func (r *SchemaReader) ReadSchema(ctx context.Context, tenantID string, version string) (sch *base.SchemaDefinition, err error) {
	txn := r.database.DB.Txn(false)
	defer txn.Abort()
	var it memdb.ResultIterator
	it, err = txn.Get(SchemaDefinitionsTable, "version", tenantID, version)
	if err != nil {
		return sch, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}

	var definitions []string
	for obj := it.Next(); obj != nil; obj = it.Next() {
		definitions = append(definitions, obj.(repositories.SchemaDefinition).Serialized())
	}

	sch, err = schema.NewSchemaFromStringDefinitions(true, definitions...)
	if err != nil {
		return nil, err
	}

	return sch, nil
}

// ReadSchemaDefinition - Reads a Schema Definition from repository
func (r *SchemaReader) ReadSchemaDefinition(ctx context.Context, tenantID, entityType, version string) (definition *base.EntityDefinition, v string, err error) {
	txn := r.database.DB.Txn(false)
	defer txn.Abort()
	var raw interface{}
	raw, err = txn.First(SchemaDefinitionsTable, "id", tenantID, entityType, version)
	if err != nil {
		return nil, "", errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}

	def, ok := raw.(repositories.SchemaDefinition)
	if ok {
		var sch *base.SchemaDefinition
		sch, err = schema.NewSchemaFromStringDefinitions(false, def.Serialized())
		if err != nil {
			return nil, "", err
		}
		definition, err = schema.GetEntityByName(sch, entityType)
		if err != nil {
			return nil, "", err
		}
		return definition, def.Version, err
	}

	return nil, "", errors.New(base.ErrorCode_ERROR_CODE_SCHEMA_NOT_FOUND.String())
}

// HeadVersion - Reads the latest version from the repository.
func (r *SchemaReader) HeadVersion(ctx context.Context, tenantID string) (string, error) {
	var err error
	txn := r.database.DB.Txn(false)
	defer txn.Abort()
	var raw interface{}
	raw, err = txn.Last(SchemaDefinitionsTable, "tenant", tenantID)
	if err != nil {
		return "", errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}
	if _, ok := raw.(repositories.SchemaDefinition); ok {
		return raw.(repositories.SchemaDefinition).Version, nil
	}
	return "", errors.New(base.ErrorCode_ERROR_CODE_SCHEMA_NOT_FOUND.String())
}
