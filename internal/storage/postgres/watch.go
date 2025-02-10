package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"

	"google.golang.org/protobuf/encoding/protojson"

	"google.golang.org/protobuf/types/known/anypb"

	"github.com/Masterminds/squirrel"

	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/internal/storage/postgres/snapshot"
	"github.com/Permify/permify/internal/storage/postgres/types"
	db "github.com/Permify/permify/pkg/database/postgres"
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
	txOptions pgx.TxOptions
}

// NewWatcher returns a new instance of the Watch.
func NewWatcher(database *db.Postgres) *Watch {
	return &Watch{
		database:  database,
		txOptions: pgx.TxOptions{IsoLevel: pgx.ReadCommitted, AccessMode: pgx.ReadOnly},
	}
}

// Watch returns a channel that emits a stream of changes to the relationship tuples in the database.
func (w *Watch) Watch(ctx context.Context, tenantID, snap string) (<-chan *base.DataChanges, <-chan error) {
	// Create channels for changes and errors.
	changes := make(chan *base.DataChanges, w.database.GetWatchBufferSize())
	errs := make(chan error, 1)

	var sleep *time.Timer
	const maxSleepDuration = 2 * time.Second
	const defaultSleepDuration = 100 * time.Millisecond
	sleepDuration := defaultSleepDuration

	slog.DebugContext(ctx, "watching for changes in the database", slog.Any("tenant_id", tenantID), slog.Any("snapshot", snap))

	// Decode the snapshot value.
	// The snapshot value represents a point in the history of the database.
	st, err := snapshot.EncodedToken{Value: snap}.Decode()
	if err != nil {
		// If there is an error in decoding the snapshot, send the error and return.
		errs <- err

		slog.Error("failed to decode snapshot", slog.Any("error", err))

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

				slog.Error("error getting recent transaction", slog.Any("error", err))

				errs <- err
				return
			}

			// Process each recent transaction ID.
			for _, id := range recentIDs {
				// Get the changes in the database associated with the current transaction ID.
				updates, err := w.getChanges(ctx, id, tenantID)
				if err != nil {
					// If there is an error in getting the changes, send the error and return.
					slog.ErrorContext(ctx, "failed to get changes for transaction", slog.Any("id", id), slog.Any("error", err))
					errs <- err
					return
				}

				// Send the changes, but respect the context cancellation.
				select {
				case <-ctx.Done(): // If the context is done, send an error and return.
					slog.ErrorContext(ctx, "context canceled, stopping watch")
					errs <- errors.New(base.ErrorCode_ERROR_CODE_CANCELLED.String())
					return
				case changes <- updates: // Send updates to the changes channel.
					slog.DebugContext(ctx, "sent updates to the changes channel for transaction", slog.Any("id", id))
				}

				// Update the transaction ID for the next round.
				cr = id.Uint
				sleepDuration = defaultSleepDuration
			}

			if len(recentIDs) == 0 {

				if sleep == nil {
					sleep = time.NewTimer(sleepDuration)
				} else {
					sleep.Reset(sleepDuration)
				}

				// Increase the sleep duration exponentially, but cap it at maxSleepDuration.
				if sleepDuration < maxSleepDuration {
					sleepDuration *= 2
				} else {
					sleepDuration = maxSleepDuration
				}

				select {
				case <-ctx.Done(): // If the context is done, send an error and return.
					slog.ErrorContext(ctx, "context canceled, stopping watch")
					errs <- errors.New(base.ErrorCode_ERROR_CODE_CANCELLED.String())
					return
				case <-sleep.C: // If the timer is done, continue the loop.
					slog.DebugContext(ctx, "no recent transaction IDs, waiting for changes")
				}
			}
		}
	}()

	slog.DebugContext(ctx, "watch started successfully")

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

		slog.ErrorContext(ctx, "error while building sql query", slog.Any("error", err))

		return nil, err
	}

	slog.DebugContext(ctx, "executing SQL query to get recent transaction", slog.Any("query", query), slog.Any("arguments", args))

	// Execute the SQL query.
	rows, err := w.database.ReadPool.Query(ctx, query, args...)
	if err != nil {
		slog.ErrorContext(ctx, "failed to execute SQL query", slog.Any("error", err))
		return nil, err
	}
	defer rows.Close()

	// Loop through the rows and append XID8 values to the results.
	var xids []types.XID8
	for rows.Next() {
		var xid types.XID8
		err := rows.Scan(&xid)
		if err != nil {
			slog.ErrorContext(ctx, "error while scanning row", slog.Any("error", err))
			return nil, err
		}
		xids = append(xids, xid)
	}

	// Check for errors that could have occurred during iteration.
	err = rows.Err()
	if err != nil {

		slog.ErrorContext(ctx, "failed to iterate over rows", slog.Any("error", err))

		return nil, err
	}

	slog.DebugContext(ctx, "successfully retrieved recent transaction", slog.Any("ids", xids))
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
func (w *Watch) getChanges(ctx context.Context, value types.XID8, tenantID string) (*base.DataChanges, error) {
	// Initialize a new TupleChanges instance.
	changes := &base.DataChanges{}

	slog.DebugContext(ctx, "retrieving changes for transaction", slog.Any("id", value), slog.Any("tenant_id", tenantID))

	// Construct the SQL SELECT statement for retrieving the changes from the RelationTuplesTable.
	tbuilder := w.database.Builder.Select("entity_type, entity_id, relation, subject_type, subject_id, subject_relation, expired_tx_id").
		From(RelationTuplesTable).
		Where(squirrel.Eq{"tenant_id": tenantID}).Where(squirrel.Or{
		squirrel.Eq{"created_tx_id": value},
		squirrel.Eq{"expired_tx_id": value},
	})

	// Generate the SQL query and arguments.
	tquery, targs, err := tbuilder.ToSql()
	if err != nil {
		slog.ErrorContext(ctx, "error while building sql query for relation tuples", slog.Any("error", err))
		return nil, err
	}

	slog.DebugContext(ctx, "executing sql query for relation tuples", slog.Any("query", tquery), slog.Any("arguments", targs))

	// Execute the SQL query and retrieve the result rows.
	var trows pgx.Rows
	trows, err = w.database.ReadPool.Query(ctx, tquery, targs...)
	if err != nil {
		slog.ErrorContext(ctx, "failed to execute sql query for relation tuples", slog.Any("error", err))
		return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}
	// Ensure the rows are closed after processing.
	defer trows.Close()

	abuilder := w.database.Builder.Select("entity_type, entity_id, attribute, value, expired_tx_id").
		From(AttributesTable).
		Where(squirrel.Eq{"tenant_id": tenantID}).Where(squirrel.Or{
		squirrel.Eq{"created_tx_id": value},
		squirrel.Eq{"expired_tx_id": value},
	})

	aquery, aargs, err := abuilder.ToSql()
	if err != nil {
		slog.ErrorContext(ctx, "error while building SQL query for attributes", slog.Any("error", err))
		return nil, err
	}

	slog.DebugContext(ctx, "executing sql query for attributes", slog.Any("query", aquery), slog.Any("arguments", aargs))

	var arows pgx.Rows
	arows, err = w.database.ReadPool.Query(ctx, aquery, aargs...)
	if err != nil {
		slog.ErrorContext(ctx, "error while executing SQL query for attributes", slog.Any("error", err))
		return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}
	// Ensure the rows are closed after processing.
	defer arows.Close()

	// Set the snapshot token for the changes.
	changes.SnapToken = snapshot.Token{Value: value}.Encode().String()

	// Iterate through the result rows.
	for trows.Next() {
		var expiredXID types.XID8

		rt := storage.RelationTuple{}
		// Scan the result row into a RelationTuple instance.
		err = trows.Scan(&rt.EntityType, &rt.EntityID, &rt.Relation, &rt.SubjectType, &rt.SubjectID, &rt.SubjectRelation, &expiredXID)
		if err != nil {
			slog.ErrorContext(ctx, "error while scanning row for relation tuples", slog.Any("error", err))
			return nil, err
		}

		// Determine the operation type based on the expired transaction ID.
		op := base.DataChange_OPERATION_CREATE
		if expiredXID.Uint == value.Uint {
			op = base.DataChange_OPERATION_DELETE
		}

		// Append the change to the list of changes.
		changes.DataChanges = append(changes.DataChanges, &base.DataChange{
			Operation: op,
			Type: &base.DataChange_Tuple{
				Tuple: rt.ToTuple(),
			},
		})
	}

	// Iterate through the result rows.
	for arows.Next() {
		var expiredXID types.XID8

		rt := storage.Attribute{}

		var valueStr string

		// Scan the result row into a RelationTuple instance.
		err = arows.Scan(&rt.EntityType, &rt.EntityID, &rt.Attribute, &valueStr, &expiredXID)
		if err != nil {
			slog.ErrorContext(ctx, "error while scanning row for attributes", slog.Any("error", err))
			return nil, err
		}

		// Unmarshal the JSON data from `valueStr` into `rt.Value`.
		rt.Value = &anypb.Any{}
		err = protojson.Unmarshal([]byte(valueStr), rt.Value)
		if err != nil {
			slog.ErrorContext(ctx, "failed to unmarshal attribute value", slog.Any("error", err))
			return nil, err
		}

		// Determine the operation type based on the expired transaction ID.
		op := base.DataChange_OPERATION_CREATE
		if expiredXID.Uint == value.Uint {
			op = base.DataChange_OPERATION_DELETE
		}

		// Append the change to the list of changes.
		changes.DataChanges = append(changes.DataChanges, &base.DataChange{
			Operation: op,
			Type: &base.DataChange_Attribute{
				Attribute: rt.ToAttribute(),
			},
		})
	}

	slog.DebugContext(ctx, "successfully retrieved changes for transaction", slog.Any("id", value))

	// Return the changes and no error.
	return changes, nil
}
