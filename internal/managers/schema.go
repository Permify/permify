package managers

import (
	"context"
	"errors"
	"fmt"

	"github.com/rs/xid"

	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/pkg/cache"
	"github.com/Permify/permify/pkg/dsl/ast"
	"github.com/Permify/permify/pkg/dsl/compiler"
	"github.com/Permify/permify/pkg/dsl/parser"
	"github.com/Permify/permify/pkg/dsl/schema"
	base "github.com/Permify/permify/pkg/pb/base/v1"
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
func (manager *EntityConfigManager) All(ctx context.Context, version string) (*base.Schema, error) {
	var sch *base.Schema
	var err error
	var cn []repositories.EntityConfig
	cn, err = manager.repository.All(ctx, version)
	if err != nil {
		return sch, err
	}
	var serializedConfigs []string
	for _, c := range cn {
		serializedConfigs = append(serializedConfigs, c.Serialized())
	}
	sch, err = compiler.NewSchema(serializedConfigs...)
	if err != nil {
		return sch, err
	}
	return sch, err
}

// Read -
func (manager *EntityConfigManager) Read(ctx context.Context, name string, version string) (entityDefinition *base.EntityDefinition, err error) {
	if manager.cache == nil {
		var config repositories.EntityConfig
		config, err = manager.repository.Read(ctx, name, version)
		if err != nil {
			return nil, err
		}

		var sch *base.Schema
		sch, err = compiler.NewSchemaWithoutReferenceValidation(config.Serialized())
		if err != nil {
			return nil, err
		}

		return schema.GetEntityByName(sch, name)
	}

	var key string
	var s interface{}
	found := false

	if version != "" {
		key = fmt.Sprintf("%s|%s", name, version)
		s, found = manager.cache.Get(key)
	}

	if !found {
		var config repositories.EntityConfig
		config, err = manager.repository.Read(ctx, name, version)
		if err != nil {
			return nil, err
		}
		key = fmt.Sprintf("%s|%s", name, config.Version)
		var sch *base.Schema
		sch, err = compiler.NewSchemaWithoutReferenceValidation(config.Serialized())
		if err != nil {
			return nil, err
		}
		var def *base.EntityDefinition
		def, err = schema.GetEntityByName(sch, name)
		if err != nil {
			return nil, err
		}
		manager.cache.Set(key, def, 1)
		return def, nil
	}

	def, ok := s.(*base.EntityDefinition)
	if !ok {
		return nil, errors.New(base.ErrorCode_type_conversation_error.String())
	}

	return def, err
}

// Write -
func (manager *EntityConfigManager) Write(ctx context.Context, schema string) (string, error) {
	version := xid.New().String()

	sch, err := parser.NewParser(schema).Parse()
	if err != nil {
		return "", err
	}

	_, err = compiler.NewCompiler(false, sch).Compile()
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
