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

type NoopRelationshipReader struct{}

func NewNoopRelationshipReader() RelationshipReader {
	return &NoopRelationshipReader{}
}

func (f *NoopRelationshipReader) QueryRelationships(_ context.Context, _ string, _ *base.TupleFilter, _ string) (*database.TupleIterator, error) {
	return database.NewTupleIterator(), nil
}

func (f *NoopRelationshipReader) ReadRelationships(_ context.Context, _ string, _ *base.TupleFilter, _ string, _ database.Pagination) (*database.TupleCollection, database.EncodedContinuousToken, error) {
	return database.NewTupleCollection(), database.NewNoopContinuousToken().Encode(), nil
}

func (f *NoopRelationshipReader) HeadSnapshot(_ context.Context, _ string) (token.SnapToken, error) {
	return token.NewNoopToken(), nil
}

// RelationshipWriter - Writes relation tuples to the storage.
type RelationshipWriter interface {
	// WriteRelationships writes relation tuples to the storage.
	WriteRelationships(ctx context.Context, tenantID string, collection *database.TupleCollection) (token token.EncodedSnapToken, err error)
	// DeleteRelationships deletes relation tuples from the storage.
	DeleteRelationships(ctx context.Context, tenantID string, filter *base.TupleFilter) (token token.EncodedSnapToken, err error)
}

type NoopRelationshipWriter struct{}

func NewNoopRelationshipWriter() RelationshipWriter {
	return &NoopRelationshipWriter{}
}

func (n *NoopRelationshipWriter) WriteRelationships(_ context.Context, _ string, _ *database.TupleCollection) (token.EncodedSnapToken, error) {
	return token.NewNoopToken().Encode(), nil
}

func (n *NoopRelationshipWriter) DeleteRelationships(_ context.Context, _ string, _ *base.TupleFilter) (token.EncodedSnapToken, error) {
	return token.NewNoopToken().Encode(), nil
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

type NoopSchemaReader struct{}

func NewNoopSchemaReader() SchemaReader {
	return &NoopSchemaReader{}
}

func (n *NoopSchemaReader) ReadSchema(_ context.Context, _ string, _ string) (*base.SchemaDefinition, error) {
	return &base.SchemaDefinition{}, nil
}

func (n *NoopSchemaReader) ReadSchemaDefinition(_ context.Context, _ string, _, _ string) (*base.EntityDefinition, string, error) {
	return &base.EntityDefinition{}, "", nil
}

func (n *NoopSchemaReader) HeadVersion(_ context.Context, _ string) (string, error) {
	return "", nil
}

// SchemaWriter - Writes schema definitions to the storage.
type SchemaWriter interface {
	// WriteSchema writes schema to the storage.
	WriteSchema(ctx context.Context, definitions []SchemaDefinition) (err error)
}

type NoopSchemaWriter struct{}

func NewNoopSchemaWriter() SchemaWriter {
	return &NoopSchemaWriter{}
}

func (n *NoopSchemaWriter) WriteSchema(_ context.Context, _ []SchemaDefinition) error {
	return nil
}

// Watcher - Watches relation tuple changes from the storage.
type Watcher interface {
	// Watch watches relation tuple changes from the storage.
	Watch(ctx context.Context, tenantID string, snap string) (<-chan *base.TupleChanges, <-chan error)
}

type NoopWatcher struct{}

func NewNoopWatcher() Watcher {
	return &NoopWatcher{}
}

func (n *NoopWatcher) Watch(_ context.Context, _ string, _ string) (<-chan *base.TupleChanges, <-chan error) {
	// Create empty channels
	tupleChanges := make(chan *base.TupleChanges)
	errs := make(chan error)

	// Close the channels immediately
	close(tupleChanges)
	close(errs)

	return tupleChanges, errs
}

// TenantReader - Reads tenants from the storage.
type TenantReader interface {
	// ListTenants reads tenants from the storage.
	ListTenants(ctx context.Context, pagination database.Pagination) (tenants []*base.Tenant, ct database.EncodedContinuousToken, err error)
}

type NoopTenantReader struct{}

func NewNoopTenantReader() TenantReader {
	return &NoopTenantReader{}
}

func (n *NoopTenantReader) ListTenants(_ context.Context, _ database.Pagination) ([]*base.Tenant, database.EncodedContinuousToken, error) {
	return []*base.Tenant{}, database.NewNoopContinuousToken().Encode(), nil
}

// TenantWriter - Writes tenants to the storage.
type TenantWriter interface {
	// CreateTenant writes tenant to the storage.
	CreateTenant(ctx context.Context, id, name string) (tenant *base.Tenant, err error)
	// DeleteTenant deletes tenant from the storage.
	DeleteTenant(ctx context.Context, tenantID string) (tenant *base.Tenant, err error)
}

type NoopTenantWriter struct{}

func NewNoopTenantWriter() TenantWriter {
	return &NoopTenantWriter{}
}

func (n *NoopTenantWriter) CreateTenant(_ context.Context, _, _ string) (*base.Tenant, error) {
	return &base.Tenant{}, nil
}

func (n *NoopTenantWriter) DeleteTenant(_ context.Context, _ string) (*base.Tenant, error) {
	return &base.Tenant{}, nil
}
