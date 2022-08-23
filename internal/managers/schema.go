package managers

import (
	"context"

	"github.com/dgraph-io/ristretto"
	"github.com/rs/xid"

	"github.com/Permify/permify/internal/entities"
	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/pkg/dsl/schema"
)

// EntityConfigManager -
type EntityConfigManager struct {
	repository repositories.IEntityConfigRepository
	cache      *ristretto.Cache
}

// NewEntityConfigManager -
func NewEntityConfigManager(repository repositories.IEntityConfigRepository, cache *ristretto.Cache) IEntityConfigManager {
	return &EntityConfigManager{
		repository: repository,
		cache:      cache,
	}
}

// All -
func (manager *EntityConfigManager) All(ctx context.Context, version string) (sch schema.Schema, err error) {
	var cn entities.EntityConfigs
	cn, err = manager.repository.All(ctx, version)
	if err != nil {
		return schema.Schema{}, err
	}
	sch, err = cn.ToSchema()
	if err != nil {
		return schema.Schema{}, err
	}
	return
}

// Read -
func (manager *EntityConfigManager) Read(ctx context.Context, name string, version string) (entity schema.Entity, err error) {
	var key string
	var s interface{}
	found := false

	if version != "" {
		key = name + "|" + version
		s, found = manager.cache.Get(key)
	}

	if !found {
		var config entities.EntityConfig
		config, err = manager.repository.Read(ctx, name, version)
		if err != nil {
			return entity, err
		}
		key = name + "|" + config.Version
		manager.cache.Set(key, config, 1)

		var sch schema.Schema
		sch, err = config.ToSchema()
		if err != nil {
			return schema.Entity{}, err
		}

		return sch.Entities[name], err
	}

	conf := s.(entities.EntityConfig)
	var sch schema.Schema
	sch, err = conf.ToSchema()
	if err != nil {
		return schema.Entity{}, err
	}

	return sch.Entities[name], err
}

// Write -
func (manager *EntityConfigManager) Write(ctx context.Context, configs entities.EntityConfigs) (version string, err error) {
	version = xid.New().String()
	return version, manager.repository.Write(ctx, configs, version)
}
