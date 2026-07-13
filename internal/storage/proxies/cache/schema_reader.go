package cache

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/pkg/cache"
	"github.com/Permify/permify/pkg/database"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// SchemaReader - Add cache behaviour to schema reader
type SchemaReader struct {
	delegate storage.SchemaReader
	cache    cache.Cache
}

// NewSchemaReader new instance of SchemaReader
func NewSchemaReader(delegate storage.SchemaReader, cache cache.Cache) *SchemaReader {
	return &SchemaReader{
		delegate: delegate,
		cache:    cache,
	}
}

// ReadSchema returns the schema definition for a specific tenant and version as a structured object.
func (r *SchemaReader) ReadSchema(ctx context.Context, tenantID, sharedSchemaID, version string) (schema *base.SchemaDefinition, err error) {
	return r.delegate.ReadSchema(ctx, tenantID, sharedSchemaID, version)
}

// ReadSchemaString returns the schema definition for a specific tenant and version as a string.
func (r *SchemaReader) ReadSchemaString(ctx context.Context, tenantID, sharedSchemaID, version string) (definitions []string, err error) {
	return r.delegate.ReadSchemaString(ctx, tenantID, sharedSchemaID, version)
}

// ReadEntityDefinition - Read entity definition from the repository
func (r *SchemaReader) ReadEntityDefinition(ctx context.Context, tenantID, sharedSchemaID, entityName, version string) (definition *base.EntityDefinition, v string, err error) {
	var s interface{}
	found := false
	if version != "" {
		s, found = r.cache.Get(schemaCacheKey(tenantID, sharedSchemaID, entityName, version))
	}
	if !found {
		definition, version, err = r.delegate.ReadEntityDefinition(ctx, tenantID, sharedSchemaID, entityName, version)
		if err != nil {
			return nil, "", err
		}
		size := reflect.TypeOf(definition).Size()
		r.cache.Set(schemaCacheKey(tenantID, sharedSchemaID, entityName, version), definition, int64(size))
		return definition, version, nil
	}
	def, ok := s.(*base.EntityDefinition)
	if !ok {
		return nil, "", errors.New(base.ErrorCode_ERROR_CODE_SCAN.String())
	}
	return def, "", err
}

// ReadRuleDefinition - Read rule definition from the repository
func (r *SchemaReader) ReadRuleDefinition(ctx context.Context, tenantID, sharedSchemaID, ruleName, version string) (definition *base.RuleDefinition, v string, err error) {
	var s interface{}
	found := false
	if version != "" {
		s, found = r.cache.Get(schemaCacheKey(tenantID, sharedSchemaID, ruleName, version))
	}
	if !found {
		definition, version, err = r.delegate.ReadRuleDefinition(ctx, tenantID, sharedSchemaID, ruleName, version)
		if err != nil {
			return nil, "", err
		}
		size := reflect.TypeOf(definition).Size()
		r.cache.Set(schemaCacheKey(tenantID, sharedSchemaID, ruleName, version), definition, int64(size))
		return definition, version, nil
	}
	def, ok := s.(*base.RuleDefinition)
	if !ok {
		return nil, "", errors.New(base.ErrorCode_ERROR_CODE_SCAN.String())
	}
	return def, "", err
}

// schemaCacheKey builds a cache key. When sharedSchemaID is set, uses it as prefix
// so all tenants sharing the same schema hit the same cache entry.
func schemaCacheKey(tenantID, sharedSchemaID, name, version string) string {
	if sharedSchemaID != "" {
		return fmt.Sprintf("|%s|%s|%s", sharedSchemaID, name, version)
	}
	return fmt.Sprintf("%s||%s|%s", tenantID, name, version)
}

// HeadVersion - Finds the latest version of the schema.
func (r *SchemaReader) HeadVersion(ctx context.Context, tenantID string) (string, string, error) {
	return r.delegate.HeadVersion(ctx, tenantID)
}

// ListSchemas - List all Schemas
func (r *SchemaReader) ListSchemas(ctx context.Context, tenantID, sharedSchemaID string, pagination database.Pagination) (schemas []*base.SchemaList, ct database.EncodedContinuousToken, err error) {
	schemas, ct, err = r.delegate.ListSchemas(ctx, tenantID, sharedSchemaID, pagination)
	if err != nil {
		return nil, nil, err
	}
	return schemas, ct, nil
}
