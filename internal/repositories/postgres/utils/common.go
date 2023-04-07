package utils

import (
	"database/sql"
	"fmt"
	"github.com/pkg/errors"
	"time"

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

// GarbageCollectQuery -
func GarbageCollectQuery(window time.Duration, tenantID string) squirrel.DeleteBuilder {
	expiredTransactions := squirrel.
		Select("id").
		From("transactions").
		Where(squirrel.Expr("timestamp < ?", time.Now().Add(-window)))
	expiredRows := squirrel.
		Select("*").
		From("relation_tuples rt").
		JoinClause(fmt.Sprintf("JOIN (%s) et ON rt.created_tx_id = et.id", expiredTransactions)).
		Where(squirrel.Or{
			squirrel.Expr("rt.expired_tx_id = '0'::xid8"),
			squirrel.Expr("rt.expired_tx_id IN ?", expiredTransactions),
		})
	deleteQuery := squirrel.
		Delete("relation_tuples").
		Where(squirrel.Expr("id IN ?", expiredRows), squirrel.Eq{"tenant_id": tenantID})
	return deleteQuery
}

// Rollback - Rollbacks a transaction and logs the error
func Rollback(tx *sql.Tx, logger logger.Interface) {
	if err := tx.Rollback(); !errors.Is(err, sql.ErrTxDone) && err != nil {
		logger.Error("failed to rollback transaction", err)
	}
}
