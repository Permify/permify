package memory

import (
	"context"
	"errors"

	"github.com/Permify/permify/internal/repositories"
	db "github.com/Permify/permify/pkg/database/memory"
	"github.com/Permify/permify/pkg/logger"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// SchemaWriter - Structure for Schema Writer
type SchemaWriter struct {
	database *db.Memory
	// logger
	logger logger.Interface
}

// NewSchemaWriter creates a new SchemaWriter
func NewSchemaWriter(database *db.Memory, logger logger.Interface) *SchemaWriter {
	return &SchemaWriter{
		database: database,
		logger:   logger,
	}
}

// WriteSchema - Write Schema to repository
func (r *SchemaWriter) WriteSchema(ctx context.Context, definitions []repositories.SchemaDefinition) error {
	var err error
	txn := r.database.DB.Txn(true)
	defer txn.Abort()
	for _, definition := range definitions {
		if err = txn.Insert(SchemaDefinitionTable, definition); err != nil {
			return errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
		}
	}
	txn.Commit()
	return nil
}
