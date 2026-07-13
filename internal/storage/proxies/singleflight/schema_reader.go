package singleflight

import (
	"context"

	"resenje.org/singleflight"

	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/pkg/database"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

type headVersionResult struct {
	version        string
	sharedSchemaID string
}

// SchemaReader - Add singleflight behaviour to schema reader
type SchemaReader struct {
	delegate storage.SchemaReader
	group    singleflight.Group[string, headVersionResult]
}

// NewSchemaReader - Add singleflight behaviour to new schema reader
func NewSchemaReader(delegate storage.SchemaReader) *SchemaReader {
	return &SchemaReader{delegate: delegate}
}

// ReadSchema returns the schema definition for a specific tenant and version as a structured object.
func (r *SchemaReader) ReadSchema(ctx context.Context, tenantID, sharedSchemaID, version string) (*base.SchemaDefinition, error) {
	return r.delegate.ReadSchema(ctx, tenantID, sharedSchemaID, version)
}

// ReadSchemaString returns the schema definition for a specific tenant and version as a string.
func (r *SchemaReader) ReadSchemaString(ctx context.Context, tenantID, sharedSchemaID, version string) (definitions []string, err error) {
	return r.delegate.ReadSchemaString(ctx, tenantID, sharedSchemaID, version)
}

// ReadEntityDefinition - Read entity definition from repository
func (r *SchemaReader) ReadEntityDefinition(ctx context.Context, tenantID, sharedSchemaID, entityName, version string) (*base.EntityDefinition, string, error) {
	return r.delegate.ReadEntityDefinition(ctx, tenantID, sharedSchemaID, entityName, version)
}

// ReadRuleDefinition - Read rule definition from repository
func (r *SchemaReader) ReadRuleDefinition(ctx context.Context, tenantID, sharedSchemaID, ruleName, version string) (*base.RuleDefinition, string, error) {
	return r.delegate.ReadRuleDefinition(ctx, tenantID, sharedSchemaID, ruleName, version)
}

// HeadVersion - Finds the latest version of the schema.
func (r *SchemaReader) HeadVersion(ctx context.Context, tenantID string) (string, string, error) {
	res, _, err := r.group.Do(ctx, tenantID, func(ctx context.Context) (headVersionResult, error) {
		sharedSchemaID, version, err := r.delegate.HeadVersion(ctx, tenantID)
		return headVersionResult{version: version, sharedSchemaID: sharedSchemaID}, err
	})
	return res.sharedSchemaID, res.version, err
}

// ListSchemas - List all Schemas
func (r *SchemaReader) ListSchemas(ctx context.Context, tenantID, sharedSchemaID string, pagination database.Pagination) (schemas []*base.SchemaList, ct database.EncodedContinuousToken, err error) {
	return r.delegate.ListSchemas(ctx, tenantID, sharedSchemaID, pagination)
}
