package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/Permify/permify/internal/repositories/entities"
	"github.com/Permify/permify/internal/repositories/filters"
	"github.com/Permify/permify/pkg/database"
	db "github.com/Permify/permify/pkg/database/mongo"
	"github.com/Permify/permify/pkg/errors"
)

// RelationTupleRepository -.
type RelationTupleRepository struct {
	Database *db.Mongo
}

// NewRelationTupleRepository -.
func NewRelationTupleRepository(mn *db.Mongo) *RelationTupleRepository {
	return &RelationTupleRepository{mn}
}

// Migrate -
func (r *RelationTupleRepository) Migrate() errors.Error {
	var err error
	command := bson.D{{"create", entities.RelationTuple{}.Collection()}}
	var result bson.M
	if err = r.Database.Database().RunCommand(context.TODO(), command).Decode(&result); err != nil {
		return errors.NewError(errors.Database).SetSubKind(database.ErrMigration)
	}
	_, err = r.Database.Database().Collection(entities.RelationTuple{}.Collection()).Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys: bson.M{
			"tuple": 1,
		},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return errors.NewError(errors.Database).SetSubKind(database.ErrMigration)
	}
	return nil
}

// QueryTuples -
func (r *RelationTupleRepository) QueryTuples(ctx context.Context, entity string, objectID string, relation string) (entities.RelationTuples, errors.Error) {
	var err error
	var tuples entities.RelationTuples
	coll := r.Database.Database().Collection(entities.RelationTuple{}.Collection())
	filter := bson.M{"entity": entity, "object_id": objectID, "relation": relation}
	opts := options.Find().SetSort(bson.D{{"userset_entity", 1}, {"userset_relation", 1}})

	var cursor *mongo.Cursor
	cursor, err = coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, errors.NewError(errors.Database).SetSubKind(database.ErrBuilder)
	}

	if err = cursor.All(ctx, &tuples); err != nil {
		return nil, errors.NewError(errors.Database).SetSubKind(database.ErrExecution)
	}
	return tuples, nil
}

// Read -.
func (r *RelationTupleRepository) Read(ctx context.Context, filter filters.RelationTupleFilter) (entities.RelationTuples, errors.Error) {
	var err error
	var tuples entities.RelationTuples

	coll := r.Database.Database().Collection(entities.RelationTuple{}.Collection())

	eq := bson.M{}
	eq["entity"] = filter.Entity

	if filter.Entity.ID != "" {
		eq["object_id"] = filter.Entity.ID
	}

	if filter.Relation != "" {
		eq["relation"] = filter.Relation
	}

	if filter.Subject.Type != "" {
		eq["userset_entity"] = filter.Subject.Type
	}

	if filter.Subject.ID != "" {
		eq["userset_object_id"] = filter.Subject.ID
	}

	if filter.Subject.Relation != "" {
		eq["userset_relation"] = filter.Subject.Relation
	}

	opts := options.Find().SetSort(bson.D{{"userset_entity", 1}, {"userset_relation", 1}})

	var cursor *mongo.Cursor
	cursor, err = coll.Find(ctx, eq, opts)
	if err != nil {
		return nil, errors.NewError(errors.Database).SetSubKind(database.ErrBuilder)
	}
	if err = cursor.All(ctx, &tuples); err != nil {
		return nil, errors.NewError(errors.Database).SetSubKind(database.ErrExecution)
	}

	return tuples, nil
}

// Write -.
func (r *RelationTupleRepository) Write(ctx context.Context, tuples entities.RelationTuples) errors.Error {
	var err error
	coll := r.Database.Database().Collection(entities.RelationTuple{}.Collection())
	var docs []interface{}
	for _, tup := range tuples {
		docs = append(docs, bson.D{{"entity", tup.Entity}, {"object_id", tup.ObjectID}, {"relation", tup.Relation}, {"userset_entity", tup.UsersetEntity}, {"userset_object_id", tup.UsersetObjectID}, {"userset_relation", tup.UsersetRelation}, {"tuple", tup.ToTuple().String()}})
	}
	_, err = coll.InsertMany(ctx, docs)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.NewError(errors.Database).SetSubKind(database.ErrUniqueConstraint)
		}
		return errors.NewError(errors.Database).SetMessage(err.Error())
	}
	return nil
}

// Delete -.
func (r *RelationTupleRepository) Delete(ctx context.Context, tuples entities.RelationTuples) errors.Error {
	coll := r.Database.Database().Collection(entities.RelationTuple{}.Collection())
	for _, tuple := range tuples {
		filter := bson.M{"entity": tuple.Entity, "object_id": tuple.ObjectID, "relation": tuple.Relation, "userset_entity": tuple.UsersetEntity, "userset_object_id": tuple.UsersetObjectID, "userset_relation": tuple.UsersetRelation}
		_, err := coll.DeleteOne(ctx, filter)
		if err != nil {
			return errors.NewError(errors.Database).SetSubKind(database.ErrExecution)
		}
	}
	return nil
}
