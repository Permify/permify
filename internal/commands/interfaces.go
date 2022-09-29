package commands

import (
	"context"

	"github.com/Permify/permify/pkg/dsl/schema"
	"github.com/Permify/permify/pkg/errors"
)

// ICheckCommand -
type ICheckCommand interface {
	Execute(ctx context.Context, q *CheckQuery, child schema.Child) (response CheckResponse, err errors.Error)
}

// IExpandCommand -
type IExpandCommand interface {
	Execute(ctx context.Context, q *ExpandQuery, child schema.Child) (response ExpandResponse, err errors.Error)
}

// ISchemaLookupCommand -
type ISchemaLookupCommand interface {
	Execute(ctx context.Context, q *SchemaLookupQuery, actions []schema.Action) (response SchemaLookupResponse, err errors.Error)
}

// ILookupQueryCommand -
type ILookupQueryCommand interface {
	Execute(ctx context.Context, q *LookupQueryQuery, child schema.Child) (response LookupQueryResponse, err errors.Error)
}
