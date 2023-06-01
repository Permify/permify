package utils

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/Masterminds/squirrel"

	"github.com/Permify/permify/pkg/logger"
)

// SnapshotQuery adds conditions to a SELECT query for checking transaction visibility based on created and expired transaction IDs.
// The query checks if transactions are visible in a snapshot associated with the provided value.
func SnapshotQuery(sl squirrel.SelectBuilder, value uint64) squirrel.SelectBuilder {
	// Convert the value to a string once to reduce redundant calls to fmt.Sprintf.
	valStr := fmt.Sprintf("'%v'::xid8", value)

	// Create a subquery for the snapshot associated with the provided value.
	snapshotQuery := fmt.Sprintf("(select snapshot from transactions where id = %s)", valStr)

	// Create an expression to check if a transaction with a specific created_tx_id is visible in the snapshot.
	visibilityExpr := squirrel.Expr(fmt.Sprintf("pg_visible_in_snapshot(created_tx_id, %s) = true", snapshotQuery))
	// Create an expression to check if the created_tx_id is equal to the provided value.
	createdExpr := squirrel.Expr(fmt.Sprintf("created_tx_id = %s", valStr))
	// Use OR condition for the created expressions.
	createdWhere := squirrel.Or{visibilityExpr, createdExpr}

	// Create an expression to check if a transaction with a specific expired_tx_id is not visible in the snapshot.
	expiredVisibilityExpr := squirrel.Expr(fmt.Sprintf("pg_visible_in_snapshot(expired_tx_id, %s) = false", snapshotQuery))
	// Create an expression to check if the expired_tx_id is equal to zero.
	expiredZeroExpr := squirrel.Expr("expired_tx_id = '0'::xid8")
	// Create an expression to check if the expired_tx_id is not equal to the provided value.
	expiredNotExpr := squirrel.Expr(fmt.Sprintf("expired_tx_id <> %s", valStr))
	// Use AND condition for the expired expressions, checking both visibility and non-equality with value.
	expiredWhere := squirrel.And{squirrel.Or{expiredVisibilityExpr, expiredZeroExpr}, expiredNotExpr}

	// Add the created and expired conditions to the SELECT query.
	return sl.Where(createdWhere).Where(expiredWhere)
}

// GarbageCollectQuery -
func GarbageCollectQuery(window time.Duration, tenantID string) squirrel.DeleteBuilder {
	return squirrel.Delete("relation_tuples").
		Where(squirrel.Expr(fmt.Sprintf("created_tx_id IN (SELECT id FROM transactions WHERE timestamp < '%v')", time.Now().Add(-window).Format(time.RFC3339)))).
		Where(squirrel.And{
			squirrel.Or{
				squirrel.Expr("expired_tx_id = '0'::xid8"),
				squirrel.Expr(fmt.Sprintf("expired_tx_id IN (SELECT id FROM transactions WHERE timestamp < '%v')", time.Now().Add(-window).Format(time.RFC3339))),
			},
			squirrel.Expr(fmt.Sprintf("tenant_id = '%v'", tenantID)),
		})
}

// Rollback - Rollbacks a transaction and logs the error
func Rollback(tx *sql.Tx, logger logger.Interface) {
	if err := tx.Rollback(); !errors.Is(err, sql.ErrTxDone) && err != nil {
		logger.Error("failed to rollback transaction", err)
	}
}
