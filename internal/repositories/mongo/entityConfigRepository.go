package mongo

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/Permify/permify/internal/entities"
	"github.com/Permify/permify/internal/internal-errors"
	db "github.com/Permify/permify/pkg/database/mongo"
)

// EntityConfigRepository -.
type EntityConfigRepository struct {
	Database *db.Mongo
}

// NewEntityConfigRepository -.
func NewEntityConfigRepository(mn *db.Mongo) *EntityConfigRepository {
	return &EntityConfigRepository{mn}
}

// Migrate -
func (r *EntityConfigRepository) Migrate() (err error) {
	command := bson.D{{"create", entities.EntityConfig{}.Collection()}}
	var result bson.M
	if err = r.Database.Database().RunCommand(context.TODO(), command).Decode(&result); err != nil {
		return nil
	}
	return nil
}

// All -
func (r *EntityConfigRepository) All(ctx context.Context) (configs entities.EntityConfigs, err error) {
	coll := r.Database.Database().Collection(entities.EntityConfig{}.Collection())
	filter := bson.M{}
	var cursor *mongo.Cursor
	cursor, err = coll.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	if err = cursor.All(ctx, &configs); err != nil {
		return nil, err
	}

	return
}

// Read -
func (r *EntityConfigRepository) Read(ctx context.Context, name string) (config entities.EntityConfig, err error) {
	coll := r.Database.Database().Collection(entities.EntityConfig{}.Collection())
	filter := bson.M{"entity": name}
	err = coll.FindOne(ctx, filter).Decode(&config)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return config, internal_errors.EntityConfigCannotFoundError
		}
		return
	}
	return
}

// Replace -
func (r *EntityConfigRepository) Replace(ctx context.Context, configs entities.EntityConfigs) (err error) {
	if len(configs) < 1 {
		return nil
	}

	err = r.Clear(ctx)
	if err != nil {
		return err
	}

	coll := r.Database.Database().Collection(entities.EntityConfig{}.Collection())

	var docs []interface{}

	for _, con := range configs {
		docs = append(docs, bson.D{{"entity", con.Entity}, {"serialized_config", con.SerializedConfig}})
	}

	_, err = coll.InsertMany(ctx, docs)
	if err != nil {
		return err
	}
	return nil
}

// Clear -
func (r *EntityConfigRepository) Clear(ctx context.Context) error {
	coll := r.Database.Database().Collection(entities.EntityConfig{}.Collection())
	filter := bson.M{}
	_, err := coll.DeleteMany(ctx, filter)
	if err != nil {
		return err
	}
	return nil
}
