package commands

import (
	"context"

	"github.com/Permify/permify/pkg/dsl/schema"
)

// ICheckCommand -
type ICheckCommand interface {
	Execute(ctx context.Context, q *CheckQuery, child schema.Child) (response CheckResponse)
}

// IExpandCommand -
type IExpandCommand interface {
	Execute(ctx context.Context, q *ExpandQuery, child schema.Child) (response ExpandResponse)
}
