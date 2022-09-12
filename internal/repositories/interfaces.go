package repositories

import (
	"context"

	"github.com/Permify/permify/internal/repositories/entities"
	"github.com/Permify/permify/internal/repositories/filters"
	"github.com/Permify/permify/pkg/errors"
)

// Migratable -
type Migratable interface {
	Migrate() errors.Error
}

// IRelationTupleRepository -
type IRelationTupleRepository interface {
	Migratable
	QueryTuples(ctx context.Context, namespace string, objectID string, relation string) (entities.RelationTuples, errors.Error)
	Read(ctx context.Context, filter filters.RelationTupleFilter) (entities.RelationTuples, errors.Error)
	Write(context.Context, entities.RelationTuples) errors.Error
	Delete(context.Context, entities.RelationTuples) errors.Error
}

// IEntityConfigRepository -
type IEntityConfigRepository interface {
	Migratable
	All(ctx context.Context, version string) (configs entities.EntityConfigs, err errors.Error)
	Read(ctx context.Context, name string, version string) (config entities.EntityConfig, err errors.Error)
	Write(ctx context.Context, configs entities.EntityConfigs, version string) (err errors.Error)
}
