package memory

import (
	"context"

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
func (r *Watch) Watch(_ context.Context, _, _ string) (<-chan *base.DataChanges, <-chan error) {
	tupleChanges := make(chan *base.DataChanges)
	errs := make(chan error)

	r.logger.Info(base.ErrorCode_ERROR_CODE_NOT_IMPLEMENTED.String())

	// Close the channels immediately
	close(tupleChanges)
	close(errs)

	return tupleChanges, errs
}
