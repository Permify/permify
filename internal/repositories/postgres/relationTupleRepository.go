package postgres

import (
	"context"
	e "errors"

	"github.com/jackc/pgconn"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4"

	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/internal/repositories/postgres/migrations"
	"github.com/Permify/permify/pkg/database"
	db "github.com/Permify/permify/pkg/database/postgres"
	"github.com/Permify/permify/pkg/errors"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
)

// RelationTupleRepository -.
type RelationTupleRepository struct {
	Database *db.Postgres
}

// NewRelationTupleRepository -.
func NewRelationTupleRepository(pg *db.Postgres) *RelationTupleRepository {
	return &RelationTupleRepository{pg}
}

// Migrate -
func (r *RelationTupleRepository) Migrate() errors.Error {
	var err error
	ctx := context.Background()

	var tx pgx.Tx
	tx, err = r.Database.Pool.Begin(ctx)
	if err != nil {
		return errors.DatabaseError.SetSubKind(database.ErrMigration)
	}

	_, err = tx.Exec(context.Background(), migrations.CreateRelationTupleMigration())
	if err != nil {
		return errors.DatabaseError.SetSubKind(database.ErrMigration)
	}

	_, err = tx.Exec(context.Background(), migrations.DropRelationTupleTypeColumnIfExistMigration())
	if err != nil {
		return errors.DatabaseError.SetSubKind(database.ErrMigration)
	}

	_, err = tx.Exec(context.Background(), migrations.CreateRelationTupleUserSetIndexIfNotExistMigration())
	if err != nil {
		return errors.DatabaseError.SetSubKind(database.ErrMigration)
	}

	_, err = tx.Exec(context.Background(), migrations.CreateRelationTupleUserSetRelationIndexIfNotExistMigration())
	if err != nil {
		return errors.DatabaseError.SetSubKind(database.ErrMigration)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return errors.DatabaseError.SetSubKind(database.ErrMigration)
	}

	return nil
}

// ReverseQueryTuples -
func (r *RelationTupleRepository) ReverseQueryTuples(ctx context.Context, entity string, relation string, subjectEntity string, subjectIDs []string, subjectRelation string) (tuple.ITupleIterator, errors.Error) {
	var err error
	var sql string
	var args []interface{}
	sql, args, err = r.Database.Builder.
		Select("entity, object_id, relation, userset_entity, userset_object_id, userset_relation").From("relation_tuple").Where(squirrel.Eq{"entity": entity, "relation": relation, "userset_entity": subjectEntity, "userset_object_id": subjectIDs, "userset_relation": subjectRelation}).
		ToSql()
	if err != nil {
		return nil, errors.DatabaseError.SetSubKind(database.ErrBuilder)
	}

	var rows pgx.Rows
	rows, err = r.Database.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, errors.DatabaseError.SetSubKind(database.ErrExecution)
	}
	defer rows.Close()

	collection := tuple.NewTupleCollection()

	for rows.Next() {
		rt := repositories.RelationTuple{}
		err = rows.Scan(&rt.Entity, &rt.EntityID, &rt.Relation, &rt.SubjectEntity, &rt.SubjectID, &rt.SubjectRelation)
		if err != nil {
			return nil, errors.DatabaseError.SetMessage(err.Error())
		}
		collection.Add(rt.ToTuple())
	}

	return collection.CreateTupleIterator(), nil
}

// QueryTuples -
func (r *RelationTupleRepository) QueryTuples(ctx context.Context, entity string, entityID string, relation string) (tuple.ITupleIterator, errors.Error) {
	var err error
	var sql string
	var args []interface{}
	sql, args, err = r.Database.Builder.
		Select("entity, object_id, relation, userset_entity, userset_object_id, userset_relation").From("relation_tuple").Where(squirrel.Eq{"entity": entity, "object_id": entityID, "relation": relation}).OrderBy("userset_entity, userset_relation ASC").
		ToSql()
	if err != nil {
		return nil, errors.DatabaseError.SetSubKind(database.ErrBuilder)
	}

	var rows pgx.Rows
	rows, err = r.Database.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, errors.DatabaseError.SetSubKind(database.ErrExecution)
	}
	defer rows.Close()

	collection := tuple.NewTupleCollection()

	for rows.Next() {
		rt := repositories.RelationTuple{}
		err = rows.Scan(&rt.Entity, &rt.EntityID, &rt.Relation, &rt.SubjectEntity, &rt.SubjectID, &rt.SubjectRelation)
		if err != nil {
			return nil, errors.DatabaseError.SetMessage(err.Error())
		}
		collection.Add(rt.ToTuple())
	}

	return collection.CreateTupleIterator(), nil
}

