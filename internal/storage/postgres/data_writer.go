package postgres

import (
	"context"
	"errors"
	"log/slog" // Structured logging

	"github.com/jackc/pgx/v5"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/Permify/permify/internal"
	"github.com/Permify/permify/internal/storage/postgres/snapshot"
	"github.com/Permify/permify/internal/storage/postgres/utils"
	"github.com/Permify/permify/internal/validation"
	"github.com/Permify/permify/pkg/bundle"
	"github.com/Permify/permify/pkg/database"
	db "github.com/Permify/permify/pkg/database/postgres"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/token"
	"github.com/Permify/permify/pkg/tuple"
)

// DataWriter - Structure for Data Writer
type DataWriter struct {
	database *db.Postgres
	// options
	txOptions pgx.TxOptions
}

func NewDataWriter(database *db.Postgres) *DataWriter {
	return &DataWriter{
		database:  database,
		txOptions: pgx.TxOptions{IsoLevel: pgx.Serializable, AccessMode: pgx.ReadWrite},
	}
}

// Write method writes a collection of tuples and attributes to the database for a specific tenant.
// It returns an EncodedSnapToken upon successful write or an error if the write fails.
func (w *DataWriter) Write(
	ctx context.Context,
	tenantID string,
	tupleCollection *database.TupleCollection,
	attributeCollection *database.AttributeCollection,
) (token token.EncodedSnapToken, err error) {
	// Start a new tracing span for this operation.
	ctx, span := internal.Tracer.Start(ctx, "data-writer.write")
	defer span.End() // Ensure that the span is ended when the function returns.

	// Log the start of a data write operation.
	slog.DebugContext(ctx, "writing data for tenant_id", slog.String("tenant_id", tenantID), "max retries", slog.Any("max_retries", w.database.GetMaxRetries()))
	// Validate data size
	// Check if the total number of tuples and attributes exceeds the maximum allowed per write.
	if len(tupleCollection.GetTuples())+len(attributeCollection.GetAttributes()) > w.database.GetMaxDataPerWrite() {
		return nil, errors.New(base.ErrorCode_ERROR_CODE_MAX_DATA_PER_WRITE_EXCEEDED.String())
	}

	// Retry loop for handling transient errors like serialization issues.
	for i := 0; i <= w.database.GetMaxRetries(); i++ {
		// Attempt to write the data to the database.
		tkn, err := w.write(ctx, tenantID, tupleCollection, attributeCollection)
		if err != nil {
			// Check if the error is due to serialization, and if so, retry.
			if utils.IsSerializationRelatedError(err) || pgconn.SafeToRetry(err) {
				slog.WarnContext(ctx, "serialization error occurred", slog.String("tenant_id", tenantID), slog.Int("retry", i))
				utils.WaitWithBackoff(ctx, tenantID, i)
				continue // Retry the operation.
			}
			// If the error is not serialization-related, handle it and return.
			return nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_DATASTORE)
		}
		// If to write is successful, return the token.
		return tkn, nil
	}

	// Log an error if the operation failed after reaching the maximum number of retries.
	slog.ErrorContext(ctx, "max retries reached", slog.Any("error", errors.New(base.ErrorCode_ERROR_CODE_ERROR_MAX_RETRIES.String())))
	// Max retries exceeded
	// Return an error indicating that the maximum number of retries has been reached.
	return nil, errors.New(base.ErrorCode_ERROR_CODE_ERROR_MAX_RETRIES.String())
}

