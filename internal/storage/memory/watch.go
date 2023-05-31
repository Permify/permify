package memory

import (
	"context"
	"errors"

	db "github.com/Permify/permify/pkg/database/memory"
	"github.com/Permify/permify/pkg/logger"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// Watch - Watches for changes in the repository.
type Watch struct {
	database *db.Memory
	// logger
	logger logger.Interface
}

// NewWatcher - Creates a new Watcher
func NewWatcher(database *db.Memory, logger logger.Interface) *Watch {
	return &Watch{
		database: database,
		logger:   logger,
	}
}

// Watch - Watches for changes in the repository.
func (r *Watch) Watch(ctx context.Context, tenantID string, snap string) (<-chan *base.TupleChanges, <-chan error) {
	errs := make(chan error, 1)
	errs <- errors.New(base.ErrorCode_ERROR_CODE_NOT_IMPLEMENTED.String())
	return nil, errs
}
