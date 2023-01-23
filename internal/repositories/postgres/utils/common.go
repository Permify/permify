package utils

import (
	"database/sql"
	"fmt"

	"github.com/pkg/errors"

	"github.com/Masterminds/squirrel"

	"github.com/Permify/permify/pkg/logger"
)

// SnapshotQuery -
func SnapshotQuery(sl squirrel.SelectBuilder, revision uint64) squirrel.SelectBuilder {
	return sl.Where(squirrel.Or{
		squirrel.Expr(fmt.Sprintf("pg_visible_in_snapshot(created_tx_id, (select snapshot from transactions where id = '%v'::xid8)) = true", revision)),
		squirrel.Expr(fmt.Sprintf("created_tx_id = '%v'::xid8", revision)),
	}).Where(squirrel.And{
		squirrel.Or{
			squirrel.Expr(fmt.Sprintf("pg_visible_in_snapshot(expired_tx_id, (select snapshot from transactions where id = '%v'::xid8)) = false", revision)),
			squirrel.Expr("expired_tx_id = '0'::xid8"),
		},
		squirrel.Expr(fmt.Sprintf("expired_tx_id <> '%v'::xid8", revision)),
	})
}

// Rollback - Rollbacks a transaction and logs the error
func Rollback(tx *sql.Tx, logger logger.Interface) {
	if err := tx.Rollback(); !errors.Is(err, sql.ErrTxDone) && err != nil {
		logger.Error("failed to rollback transaction", err)
	}
}