// Delete method removes data from the database based on the provided tuple and attribute filters.
// It returns an EncodedSnapToken upon successful deletion or an error if the deletion fails.
func (w *DataWriter) Delete(
	ctx context.Context,
	tenantID string,
	tupleFilter *base.TupleFilter,
	attributeFilter *base.AttributeFilter,
) (token.EncodedSnapToken, error) {
	// Start a new tracing span for this delete operation.
	ctx, span := internal.Tracer.Start(ctx, "data-writer.delete")
	defer span.End() // Ensure that the span is ended when the function returns.

	// Log the start of a data deletion operation.
	slog.DebugContext(ctx, "deleting data for tenant_id", slog.String("tenant_id", tenantID), "max retries", slog.Any("max_retries", w.database.GetMaxRetries()))
	// Retry loop with backoff
	// Retry loop for handling transient errors like serialization issues.
	for i := 0; i <= w.database.GetMaxRetries(); i++ {
		// Attempt to delete the data from the database.
		tkn, err := w.delete(ctx, tenantID, tupleFilter, attributeFilter)
		if err != nil {
			// Check if the error is due to serialization, and if so, retry.
			if utils.IsSerializationRelatedError(err) || pgconn.SafeToRetry(err) {
				slog.WarnContext(ctx, "serialization error occurred", slog.String("tenant_id", tenantID), slog.Int("retry", i))
				utils.WaitWithBackoff(ctx, tenantID, i)
				continue // Retry the operation.
			}
			// If the error is not serialization-related, handle it and return.
			return nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_DATASTORE)
		}
		// If the delete operation is successful, return the token.
		return tkn, nil
	}

	// Log an error if the operation failed after reaching the maximum number of retries.
	slog.DebugContext(ctx, "max retries reached", slog.Any("error", errors.New(base.ErrorCode_ERROR_CODE_ERROR_MAX_RETRIES.String())))
	// Max retries exceeded
	// Return an error indicating that the maximum number of retries has been reached.
	return nil, errors.New(base.ErrorCode_ERROR_CODE_ERROR_MAX_RETRIES.String())
}

// RunBundle executes a bundle of operations in the context of a given tenant.
// It returns an EncodedSnapToken upon successful completion or an error if the operation fails.
func (w *DataWriter) RunBundle(
	ctx context.Context,
	tenantID string,
	arguments map[string]string,
	b *base.DataBundle,
) (token.EncodedSnapToken, error) {
	// Start a new tracing span for this operation.
	ctx, span := internal.Tracer.Start(ctx, "data-writer.run-bundle")
	defer span.End() // Ensure that the span is ended when the function returns.

	// Log the start of running a bundle operation.
	slog.DebugContext(ctx, "running bundle for tenant_id", slog.String("tenant_id", tenantID), "max retries", slog.Any("max_retries", w.database.GetMaxRetries()))
	// Retry loop with backoff
	// Retry loop for handling transient errors like serialization issues.
	for i := 0; i <= w.database.GetMaxRetries(); i++ {
		// Attempt to run the bundle operation.
		tkn, err := w.runBundle(ctx, tenantID, arguments, b)
		if err != nil {
			// Check if the error is due to serialization, and if so, retry.
			if utils.IsSerializationRelatedError(err) || pgconn.SafeToRetry(err) {
				slog.WarnContext(ctx, "serialization error occurred", slog.String("tenant_id", tenantID), slog.Int("retry", i))
				utils.WaitWithBackoff(ctx, tenantID, i)
				continue // Retry the operation.
			}
			// If the error is not serialization-related, handle it and return.
			return nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_DATASTORE)
		}
		// If the operation is successful, return the token.
		return tkn, nil
	}

	// Log an error if the operation failed after reaching the maximum number of retries.
	slog.ErrorContext(ctx, "max retries reached", slog.Any("error", errors.New(base.ErrorCode_ERROR_CODE_ERROR_MAX_RETRIES.String())))
	// Max retries exceeded
	// Return an error indicating that the maximum number of retries has been reached.
	return nil, errors.New(base.ErrorCode_ERROR_CODE_ERROR_MAX_RETRIES.String())
}

// write handles the database writing of tuple and attribute collections for a given tenant.
// It returns an EncodedSnapToken upon successful write or an error if the write fails.
func (w *DataWriter) write(
	ctx context.Context,
	tenantID string,
	tupleCollection *database.TupleCollection,
	attributeCollection *database.AttributeCollection,
) (token token.EncodedSnapToken, err error) {
	var tx pgx.Tx
	tx, err = w.database.WritePool.BeginTx(ctx, w.txOptions)
	if err != nil {
		return nil, err
	}
	// Defer rollback
	defer func() {
		_ = tx.Rollback(ctx)
	}()
	// Get transaction ID
	var xid db.XID8
	err = tx.QueryRow(ctx, utils.TransactionTemplate, tenantID).Scan(&xid)
	if err != nil {
		return nil, err
	}

	slog.DebugContext(ctx, "retrieved transaction", slog.Any("xid", xid), "for tenant", slog.Any("tenant_id", tenantID))

	slog.DebugContext(ctx, "processing tuples and executing insert query")

	batch := &pgx.Batch{}

	if len(tupleCollection.GetTuples()) > 0 {
		err = w.batchInsertRelationships(batch, xid, tenantID, tupleCollection)
		if err != nil {
			return nil, err
		}
	}

	if len(attributeCollection.GetAttributes()) > 0 {
		err = w.batchUpdateAttributes(batch, xid, tenantID, buildDeleteClausesForAttributes(attributeCollection))
		if err != nil {
			return nil, err
		}
		err = w.batchInsertAttributes(batch, xid, tenantID, attributeCollection)
		if err != nil {
			return nil, err
		}
	}

	batchResult := tx.SendBatch(ctx, batch)
	for i := 0; i < batch.Len(); i++ {
		_, err = batchResult.Exec()
		if err != nil {
			err = batchResult.Close()
			if err != nil {
				return nil, err
			}
			return nil, err
		}
	}

	err = batchResult.Close()
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	// Log success
	slog.DebugContext(ctx, "data successfully written to the database")
	// Return snapshot token
	return snapshot.NewToken(xid).Encode(), nil
}

