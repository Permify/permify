package postgres

import (
	"context"
	"errors"
	"log/slog"
	"strconv"

	"github.com/jackc/pgx/v5"

	"github.com/Masterminds/squirrel"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/Permify/permify/internal"
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
	txOptions pgx.TxOptions // txOptions specifies the isolation level for database transaction and sets it as read only
}

// NewDataReader is a constructor function for DataReader.
// It initializes a new DataReader with a given database, a logger, and sets transaction options to be read-only with Repeatable Read isolation level.
func NewDataReader(database *db.Postgres) *DataReader {
	return &DataReader{
		database:  database,                                                             // Set the database to the passed in PostgreSQL instance
		txOptions: pgx.TxOptions{IsoLevel: pgx.ReadCommitted, AccessMode: pgx.ReadOnly}, // Set the transaction options
	}
}

// QueryRelationships reads relation tuples from the storage based on the given filter.
func (r *DataReader) QueryRelationships(ctx context.Context, tenantID string, filter *base.TupleFilter, snap string, pagination database.CursorPagination) (it *database.TupleIterator, err error) {
	// Start a new trace span and end it when the function exits.
	ctx, span := internal.Tracer.Start(ctx, "data-reader.query-relationships")
	defer span.End()

	slog.DebugContext(ctx, "querying relationships for tenant_id", slog.String("tenant_id", tenantID))

	// Decode the snapshot value.
	var st token.SnapToken
	st, err = snapshot.EncodedToken{Value: snap}.Decode()
	if err != nil {
		return nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_INTERNAL)
	}

	// Build the relationships query based on the provided filter and snapshot value.
	var args []interface{}
	builder := r.database.Builder.Select("entity_type, entity_id, relation, subject_type, subject_id, subject_relation").From(RelationTuplesTable).Where(squirrel.Eq{"tenant_id": tenantID})
	builder = utils.TuplesFilterQueryForSelectBuilder(builder, filter)
	builder = utils.SnapshotQuery(builder, st.(snapshot.Token).Value.Uint)

	if pagination.Cursor() != "" {
		var t database.ContinuousToken
		t, err = utils.EncodedContinuousToken{Value: pagination.Cursor()}.Decode()
		if err != nil {
			return nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_INVALID_CONTINUOUS_TOKEN)
		}
		builder = builder.Where(squirrel.GtOrEq{pagination.Sort(): t.(utils.ContinuousToken).Value})
	}

	if pagination.Sort() != "" {
		builder = builder.OrderBy(pagination.Sort())
	}

	// Apply limit if specified in pagination
	limit := pagination.Limit()
	if limit > 0 {
		builder = builder.Limit(uint64(limit))
	}

	// Generate the SQL query and arguments.
	var query string
	query, args, err = builder.ToSql()
	if err != nil {
		return nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_SQL_BUILDER)
	}

	slog.DebugContext(ctx, "generated sql query", slog.String("query", query), "with args", slog.Any("arguments", args))

	// Execute the SQL query and retrieve the result rows.
	var rows pgx.Rows
	rows, err = r.database.ReadPool.Query(ctx, query, args...)
	if err != nil {
		return nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_EXECUTION)
	}
	defer rows.Close()

	// Process the result rows and store the relationships in a TupleCollection.
	collection := database.NewTupleCollection()
	for rows.Next() {
		rt := storage.RelationTuple{}
		err = rows.Scan(&rt.EntityType, &rt.EntityID, &rt.Relation, &rt.SubjectType, &rt.SubjectID, &rt.SubjectRelation)
		if err != nil {
			return nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_SCAN)
		}
		collection.Add(rt.ToTuple())
	}
	if err = rows.Err(); err != nil {
		return nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_SCAN)
	}

	slog.DebugContext(ctx, "successfully retrieved relation tuples from the database")

	// Return a TupleIterator created from the TupleCollection.
	return collection.CreateTupleIterator(), nil
}

