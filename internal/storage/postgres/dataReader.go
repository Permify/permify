package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/golang/protobuf/jsonpb"
	"google.golang.org/protobuf/types/known/anypb"

	"go.opentelemetry.io/otel/codes"

	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/internal/storage/postgres/snapshot"
	"github.com/Permify/permify/internal/storage/postgres/types"
	"github.com/Permify/permify/internal/storage/postgres/utils"
	"github.com/Permify/permify/pkg/database"
	db "github.com/Permify/permify/pkg/database/postgres"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/token"
)

// DataReader is a struct which holds a reference to the database, transaction options and a logger.
// It is responsible for reading data from the database.
type DataReader struct {
	database  *db.Postgres  // database is an instance of the PostgreSQL database
	txOptions sql.TxOptions // txOptions specifies the isolation level for database transaction and sets it as read only
}

// NewDataReader is a constructor function for DataReader.
// It initializes a new DataReader with a given database, a logger, and sets transaction options to be read-only with Repeatable Read isolation level.
func NewDataReader(database *db.Postgres) *DataReader {
	return &DataReader{
		database:  database,                                                          // Set the database to the passed in PostgreSQL instance
		txOptions: sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true}, // Set the transaction options
	}
}

// QueryRelationships reads relation tuples from the storage based on the given filter.
func (r *DataReader) QueryRelationships(ctx context.Context, tenantID string, filter *base.TupleFilter, snap string) (it *database.TupleIterator, err error) {
	// Start a new trace span and end it when the function exits.
	ctx, span := tracer.Start(ctx, "data-reader.query-relationships")
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
	defer utils.Rollback(tx)

	// Build the relationships query based on the provided filter and snapshot value.
	var args []interface{}
	builder := r.database.Builder.Select("entity_type, entity_id, relation, subject_type, subject_id, subject_relation").From(RelationTuplesTable).Where(squirrel.Eq{"tenant_id": tenantID})
	builder = utils.TuplesFilterQueryForSelectBuilder(builder, filter)
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

// ReadRelationships reads relation tuples from the storage based on the given filter and pagination.
func (r *DataReader) ReadRelationships(ctx context.Context, tenantID string, filter *base.TupleFilter, snap string, pagination database.Pagination) (collection *database.TupleCollection, ct database.EncodedContinuousToken, err error) {
	// Start a new trace span and end it when the function exits.
	ctx, span := tracer.Start(ctx, "data-reader.read-relationships")
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
	defer utils.Rollback(tx)

	// Build the relationships query based on the provided filter, snapshot value, and pagination settings.
	builder := r.database.Builder.Select("id, entity_type, entity_id, relation, subject_type, subject_id, subject_relation").From(RelationTuplesTable).Where(squirrel.Eq{"tenant_id": tenantID})
	builder = utils.TuplesFilterQueryForSelectBuilder(builder, filter)
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
		return nil, database.NewNoopContinuousToken().Encode(), errors.New(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String())
	}

	// Execute the query and retrieve the rows.
	var rows *sql.Rows
	rows, err = tx.QueryContext(ctx, query, args...)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, database.NewNoopContinuousToken().Encode(), errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
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

	return database.NewTupleCollection(tuples...), database.NewNoopContinuousToken().Encode(), nil
}

// QuerySingleAttribute retrieves a single attribute from the storage based on the given filter.
func (r *DataReader) QuerySingleAttribute(ctx context.Context, tenantID string, filter *base.AttributeFilter, snap string) (attribute *base.Attribute, err error) {
	// Start a new trace span and end it when the function exits.
	ctx, span := tracer.Start(ctx, "data-reader.query-single-attribute")
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
	defer utils.Rollback(tx)

	// Build the relationships query based on the provided filter and snapshot value.
	var args []interface{}
	builder := r.database.Builder.Select("entity_type, entity_id, attribute, value").From(AttributesTable).Where(squirrel.Eq{"tenant_id": tenantID})
	builder = utils.AttributesFilterQueryForSelectBuilder(builder, filter)
	builder = utils.SnapshotQuery(builder, st.(snapshot.Token).Value.Uint)

	// Generate the SQL query and arguments.
	var query string
	query, args, err = builder.ToSql()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, errors.New(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String())
	}

	row := tx.QueryRowContext(ctx, query, args...)

	rt := storage.Attribute{}

	// Suppose you have a struct `rt` with a field `Value` of type `*anypb.Any`.
	var valueStr string

	// Scan the row from the database into the fields of `rt` and `valueStr`.
	err = row.Scan(&rt.EntityType, &rt.EntityID, &rt.Attribute, &valueStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		} else {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, err
		}
	}

	// Unmarshal the JSON data from `valueStr` into `rt.Value`.
	rt.Value = &anypb.Any{}
	unmarshaler := &jsonpb.Unmarshaler{}
	err = unmarshaler.Unmarshal(strings.NewReader(valueStr), rt.Value)
	if err != nil {
		return nil, err
	}

	// Commit the transaction.
	err = tx.Commit()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return rt.ToAttribute(), nil
}

