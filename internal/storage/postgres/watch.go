package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"

	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/internal/storage/postgres/snapshot"
	"github.com/Permify/permify/internal/storage/postgres/types"
	db "github.com/Permify/permify/pkg/database/postgres"
	"github.com/Permify/permify/pkg/logger"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// Watch is an implementation of the storage.Watch interface, which is used
type Watch struct {
	// database is a pointer to a Postgres database instance, which is used
	// to perform operations on the relationship data.
	database *db.Postgres

	// txOptions holds the configuration for database transactions, such as
	// isolation level and read-only mode, to be applied when performing
	// operations on the relationship data.
	txOptions sql.TxOptions

	// options
	bufferSize int

	// logger is an instance of a logger that implements the logger.Interface
	// and is used to log messages related to the operations performed by
	// the RelationshipReader.
	logger logger.Interface
}

// NewWatcher returns a new instance of the Watch.
func NewWatcher(database *db.Postgres, logger logger.Interface) *Watch {
	return &Watch{
		database:   database,
		txOptions:  sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true},
		logger:     logger,
		bufferSize: _defaultWatchBufferSize,
	}
}

// Watch returns a channel that emits a stream of changes to the relationship tuples in the database.
func (w *Watch) Watch(ctx context.Context, tenantID string, snap string) (<-chan *base.TupleChanges, <-chan error) {
	// Create channels for changes and errors.
	changes := make(chan *base.TupleChanges, w.bufferSize)
	errs := make(chan error, 1)

	// Decode the snapshot value.
	// The snapshot value represents a point in the history of the database.
	st, err := snapshot.EncodedToken{Value: snap}.Decode()
	if err != nil {
		// If there is an error in decoding the snapshot, send the error and return.
		errs <- err
		return changes, errs
	}

	// Start a goroutine to watch for changes in the database.
	go func() {
		// Ensure to close the channels when we're done.
		defer close(changes)
		defer close(errs)

		// Get the transaction ID from the snapshot.
		cr := st.(snapshot.Token).Value.Uint

		// Continuously watch for changes.
		for {
			// Get the list of recent transaction IDs.
			recentIDs, err := w.getRecentXIDs(ctx, cr, tenantID)
			if err != nil {
				// If there is an error in getting recent transaction IDs, send the error and return.
				errs <- err
				return
			}

			// Process each recent transaction ID.
			for _, id := range recentIDs {
				// Get the changes in the database associated with the current transaction ID.
				updates, err := w.getChanges(ctx, id, tenantID)
				if err != nil {
					// If there is an error in getting the changes, send the error and return.
					errs <- err
					return
				}

				// Send the changes, but respect the context cancellation.
				select {
				case changes <- updates: // Send updates to the changes channel.
				case <-ctx.Done(): // If the context is done, send an error and return.
					errs <- errors.New(base.ErrorCode_ERROR_CODE_CANCELLED.String())
					return
				}

				// Update the transaction ID for the next round.
				cr = id.Uint
			}

			// If there are no recent transaction IDs, wait for a short period before trying again.
			if len(recentIDs) == 0 {
				sleep := time.NewTimer(100 * time.Millisecond)

				select {
				case <-sleep.C: // If the timer is done, continue the loop.
				case <-ctx.Done(): // If the context is done, send an error and return.
					errs <- errors.New(base.ErrorCode_ERROR_CODE_CANCELLED.String())
					return
				}
			}
		}
	}()

	// Return the channels that the caller will listen to for changes and errors.
	return changes, errs
}

// getRecentXIDs fetches a list of XID8 identifiers from the 'transactions' table
// for all transactions committed after a specified XID value.
//
// Parameters:
//   - ctx:       A context to control the execution lifetime.
//   - value:     The transaction XID after which we need the changes.
//   - tenantID:  The ID of the tenant to filter the transactions for.
//
// Returns:
//   - A slice of XID8 identifiers.
//   - An error if the query fails to execute, or other error occurs during its execution.
func (w *Watch) getRecentXIDs(ctx context.Context, value uint64, tenantID string) ([]types.XID8, error) {
	// Convert the value to a string formatted as a Postgresql XID8 type.
	valStr := fmt.Sprintf("'%v'::xid8", value)

	subquery := fmt.Sprintf("(select pg_xact_commit_timestamp(id::xid) from transactions where id = %s)", valStr)

	// Build the main query to get transactions committed after the one with a given XID,
	// still visible in the current snapshot, ordered by their commit timestamps.
	builder := w.database.Builder.Select("id").
		From(TransactionsTable).
		Where(fmt.Sprintf("pg_xact_commit_timestamp(id::xid) > (%s)", subquery)).
		Where("id < pg_snapshot_xmin(pg_current_snapshot())").
		Where(squirrel.Eq{"tenant_id": tenantID}).
		OrderBy("pg_xact_commit_timestamp(id::xid)")

	// Convert the builder to a SQL query and arguments.
	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	// Execute the SQL query.
	rows, err := w.database.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Loop through the rows and append XID8 values to the results.
	var xids []types.XID8
	for rows.Next() {
		var xid types.XID8
		err := rows.Scan(&xid)
		if err != nil {
			return nil, err
		}
		xids = append(xids, xid)
	}

	// Check for errors that could have occurred during iteration.
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return xids, nil
}

// getChanges is a method that retrieves the changes that occurred in the relation tuples within a specified transaction.
//
// ctx: The context.Context instance for managing the life-cycle of this function.
// value: The ID of the transaction for which to retrieve the changes.
// tenantID: The ID of the tenant for which to retrieve the changes.
//
// This method returns a TupleChanges instance that encapsulates the changes in the relation tuples within the specified
// transaction, or an error if something went wrong during execution.
func (w *Watch) getChanges(ctx context.Context, value types.XID8, tenantID string) (*base.TupleChanges, error) {
	// Initialize a new TupleChanges instance.
	changes := &base.TupleChanges{}

	// Construct the SQL SELECT statement for retrieving the changes from the RelationTuplesTable.
	builder := w.database.Builder.Select("entity_type, entity_id, relation, subject_type, subject_id, subject_relation, expired_tx_id").
		From(RelationTuplesTable).
		Where(squirrel.Eq{"tenant_id": tenantID}).Where(squirrel.Or{
		squirrel.Eq{"created_tx_id": value},
		squirrel.Eq{"expired_tx_id": value},
	})

	// Generate the SQL query and arguments.
	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	// Execute the SQL query and retrieve the result rows.
	var rows *sql.Rows
	rows, err = w.database.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}
	// Ensure the rows are closed after processing.
	defer rows.Close()

	// Set the snapshot token for the changes.
	changes.SnapToken = snapshot.Token{Value: value}.Encode().String()

	// Iterate through the result rows.
	for rows.Next() {
		var expiredXID types.XID8

		rt := storage.RelationTuple{}
		// Scan the result row into a RelationTuple instance.
		err = rows.Scan(&rt.EntityType, &rt.EntityID, &rt.Relation, &rt.SubjectType, &rt.SubjectID, &rt.SubjectRelation, &expiredXID)
		if err != nil {
			return nil, err
		}

		// Determine the operation type based on the expired transaction ID.
		op := base.TupleChange_OPERATION_CREATE
		if expiredXID.Uint == value.Uint {
			op = base.TupleChange_OPERATION_DELETE
		}

		// Append the change to the list of changes.
		changes.TupleChanges = append(changes.TupleChanges, &base.TupleChange{
			Operation: op,
			Tuple:     rt.ToTuple(),
		})
	}

	// Return the changes and no error.
	return changes, nil
}
