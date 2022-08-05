package repositories

import (
	"context"

	"github.com/Permify/permify/internal/entities"
	`github.com/Permify/permify/internal/repositories/filters`
)

// Migratable -
type Migratable interface {
	Migrate() error
}

// IRelationTupleRepository -
type IRelationTupleRepository interface {
	Migratable
	QueryTuples(ctx context.Context, namespace string, objectID string, relation string) ([]entities.RelationTuple, error)
	Read(ctx context.Context, filter filters.RelationTupleFilter) (tuples []entities.RelationTuple, err error)
	Write(context.Context, []entities.RelationTuple) error
	Delete(context.Context, []entities.RelationTuple) error
}

// IEntityConfigRepository -
type IEntityConfigRepository interface {
	Migratable
	All(ctx context.Context) (configs []entities.EntityConfig, err error)
	Replace(ctx context.Context, configs []entities.EntityConfig) (err error)
	Clear(ctx context.Context) (err error)
}
