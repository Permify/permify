package utils

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/codes"

	"go.opentelemetry.io/otel/trace"

	"github.com/Masterminds/squirrel"

	base "github.com/Permify/permify/pkg/pb/base/v1"
)

const (
	TransactionTemplate       = `INSERT INTO transactions (tenant_id) VALUES ($1) RETURNING id, snapshot`
	InsertTenantTemplate      = `INSERT INTO tenants (id, name) VALUES ($1, $2) RETURNING created_at`
	DeleteTenantTemplate      = `DELETE FROM tenants WHERE id = $1 RETURNING name, created_at`
	DeleteAllByTenantTemplate = `DELETE FROM %s WHERE tenant_id = $1`

	// ActiveRecordTxnID represents the maximum XID8 value used for active records
	// to avoid XID wraparound issues (instead of using 0)
	ActiveRecordTxnID = uint64(9223372036854775807)
	MaxXID8Value      = "'9223372036854775807'::xid8"

	// earliestPostgresVersion represents the earliest supported version of PostgreSQL is 13.8
	earliestPostgresVersion = 130008 // The earliest supported version of PostgreSQL is 13.8
)

// createFinalSnapshot creates a final snapshot string for proper transaction visibility.
// If xmax != xid, it adds xid to the xip_list to make the snapshot unique.
func createFinalSnapshot(snapshotValue string, xid uint64) string {
	// Parse snapshot: "xmin:xmax:xip_list"
	parts := strings.Split(snapshotValue, ":")
	if len(parts) < 2 {
		// Invalid snapshot format, return original
		return snapshotValue
	}

	// Parse xmax as integer for proper comparison
	xmaxStr := parts[1]
	xmax, err := strconv.ParseUint(xmaxStr, 10, 64)
	if err != nil {
		// If parsing fails, use original snapshot
		return snapshotValue
	}

	// If xmax == xid, no need to modify snapshot
	if xmax == xid {
		return snapshotValue
	}

	// If xid > xmax, this is invalid - xid should be visible in snapshot
	// Return original snapshot to avoid creating invalid format
	if xid > xmax {
		return snapshotValue
	}

	// xmax > xid, need to add xid to xip_list for uniqueness
	xmin := parts[0]

	// Check if xip_list exists and is not empty
	if len(parts) == 3 && parts[2] != "" {
		// Check if xid is already in xip_list
		xipList := parts[2]
		for xip := range strings.SplitSeq(xipList, ",") {
			if strings.TrimSpace(xip) == fmt.Sprintf("%d", xid) {
				// xid already in xip_list, return original snapshot
				return snapshotValue
			}
		}
		// xid not in xip_list, append it
		return fmt.Sprintf("%s:%s:%s,%d", xmin, xmaxStr, parts[2], xid)
	} else {
		// xip_list is empty, add xid
		return fmt.Sprintf("%s:%s:%d", xmin, xmaxStr, xid)
	}
}

// SnapshotQuery adds conditions to a SELECT query for checking transaction visibility based on created and expired transaction IDs.
// Optimized version with parameterized queries for security.
func SnapshotQuery(sl squirrel.SelectBuilder, value uint64, snapshotValue string) squirrel.SelectBuilder {
	// Backward compatibility: if snapshot is empty, use old method
	if snapshotValue == "" {
		// Create a subquery for the snapshot associated with the provided value.
		snapshotQuery := "(select snapshot from transactions where id = ?::xid8)"

		// Records that were created and are visible in the snapshot
		createdWhere := squirrel.Or{
			squirrel.Expr("pg_visible_in_snapshot(created_tx_id, ?) = true", squirrel.Expr(snapshotQuery, value)),
			squirrel.Expr("created_tx_id = ?::xid8", value), // Include current transaction
		}

		// Records that are still active (not expired) at snapshot time
		expiredWhere := squirrel.And{
			squirrel.Or{
				squirrel.Expr("pg_visible_in_snapshot(expired_tx_id, ?) = false", squirrel.Expr(snapshotQuery, value)),
				squirrel.Expr("expired_tx_id = ?::xid8", ActiveRecordTxnID), // Never expired
			},
			squirrel.Expr("expired_tx_id <> ?::xid8", value), // Not expired by current transaction
		}

		// Add the created and expired conditions to the SELECT query.
		return sl.Where(createdWhere).Where(expiredWhere)
	}

	// Create final snapshot with proper visibility
	finalSnapshot := createFinalSnapshot(snapshotValue, value)

	// Records that were created and are visible in the snapshot
	createdWhere := squirrel.Or{
		squirrel.Expr("pg_visible_in_snapshot(created_tx_id, ?) = true", finalSnapshot),
		squirrel.Expr("created_tx_id = ?::xid8", value), // Include current transaction
	}

	// Records that are still active (not expired) at snapshot time
	expiredWhere := squirrel.And{
		squirrel.Or{
			squirrel.Expr("pg_visible_in_snapshot(expired_tx_id, ?) = false", finalSnapshot),
			squirrel.Expr("expired_tx_id = ?::xid8", ActiveRecordTxnID), // Never expired
		},
		squirrel.Expr("expired_tx_id <> ?::xid8", value), // Not expired by current transaction
	}

	// Add the created and expired conditions to the SELECT query.
	return sl.Where(createdWhere).Where(expiredWhere)
}

