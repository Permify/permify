package repositories

import (
	"context"

	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
)

// Migratable -
type Migratable interface {
	Migrate() error
}

// IRelationTupleRepository -
type IRelationTupleRepository interface {
	Migratable
	QueryTuples(ctx context.Context, entityType string, entityID string, relation string) (tuple.ITupleIterator, error)
	ReverseQueryTuples(ctx context.Context, entity string, relation string, subjectEntity string, subjectIDs []string, subjectRelation string) (tuple.ITupleIterator, error)
	Read(ctx context.Context, filter *base.TupleFilter) (tuple.ITupleCollection, error)
	Write(context.Context, tuple.ITupleIterator) error
	Delete(context.Context, tuple.ITupleIterator) error
}

// IEntityConfigRepository -
type IEntityConfigRepository interface {
	Migratable
	All(ctx context.Context, version string) (configs []EntityConfig, err error)
	Read(ctx context.Context, name string, version string) (config EntityConfig, err error)
	Write(ctx context.Context, configs []EntityConfig, version string) (err error)
}
