package postgres

import (
	"context"
	"fmt"
	"log/slog"
)

const (
	// ActiveRecordTxnID represents the maximum XID8 value used for active records
	// to avoid XID wraparound issues (instead of using 0)
	ActiveRecordTxnID = uint64(9223372036854775807)
	MaxXID8Value      = "'9223372036854775807'::xid8"
)

// RepairConfig holds configuration for the XID counter repair operation
type RepairConfig struct {
	BatchSize  int  // batch size for XID advancement
	MaxRetries int  // maximum number of retries
	RetryDelay int  // milliseconds
	DryRun     bool // perform a dry run without making changes
	Verbose    bool // enable verbose logging
}

// DefaultRepairConfig returns default configuration for XID counter repair
func DefaultRepairConfig() *RepairConfig {
	return &RepairConfig{
		BatchSize:  1000, // default batch size for XID advancement
		MaxRetries: 3,
		RetryDelay: 100,
		DryRun:     false,
		Verbose:    true,
	}
}

// RepairResult holds the results of the XID counter repair operation
type RepairResult struct {
	CreatedTxIdFixed int // Number of XIDs advanced in counter
	Errors           []error
	Duration         string
}

// Repair performs XID counter repair to prevent XID wraparound issues
// This function uses a safe approach: only advance XID counter, don't modify existing data
func (p *Postgres) Repair(ctx context.Context, config *RepairConfig) (*RepairResult, error) {
	if config == nil {
		config = DefaultRepairConfig()
	}

	// Validate BatchSize - don't accept 0 or negative values
	if config.BatchSize <= 0 {
		config.BatchSize = 1000 // Use default value
		slog.InfoContext(ctx, "Invalid BatchSize provided, using default", slog.Int("default_batch_size", 1000))
	}

	result := &RepairResult{
		Errors: make([]error, 0),
	}

	slog.InfoContext(ctx, "Starting PostgreSQL transaction ID counter repair",
		slog.Bool("dry_run", config.DryRun),
		slog.Int("batch_size", config.BatchSize))

	// Step 1: Get current PostgreSQL XID
	currentXID, err := p.getCurrentPostgreSQLXID(ctx)
	if err != nil {
		return result, fmt.Errorf("failed to get current PostgreSQL XID: %w", err)
	}

	if config.Verbose {
		slog.InfoContext(ctx, "Current PostgreSQL transaction ID", slog.Uint64("current_xid", currentXID))
	}

	// Step 2: Get maximum referenced XID from transactions table
	maxReferencedXID, err := p.getMaxReferencedXID(ctx)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("failed to get max referenced XID: %w", err))
		return result, nil
	}

	if config.Verbose {
		slog.InfoContext(ctx, "Maximum referenced transaction ID", slog.Uint64("max_referenced_xid", maxReferencedXID))
	}

	// Step 3: Advance XID counter if needed
	if maxReferencedXID > currentXID {
		counterDelta := int(maxReferencedXID - currentXID + 1000) // Add safety buffer

		if config.DryRun {
			slog.InfoContext(ctx, "Would advance XID counter",
				slog.Int("counter_delta", counterDelta),
				slog.Uint64("target_xid", maxReferencedXID+1000))
			result.CreatedTxIdFixed = counterDelta
		} else {
			if err := p.advanceXIDCounterByDelta(ctx, counterDelta, config); err != nil {
				result.Errors = append(result.Errors, fmt.Errorf("failed to advance XID counter: %w", err))
			} else {
				result.CreatedTxIdFixed = counterDelta
			}
		}
	} else {
		slog.InfoContext(ctx, "No XID counter advancement needed")
	}

	slog.InfoContext(ctx, "PostgreSQL transaction ID counter repair completed",
		slog.Int("transactions_advanced", result.CreatedTxIdFixed),
		slog.Int("errors", len(result.Errors)))

	return result, nil
}

// getCurrentPostgreSQLXID gets the current PostgreSQL transaction ID
func (p *Postgres) getCurrentPostgreSQLXID(ctx context.Context) (uint64, error) {
	var x int64
	query := "SELECT (pg_current_xact_id()::text)::bigint"
	if err := p.ReadPool.QueryRow(ctx, query).Scan(&x); err != nil {
		return 0, err
	}
	return uint64(x), nil
}

// getMaxReferencedXID gets the maximum transaction ID referenced in the transactions table
func (p *Postgres) getMaxReferencedXID(ctx context.Context) (uint64, error) {
	query := "SELECT (MAX(id)::text)::bigint FROM transactions"
	var x int64
	if err := p.ReadPool.QueryRow(ctx, query).Scan(&x); err != nil {
		return 0, err
	}
	return uint64(x), nil
}

// advanceXIDCounterByDelta advances the PostgreSQL XID counter by specified delta
func (p *Postgres) advanceXIDCounterByDelta(ctx context.Context, counterDelta int, config *RepairConfig) error {
	if counterDelta <= 0 {
		return nil
	}
	slog.InfoContext(ctx, "Advancing transaction ID counter by delta", slog.Int("counter_delta", counterDelta))

	batchSize := config.BatchSize
	if batchSize <= 0 {
		batchSize = 1000
	}

	conn, err := p.WritePool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire connection: %w", err)
	}
	defer conn.Release()

	remaining := counterDelta
	for remaining > 0 {
		if err := ctx.Err(); err != nil {
			return err
		}
		currentBatch := remaining
		if currentBatch > batchSize {
			currentBatch = batchSize
		}

		for i := 0; i < currentBatch; i++ {
			if err := ctx.Err(); err != nil {
				return err
			}
			tx, err := conn.Begin(ctx)
			if err != nil {
				return fmt.Errorf("begin tx: %w", err)
			}
			if _, err := tx.Exec(ctx, "SELECT pg_current_xact_id()"); err != nil {
				_ = tx.Rollback(ctx)
				return fmt.Errorf("advance xid (iter %d): %w", i+1, err)
			}
			// Rolling back is fine — XID assignment happens on first reference.
			if err := tx.Rollback(ctx); err != nil {
				return fmt.Errorf("rollback tx: %w", err)
			}
		}

		remaining -= currentBatch
		if config.Verbose {
			slog.InfoContext(ctx, "Advanced XID counter batch",
				slog.Int("batch_size", currentBatch),
				slog.Int("remaining", remaining))
		}
	}
	slog.InfoContext(ctx, "Transaction ID counter advancement completed", slog.Int("total_advanced", counterDelta))
	return nil
}
