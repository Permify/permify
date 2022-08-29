package commands

import (
	"context"

	"github.com/Permify/permify/pkg/dsl/schema"
)

// ICheckCommand -
type ICheckCommand interface {
	Execute(ctx context.Context, q *CheckQuery, child schema.Child) (response CheckResponse, err error)
}

// IExpandCommand -
type IExpandCommand interface {
	Execute(ctx context.Context, q *ExpandQuery, child schema.Child) (response ExpandResponse, err error)
}

// ISchemaLookupCommand -
type ISchemaLookupCommand interface {
	Execute(ctx context.Context, q *SchemaLookupQuery, actions []schema.Action) (response SchemaLookupResponse, err error)
}