// QueryAttributes reads multiple attributes from the storage based on the given filter.
func (r *DataReader) QueryAttributes(ctx context.Context, tenantID string, filter *base.AttributeFilter, snap string) (it *database.AttributeIterator, err error) {
	// Start a new trace span and end it when the function exits.
	ctx, span := tracer.Start(ctx, "data-reader.query-attributes")
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
	defer utils.Rollback(tx)

	// Build the relationships query based on the provided filter and snapshot value.
	var args []interface{}
	builder := r.database.Builder.Select("entity_type, entity_id, attribute, value").From(AttributesTable).Where(squirrel.Eq{"tenant_id": tenantID})
	builder = utils.AttributesFilterQueryForSelectBuilder(builder, filter)
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
	collection := database.NewAttributeCollection()
	for rows.Next() {
		rt := storage.Attribute{}

		// Suppose you have a struct `rt` with a field `Value` of type `*anypb.Any`.
		var valueStr string

		// Scan the row from the database into the fields of `rt` and `valueStr`.
		err := rows.Scan(&rt.EntityType, &rt.EntityID, &rt.Attribute, &valueStr)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, err
		}

		// Unmarshal the JSON data from `valueStr` into `rt.Value`.
		rt.Value = &anypb.Any{}
		unmarshaler := &jsonpb.Unmarshaler{}
		err = unmarshaler.Unmarshal(strings.NewReader(valueStr), rt.Value)
		if err != nil {
			return nil, err
		}

		collection.Add(rt.ToAttribute())
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
	return collection.CreateAttributeIterator(), nil
}

// ReadAttributes reads multiple attributes from the storage based on the given filter and pagination.
func (r *DataReader) ReadAttributes(ctx context.Context, tenantID string, filter *base.AttributeFilter, snap string, pagination database.Pagination) (collection *database.AttributeCollection, ct database.EncodedContinuousToken, err error) {
	// Start a new trace span and end it when the function exits.
	ctx, span := tracer.Start(ctx, "data-reader.read-attributes")
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
	defer utils.Rollback(tx)

	// Build the relationships query based on the provided filter, snapshot value, and pagination settings.
	builder := r.database.Builder.Select("id, entity_type, entity_id, attribute, value").From(AttributesTable).Where(squirrel.Eq{"tenant_id": tenantID})
	builder = utils.AttributesFilterQueryForSelectBuilder(builder, filter)
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
		return nil, database.NewNoopContinuousToken().Encode(), errors.New(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String())
	}

	// Execute the query and retrieve the rows.
	var rows *sql.Rows
	rows, err = tx.QueryContext(ctx, query, args...)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, database.NewNoopContinuousToken().Encode(), errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}
	defer rows.Close()

	var lastID uint64

	// Iterate through the rows and scan the result into a RelationTuple struct.
	attributes := make([]*base.Attribute, 0, pagination.PageSize()+1)
	for rows.Next() {
		rt := storage.Attribute{}

		// Suppose you have a struct `rt` with a field `Value` of type `*anypb.Any`.
		var valueStr string

		// Scan the row from the database into the fields of `rt` and `valueStr`.
		err := rows.Scan(&rt.ID, &rt.EntityType, &rt.EntityID, &rt.Attribute, &valueStr)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, nil, err
		}
		lastID = rt.ID

		// Unmarshal the JSON data from `valueStr` into `rt.Value`.
		rt.Value = &anypb.Any{}
		unmarshaler := &jsonpb.Unmarshaler{}
		err = unmarshaler.Unmarshal(strings.NewReader(valueStr), rt.Value)
		if err != nil {
			return nil, nil, err
		}

		attributes = append(attributes, rt.ToAttribute())
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
	if len(attributes) > int(pagination.PageSize()) {
		return database.NewAttributeCollection(attributes[:pagination.PageSize()]...), utils.NewContinuousToken(strconv.FormatUint(lastID, 10)).Encode(), nil
	}

	return database.NewAttributeCollection(attributes...), database.NewNoopContinuousToken().Encode(), nil
}

