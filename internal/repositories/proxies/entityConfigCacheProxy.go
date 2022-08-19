package proxies

import (
	"context"

	"github.com/dgraph-io/ristretto"

	"github.com/Permify/permify/internal/entities"
	"github.com/Permify/permify/internal/repositories"
)

type EntityConfigCacheProxy struct {
	repository repositories.IEntityConfigRepository
	cache      *ristretto.Cache
}

// NewEntityConfigCacheProxy -.
func NewEntityConfigCacheProxy(entityConfigRepository repositories.IEntityConfigRepository, cache *ristretto.Cache) *EntityConfigCacheProxy {
	return &EntityConfigCacheProxy{repository: entityConfigRepository, cache: cache}
}

// Migrate -
func (r *EntityConfigCacheProxy) Migrate() (err error) {
	return nil
}

// All -
func (r *EntityConfigCacheProxy) All(ctx context.Context, version string) (configs entities.EntityConfigs, err error) {
	return r.repository.All(ctx, version)
}

// Read -
func (r *EntityConfigCacheProxy) Read(ctx context.Context, name string, version string) (config entities.EntityConfig, err error) {
	var key string
	var s interface{}
	found := false

	if version != "" {
		key = name + "|" + version
		s, found = r.cache.Get(key)
	}

	if !found {
		config, err = r.repository.Read(ctx, name, version)
		if err != nil {
			return entities.EntityConfig{}, err
		}
		key = name + "|" + config.Version
		r.cache.Set(key, config, 1)
		return config, err
	}

	return s.(entities.EntityConfig), err
}

// Write -
func (r *EntityConfigCacheProxy) Write(ctx context.Context, configs entities.EntityConfigs, version string) (err error) {
	return r.repository.Write(ctx, configs, version)
}

// Clear -
func (r *EntityConfigCacheProxy) Clear(ctx context.Context, version string) error {
	return r.repository.Clear(ctx, version)
}
