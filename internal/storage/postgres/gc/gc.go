package gc

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"

	"github.com/Permify/permify/internal/storage/postgres"
	"github.com/Permify/permify/internal/storage/postgres/utils"
	db "github.com/Permify/permify/pkg/database/postgres"
	"github.com/Permify/permify/pkg/logger"
)

type GC struct {
	database *db.Postgres
	logger   logger.Interface
	interval time.Duration
	window   time.Duration
	timeout  time.Duration
}

// NewGC creates a new GC instance with the provided configuration.
func NewGC(db *db.Postgres, logger logger.Interface, opts ...Option) *GC {
	gc := &GC{
		interval: _defaultInterval,
		window:   _defaultWindow,
		timeout:  _defaultTimeout,
		database: db,
		logger:   logger,
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
				gc.logger.Error("Garbage collection failed:", err)
				continue
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
	err := gc.database.DB.QueryRowContext(ctx, "SELECT NOW() AT TIME ZONE 'UTC'").Scan(&dbNow)
	if err != nil {
		return err
	}

	// Calculate the cutoff timestamp based on the window duration.
	cutoffTime := dbNow.Add(-gc.window)

	// Retrieve the last transaction ID that occurred before the cutoff time.
	lastTransactionID, err := gc.getLastTransactionID(ctx, cutoffTime)
	if err != nil {
		return err
	}

	// Delete records in relation_tuples, attributes, and transactions tables based on the lastTransactionID.
	if err := gc.deleteRecords(ctx, postgres.RelationTuplesTable, lastTransactionID); err != nil {
		return err
	}
	if err := gc.deleteRecords(ctx, postgres.AttributesTable, lastTransactionID); err != nil {
		return err
	}
	if err := gc.deleteTransactions(ctx, lastTransactionID); err != nil {
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
	row := gc.database.DB.QueryRowContext(ctx, tquery, targs...)
	err := row.Scan(&lastTransactionID)
	if err != nil {
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

	_, err = gc.database.DB.ExecContext(ctx, query, args...)
	return err
}

// deleteTransactions generates and executes DELETE queries for the transactions table based on the lastTransactionID.
func (gc *GC) deleteTransactions(ctx context.Context, lastTransactionID uint64) error {
	valStr := fmt.Sprintf("'%v'::xid8", lastTransactionID)

	queryBuilder := gc.database.Builder.
		Delete(postgres.TransactionsTable).
		Where(squirrel.Lt{"id": valStr})

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return err
	}

	_, err = gc.database.DB.ExecContext(ctx, query, args...)
	return err
}
