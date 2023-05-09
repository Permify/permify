package storage

import (
	"context"

	"github.com/Permify/permify/pkg/database"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/token"
)

// RelationshipReader -
type RelationshipReader interface {
	// QueryRelationships reads relation tuples from the repository.
	QueryRelationships(ctx context.Context, tenantID string, filter *base.TupleFilter, snap string) (iterator *database.TupleIterator, err error)
	// ReadRelationships reads relation tuples from the repository with different options.
	ReadRelationships(ctx context.Context, tenantID string, filter *base.TupleFilter, snap string, pagination database.Pagination) (collection *database.TupleCollection, ct database.EncodedContinuousToken, err error)
	// HeadSnapshot reads the latest version of the snapshot from the repository.
	HeadSnapshot(ctx context.Context, tenantID string) (token.SnapToken, error)
}

// RelationshipWriter -
type RelationshipWriter interface {
	// WriteRelationships writes relation tuples to the repository.
	WriteRelationships(ctx context.Context, tenantID string, collection *database.TupleCollection) (token token.EncodedSnapToken, err error)
	// DeleteRelationships deletes relation tuples from the repository.
	DeleteRelationships(ctx context.Context, tenantID string, filter *base.TupleFilter) (token token.EncodedSnapToken, err error)
}

// SchemaReader -
type SchemaReader interface {
	// ReadSchema reads entity config from the repository.
	ReadSchema(ctx context.Context, tenantID string, version string) (schema *base.SchemaDefinition, err error)
	// ReadSchemaDefinition reads entity config from the repository.
	ReadSchemaDefinition(ctx context.Context, tenantID string, entityType, version string) (definition *base.EntityDefinition, v string, err error)
	// HeadVersion reads the latest version of the schema from the repository.
	HeadVersion(ctx context.Context, tenantID string) (version string, err error)
}

// SchemaWriter -
type SchemaWriter interface {
	// WriteSchema writes schema to the repository.
	WriteSchema(ctx context.Context, definitions []SchemaDefinition) (err error)
}

// TenantReader -
type TenantReader interface {
	// ListTenants reads tenants from the repository.
	ListTenants(ctx context.Context, pagination database.Pagination) (tenants []*base.Tenant, ct database.EncodedContinuousToken, err error)
}

// TenantWriter -
type TenantWriter interface {
	// CreateTenant writes tenant to the repository.
	CreateTenant(ctx context.Context, id, name string) (tenant *base.Tenant, err error)
	// DeleteTenant deletes tenant from the repository.
	DeleteTenant(ctx context.Context, tenantID string) (tenant *base.Tenant, err error)
}
