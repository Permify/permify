package commands

import (
	"context"

	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/pkg/errors"
	base `github.com/Permify/permify/pkg/pb/base/v1`
)

type ICommand interface {
	GetRelationTupleRepository() repositories.IRelationTupleRepository
}

// ICheckCommand -
type ICheckCommand interface {
	Execute(ctx context.Context, q *CheckQuery, child *base.Child) (response CheckResponse, err errors.Error)
}

// IExpandCommand -
type IExpandCommand interface {
	Execute(ctx context.Context, q *ExpandQuery, child *base.Child) (response ExpandResponse, err errors.Error)
}

// ISchemaLookupCommand -
type ISchemaLookupCommand interface {
	Execute(ctx context.Context, q *SchemaLookupQuery, actions []*base.ActionDefinition) (response SchemaLookupResponse, err errors.Error)
}

// ILookupQueryCommand -
type ILookupQueryCommand interface {
	Execute(ctx context.Context, q *LookupQueryQuery, child *base.Child) (response LookupQueryResponse, err errors.Error)
}
