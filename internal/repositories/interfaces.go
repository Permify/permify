package repositories

import (
	"context"

	"github.com/Permify/permify/internal/entities"
	"github.com/Permify/permify/internal/repositories/filters"
)

// Migratable -
type Migratable interface {
	Migrate() error
}

// IRelationTupleRepository -
type IRelationTupleRepository interface {
	Migratable
	QueryTuples(ctx context.Context, namespace string, objectID string, relation string) (entities.RelationTuples, error)
	Read(ctx context.Context, filter filters.RelationTupleFilter) (entities.RelationTuples, error)
	Write(context.Context, entities.RelationTuples) error
	Delete(context.Context, entities.RelationTuples) error
}

// IEntityConfigRepository -
type IEntityConfigRepository interface {
	Migratable
	All(ctx context.Context, version string) (configs entities.EntityConfigs, err error)
	Read(ctx context.Context, name string, version string) (config entities.EntityConfig, err error)
	Write(ctx context.Context, configs entities.EntityConfigs, version string) (err error)
	Clear(ctx context.Context, version string) (err error)
}
