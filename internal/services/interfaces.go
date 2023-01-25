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
	ReadRelationships(ctx context.Context, tenantID string, filter *base.TupleFilter, snap string, size uint32, continuousToken string) (*database.TupleCollection, database.EncodedContinuousToken, error)
	WriteRelationships(ctx context.Context, tenantID string, tuples []*base.Tuple, version string) (token.EncodedSnapToken, error)
	DeleteRelationships(ctx context.Context, tenantID string, filter *base.TupleFilter) (token.EncodedSnapToken, error)
}

// ISchemaService -
type ISchemaService interface {
	ReadSchema(ctx context.Context, tenantID string, version string) (response *base.IndexedSchema, err error)
	WriteSchema(ctx context.Context, tenantID string, schema string) (version string, err error)
}

// ITenancyService -
type ITenancyService interface {
	CreateTenant(ctx context.Context, id, name string) (tenant *base.Tenant, err error)
	DeleteTenant(ctx context.Context, tenantID string) (tenant *base.Tenant, err error)
	ListTenants(ctx context.Context, size uint32, ct string) (tenants []*base.Tenant, continuousToken database.EncodedContinuousToken, err error)
}