// GenerateGCQuery generates a Squirrel DELETE query builder for garbage collection.
// It constructs a query to delete expired records from the specified table
// based on the provided value, which represents a transaction ID.
func GenerateGCQuery(table string, value uint64) squirrel.DeleteBuilder {
	// Create a Squirrel DELETE builder for the specified table.
	deleteBuilder := squirrel.Delete(table)

	// Create an expression to check if 'expired_tx_id' is not equal to ActiveRecordTxnID (expired records).
	expiredNotActiveExpr := squirrel.Expr("expired_tx_id <> ?::xid8", ActiveRecordTxnID)

	// Create an expression to check if 'expired_tx_id' is less than the provided value (before the cutoff).
	beforeExpr := squirrel.Expr("expired_tx_id < ?::xid8", value)

	// Add the WHERE clauses to the DELETE query builder to filter and delete expired data.
	return deleteBuilder.Where(expiredNotActiveExpr).Where(beforeExpr)
}

// GenerateGCQueryForTenant generates a Squirrel DELETE query builder for tenant-aware garbage collection.
// It constructs a query to delete expired records from the specified table for a specific tenant
// based on the provided value, which represents a transaction ID.
func GenerateGCQueryForTenant(table, tenantID string, value uint64) squirrel.DeleteBuilder {
	// Create a Squirrel DELETE builder for the specified table.
	deleteBuilder := squirrel.Delete(table)

	// Create an expression to check if 'expired_tx_id' is not equal to ActiveRecordTxnID (expired records).
	expiredNotActiveExpr := squirrel.Expr("expired_tx_id <> ?::xid8", ActiveRecordTxnID)

	// Create an expression to check if 'expired_tx_id' is less than the provided value (before the cutoff).
	beforeExpr := squirrel.Expr("expired_tx_id < ?::xid8", value)

	// Add the WHERE clauses to the DELETE query builder to filter and delete expired data for the specific tenant.
	return deleteBuilder.Where(squirrel.Eq{"tenant_id": tenantID}).Where(expiredNotActiveExpr).Where(beforeExpr)
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
	// Calculate the base backoff with bit shifting for better performance
	baseBackoff := 20 * time.Millisecond
	if retries > 0 {
		// Use bit shifting instead of math.Pow for better performance
		shift := min(retries, 5) // Cap at 2^5 = 32, so max backoff is 640ms
		baseBackoff = baseBackoff << shift
	}

	// Cap at 1 second
	if baseBackoff > time.Second {
		baseBackoff = time.Second
	}

	// Generate jitter using crypto/rand
	jitter := time.Duration(secureRandomFloat64() * float64(baseBackoff) * 0.5)
	nextBackoff := baseBackoff + jitter

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
// Optimized version with better error handling and performance.
func secureRandomFloat64() float64 {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		// Use a fallback random value instead of 0 to maintain jitter
		return 0.5 // Middle value for consistent jitter behavior
	}
	// Use bit shifting instead of division for better performance
	return float64(binary.BigEndian.Uint64(b[:])) / (1 << 63) / 2.0
}

// EnsureDBVersion checks the version of the given database connection
// and returns an error if the version is not supported.
func EnsureDBVersion(db *pgxpool.Pool) (version string, err error) {
	err = db.QueryRow(context.Background(), "SHOW server_version_num;").Scan(&version)
	if err != nil {
		return version, err
	}
	v, err := strconv.Atoi(version)
	if v < earliestPostgresVersion {
		err = fmt.Errorf("unsupported postgres version: %s, expected >= %d", version, earliestPostgresVersion)
	}
	return version, err
}
