package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strconv"

	"github.com/Masterminds/squirrel"

	"go.opentelemetry.io/otel/codes"

	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/internal/storage/postgres/snapshot"
	"github.com/Permify/permify/internal/storage/postgres/types"
	"github.com/Permify/permify/internal/storage/postgres/utils"
	"github.com/Permify/permify/pkg/database"
	db "github.com/Permify/permify/pkg/database/postgres"
	"github.com/Permify/permify/pkg/logger"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/token"
)

// RelationshipReader is a structure that holds information and dependencies
// required for reading relationship data from the database.
type RelationshipReader struct {
	// database is a pointer to a Postgres database instance, which is used
	// to perform operations on the relationship data.
	database *db.Postgres

	// txOptions holds the configuration for database transactions, such as
	// isolation level and read-only mode, to be applied when performing
	// operations on the relationship data.
	txOptions sql.TxOptions

	// logger is an instance of a logger that implements the logger.Interface
	// and is used to log messages related to the operations performed by
	// the RelationshipReader.
	logger logger.Interface
}

// NewRelationshipReader creates a new instance of the RelationshipReader struct
// with the given database and logger instances. It also sets the default transaction
// options for the RelationshipReader.
//
// Parameters:
//   - database: A pointer to a Postgres database instance, which will be used
//     to perform operations on the relationship data.
//   - logger:   An instance of a logger that implements the logger.Interface, which
//     will be used to log messages related to the operations performed by
//     the RelationshipReader.
//
// Returns:
//   - A pointer to a new RelationshipReader instance, initialized with the given
//     database and logger instances, and the default transaction options.
func NewRelationshipReader(database *db.Postgres, logger logger.Interface) *RelationshipReader {
	return &RelationshipReader{
		database:  database,
		txOptions: sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true},
		logger:    logger,
	}
}

// QueryRelationships retrieves relationships from the database based on a given filter,
// tenant ID, and snapshot value. It returns a TupleIterator containing the filtered results.
//
// Parameters:
//   - ctx:       The context used for tracing and cancellation.
//   - tenantID:  The tenant ID for which the relationships should be queried.
//   - filter:    A pointer to a TupleFilter struct that defines the filtering criteria
//     for the relationships query.
//   - snap:      A string representing the snapshot value to be used for the query.
//
// Returns:
// - it:        A pointer to a TupleIterator containing the filtered relationships.
// - err:       An error, if any occurred during the execution of the query.
func (r *RelationshipReader) QueryRelationships(ctx context.Context, tenantID string, filter *base.TupleFilter, snap string) (it *database.TupleIterator, err error) {
	// Start a new trace span and end it when the function exits.
	ctx, span := tracer.Start(ctx, "relationship-reader.query-relationships")
	defer span.End()

	// Decode the snapshot value.
	var st token.SnapToken
	st, err = snapshot.EncodedToken{Value: snap}.Decode()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	// Begin a new read-only transaction with the specified isolation level.
	var tx *sql.Tx
	tx, err = r.database.DB.BeginTx(ctx, &r.txOptions)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	// Rollback the transaction in case of any error.
	defer utils.Rollback(tx, r.logger)

	// Build the relationships query based on the provided filter and snapshot value.
	var args []interface{}
	builder := r.database.Builder.Select("entity_type, entity_id, relation, subject_type, subject_id, subject_relation").From(RelationTuplesTable).Where(squirrel.Eq{"tenant_id": tenantID})
	builder = utils.FilterQueryForSelectBuilder(builder, filter)
	builder = utils.SnapshotQuery(builder, st.(snapshot.Token).Value.Uint)

	// Generate the SQL query and arguments.
	var query string
	query, args, err = builder.ToSql()

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, errors.New(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String())
	}

	// Execute the SQL query and retrieve the result rows.
	var rows *sql.Rows
	rows, err = tx.QueryContext(ctx, query, args...)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}
	defer rows.Close()

	// Process the result rows and store the relationships in a TupleCollection.
	collection := database.NewTupleCollection()
	for rows.Next() {
		rt := storage.RelationTuple{}
		err = rows.Scan(&rt.EntityType, &rt.EntityID, &rt.Relation, &rt.SubjectType, &rt.SubjectID, &rt.SubjectRelation)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, err
		}
		collection.Add(rt.ToTuple())
	}
	if err = rows.Err(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	// Commit the transaction.
	err = tx.Commit()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	// Return a TupleIterator created from the TupleCollection.
	return collection.CreateTupleIterator(), nil
}

