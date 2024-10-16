package storage

import (
	"context"

	"github.com/Permify/permify/pkg/database"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/token"
)

// DataReader - Interface for reading Data from the storage.
type DataReader interface {
	// QueryRelationships reads relation tuples from the storage based on the given filter.
	// It returns an iterator to iterate over the tuples and any error encountered.
	QueryRelationships(ctx context.Context, tenantID string, filter *base.TupleFilter, snap string, pagination database.CursorPagination) (iterator *database.TupleIterator, err error)

	// ReadRelationships reads relation tuples from the storage based on the given filter and pagination.
	// It returns a collection of tuples, a continuous token indicating the position in the data set, and any error encountered.
	ReadRelationships(ctx context.Context, tenantID string, filter *base.TupleFilter, snap string, pagination database.Pagination) (collection *database.TupleCollection, ct database.EncodedContinuousToken, err error)

	// QuerySingleAttribute retrieves a single attribute from the storage based on the given filter.
	// It returns the retrieved attribute and any error encountered.
	QuerySingleAttribute(ctx context.Context, tenantID string, filter *base.AttributeFilter, snap string) (attribute *base.Attribute, err error)

	// QueryAttributes reads multiple attributes from the storage based on the given filter.
	// It returns an iterator to iterate over the attributes and any error encountered.
	QueryAttributes(ctx context.Context, tenantID string, filter *base.AttributeFilter, snap string, pagination database.CursorPagination) (iterator *database.AttributeIterator, err error)

	// ReadAttributes reads multiple attributes from the storage based on the given filter and pagination.
	// It returns a collection of attributes, a continuous token indicating the position in the data set, and any error encountered.
	ReadAttributes(ctx context.Context, tenantID string, filter *base.AttributeFilter, snap string, pagination database.Pagination) (collection *database.AttributeCollection, ct database.EncodedContinuousToken, err error)

	// QueryUniqueSubjectReferences reads unique subject references from the storage based on the given filter and pagination.
	// It returns a slice of subject reference IDs, a continuous token indicating the position in the data set, and any error encountered.
	QueryUniqueSubjectReferences(ctx context.Context, tenantID string, subjectReference *base.RelationReference, excluded []string, snap string, pagination database.Pagination) (ids []string, ct database.EncodedContinuousToken, err error)

	// HeadSnapshot reads the latest version of the snapshot from the storage for a specific tenant.
	// It returns the snapshot token representing the version of the snapshot and any error encountered.
	HeadSnapshot(ctx context.Context, tenantID string) (token.SnapToken, error)
}

type NoopDataReader struct{}

func NewNoopRelationshipReader() DataReader {
	return &NoopDataReader{}
}

func (f *NoopDataReader) QueryRelationships(_ context.Context, _ string, _ *base.TupleFilter, _ string, _ database.CursorPagination) (*database.TupleIterator, error) {
	return database.NewTupleIterator(), nil
}

func (f *NoopDataReader) ReadRelationships(_ context.Context, _ string, _ *base.TupleFilter, _ string, _ database.Pagination) (*database.TupleCollection, database.EncodedContinuousToken, error) {
	return database.NewTupleCollection(), database.NewNoopContinuousToken().Encode(), nil
}

func (f *NoopDataReader) QuerySingleAttribute(_ context.Context, _ string, _ *base.AttributeFilter, _ string) (*base.Attribute, error) {
	return &base.Attribute{}, nil
}

func (f *NoopDataReader) QueryAttributes(_ context.Context, _ string, _ *base.AttributeFilter, _ string, _ database.CursorPagination) (*database.AttributeIterator, error) {
	return database.NewAttributeIterator(), nil
}

func (f *NoopDataReader) ReadAttributes(_ context.Context, _ string, _ *base.AttributeFilter, _ string, _ database.Pagination) (*database.AttributeCollection, database.EncodedContinuousToken, error) {
	return database.NewAttributeCollection(), database.NewNoopContinuousToken().Encode(), nil
}

func (f *NoopDataReader) QueryUniqueSubjectReferences(_ context.Context, _ string, _ *base.RelationReference, _ []string, _ string, _ database.Pagination) ([]string, database.EncodedContinuousToken, error) {
	return []string{}, database.NewNoopContinuousToken().Encode(), nil
}

func (f *NoopDataReader) HeadSnapshot(_ context.Context, _ string) (token.SnapToken, error) {
	return token.NewNoopToken(), nil
}

type DataWriter interface {
	// Write inserts a new TupleCollection and AttributeCollection into the database for a specified tenant.
	// Returns an encoded snapshot token representing the state of the database after the write operation and any error encountered.
	Write(ctx context.Context, tenantID string, tupleCollection *database.TupleCollection, attributesCollection *database.AttributeCollection) (token token.EncodedSnapToken, err error)

	// Delete removes data from the database based on the provided tuple and attribute filters for a specified tenant.
	// Returns an encoded snapshot token representing the state of the database after the delete operation and any error encountered.
	Delete(ctx context.Context, tenantID string, tupleFilter *base.TupleFilter, attributeFilter *base.AttributeFilter) (token token.EncodedSnapToken, err error)

	// RunBundle executes a specified data bundle for a given tenant.
	// Returns an encoded snapshot token representing the state of the database after running the bundle and any error encountered.
	RunBundle(ctx context.Context, tenantID string, arguments map[string]string, bundle *base.DataBundle) (token token.EncodedSnapToken, err error)
}