// ReadRelationships reads relation tuples from the storage based on the given filter and pagination.
func (r *DataReader) ReadRelationships(ctx context.Context, tenantID string, filter *base.TupleFilter, snap string, pagination database.Pagination) (collection *database.TupleCollection, ct database.EncodedContinuousToken, err error) {
	// Start a new trace span and end it when the function exits.
	ctx, span := internal.Tracer.Start(ctx, "data-reader.read-relationships")
	defer span.End()

	slog.DebugContext(ctx, "reading relationships for tenant_id", slog.String("tenant_id", tenantID))

	// Decode the snapshot value.
	var st token.SnapToken
	st, err = snapshot.EncodedToken{Value: snap}.Decode()
	if err != nil {
		return nil, nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_INTERNAL)
	}

	// Build the relationships query based on the provided filter, snapshot value, and pagination settings.
	builder := r.database.Builder.Select("id, entity_type, entity_id, relation, subject_type, subject_id, subject_relation").From(RelationTuplesTable).Where(squirrel.Eq{"tenant_id": tenantID})
	builder = utils.TuplesFilterQueryForSelectBuilder(builder, filter)
	builder = utils.SnapshotQuery(builder, st.(snapshot.Token).Value.Uint)

	// Apply the pagination token and limit to the query.
	if pagination.Token() != "" {
		var t database.ContinuousToken
		t, err = utils.EncodedContinuousToken{Value: pagination.Token()}.Decode()
		if err != nil {
			return nil, nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_INVALID_CONTINUOUS_TOKEN)
		}
		var v uint64
		v, err = strconv.ParseUint(t.(utils.ContinuousToken).Value, 10, 64)
		if err != nil {
			return nil, nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_INVALID_CONTINUOUS_TOKEN)
		}
		builder = builder.Where(squirrel.GtOrEq{"id": v})
	}

	builder = builder.OrderBy("id")

	if pagination.PageSize() != 0 {
		builder = builder.Limit(uint64(pagination.PageSize() + 1))
	}

	// Generate the SQL query and arguments.
	var query string
	var args []interface{}
	query, args, err = builder.ToSql()
	if err != nil {
		return nil, database.NewNoopContinuousToken().Encode(), utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_SQL_BUILDER)
	}

	slog.DebugContext(ctx, "generated sql query", slog.String("query", query), "with args", slog.Any("arguments", args))

	// Execute the query and retrieve the rows.
	var rows pgx.Rows
	rows, err = r.database.ReadPool.Query(ctx, query, args...)
	if err != nil {
		return nil, database.NewNoopContinuousToken().Encode(), utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_EXECUTION)
	}
	defer rows.Close()

	var lastID uint64

	// Iterate through the rows and scan the result into a RelationTuple struct.
	tuples := make([]*base.Tuple, 0, pagination.PageSize()+1)
	for rows.Next() {
		rt := storage.RelationTuple{}
		err = rows.Scan(&rt.ID, &rt.EntityType, &rt.EntityID, &rt.Relation, &rt.SubjectType, &rt.SubjectID, &rt.SubjectRelation)
		if err != nil {
			return nil, nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_SCAN)
		}
		lastID = rt.ID
		tuples = append(tuples, rt.ToTuple())
	}
	// Check for any errors during iteration.
	if err = rows.Err(); err != nil {
		return nil, nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_SCAN)
	}

	slog.DebugContext(ctx, "successfully read relation tuples from database")

	// Return the results and encoded continuous token for pagination.
	if pagination.PageSize() != 0 && len(tuples) > int(pagination.PageSize()) {
		return database.NewTupleCollection(tuples[:pagination.PageSize()]...), utils.NewContinuousToken(strconv.FormatUint(lastID, 10)).Encode(), nil
	}

	return database.NewTupleCollection(tuples...), database.NewNoopContinuousToken().Encode(), nil
}

// QuerySingleAttribute retrieves a single attribute from the storage based on the given filter.
func (r *DataReader) QuerySingleAttribute(ctx context.Context, tenantID string, filter *base.AttributeFilter, snap string) (attribute *base.Attribute, err error) {
	// Start a new trace span and end it when the function exits.
	ctx, span := internal.Tracer.Start(ctx, "data-reader.query-single-attribute")
	defer span.End()

	slog.DebugContext(ctx, "querying single attribute for tenant_id", slog.String("tenant_id", tenantID))

	// Decode the snapshot value.
	var st token.SnapToken
	st, err = snapshot.EncodedToken{Value: snap}.Decode()
	if err != nil {
		return nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_INTERNAL)
	}

	// Build the relationships query based on the provided filter and snapshot value.
	var args []interface{}
	builder := r.database.Builder.Select("entity_type, entity_id, attribute, value").From(AttributesTable).Where(squirrel.Eq{"tenant_id": tenantID})
	builder = utils.AttributesFilterQueryForSelectBuilder(builder, filter)
	builder = utils.SnapshotQuery(builder, st.(snapshot.Token).Value.Uint)

	// Generate the SQL query and arguments.
	var query string
	query, args, err = builder.ToSql()
	if err != nil {
		return nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_SQL_BUILDER)
	}

	slog.DebugContext(ctx, "generated sql query", slog.String("query", query), "with args", slog.Any("arguments", args))

	row := r.database.ReadPool.QueryRow(ctx, query, args...)

	rt := storage.Attribute{}

	// Suppose you have a struct `rt` with a field `Value` of type `*anypb.Any`.
	var valueStr string

	// Scan the row from the database into the fields of `rt` and `valueStr`.
	err = row.Scan(&rt.EntityType, &rt.EntityID, &rt.Attribute, &valueStr)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		} else {
			return nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_SCAN)
		}
	}

	// Unmarshal the JSON data from `valueStr` into `rt.Value`.
	rt.Value = &anypb.Any{}
	err = protojson.Unmarshal([]byte(valueStr), rt.Value)
	if err != nil {
		return nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_INTERNAL)
	}

	slog.DebugContext(ctx, "successfully retrieved Single attribute from the database")

	return rt.ToAttribute(), nil
}

