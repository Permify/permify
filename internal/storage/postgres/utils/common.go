package utils

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"strings"
	"time"

	"go.opentelemetry.io/otel/codes"

	"go.opentelemetry.io/otel/trace"

	"github.com/Masterminds/squirrel"

	base "github.com/Permify/permify/pkg/pb/base/v1"
)

const (
	TransactionTemplate       = `INSERT INTO transactions (tenant_id) VALUES ($1) RETURNING id`
	InsertTenantTemplate      = `INSERT INTO tenants (id, name) VALUES ($1, $2) RETURNING created_at`
	DeleteTenantTemplate      = `DELETE FROM tenants WHERE id = $1 RETURNING name, created_at`
	DeleteAllByTenantTemplate = `DELETE FROM %s WHERE tenant_id = $1`
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
func HandleError(ctx context.Context, span trace.Span, err error, errorCode base.ErrorCode) error {
	// Check if the error is context-related
	if IsContextRelatedError(ctx, err) {
		slog.DebugContext(ctx, "A context-related error occurred",
			slog.String("error", err.Error()))
		return errors.New(base.ErrorCode_ERROR_CODE_CANCELLED.String())
	}

	// Check if the error is serialization-related
	if IsSerializationRelatedError(err) {
		slog.DebugContext(ctx, "A serialization-related error occurred",
			slog.String("error", err.Error()))
		return errors.New(base.ErrorCode_ERROR_CODE_SERIALIZATION.String())
	}

	// For all other types of errors, log them at the error level and record them in the span
	slog.ErrorContext(ctx, "An operational error occurred",
		slog.Any("error", err))
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())

	// Return a new error with the standard error code provided
	return errors.New(errorCode.String())
}

// IsContextRelatedError checks if the error is due to context cancellation, deadline exceedance, or closed connection
func IsContextRelatedError(ctx context.Context, err error) bool {
	if errors.Is(ctx.Err(), context.Canceled) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return true
	}
	if errors.Is(err, context.Canceled) ||
		errors.Is(err, context.DeadlineExceeded) ||
		strings.Contains(err.Error(), "conn closed") {
		return true
	}
	return false
}

// IsSerializationRelatedError checks if the error is a serialization failure, typically in database transactions.
func IsSerializationRelatedError(err error) bool {
	if strings.Contains(err.Error(), "could not serialize") ||
		strings.Contains(err.Error(), "duplicate key value") {
		return true
	}
	return false
}

// WaitWithBackoff implements an exponential backoff strategy with jitter for retries.
// It waits for a calculated duration or until the context is cancelled, whichever comes first.
func WaitWithBackoff(ctx context.Context, tenantID string, retries int) {
	// Calculate the base backoff
	backoff := time.Duration(math.Min(float64(20*time.Millisecond)*math.Pow(2, float64(retries)), float64(1*time.Second)))

	// Generate jitter using crypto/rand
	jitter := time.Duration(secureRandomFloat64() * float64(backoff) * 0.5)
	nextBackoff := backoff + jitter

	// Log the retry wait
	slog.WarnContext(ctx, "waiting before retry",
		slog.String("tenant_id", tenantID),
		slog.Int64("backoff_duration", nextBackoff.Milliseconds()))

	// Wait or exit on context cancellation
	select {
	case <-time.After(nextBackoff):
	case <-ctx.Done():
	}
}

// secureRandomFloat64 generates a float64 value in the range [0, 1) using crypto/rand.
func secureRandomFloat64() float64 {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return 0 // Default to 0 jitter on error
	}
	return float64(binary.BigEndian.Uint64(b[:])) / (1 << 64)
}
