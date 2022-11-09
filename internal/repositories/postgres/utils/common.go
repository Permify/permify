package utils

import (
	"fmt"

	"github.com/Masterminds/squirrel"
)

// SnapshotQuery -
func SnapshotQuery(sl squirrel.SelectBuilder, revision uint64) squirrel.SelectBuilder {
	return sl.Where(squirrel.Or{
		squirrel.Expr(fmt.Sprintf("pg_visible_in_snapshot(created_tx_id, (select snapshot from transactions where id = '%v'::xid8)) = true", revision)),
		squirrel.Expr(fmt.Sprintf("created_tx_id = '%v'::xid8", revision)),
	}).Where(squirrel.And{
		squirrel.Expr(fmt.Sprintf("pg_visible_in_snapshot(expired_tx_id, (select snapshot from transactions where id = '%v'::xid8)) OR (expired_tx_id = '0'::xid8) = false", revision)),
		squirrel.Expr(fmt.Sprintf("expired_tx_id <> '%v'::xid8", revision)),
	})
}

// NewTransactionQuery -
func NewTransactionQuery() string {
	return `INSERT INTO transactions DEFAULT VALUES RETURNING id`
}
