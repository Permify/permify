package utils

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/pkg/errors"

	"github.com/Masterminds/squirrel"

	base "github.com/Permify/permify/pkg/pb/base/v1"
)

const (
	BulkEntityFilterTemplate = `
    WITH entities AS (
        (SELECT id, entity_id, entity_type, tenant_id, created_tx_id, expired_tx_id FROM relation_tuples)
        UNION ALL
        (SELECT id, entity_id, entity_type, tenant_id, created_tx_id, expired_tx_id FROM attributes)
    ), filtered_entities AS (
        SELECT DISTINCT ON (entity_id) id, entity_id
        FROM entities
        WHERE tenant_id = '%s'
        AND entity_type = '%s'
        AND %s
        AND %s
    )
    SELECT id, entity_id
    FROM filtered_entities`
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

// snapshotQuery function generates two strings representing conditions to be applied in a SQL query to filter data based on visibility of transactions.
func snapshotQuery(value uint64) (string, string) {
	// Convert the provided value into a string format suitable for our SQL query, formatted as a transaction ID.
	valStr := fmt.Sprintf("'%v'::xid8", value)

	// Create a subquery that fetches the snapshot associated with the transaction ID.
	snapshotQ := fmt.Sprintf("(SELECT snapshot FROM transactions WHERE id = %s)", valStr)

	// Create an expression that checks whether a transaction (represented by 'created_tx_id') is visible in the snapshot.
	visibilityExpr := fmt.Sprintf("pg_visible_in_snapshot(created_tx_id, %s) = true", snapshotQ)
	// Create an expression that checks if the 'created_tx_id' is the same as our transaction ID.
	createdExpr := fmt.Sprintf("created_tx_id = %s", valStr)
	// Combine these expressions to form a condition. A row will satisfy this condition if either of the expressions are true.
	createdWhere := fmt.Sprintf("(%s OR %s)", visibilityExpr, createdExpr)

	// Create an expression that checks whether a transaction (represented by 'expired_tx_id') is not visible in the snapshot.
	expiredVisibilityExpr := fmt.Sprintf("pg_visible_in_snapshot(expired_tx_id, %s) = false", snapshotQ)
	// Create an expression that checks if the 'expired_tx_id' is zero. This handles cases where the transaction hasn't expired.
	expiredZeroExpr := "expired_tx_id = '0'::xid8"
	// Create an expression that checks if the 'expired_tx_id' is not the same as our transaction ID.
	expiredNotExpr := fmt.Sprintf("expired_tx_id <> %s", valStr)
	// Combine these expressions to form a condition. A row will satisfy this condition if the first set of expressions are true (either the transaction hasn't expired, or if it has, it's not visible in the snapshot) and the second expression is also true (the 'expired_tx_id' is not the same as our transaction ID).
	expiredWhere := fmt.Sprintf("(%s AND %s)", fmt.Sprintf("(%s OR %s)", expiredVisibilityExpr, expiredZeroExpr), expiredNotExpr)

	// Return the conditions for both 'created' and 'expired' transactions. These can be used in a WHERE clause of a SQL query to filter results.
	return createdWhere, expiredWhere
}

// BulkEntityFilterQuery -
func BulkEntityFilterQuery(tenantID, entityType string, snap uint64) string {
	createdWhere, expiredWhere := snapshotQuery(snap)
	return fmt.Sprintf(BulkEntityFilterTemplate, tenantID, entityType, createdWhere, expiredWhere)
}

// GenerateGCQuery generates a Squirrel DELETE query builder for garbage collection.
// It constructs a query to delete expired records from the specified table
// based on the provided value, which represents a transaction ID.
func GenerateGCQuery(table string, value uint64) squirrel.DeleteBuilder {
	// Convert the provided value into a string format suitable for our SQL query, formatted as a transaction ID.
	valStr := fmt.Sprintf("'%v'::xid8", value)

	// Create a Squirrel DELETE builder for the specified table.
	deleteBuilder := squirrel.Delete(table)

	// Create an expression to check if 'expired_tx_id' is not equal to '0' (not expired).
	expiredZeroExpr := squirrel.Expr("expired_tx_id <> '0'::xid8")

	// Create an expression to check if 'expired_tx_id' is less than the provided value (before the cutoff).
	beforeExpr := squirrel.Expr(fmt.Sprintf("expired_tx_id < %s", valStr))

	// Add the WHERE clauses to the DELETE query builder to filter and delete expired data.
	return deleteBuilder.Where(expiredZeroExpr).Where(beforeExpr)
}

// HandleError records an error in the given span, logs the error, and returns a standardized error.
// This function is used for consistent error handling across different parts of the application.
func HandleError(span trace.Span, err error, errorCode base.ErrorCode) error {
	// Record the error on the span
	span.RecordError(err)

	// Set the status of the span
	span.SetStatus(codes.Error, err.Error())

	// Check if the error is context-related
	if IsContextRelatedError(err) {
		// Use debug level logging for context-related errors
		slog.Debug("Context-related error encountered", slog.Any("error", err), slog.Any("errorCode", errorCode))
	} else {
		// Use error level logging for all other errors
		slog.Error("Error encountered", slog.Any("error", err), slog.Any("errorCode", errorCode))
	}

	// Return a new standardized error with the provided error code
	return errors.New(errorCode.String())
}

// IsContextRelatedError checks if the error is due to context cancellation, deadline exceedance, or closed connection
func IsContextRelatedError(err error) bool {
	return errors.Is(err, context.Canceled) ||
		errors.Is(err, context.DeadlineExceeded) ||
		err.Error() == "conn closed"
}
