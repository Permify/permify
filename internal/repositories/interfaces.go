package repositories

import (
	"context"

	"github.com/Permify/permify/pkg/errors"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
)

// Migratable -
type Migratable interface {
	Migrate() errors.Error
}

// IRelationTupleRepository -
type IRelationTupleRepository interface {
	Migratable
	QueryTuples(ctx context.Context, entityType string, entityID string, relation string) (tuple.ITupleIterator, errors.Error)
	ReverseQueryTuples(ctx context.Context, entity string, relation string, subjectEntity string, subjectIDs []string, subjectRelation string) (tuple.ITupleIterator, errors.Error)
	Read(ctx context.Context, filter *base.TupleFilter) (tuple.ITupleCollection, errors.Error)
	Write(context.Context, tuple.ITupleIterator) errors.Error
	Delete(context.Context, tuple.ITupleIterator) errors.Error
}

// IEntityConfigRepository -
type IEntityConfigRepository interface {
	Migratable
	All(ctx context.Context, version string) (configs []EntityConfig, err errors.Error)
	Read(ctx context.Context, name string, version string) (config EntityConfig, err errors.Error)
	Write(ctx context.Context, configs []EntityConfig, version string) (err errors.Error)
}
