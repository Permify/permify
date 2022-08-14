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
	All(ctx context.Context) (configs entities.EntityConfigs, err error)
	Read(ctx context.Context, name string) (config entities.EntityConfig, err error)
	Replace(ctx context.Context, configs entities.EntityConfigs) (err error)
	Clear(ctx context.Context) (err error)
}
