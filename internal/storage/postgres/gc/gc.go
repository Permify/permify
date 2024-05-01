package gc

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Masterminds/squirrel"

	"github.com/Permify/permify/internal/storage/postgres"
	"github.com/Permify/permify/internal/storage/postgres/utils"
	db "github.com/Permify/permify/pkg/database/postgres"
)

// GC represents a Garbage Collector configuration for database cleanup.
type GC struct {
	// database is the database instance used for garbage collection.
	database *db.Postgres
	// interval is the duration between garbage collection runs.
	interval time.Duration
	// window is the time window for data considered for cleanup.
	window time.Duration
	// timeout is the maximum time allowed for a single GC run.
	timeout time.Duration
}

// NewGC creates a new GC instance with the provided configuration.
func NewGC(db *db.Postgres, opts ...Option) *GC {
	gc := &GC{
		interval: _defaultInterval,
		window:   _defaultWindow,
		timeout:  _defaultTimeout,
		database: db,
	}

	// Custom options
	for _, opt := range opts {
		opt(gc)
	}

	return gc
}

// Start initiates the garbage collection process periodically.
func (gc *GC) Start(ctx context.Context) error {
	ticker := time.NewTicker(gc.interval)
	defer ticker.Stop() // Ensure the ticker is stopped when the function exits.

	for {
		select {
		case <-ticker.C: // Periodically trigger garbage collection.
			if err := gc.Run(); err != nil {
				slog.Error("Garbage collection failed:", slog.Any("error", err))
				continue
			} else {
				slog.Info("Garbage collection completed successfully")
			}
		case <-ctx.Done():
			return ctx.Err() // Return context error if cancellation is requested.
		}
	}
}

// Run performs the garbage collection process.
func (gc *GC) Run() error {
	ctx, cancel := context.WithTimeout(context.Background(), gc.timeout)
	defer cancel()

	// Get the current time from the database timezone.
	var dbNow time.Time
	err := gc.database.WritePool.QueryRow(ctx, "SELECT NOW() AT TIME ZONE 'UTC'").Scan(&dbNow)
	if err != nil {
		slog.Error("Failed to get current time from the database:", slog.Any("error", err))
		return err
	}

	// Calculate the cutoff timestamp based on the window duration.
	cutoffTime := dbNow.Add(-gc.window)

	// Retrieve the last transaction ID that occurred before the cutoff time.
	lastTransactionID, err := gc.getLastTransactionID(ctx, cutoffTime)
	if err != nil {
		slog.Error("Failed to retrieve last transaction ID:", slog.Any("error", err))
		return err
	}

	if lastTransactionID == 0 {
		return nil
	}

	// Delete records in relation_tuples, attributes, and transactions tables based on the lastTransactionID.
	if err := gc.deleteRecords(ctx, postgres.RelationTuplesTable, lastTransactionID); err != nil {
		slog.Error("Failed to delete records in relation_tuples:", slog.Any("error", err))
		return err
	}
	if err := gc.deleteRecords(ctx, postgres.AttributesTable, lastTransactionID); err != nil {
		slog.Error("Failed to delete records in attributes:", slog.Any("error", err))
		return err
	}
	if err := gc.deleteTransactions(ctx, lastTransactionID); err != nil {
		slog.Error("Failed to delete transactions:", slog.Any("error", err))
		return err
	}

	return nil
}

// getLastTransactionID retrieves the last transaction ID from the transactions table that occurred before the provided timestamp.
func (gc *GC) getLastTransactionID(ctx context.Context, before time.Time) (uint64, error) {
	builder := gc.database.Builder.
		Select("id").
		From(postgres.TransactionsTable).
		Where(squirrel.Lt{"timestamp": before}).
		OrderBy("id DESC").
		Limit(1)

	tquery, targs, terr := builder.ToSql()
	if terr != nil {
		return 0, terr
	}

	var lastTransactionID uint64
	row := gc.database.WritePool.QueryRow(ctx, tquery, targs...)
	err := row.Scan(&lastTransactionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}

	return lastTransactionID, nil
}

// deleteRecords generates and executes DELETE queries for relation_tuples and attributes tables based on the lastTransactionID.
func (gc *GC) deleteRecords(ctx context.Context, table string, lastTransactionID uint64) error {
	queryBuilder := utils.GenerateGCQuery(table, lastTransactionID)
	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return err
	}

	_, err = gc.database.WritePool.Exec(ctx, query, args...)
	return err
}

// deleteTransactions deletes transactions older than the provided lastTransactionID.
// It constructs a DELETE query to remove transactions from the database table
// that have a transaction ID less than the provided value.
func (gc *GC) deleteTransactions(ctx context.Context, lastTransactionID uint64) error {
	// Convert the provided lastTransactionID into a string format suitable for SQL queries.
	valStr := fmt.Sprintf("'%v'::xid8", lastTransactionID)

	// Create a Squirrel DELETE query builder for the 'transactions' table.
	queryBuilder := gc.database.Builder.Delete(postgres.TransactionsTable)

	// Create an expression to compare the 'id' column with the lastTransactionID using Lt.
	idExpr := squirrel.Expr(fmt.Sprintf("id < %s", valStr))

	// Add the WHERE clause to filter transactions based on the expression.
	queryBuilder = queryBuilder.Where(idExpr)

	// Generate the SQL query and its arguments from the query builder.
	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return err
	}

	// Execute the DELETE query with the provided context.
	_, err = gc.database.WritePool.Exec(ctx, query, args...)
	return err
}