// QueryAttributes reads multiple attributes from the storage based on the given filter.
func (r *DataReader) QueryAttributes(ctx context.Context, tenantID string, filter *base.AttributeFilter, snap string, pagination database.CursorPagination) (it *database.AttributeIterator, err error) {
	// Start a new trace span and end it when the function exits.
	ctx, span := internal.Tracer.Start(ctx, "data-reader.query-attributes")
	defer span.End()

	slog.DebugContext(ctx, "querying Attributes for tenant_id", slog.String("tenant_id", tenantID))

	// Decode the snapshot value.
	var st token.SnapToken
	st, err = snapshot.EncodedToken{Value: snap}.Decode()
	if err != nil {
		return nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_INTERNAL)
	}

	// Build the attributes query based on the provided filter and snapshot value.
	var args []interface{}
	builder := r.database.Builder.Select("entity_type, entity_id, attribute, value").From(AttributesTable).Where(squirrel.Eq{"tenant_id": tenantID})
	builder = utils.AttributesFilterQueryForSelectBuilder(builder, filter)
	builder = utils.SnapshotQuery(builder, st.(snapshot.Token).Value.Uint)

	if pagination.Cursor() != "" {
		var t database.ContinuousToken
		t, err = utils.EncodedContinuousToken{Value: pagination.Cursor()}.Decode()
		if err != nil {
			return nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_INVALID_CONTINUOUS_TOKEN)
		}
		builder = builder.Where(squirrel.GtOrEq{pagination.Sort(): t.(utils.ContinuousToken).Value})
	}

	if pagination.Sort() != "" {
		builder = builder.OrderBy(pagination.Sort())
	}

	// Apply limit if specified in pagination
	limit := pagination.Limit()
	if limit > 0 {
		builder = builder.Limit(uint64(limit))
	}

	// Generate the SQL query and arguments.
	var query string
	query, args, err = builder.ToSql()
	if err != nil {
		return nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_SQL_BUILDER)
	}

	slog.DebugContext(ctx, "generated sql query", slog.String("query", query), "with args", slog.Any("arguments", args))

	// Execute the SQL query and retrieve the result rows.
	var rows pgx.Rows
	rows, err = r.database.ReadPool.Query(ctx, query, args...)
	if err != nil {
		return nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_EXECUTION)
	}
	defer rows.Close()

	// Process the result rows and store the attributes in an AttributeCollection.
	collection := database.NewAttributeCollection()
	for rows.Next() {
		rt := storage.Attribute{}

		// Suppose you have a struct `rt` with a field `Value` of type `*anypb.Any`.
		var valueStr string

		// Scan the row from the database into the fields of `rt` and `valueStr`.
		err := rows.Scan(&rt.EntityType, &rt.EntityID, &rt.Attribute, &valueStr)
		if err != nil {
			return nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_SCAN)
		}

		// Unmarshal the JSON data from `valueStr` into `rt.Value`.
		rt.Value = &anypb.Any{}
		err = protojson.Unmarshal([]byte(valueStr), rt.Value)
		if err != nil {
			return nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_INTERNAL)
		}

		collection.Add(rt.ToAttribute())
	}
	if err = rows.Err(); err != nil {
		return nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_SCAN)
	}

	slog.DebugContext(ctx, "successfully retrieved attributes tuples from the database")

	// Return an AttributeIterator created from the AttributeCollection.
	return collection.CreateAttributeIterator(), nil
}

