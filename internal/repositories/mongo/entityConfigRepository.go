package mongo

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	`go.mongodb.org/mongo-driver/mongo/options`

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
func (r *EntityConfigRepository) All(ctx context.Context, version string) (configs entities.EntityConfigs, err error) {

	if version == "" {
		version, err = r.findLastVersion(ctx)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				return configs, internal_errors.EntityConfigCannotFoundError
			}
			return configs, err
		}
	}

	coll := r.Database.Database().Collection(entities.EntityConfig{}.Collection())
	filter := bson.M{"version": version}
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
func (r *EntityConfigRepository) Read(ctx context.Context, name string, version string) (config entities.EntityConfig, err error) {

	if version == "" {
		version, err = r.findLastVersion(ctx)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				return config, internal_errors.EntityConfigCannotFoundError
			}
			return config, err
		}
	}

	coll := r.Database.Database().Collection(entities.EntityConfig{}.Collection())
	filter := bson.M{"entity": name, "version": version}
	err = coll.FindOne(ctx, filter).Decode(&config)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return config, internal_errors.EntityConfigCannotFoundError
		}
		return
	}
	return
}

// findLastVersion -
func (r *EntityConfigRepository) findLastVersion(ctx context.Context) (version string, err error) {
	coll := r.Database.Database().Collection(entities.EntityConfig{}.Collection())
	filter := bson.M{}
	opts := options.FindOne().SetSort(bson.D{{"version", -1}})
	var entityConfig entities.EntityConfig
	err = coll.FindOne(ctx, filter, opts).Decode(&entityConfig)
	if err != nil {
		return version, err
	}
	return entityConfig.Version, err
}

// Replace -
func (r *EntityConfigRepository) Write(ctx context.Context, configs entities.EntityConfigs, version string) (err error) {
	if len(configs) < 1 {
		return nil
	}

	coll := r.Database.Database().Collection(entities.EntityConfig{}.Collection())

	var docs []interface{}

	for _, config := range configs {
		docs = append(docs, bson.D{{"entity", config.Entity}, {"serialized_config", config.SerializedConfig}, {"version", version}})
	}

	_, err = coll.InsertMany(ctx, docs)
	if err != nil {
		return err
	}
	return nil
}

// Clear -
func (r *EntityConfigRepository) Clear(ctx context.Context, version string) error {
	coll := r.Database.Database().Collection(entities.EntityConfig{}.Collection())
	filter := bson.M{"version": version}
	_, err := coll.DeleteMany(ctx, filter)
	if err != nil {
		return err
	}
	return nil
}
