package proxies

import (
	`context`
	`github.com/dgraph-io/ristretto`

	`github.com/Permify/permify/internal/entities`
	`github.com/Permify/permify/internal/repositories`
)

type EntityConfigProxy struct {
	repository repositories.IEntityConfigRepository
	cache      *ristretto.Cache
}

// NewEntityConfigProxy -.
func NewEntityConfigProxy(entityConfigRepository repositories.IEntityConfigRepository, cache *ristretto.Cache) *EntityConfigProxy {
	return &EntityConfigProxy{repository: entityConfigRepository, cache: cache}
}

// Migrate -
func (r *EntityConfigProxy) Migrate() (err error) {
	return nil
}

// All -
func (r *EntityConfigProxy) All(ctx context.Context, version string) (configs entities.EntityConfigs, err error) {
	return r.repository.All(ctx, version)
}

// Read -
func (r *EntityConfigProxy) Read(ctx context.Context, name string, version string) (config entities.EntityConfig, err error) {
	var key string
	var s interface{}
	var found = false

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
func (r *EntityConfigProxy) Write(ctx context.Context, configs entities.EntityConfigs, version string) (err error) {
	return r.repository.Write(ctx, configs, version)
}

// Clear -
func (r *EntityConfigProxy) Clear(ctx context.Context, version string) error {
	return r.repository.Clear(ctx, version)
}
