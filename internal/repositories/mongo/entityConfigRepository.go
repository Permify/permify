package mongo

import (
	"context"
	e "errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/Permify/permify/internal/repositories/entities"
	"github.com/Permify/permify/pkg/database"
	db "github.com/Permify/permify/pkg/database/mongo"
	"github.com/Permify/permify/pkg/errors"
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
func (r *EntityConfigRepository) Migrate() errors.Error {
	var err error
	command := bson.D{{"create", entities.EntityConfig{}.Collection()}}
	var result bson.M
	err = r.Database.Database().RunCommand(context.TODO(), command).Decode(&result)
	if err != nil {
		return errors.NewError(errors.Database).SetSubKind(database.ErrMigration)
	}
	return nil
}

// All -
func (r *EntityConfigRepository) All(ctx context.Context, version string) (entities.EntityConfigs, errors.Error) {
	var configs entities.EntityConfigs
	var err error

	if version == "" {
		version, err = r.findLastVersion(ctx)
		if err != nil {
			return configs, errors.NewError(errors.Database).SetMessage(err.Error())
		}
	}

	coll := r.Database.Database().Collection(entities.EntityConfig{}.Collection())
	filter := bson.M{"version": version}
	var cursor *mongo.Cursor
	cursor, err = coll.Find(ctx, filter)
	if err != nil {
		return nil, errors.NewError(errors.Database).SetMessage(err.Error())
	}

	if err = cursor.All(ctx, &configs); err != nil {
		return nil, errors.NewError(errors.Database).SetMessage(err.Error())
	}

	return configs, nil
}

// Read -
func (r *EntityConfigRepository) Read(ctx context.Context, name string, version string) (entities.EntityConfig, errors.Error) {
	var config entities.EntityConfig
	var err error

	if version == "" {
		version, err = r.findLastVersion(ctx)
		if err != nil {
			if e.Is(err, mongo.ErrNoDocuments) {
				return config, errors.NewError(errors.Database).SetSubKind(database.ErrRecordNotFound)
			}
			return config, errors.NewError(errors.Database).SetMessage(err.Error())
		}
	}

	coll := r.Database.Database().Collection(entities.EntityConfig{}.Collection())
	filter := bson.M{"entity": name, "version": version}
	err = coll.FindOne(ctx, filter).Decode(&config)
	if err != nil {
		if e.Is(err, mongo.ErrNoDocuments) {
			return config, errors.NewError(errors.Database).SetSubKind(database.ErrRecordNotFound)
		}
		return config, errors.NewError(errors.Database).SetMessage(err.Error())
	}
	return config, nil
}

// findLastVersion -
func (r *EntityConfigRepository) findLastVersion(ctx context.Context) (string, errors.Error) {
	var version string
	var err error
	coll := r.Database.Database().Collection(entities.EntityConfig{}.Collection())
	filter := bson.M{}
	opts := options.FindOne().SetSort(bson.D{{"version", -1}})
	var entityConfig entities.EntityConfig
	err = coll.FindOne(ctx, filter, opts).Decode(&entityConfig)
	if err != nil {
		if e.Is(err, mongo.ErrNoDocuments) {
			return version, errors.NewError(errors.Database).SetSubKind(database.ErrRecordNotFound)
		}
		return version, errors.NewError(errors.Database).SetMessage(err.Error())
	}
	return entityConfig.Version, nil
}

// Write -
func (r *EntityConfigRepository) Write(ctx context.Context, configs entities.EntityConfigs, version string) errors.Error {
	var err error
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
		return errors.NewError(errors.Database).SetMessage(err.Error())
	}
	return nil
}
