package managers

import (
	"context"

	"github.com/rs/xid"

	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/internal/repositories/entities"
	"github.com/Permify/permify/pkg/cache"
	"github.com/Permify/permify/pkg/dsl/parser"
	"github.com/Permify/permify/pkg/dsl/schema"
	"github.com/Permify/permify/pkg/errors"
	`github.com/Permify/permify/pkg/helper`
)

// EntityConfigManager -
type EntityConfigManager struct {
	repository repositories.IEntityConfigRepository
	cache      cache.Cache
}

// NewEntityConfigManager -
func NewEntityConfigManager(repository repositories.IEntityConfigRepository, cache cache.Cache) IEntityConfigManager {
	return &EntityConfigManager{
		repository: repository,
		cache:      cache,
	}
}

// All -
func (manager *EntityConfigManager) All(ctx context.Context, version string) (schema.Schema, errors.Error) {
	var sch schema.Schema
	var err errors.Error
	var cn entities.EntityConfigs
	cn, err = manager.repository.All(ctx, version)
	if err != nil {
		return sch, err
	}
	sch, err = cn.ToSchema()
	if err != nil {
		return sch, err
	}
	return sch, err
}

// Read -
func (manager *EntityConfigManager) Read(ctx context.Context, name string, version string) (entity schema.Entity, err errors.Error) {
	if manager.cache == nil {
		var config entities.EntityConfig
		config, err = manager.repository.Read(ctx, name, version)

		var sch schema.Schema
		sch, err = config.ToSchema()
		if err != nil {
			return entity, err
		}

		return sch.Entities[name], err
	}

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
			return entity, err
		}

		return sch.Entities[name], err
	}

	conf := s.(entities.EntityConfig)
	var sch schema.Schema
	sch, err = conf.ToSchema()
	if err != nil {
		return entity, err
	}

	return sch.Entities[name], err
}

// Write -
func (manager *EntityConfigManager) Write(ctx context.Context, configs string) (string, errors.Error) {
	version := xid.New().String()

	_, err := parser.NewParser(configs).Parse()
	if err != nil {
		return "", err
	}

	helper.Pre("no error")

	//err := sch.Validate()
	//if err != nil {
	//	return "", err
	//}

	var cnf []entities.EntityConfig
	//for _, st := range sch.Statements {
	//	cnf = append(cnf, entities.EntityConfig{
	//		Entity:           st.(*ast.EntityStatement).Name.Literal,
	//		SerializedConfig: []byte(st.String()),
	//	})
	//}

	return version, manager.repository.Write(ctx, cnf, version)
}