// QueryUniqueEntities reads unique entities from the storage based on the given filter and pagination.
func (r *DataReader) QueryUniqueEntities(ctx context.Context, tenantID, name, snap string, pagination database.Pagination) (ids []string, ct database.EncodedContinuousToken, err error) {
	// Start a new trace span and end it when the function exits.
	ctx, span := tracer.Start(ctx, "data-reader.query-unique-entities")
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
	defer utils.Rollback(tx)

	query := utils.BulkEntityFilterQuery(tenantID, name, st.(snapshot.Token).Value.Uint)

	// Apply the pagination token and limit to the subQuery.
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

		query = fmt.Sprintf("%s WHERE id >= %s", query, strconv.FormatUint(v, 10))
	}

	// Append ORDER BY and LIMIT clauses.
	query = fmt.Sprintf("%s ORDER BY id LIMIT %d", query, pagination.PageSize()+1)

	// Execute the query and retrieve the rows.
	var rows *sql.Rows
	rows, err = tx.QueryContext(ctx, query)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, database.NewNoopContinuousToken().Encode(), errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}
	defer rows.Close()

	var lastID uint64

	// Iterate through the rows and scan the result into a RelationTuple struct.
	entityIDs := make([]string, 0, pagination.PageSize()+1)
	for rows.Next() {
		var entityId string
		err = rows.Scan(&lastID, &entityId)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, nil, err
		}

		entityIDs = append(entityIDs, entityId)
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
	if len(entityIDs) > int(pagination.PageSize()) {
		return entityIDs[:pagination.PageSize()], utils.NewContinuousToken(strconv.FormatUint(lastID, 10)).Encode(), nil
	}

	return entityIDs, database.NewNoopContinuousToken().Encode(), nil
}

// QueryUniqueSubjectReferences reads unique subject references from the storage based on the given filter and pagination.
func (r *DataReader) QueryUniqueSubjectReferences(ctx context.Context, tenantID string, subjectReference *base.RelationReference, snap string, pagination database.Pagination) (ids []string, ct database.EncodedContinuousToken, err error) {
	// Start a new trace span and end it when the function exits.
	ctx, span := tracer.Start(ctx, "data-reader.query-unique-subject-reference")
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
	defer utils.Rollback(tx)

	// Build the relationships query based on the provided filter, snapshot value, and pagination settings.
	builder := r.database.Builder.
		Select("MIN(id) as id, subject_id"). // This will pick the smallest `id` for each unique `subject_id`.
		From(RelationTuplesTable).
		Where(squirrel.Eq{"tenant_id": tenantID}).
		GroupBy("subject_id")
	builder = utils.TuplesFilterQueryForSelectBuilder(builder, &base.TupleFilter{Subject: &base.SubjectFilter{Type: subjectReference.GetType(), Relation: subjectReference.GetRelation()}})
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
		return nil, database.NewNoopContinuousToken().Encode(), errors.New(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String())
	}

	// Execute the query and retrieve the rows.
	var rows *sql.Rows
	rows, err = tx.QueryContext(ctx, query, args...)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, database.NewNoopContinuousToken().Encode(), errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}
	defer rows.Close()

	var lastID uint64

	// Iterate through the rows and scan the result into a RelationTuple struct.
	subjectIDs := make([]string, 0, pagination.PageSize()+1)
	for rows.Next() {
		var subjectID string
		err = rows.Scan(&lastID, &subjectID)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, nil, err
		}
		subjectIDs = append(subjectIDs, subjectID)
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
	if len(subjectIDs) > int(pagination.PageSize()) {
		return subjectIDs[:pagination.PageSize()], utils.NewContinuousToken(strconv.FormatUint(lastID, 10)).Encode(), nil
	}

	return subjectIDs, database.NewNoopContinuousToken().Encode(), nil
}

// HeadSnapshot retrieves the latest snapshot token associated with the tenant.
func (r *DataReader) HeadSnapshot(ctx context.Context, tenantID string) (token.SnapToken, error) {
	// Start a new trace span and end it when the function exits.
	ctx, span := tracer.Start(ctx, "data-reader.head-snapshot")
	defer span.End()

	var xid types.XID8

	// Build the query to find the highest transaction ID associated with the tenant.
	builder := r.database.Builder.Select("id").From(TransactionsTable).Where(squirrel.Eq{"tenant_id": tenantID}).OrderBy("id DESC").Limit(1)
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
