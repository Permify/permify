package utils

import (
	"database/sql"
	"errors"

	"github.com/Permify/permify/pkg/logger"
)

// Rollback - Rollbacks a transaction and logs the error
func Rollback(tx *sql.Tx, logger logger.Interface) {
	if err := tx.Rollback(); !errors.Is(err, sql.ErrTxDone) && err != nil {
		logger.Error("failed to rollback transaction", err)
	}
}