// ReadRelationships retrieves relationships from the database based on a given filter,
// tenant ID, snapshot value, and pagination settings. It returns a TupleCollection
// containing the filtered results and an encoded continuous token for pagination.
//
// Parameters:
//   - ctx:        The context used for tracing and cancellation.
//   - tenantID:   The tenant ID for which the relationships should be queried.
//   - filter:     A pointer to a TupleFilter struct that defines the filtering criteria
//     for the relationships query.
//   - snap:       A string representing the snapshot value to be used for the query.
//   - pagination: A Pagination struct containing the page size and token for the query.
//
// Returns:
// - collection: A pointer to a TupleCollection containing the filtered relationships.
// - ct:         An EncodedContinuousToken representing the next token for pagination.
// - err:        An error, if any occurred during the execution of the query.
func (r *RelationshipReader) ReadRelationships(ctx context.Context, tenantID string, filter *base.TupleFilter, snap string, pagination database.Pagination) (collection *database.TupleCollection, ct database.EncodedContinuousToken, err error) {
	// Start a new trace span and end it when the function exits.
	ctx, span := tracer.Start(ctx, "relationship-reader.read-relationships")
	defer span.End()

	// Decode the snapshot value.
	var st token.SnapToken
	st, err = snapshot.EncodedToken{Value: snap}.Decode()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, nil, err
	}

	// Begin a new read-only transaction with the specified isolation level.
	var tx *sql.Tx
	tx, err = r.database.DB.BeginTx(ctx, &r.txOptions)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, nil, err
	}

	// Rollback the transaction in case of any error.
	defer utils.Rollback(tx, r.logger)

	// Build the relationships query based on the provided filter, snapshot value, and pagination settings.
	builder := r.database.Builder.Select("id, entity_type, entity_id, relation, subject_type, subject_id, subject_relation").From(RelationTuplesTable).Where(squirrel.Eq{"tenant_id": tenantID})
	builder = utils.FilterQueryForSelectBuilder(builder, filter)
	builder = utils.SnapshotQuery(builder, st.(snapshot.Token).Value.Uint)

	// Apply the pagination token and limit to the query.
	if pagination.Token() != "" {
		var t database.ContinuousToken
		t, err = utils.EncodedContinuousToken{Value: pagination.Token()}.Decode()
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, nil, err
		}
		var v uint64
		v, err = strconv.ParseUint(t.(utils.ContinuousToken).Value, 10, 64)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, nil, errors.New(base.ErrorCode_ERROR_CODE_INVALID_CONTINUOUS_TOKEN.String())
		}
		builder = builder.Where(squirrel.GtOrEq{"id": v})
	}

	builder = builder.OrderBy("id").Limit(uint64(pagination.PageSize() + 1))

	// Generate the SQL query and arguments.
	var query string
	var args []interface{}
	query, args, err = builder.ToSql()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, utils.NewNoopContinuousToken().Encode(), errors.New(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String())
	}

	// Execute the query and retrieve the rows.
	var rows *sql.Rows
	rows, err = tx.QueryContext(ctx, query, args...)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, utils.NewNoopContinuousToken().Encode(), errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}
	defer rows.Close()

	var lastID uint64

	// Iterate through the rows and scan the result into a RelationTuple struct.
	tuples := make([]*base.Tuple, 0, pagination.PageSize()+1)
	for rows.Next() {
		rt := storage.RelationTuple{}
		err = rows.Scan(&rt.ID, &rt.EntityType, &rt.EntityID, &rt.Relation, &rt.SubjectType, &rt.SubjectID, &rt.SubjectRelation)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, nil, err
		}
		lastID = rt.ID
		tuples = append(tuples, rt.ToTuple())
	}
	// Check for any errors during iteration.
	if err = rows.Err(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, nil, err
	}

	// Commit the transaction.
	err = tx.Commit()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, nil, err
	}

	// Return the results and encoded continuous token for pagination.
	if len(tuples) > int(pagination.PageSize()) {
		return database.NewTupleCollection(tuples[:pagination.PageSize()]...), utils.NewContinuousToken(strconv.FormatUint(lastID, 10)).Encode(), nil
	}

	return database.NewTupleCollection(tuples...), utils.NewNoopContinuousToken().Encode(), nil
}

// HeadSnapshot retrieves the latest snapshot token for a given tenant ID.
// It queries the transaction table to find the highest transaction ID associated with the tenant.
//
// Parameters:
// - ctx:      The context used for tracing and cancellation.
// - tenantID: The tenant ID for which the latest snapshot token should be retrieved.
//
// Returns:
// - token.SnapToken: The latest snapshot token associated with the tenant.
// - error:           An error, if any occurred during the execution of the query.
func (r *RelationshipReader) HeadSnapshot(ctx context.Context, tenantID string) (token.SnapToken, error) {
	// Start a new trace span and end it when the function exits.
	ctx, span := tracer.Start(ctx, "relationship-reader.head-snapshot")
	defer span.End()

	var xid types.XID8

	// Build the query to find the highest transaction ID associated with the tenant.
	builder := r.database.Builder.Select("MAX(id)").From(TransactionsTable).Where(squirrel.Eq{"tenant_id": tenantID})
	query, args, err := builder.ToSql()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, errors.New(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String())
	}

	// Execute the query and retrieve the highest transaction ID.
	row := r.database.DB.QueryRowContext(ctx, query, args...)
	err = row.Scan(&xid)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		// If no rows are found, return a snapshot token with a value of 0.
		if errors.Is(err, sql.ErrNoRows) {
			return snapshot.Token{Value: types.XID8{Uint: 0}}, nil
		}
		return nil, err
	}

	// Return the latest snapshot token associated with the tenant.
	return snapshot.Token{Value: xid}, nil
}