// Read -.
func (r *RelationTupleRepository) Read(ctx context.Context, filter *base.TupleFilter) (tuple.ITupleCollection, errors.Error) {
	var err error
	var sql string

	eq := squirrel.Eq{}
	eq["entity"] = filter.Entity.Type

	if filter.GetEntity().GetId() != "" {
		eq["object_id"] = filter.GetEntity().GetId()
	}

	if filter.GetRelation() != "" {
		eq["relation"] = filter.GetRelation()
	}

	if filter.GetSubject().GetType() != "" {
		eq["userset_entity"] = filter.GetSubject().GetType()
	}

	if filter.GetSubject().GetId() != "" {
		eq["userset_object_id"] = filter.GetSubject().GetId()
	}

	if filter.GetSubject().GetRelation() != "" {
		eq["userset_relation"] = filter.GetSubject().GetRelation()
	}

	var args []interface{}
	sql, args, err = r.Database.Builder.
		Select("entity, object_id, relation, userset_entity, userset_object_id, userset_relation, commit_time").From("relation_tuple").Where(eq).OrderBy("userset_entity, userset_relation ASC").
		ToSql()
	if err != nil {
		return nil, errors.DatabaseError.SetSubKind(database.ErrBuilder)
	}

	var rows pgx.Rows
	rows, err = r.Database.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, errors.DatabaseError.SetSubKind(database.ErrExecution)
	}
	defer rows.Close()

	collection := tuple.NewTupleCollection()

	for rows.Next() {
		rt := repositories.RelationTuple{}
		err = rows.Scan(&rt.Entity, &rt.EntityID, &rt.Relation, &rt.SubjectEntity, &rt.SubjectID, &rt.SubjectRelation)
		if err != nil {
			return nil, errors.DatabaseError.SetMessage(err.Error())
		}
		collection.Add(rt.ToTuple())
	}

	return collection, nil
}

// Write -.
func (r *RelationTupleRepository) Write(ctx context.Context, tuples tuple.ITupleIterator) errors.Error {
	var err error

	if !tuples.HasNext() {
		return nil
	}

	sql := r.Database.Builder.
		Insert("relation_tuple").
		Columns("entity, object_id, relation, userset_entity, userset_object_id, userset_relation")

	for tuples.HasNext() {
		t := tuples.GetNext()
		sql = sql.Values(t.GetEntity().GetType(), t.GetEntity().GetId(), t.GetRelation(), t.GetSubject().GetType(), t.GetSubject().GetId(), t.GetSubject().GetRelation())
	}

	var query string
	var args []interface{}
	query, args, err = sql.ToSql()
	if err != nil {
		return errors.DatabaseError.SetSubKind(database.ErrBuilder)
	}

	_, err = r.Database.Pool.Exec(ctx, query, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if e.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				return errors.DatabaseError.SetSubKind(database.ErrUniqueConstraint)
			default:
				return errors.DatabaseError.SetSubKind(database.ErrExecution)
			}
		}
		return errors.DatabaseError.SetMessage(err.Error())
	}

	return nil
}

// Delete -.
func (r *RelationTupleRepository) Delete(ctx context.Context, tuples tuple.ITupleIterator) errors.Error {
	tx, err := r.Database.Pool.Begin(ctx)
	if err != nil {
		return errors.DatabaseError.SetMessage(err.Error())
	}
	batch := &pgx.Batch{}
	for tuples.HasNext() {
		t := tuples.GetNext()
		sql, args, err := r.Database.Builder.
			Delete("relation_tuple").Where(squirrel.Eq{"entity": t.GetEntity().GetType(), "object_id": t.GetEntity().GetId(), "relation": t.GetRelation(), "userset_entity": t.GetSubject().GetType(), "userset_object_id": t.GetSubject().GetId(), "userset_relation": t.GetSubject().GetRelation()}).
			ToSql()
		if err != nil {
			return errors.DatabaseError.SetSubKind(database.ErrBuilder)
		}
		batch.Queue(sql, args...)
	}
	results := tx.SendBatch(ctx, batch)
	if err = results.Close(); err != nil {
		return errors.DatabaseError.SetMessage(err.Error())
	}
	if err = tx.Commit(ctx); err != nil {
		return errors.DatabaseError.SetSubKind(database.ErrExecution)
	}
	return nil
}
