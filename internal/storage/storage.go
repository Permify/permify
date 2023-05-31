package storage

import (
	"context"

	"github.com/Permify/permify/pkg/database"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/token"
)

// RelationshipReader - Reads relation tuples from the storage.
type RelationshipReader interface {
	// QueryRelationships reads relation tuples from the storage.
	QueryRelationships(ctx context.Context, tenantID string, filter *base.TupleFilter, snap string) (iterator *database.TupleIterator, err error)
	// ReadRelationships reads relation tuples from the storage with different options.
	ReadRelationships(ctx context.Context, tenantID string, filter *base.TupleFilter, snap string, pagination database.Pagination) (collection *database.TupleCollection, ct database.EncodedContinuousToken, err error)
	// HeadSnapshot reads the latest version of the snapshot from the storage.
	HeadSnapshot(ctx context.Context, tenantID string) (token.SnapToken, error)
}

// RelationshipWriter - Writes relation tuples to the storage.
type RelationshipWriter interface {
	// WriteRelationships writes relation tuples to the storage.
	WriteRelationships(ctx context.Context, tenantID string, collection *database.TupleCollection) (token token.EncodedSnapToken, err error)
	// DeleteRelationships deletes relation tuples from the storage.
	DeleteRelationships(ctx context.Context, tenantID string, filter *base.TupleFilter) (token token.EncodedSnapToken, err error)
}

// SchemaReader - Reads schema definitions from the storage.
type SchemaReader interface {
	// ReadSchema reads entity config from the storage.
	ReadSchema(ctx context.Context, tenantID string, version string) (schema *base.SchemaDefinition, err error)
	// ReadSchemaDefinition reads entity config from the storage.
	ReadSchemaDefinition(ctx context.Context, tenantID string, entityType, version string) (definition *base.EntityDefinition, v string, err error)
	// HeadVersion reads the latest version of the schema from the storage.
	HeadVersion(ctx context.Context, tenantID string) (version string, err error)
}

// SchemaWriter - Writes schema definitions to the storage.
type SchemaWriter interface {
	// WriteSchema writes schema to the storage.
	WriteSchema(ctx context.Context, definitions []SchemaDefinition) (err error)
}

// Watcher - Watches relation tuple changes from the storage.
type Watcher interface {
	// Watch watches relation tuple changes from the storage.
	Watch(ctx context.Context, tenantID string, snap string) (<-chan *base.TupleChanges, <-chan error)
}

// TenantReader - Reads tenants from the storage.
type TenantReader interface {
	// ListTenants reads tenants from the storage.
	ListTenants(ctx context.Context, pagination database.Pagination) (tenants []*base.Tenant, ct database.EncodedContinuousToken, err error)
}

// TenantWriter - Writes tenants to the storage.
type TenantWriter interface {
	// CreateTenant writes tenant to the storage.
	CreateTenant(ctx context.Context, id, name string) (tenant *base.Tenant, err error)
	// DeleteTenant deletes tenant from the storage.
	DeleteTenant(ctx context.Context, tenantID string) (tenant *base.Tenant, err error)
}
