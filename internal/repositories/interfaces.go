package repositories

import (
	"context"

	"github.com/Permify/permify/internal/entities"
)

// IRelationTupleRepository -
type IRelationTupleRepository interface {
	QueryTuples(ctx context.Context, namespace string, objectID string, relation string) ([]entities.RelationTuple, error)
	Write(context.Context, []entities.RelationTuple) error
	Delete(context.Context, []entities.RelationTuple) error
}

// IEntityConfigRepository -
type IEntityConfigRepository interface {
	All(ctx context.Context) (configs []entities.EntityConfig, err error)
	First(ctx context.Context, entity string) (config entities.EntityConfig, err error)
	Replace(ctx context.Context, configs []entities.EntityConfig) (err error)
}