// delete handles the deletion of tuples and attributes from the database based on provided filters.
// It returns an EncodedSnapToken upon successful deletion or an error if the deletion fails.
func (w *DataWriter) delete(
	ctx context.Context,
	tenantID string,
	tupleFilter *base.TupleFilter,
	attributeFilter *base.AttributeFilter,
) (token token.EncodedSnapToken, err error) {
	var tx pgx.Tx
	tx, err = w.database.WritePool.BeginTx(ctx, w.txOptions)
	if err != nil {
		return nil, err
	}
	// Defer rollback
	defer func() {
		_ = tx.Rollback(ctx)
	}()
	// Get transaction ID
	var xid db.XID8
	err = tx.QueryRow(ctx, utils.TransactionTemplate, tenantID).Scan(&xid)
	if err != nil {
		return nil, err
	}

	slog.DebugContext(ctx, "retrieved transaction", slog.Any("xid", xid), "for tenant", slog.Any("tenant_id", tenantID))

	slog.DebugContext(ctx, "processing tuple and executing update query")
	// Process tuple filter
	if !validation.IsTupleFilterEmpty(tupleFilter) {
		tbuilder := w.database.Builder.Update(RelationTuplesTable).Set("expired_tx_id", xid).Where(squirrel.Eq{"expired_tx_id": utils.ActiveRecordTxnID, "tenant_id": tenantID})
		tbuilder = utils.TuplesFilterQueryForUpdateBuilder(tbuilder, tupleFilter)

		var tquery string
		var targs []interface{}

		tquery, targs, err = tbuilder.ToSql()
		if err != nil {
			return nil, err
		}

		_, err = tx.Exec(ctx, tquery, targs...)
		if err != nil {
			return nil, err
		}
	}

	slog.DebugContext(ctx, "processing attribute and executing update query")
	// Process attribute filter
	if !validation.IsAttributeFilterEmpty(attributeFilter) {
		abuilder := w.database.Builder.Update(AttributesTable).Set("expired_tx_id", xid).Where(squirrel.Eq{"expired_tx_id": utils.ActiveRecordTxnID, "tenant_id": tenantID})
		abuilder = utils.AttributesFilterQueryForUpdateBuilder(abuilder, attributeFilter)
		// Declare query variables
		var aquery string
		var aargs []interface{}
		// Generate SQL query
		aquery, aargs, err = abuilder.ToSql()
		if err != nil {
			return nil, err
		}

		_, err = tx.Exec(ctx, aquery, aargs...)
		if err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	slog.DebugContext(ctx, "data successfully deleted from the database")
	// Return snapshot token
	return snapshot.NewToken(xid).Encode(), nil
} // End Delete
// Helper functions for RunBundle
// RunBundle helper function
// RunBundle implementation
// runBundle executes a series of operations defined in a DataBundle within a single database transaction.
// It returns an EncodedSnapToken upon successful execution of all operations or an error if any operation fails.
func (w *DataWriter) runBundle(
	ctx context.Context,
	tenantID string,
	arguments map[string]string,
	b *base.DataBundle,
) (token token.EncodedSnapToken, err error) {
	var tx pgx.Tx
	tx, err = w.database.WritePool.BeginTx(ctx, w.txOptions)
	if err != nil {
		return nil, err
	}
	// Defer rollback
	defer func() {
		_ = tx.Rollback(ctx)
	}()
	// Get transaction ID
	var xid db.XID8
	err = tx.QueryRow(ctx, utils.TransactionTemplate, tenantID).Scan(&xid)
	if err != nil {
		return nil, err
	}

	slog.DebugContext(ctx, "retrieved transaction", slog.Any("xid", xid), "for tenant", slog.Any("tenant_id", tenantID))
	// Create batch for operations
	batch := &pgx.Batch{}

	for _, op := range b.GetOperations() {
		tb, ab, err := bundle.Operation(arguments, op)
		if err != nil {
			return nil, err
		}

		err = w.runOperation(batch, xid, tenantID, tb, ab)
		if err != nil {
			return nil, err
		}
	}
	// Send batch to database
	batchResult := tx.SendBatch(ctx, batch)
	for i := 0; i < batch.Len(); i++ {
		_, err = batchResult.Exec()
		if err != nil {
			err = batchResult.Close()
			if err != nil {
				return nil, err
			}
			return nil, err
		}
	}

	err = batchResult.Close()
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	// Return snapshot token
	return snapshot.NewToken(xid).Encode(), nil
}

// runOperation processes and executes database operations defined in TupleBundle and AttributeBundle within a given transaction.
func (w *DataWriter) runOperation(
	batch *pgx.Batch,
	xid db.XID8,
	tenantID string,
	tb database.TupleBundle,
	ab database.AttributeBundle,
) (err error) {
	slog.Debug("processing bundles queries")
	if len(tb.Write.GetTuples()) > 0 {
		err = w.batchInsertRelationships(batch, xid, tenantID, &tb.Write)
		if err != nil {
			return err
		}
	}

	if len(ab.Write.GetAttributes()) > 0 {
		deleteClauses := buildDeleteClausesForAttributes(&ab.Write)
		err = w.batchUpdateAttributes(batch, xid, tenantID, deleteClauses)
		if err != nil {
			return err
		}

		err = w.batchInsertAttributes(batch, xid, tenantID, &ab.Write)
		if err != nil {
			return err
		}
	}

	if len(tb.Delete.GetTuples()) > 0 {
		deleteClauses := buildDeleteClausesForRelationships(&tb.Delete)
		err = w.batchUpdateRelationships(batch, xid, tenantID, deleteClauses)
		if err != nil {
			return err
		}
	}

	if len(ab.Delete.GetAttributes()) > 0 {
		deleteClauses := buildDeleteClausesForAttributes(&ab.Delete)
		err = w.batchUpdateAttributes(batch, xid, tenantID, deleteClauses)
		if err != nil {
			return err
		}
	}

	return nil
} // End runBundle
// Batch operations helper functions
// Batch insert helper functions
// Batch insert relationships
// Batch insert relationships to database
func (w *DataWriter) batchInsertRelationships(batch *pgx.Batch, xid db.XID8, tenantID string, tupleCollection *database.TupleCollection) error { // Insert relationships in batch
	titer := tupleCollection.CreateTupleIterator() // Create iterator
	for titer.HasNext() {                          // Iterate tuples
		t := titer.GetNext()                      // Get next tuple
		srelation := t.GetSubject().GetRelation() // Get subject relation
		if srelation == tuple.ELLIPSIS {          // Check ellipsis
			srelation = "" // Reset relation
		} // End of ellipsis check
		batch.Queue( // Queue insert
			"INSERT INTO relation_tuples (entity_type, entity_id, relation, subject_type, subject_id, subject_relation, created_tx_id, tenant_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) ON CONFLICT ON CONSTRAINT uq_relation_tuple_not_expired DO NOTHING", // Insert query
			t.GetEntity().GetType(), t.GetEntity().GetId(), t.GetRelation(), t.GetSubject().GetType(), t.GetSubject().GetId(), srelation, xid, tenantID, // Values
		) // End of queue
	} // End of iteration
	return nil // Success
} // End of batchInsertRelationships

// Batch update relationships
// Batch update relationships in database
func (w *DataWriter) batchUpdateRelationships(batch *pgx.Batch, xid db.XID8, tenantID string, deleteClauses []squirrel.Eq) error { // Update relationships in batch
	for _, condition := range deleteClauses { // Iterate delete clauses
		query, args, err := w.database.Builder.Update(RelationTuplesTable). // Build update query
											Set("expired_tx_id", xid).                                                           // Set expiry
											Where(squirrel.Eq{"expired_tx_id": utils.ActiveRecordTxnID, "tenant_id": tenantID}). // Where active
											Where(condition).                                                                    // Where condition
											ToSql()                                                                              // Generate SQL
		if err != nil { // Check SQL generation error
			return err // Return error
		} // End of error check
		batch.Queue(query, args...) // Queue update
	} // End of iteration
	return nil // Success
} // End of batchUpdateRelationships

// Build delete clauses for relationships
// Build delete clauses for relationship tuples
func buildDeleteClausesForRelationships(tupleCollection *database.TupleCollection) []squirrel.Eq { // Build delete clauses
	deleteClauses := make([]squirrel.Eq, 0) // Initialize clauses
	// Iterate tuples
	titer := tupleCollection.CreateTupleIterator() // Create iterator
	for titer.HasNext() {                          // Iterate tuples
		t := titer.GetNext()                      // Get next tuple
		srelation := t.GetSubject().GetRelation() // Get subject relation
		if srelation == tuple.ELLIPSIS {          // Check ellipsis
			srelation = "" // Reset relation
		} // End of ellipsis check
		// Build condition
		condition := squirrel.Eq{ // Create condition
			"entity_type":      t.GetEntity().GetType(),  // Entity type
			"entity_id":        t.GetEntity().GetId(),    // Entity ID
			"relation":         t.GetRelation(),          // Relation
			"subject_type":     t.GetSubject().GetType(), // Subject type
			"subject_id":       t.GetSubject().GetId(),   // Subject ID
			"subject_relation": srelation,                // Subject relation
		} // End of condition
		// Append to clauses
		deleteClauses = append(deleteClauses, condition) // Append condition
	} // End of iteration
	// Return clauses
	return deleteClauses // Return delete clauses
} // End of buildDeleteClausesForRelationships

// Batch insert attributes
// Batch insert attributes to database
func (w *DataWriter) batchInsertAttributes(batch *pgx.Batch, xid db.XID8, tenantID string, attributeCollection *database.AttributeCollection) error { // Insert attributes in batch
	aiter := attributeCollection.CreateAttributeIterator() // Create iterator
	for aiter.HasNext() {                                  // Iterate attributes
		a := aiter.GetNext() // Get next attribute
		// Marshal attribute value to JSON
		jsonBytes, err := protojson.Marshal(a.GetValue()) // Marshal to JSON
		if err != nil {                                   // Check marshal error
			return err // Return error
		} // End of marshal error check
		// Convert to string
		jsonStr := string(jsonBytes) // Convert to string
		// Queue insert
		batch.Queue( // Queue insert
			"INSERT INTO attributes (entity_type, entity_id, attribute, value, created_tx_id, tenant_id) VALUES ($1, $2, $3, $4, $5, $6)", // Insert query
			a.GetEntity().GetType(), a.GetEntity().GetId(), a.GetAttribute(), jsonStr, xid, tenantID, // Values
		) // End of queue
	} // End of iteration
	return nil // Success
} // End of batchInsertAttributes

// Batch update attributes
// Batch update attributes in database
func (w *DataWriter) batchUpdateAttributes(batch *pgx.Batch, xid db.XID8, tenantID string, deleteClauses []squirrel.Eq) error { // Update attributes in batch
	for _, condition := range deleteClauses { // Iterate delete clauses
		query, args, err := w.database.Builder.Update(AttributesTable). // Build update query
										Set("expired_tx_id", xid).                                                           // Set expiry
										Where(squirrel.Eq{"expired_tx_id": utils.ActiveRecordTxnID, "tenant_id": tenantID}). // Where active
										Where(condition).                                                                    // Where condition
										ToSql()                                                                              // Generate SQL
		if err != nil { // Check SQL generation error
			return err // Return error
		} // End of error check
		// Queue update
		batch.Queue(query, args...) // Queue update
	} // End of iteration
	return nil // Success
} // End of batchUpdateAttributes

// Build delete clauses for attributes
// Build delete clauses for attribute records
func buildDeleteClausesForAttributes(attributeCollection *database.AttributeCollection) []squirrel.Eq { // Build delete clauses
	deleteClauses := make([]squirrel.Eq, 0) // Initialize clauses
	// Iterate attributes
	aiter := attributeCollection.CreateAttributeIterator() // Create iterator
	for aiter.HasNext() {                                  // Iterate attributes
		a := aiter.GetNext() // Get next attribute
		// Build condition
		condition := squirrel.Eq{ // Create condition
			"entity_type": a.GetEntity().GetType(), // Entity type
			"entity_id":   a.GetEntity().GetId(),   // Entity ID
			"attribute":   a.GetAttribute(),        // Attribute name
		} // End of condition
		// Append to clauses
		deleteClauses = append(deleteClauses, condition) // Append condition
	} // End of iteration
	// Return clauses
	return deleteClauses // Return delete clauses
} // End of buildDeleteClausesForAttributes
