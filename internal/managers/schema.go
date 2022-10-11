package managers

import (
	"context"

	"github.com/rs/xid"

	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/pkg/cache"
	"github.com/Permify/permify/pkg/dsl/ast"
	"github.com/Permify/permify/pkg/dsl/parser"
	"github.com/Permify/permify/pkg/dsl/schema"
	`github.com/Permify/permify/pkg/dsl/translator`
	"github.com/Permify/permify/pkg/errors"
	base `github.com/Permify/permify/pkg/pb/base/v1`
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
func (manager *EntityConfigManager) All(ctx context.Context, version string) (*base.Schema, errors.Error) {
	var sch *base.Schema
	var err errors.Error
	var cn []repositories.EntityConfig
	cn, err = manager.repository.All(ctx, version)
	if err != nil {
		return sch, err
	}
	var serializedConfigs []string
	for _, c := range cn {
		serializedConfigs = append(serializedConfigs, c.Serialized())
	}
	sch, err = translator.StringToSchema(serializedConfigs...)
	if err != nil {
		return sch, err
	}
	return sch, err
}

// Read -
func (manager *EntityConfigManager) Read(ctx context.Context, name string, version string) (entityDefinition *base.EntityDefinition, err errors.Error) {
	if manager.cache == nil {
		var config repositories.EntityConfig
		config, err = manager.repository.Read(ctx, name, version)

		var sch *base.Schema
		sch, err = translator.StringToSchema(config.Serialized())
		if err != nil {
			return nil, err
		}

		return schema.GetEntityByName(sch, name)
	}

	var key string
	var s interface{}
	found := false

	if version != "" {
		key = name + "|" + version
		s, found = manager.cache.Get(key)
	}

	if !found {
		var config repositories.EntityConfig
		config, err = manager.repository.Read(ctx, name, version)
		if err != nil {
			return nil, err
		}
		key = name + "|" + config.Version
		manager.cache.Set(key, config, 1)

		var sch *base.Schema
		sch, err = translator.StringToSchema(config.Serialized())
		if err != nil {
			return nil, err
		}

		return schema.GetEntityByName(sch, name)
	}

	conf := s.(repositories.EntityConfig)
	var sch *base.Schema
	sch, err = translator.StringToSchema(conf.Serialized())
	if err != nil {
		return nil, err
	}

	return schema.GetEntityByName(sch, name)
}

// Write -
func (manager *EntityConfigManager) Write(ctx context.Context, configs string) (string, errors.Error) {
	version := xid.New().String()

	sch, err := parser.NewParser(configs).Parse()
	if err != nil {
		return "", err
	}

	err = sch.Validate()
	if err != nil {
		return "", err
	}

	var cnf []repositories.EntityConfig
	for _, st := range sch.Statements {
		cnf = append(cnf, repositories.EntityConfig{
			Entity:           st.(*ast.EntityStatement).Name.Literal,
			SerializedConfig: []byte(st.String()),
		})
	}

	return version, manager.repository.Write(ctx, cnf, version)
}
