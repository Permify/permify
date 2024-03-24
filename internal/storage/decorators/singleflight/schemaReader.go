package singleflight

import (
	"context"

	"resenje.org/singleflight"

	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/pkg/database"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// SchemaReader - Add singleflight behaviour to schema reader
type SchemaReader struct {
	delegate storage.SchemaReader
	group    singleflight.Group[string, string]
}

// NewSchemaReader - Add singleflight behaviour to new schema reader
func NewSchemaReader(delegate storage.SchemaReader) *SchemaReader {
	return &SchemaReader{delegate: delegate}
}

// ReadSchema returns the schema definition for a specific tenant and version as a structured object.
func (r *SchemaReader) ReadSchema(ctx context.Context, tenantID, version string) (*base.SchemaDefinition, error) {
	return r.delegate.ReadSchema(ctx, tenantID, version)
}

// ReadSchemaString returns the schema definition for a specific tenant and version as a string.
func (r *SchemaReader) ReadSchemaString(ctx context.Context, tenantID, version string) (definitions []string, err error) {
	return r.delegate.ReadSchemaString(ctx, tenantID, version)
}

// ReadEntityDefinition - Read entity definition from repository
func (r *SchemaReader) ReadEntityDefinition(ctx context.Context, tenantID, entityName, version string) (*base.EntityDefinition, string, error) {
	return r.delegate.ReadEntityDefinition(ctx, tenantID, entityName, version)
}

// ReadRuleDefinition - Read rule definition from repository
func (r *SchemaReader) ReadRuleDefinition(ctx context.Context, tenantID, ruleName, version string) (*base.RuleDefinition, string, error) {
	return r.delegate.ReadRuleDefinition(ctx, tenantID, ruleName, version)
}

// HeadVersion - Finds the latest version of the schema.
func (r *SchemaReader) HeadVersion(ctx context.Context, tenantID string) (version string, err error) {
	rev, _, err := r.group.Do(ctx, "", func(ctx context.Context) (string, error) {
		return r.delegate.HeadVersion(ctx, tenantID)
	})
	return rev, err
}

// ListSchemas - List all Schemas
func (r *SchemaReader) ListSchemas(ctx context.Context, tenantID string, pagination database.Pagination) (schemas []*base.SchemaList, ct database.EncodedContinuousToken, err error) {
	return r.delegate.ListSchemas(ctx, tenantID, pagination)
}
