package memory

import (
	"context"
	"errors"
	"sync"

	"github.com/Permify/permify/internal/storage"
	db "github.com/Permify/permify/pkg/database/memory"
	"github.com/Permify/permify/pkg/logger"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

var (
	headVersion map[string]string
	mu          sync.Mutex
)

func init() {
	headVersion = make(map[string]string)
}

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
func (w *SchemaWriter) WriteSchema(ctx context.Context, definitions []storage.SchemaDefinition) error {
	var err error
	txn := w.database.DB.Txn(true)
	defer txn.Abort()

	var tenantID string
	var version string

	for _, definition := range definitions {
		if err = txn.Insert(SchemaDefinitionsTable, definition); err != nil {
			return errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
		}
		tenantID = definition.TenantID
		version = definition.Version
	}
	txn.Commit()

	mu.Lock()
	headVersion[tenantID] = version
	mu.Unlock()

	return nil
}