type NoopDataWriter struct{}

func NewNoopDataWriter() DataWriter {
	return &NoopDataWriter{}
}

func (n *NoopDataWriter) Write(_ context.Context, _ string, _ *database.TupleCollection, _ *database.AttributeCollection) (token.EncodedSnapToken, error) {
	return token.NewNoopToken().Encode(), nil
}

func (n *NoopDataWriter) Delete(_ context.Context, _ string, _ *base.TupleFilter, _ *base.AttributeFilter) (token.EncodedSnapToken, error) {
	return token.NewNoopToken().Encode(), nil
}

func (n *NoopDataWriter) RunBundle(_ context.Context, _ string, _ map[string]string, _ *base.DataBundle) (token.EncodedSnapToken, error) {
	return nil, nil
}

// SchemaReader - Reads schema definitions from the storage.
type SchemaReader interface {
	// ReadSchema returns the schema definition for a specific tenant and version as a structured object.
	ReadSchema(ctx context.Context, tenantID, version string) (schema *base.SchemaDefinition, err error)
	// ReadSchemaString returns the schema definition for a specific tenant and version as a string.
	ReadSchemaString(ctx context.Context, tenantID, version string) (definitions []string, err error)
	// ReadEntityDefinition reads entity config from the storage.
	ReadEntityDefinition(ctx context.Context, tenantID, entityName, version string) (definition *base.EntityDefinition, v string, err error)
	// ReadRuleDefinition reads rule config from the storage.
	ReadRuleDefinition(ctx context.Context, tenantID, ruleName, version string) (definition *base.RuleDefinition, v string, err error)
	// HeadVersion reads the latest version of the schema from the storage.
	HeadVersion(ctx context.Context, tenantID string) (version string, err error)
	// ListSchemas lists all schemas from the storage
	ListSchemas(ctx context.Context, tenantID string, pagination database.Pagination) (schemas []*base.SchemaList, ct database.EncodedContinuousToken, err error)
}

type NoopSchemaReader struct{}

func NewNoopSchemaReader() SchemaReader {
	return &NoopSchemaReader{}
}

func (n *NoopSchemaReader) ReadSchema(_ context.Context, _, _ string) (*base.SchemaDefinition, error) {
	return &base.SchemaDefinition{}, nil
}

func (n *NoopSchemaReader) ReadSchemaString(_ context.Context, _, _ string) ([]string, error) {
	return []string{}, nil
}

func (n *NoopSchemaReader) ReadEntityDefinition(_ context.Context, _, _, _ string) (*base.EntityDefinition, string, error) {
	return &base.EntityDefinition{}, "", nil
}

func (n *NoopSchemaReader) ReadRuleDefinition(_ context.Context, _, _, _ string) (*base.RuleDefinition, string, error) {
	return &base.RuleDefinition{}, "", nil
}

func (n *NoopSchemaReader) HeadVersion(_ context.Context, _ string) (string, error) {
	return "", nil
}

func (n *NoopSchemaReader) ListSchemas(_ context.Context, _ string, _ database.Pagination) (tenants []*base.SchemaList, ct database.EncodedContinuousToken, err error) {
	return nil, nil, nil
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

// BundleReader - Reads data bundles from storage.
type BundleReader interface {
	// Read retrieves a data bundle based on tenant ID and name.
	Read(ctx context.Context, tenantID, name string) (bundle *base.DataBundle, err error)
}

type NoopBundleReader struct{}

func NewNoopBundleReader() BundleReader {
	return &NoopBundleReader{}
}

func (n *NoopBundleReader) Read(_ context.Context, _, _ string) (*base.DataBundle, error) {
	return nil, nil
}

// BundleWriter - Manages writing and deletion of data bundles.
type BundleWriter interface {
	// Write stores bundles in storage for a tenant.
	Write(ctx context.Context, bundles []Bundle) (names []string, err error)

	// Delete removes a bundle from storage for a tenant.
	Delete(ctx context.Context, tenantID, name string) (err error)
}

type NoopBundleWriter struct{}

func NewNoopBundleWriter() BundleWriter {
	return &NoopBundleWriter{}
}

func (n *NoopBundleWriter) Write(_ context.Context, _ []Bundle) (names []string, err error) {
	return nil, nil
}

func (n *NoopBundleWriter) Delete(_ context.Context, _, _ string) error {
	return nil
}

// Watcher - Watches relation tuple changes from the storage.
type Watcher interface {
	// Watch watches relation tuple changes from the storage.
	Watch(ctx context.Context, tenantID, snap string) (<-chan *base.DataChanges, <-chan error)
}

type NoopWatcher struct{}

func NewNoopWatcher() Watcher {
	return &NoopWatcher{}
}

func (n *NoopWatcher) Watch(_ context.Context, _, _ string) (<-chan *base.DataChanges, <-chan error) {
	// Create empty channels
	aclChanges := make(chan *base.DataChanges)
	errs := make(chan error)

	// Close the channels immediately
	close(aclChanges)
	close(errs)

	return aclChanges, errs
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
