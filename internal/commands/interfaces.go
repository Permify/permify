package commands

import (
	"context"

	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// ICheckCommand -
type ICheckCommand interface {
	Execute(ctx context.Context, request *base.PermissionCheckRequest) (response *base.PermissionCheckResponse, err error)
}

// IExpandCommand -
type IExpandCommand interface {
	Execute(ctx context.Context, request *base.PermissionExpandRequest) (response *base.PermissionExpandResponse, err error)
}

// ILookupSchemaCommand -
type ILookupSchemaCommand interface {
	Execute(ctx context.Context, request *base.PermissionLookupSchemaRequest) (response *base.PermissionLookupSchemaResponse, err error)
}

// ILookupEntityCommand -
type ILookupEntityCommand interface {
	Execute(ctx context.Context, request *base.PermissionLookupEntityRequest) (response *base.PermissionLookupEntityResponse, err error)
}
