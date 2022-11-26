package services

import (
	"context"

	"github.com/Permify/permify/pkg/database"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/token"
)

// IPermissionService -
type IPermissionService interface {
	CheckPermissions(ctx context.Context, request *base.PermissionCheckRequest) (response *base.PermissionCheckResponse, err error)
	ExpandPermissions(ctx context.Context, request *base.PermissionExpandRequest) (response *base.PermissionExpandResponse, err error)
	LookupSchema(ctx context.Context, request *base.PermissionLookupSchemaRequest) (response *base.PermissionLookupSchemaResponse, err error)
	LookupEntity(ctx context.Context, request *base.PermissionLookupEntityRequest) (response *base.PermissionLookupEntityResponse, err error)
	LookupEntityStream(ctx context.Context, request *base.PermissionLookupEntityRequest, server base.Permission_LookupEntityStreamServer) (err error)
}

// IRelationshipService -
type IRelationshipService interface {
	ReadRelationships(ctx context.Context, filter *base.TupleFilter, token string) (database.ITupleCollection, error)
	WriteRelationships(ctx context.Context, tuples []*base.Tuple, version string) (token.EncodedSnapToken, error)
	DeleteRelationships(ctx context.Context, filter *base.TupleFilter) (token.EncodedSnapToken, error)
}

// ISchemaService -
type ISchemaService interface {
	ReadSchema(ctx context.Context, version string) (response *base.IndexedSchema, err error)
	WriteSchema(ctx context.Context, schema string) (version string, err error)
}