// ReadAttributes reads multiple attributes from the storage based on the given filter and pagination.
func (r *DataReader) ReadAttributes(ctx context.Context, tenantID string, filter *base.AttributeFilter, snap string, pagination database.Pagination) (collection *database.AttributeCollection, ct database.EncodedContinuousToken, err error) {
	// Start a new trace span and end it when the function exits.
	ctx, span := internal.Tracer.Start(ctx, "data-reader.read-attributes")
	defer span.End()

	slog.DebugContext(ctx, "reading attributes for tenant_id", slog.String("tenant_id", tenantID))

	// Decode the snapshot value.
	var st token.SnapToken
	st, err = snapshot.EncodedToken{Value: snap}.Decode()
	if err != nil {
		return nil, nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_INTERNAL)
	}

	// Build the relationships query based on the provided filter, snapshot value, and pagination settings.
	builder := r.database.Builder.Select("id, entity_type, entity_id, attribute, value").From(AttributesTable).Where(squirrel.Eq{"tenant_id": tenantID})
	builder = utils.AttributesFilterQueryForSelectBuilder(builder, filter)
	builder = utils.SnapshotQuery(builder, st.(snapshot.Token).Value.Uint)

	// Apply the pagination token and limit to the query.
	if pagination.Token() != "" {
		var t database.ContinuousToken
		t, err = utils.EncodedContinuousToken{Value: pagination.Token()}.Decode()
		if err != nil {
			return nil, nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_INVALID_CONTINUOUS_TOKEN)
		}
		var v uint64
		v, err = strconv.ParseUint(t.(utils.ContinuousToken).Value, 10, 64)
		if err != nil {
			return nil, nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_INVALID_CONTINUOUS_TOKEN)
		}
		builder = builder.Where(squirrel.GtOrEq{"id": v})
	}

	builder = builder.OrderBy("id")

	if pagination.PageSize() != 0 {
		builder = builder.Limit(uint64(pagination.PageSize() + 1))
	}

	// Generate the SQL query and arguments.
	var query string
	var args []interface{}
	query, args, err = builder.ToSql()
	if err != nil {
		return nil, database.NewNoopContinuousToken().Encode(), utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_SQL_BUILDER)
	}

	slog.DebugContext(ctx, "generated sql query", slog.String("query", query), "with args", slog.Any("arguments", args))

	// Execute the query and retrieve the rows.
	var rows pgx.Rows
	rows, err = r.database.ReadPool.Query(ctx, query, args...)
	if err != nil {
		return nil, database.NewNoopContinuousToken().Encode(), utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_EXECUTION)
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
			return nil, nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_SCAN)
		}
		lastID = rt.ID

		// Unmarshal the JSON data from `valueStr` into `rt.Value`.
		rt.Value = &anypb.Any{}
		err = protojson.Unmarshal([]byte(valueStr), rt.Value)
		if err != nil {
			return nil, nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_INTERNAL)
		}

		attributes = append(attributes, rt.ToAttribute())
	}
	// Check for any errors during iteration.
	if err = rows.Err(); err != nil {
		return nil, nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_SCAN)
	}

	slog.DebugContext(ctx, "successfully read attributes from the database")

	// Return the results and encoded continuous token for pagination.
	if len(attributes) > int(pagination.PageSize()) {
		return database.NewAttributeCollection(attributes[:pagination.PageSize()]...), utils.NewContinuousToken(strconv.FormatUint(lastID, 10)).Encode(), nil
	}

	return database.NewAttributeCollection(attributes...), database.NewNoopContinuousToken().Encode(), nil
}

// QueryUniqueSubjectReferences reads unique subject references from the storage based on the given filter and pagination.
func (r *DataReader) QueryUniqueSubjectReferences(ctx context.Context, tenantID string, subjectReference *base.RelationReference, excluded []string, snap string, pagination database.Pagination) (ids []string, ct database.EncodedContinuousToken, err error) {
	// Start a new trace span and end it when the function exits.
	ctx, span := internal.Tracer.Start(ctx, "data-reader.query-unique-subject-reference")
	defer span.End()

	slog.DebugContext(ctx, "querying unique subject references for tenant_id", slog.String("tenant_id", tenantID))

	// Decode the snapshot value.
	var st token.SnapToken
	st, err = snapshot.EncodedToken{Value: snap}.Decode()
	if err != nil {
		return nil, nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_INTERNAL)
	}

	// Build the relationships query based on the provided filter, snapshot value, and pagination settings.
	builder := r.database.Builder.
		Select("subject_id"). // This will pick the smallest `id` for each unique `subject_id`.
		From(RelationTuplesTable).
		Where(squirrel.Eq{"tenant_id": tenantID}).
		GroupBy("subject_id")

	// Apply subject filter
	builder = utils.TuplesFilterQueryForSelectBuilder(builder, &base.TupleFilter{
		Subject: &base.SubjectFilter{
			Type:     subjectReference.GetType(),
			Relation: subjectReference.GetRelation(),
		},
	})

	// Apply snapshot filter
	builder = utils.SnapshotQuery(builder, st.(snapshot.Token).Value.Uint)

	// Apply exclusion if the list is not empty
	if len(excluded) > 0 {
		builder = builder.Where(squirrel.NotEq{"subject_id": excluded})
	}

	// Apply the pagination token and limit to the query.
	if pagination.Token() != "" {
		var t database.ContinuousToken
		t, err = utils.EncodedContinuousToken{Value: pagination.Token()}.Decode()
		if err != nil {
			return nil, nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_INVALID_CONTINUOUS_TOKEN)
		}
		builder = builder.Where(squirrel.GtOrEq{"subject_id": t.(utils.ContinuousToken).Value})
	}

	builder = builder.OrderBy("subject_id")

	if pagination.PageSize() != 0 {
		builder = builder.Limit(uint64(pagination.PageSize() + 1))
	}

	// Generate the SQL query and arguments.
	var query string
	var args []interface{}
	query, args, err = builder.ToSql()
	if err != nil {
		return nil, database.NewNoopContinuousToken().Encode(), utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_SQL_BUILDER)
	}

	slog.DebugContext(ctx, "generated sql query", slog.String("query", query), "with args", slog.Any("arguments", args))

	// Execute the query and retrieve the rows.
	var rows pgx.Rows
	rows, err = r.database.ReadPool.Query(ctx, query, args...)
	if err != nil {
		return nil, database.NewNoopContinuousToken().Encode(), utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_EXECUTION)
	}
	defer rows.Close()

	var lastID string

	// Iterate through the rows and scan the result into a RelationTuple struct.
	subjectIDs := make([]string, 0, pagination.PageSize()+1)
	for rows.Next() {
		var subjectID string
		err = rows.Scan(&subjectID)
		if err != nil {
			return nil, nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_INTERNAL)
		}

		subjectIDs = append(subjectIDs, subjectID)
		lastID = subjectID
	}
	// Check for any errors during iteration.
	if err = rows.Err(); err != nil {
		return nil, nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_SCAN)
	}

	slog.DebugContext(ctx, "successfully retrieved unique subject references from the database")

	// Return the results and encoded continuous token for pagination.
	if pagination.PageSize() != 0 && len(subjectIDs) > int(pagination.PageSize()) {
		return subjectIDs[:pagination.PageSize()], utils.NewContinuousToken(lastID).Encode(), nil
	}

	return subjectIDs, database.NewNoopContinuousToken().Encode(), nil
}

// HeadSnapshot retrieves the latest snapshot token associated with the tenant.
func (r *DataReader) HeadSnapshot(ctx context.Context, tenantID string) (token.SnapToken, error) {
	// Start a new trace span and end it when the function exits.
	ctx, span := internal.Tracer.Start(ctx, "data-reader.head-snapshot")
	defer span.End()

	slog.DebugContext(ctx, "getting head snapshot for tenant_id", slog.String("tenant_id", tenantID))

	var xid types.XID8

	// Build the query to find the highest transaction ID associated with the tenant.
	builder := r.database.Builder.Select("id").From(TransactionsTable).Where(squirrel.Eq{"tenant_id": tenantID}).OrderBy("id DESC").Limit(1)
	query, args, err := builder.ToSql()
	if err != nil {
		return nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_SQL_BUILDER)
	}

	// TODO: To optimize this query, create the following index concurrently to avoid table locks:
	// CREATE INDEX CONCURRENTLY idx_transactions_tenant_id_id ON transactions(tenant_id, id DESC);

	// Execute the query and retrieve the highest transaction ID.
	err = r.database.ReadPool.QueryRow(ctx, query, args...).Scan(&xid)
	if err != nil {
		// If no rows are found, return a snapshot token with a value of 0.
		if errors.Is(err, pgx.ErrNoRows) {
			return snapshot.Token{Value: types.XID8{Uint: 0}}, nil
		}
		return nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_SCAN)
	}

	slog.DebugContext(ctx, "successfully retrieved latest snapshot token")

	// Return the latest snapshot token associated with the tenant.
	return snapshot.Token{Value: xid}, nil
}
