package decorators

import (
	"context"
	"errors"
	"fmt"

	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/pkg/cache"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

type SchemaReaderWithCache struct {
	delegate repositories.SchemaReader
	cache    cache.Cache
}

// NewSchemaReaderWithCache new instance of SchemaReaderWithCache
func NewSchemaReaderWithCache(delegate repositories.SchemaReader, cache cache.Cache) *SchemaReaderWithCache {
	return &SchemaReaderWithCache{
		delegate: delegate,
		cache:    cache,
	}
}

// ReadSchema -
func (r *SchemaReaderWithCache) ReadSchema(ctx context.Context, version string) (schema *base.IndexedSchema, err error) {
	return r.delegate.ReadSchema(ctx, version)
}

// ReadSchemaDefinition -
func (r *SchemaReaderWithCache) ReadSchemaDefinition(ctx context.Context, entityType string, version string) (definition *base.EntityDefinition, err error) {
	var key string
	var s interface{}
	found := false
	if version != "" {
		key = fmt.Sprintf("%s|%s", entityType, version)
		s, found = r.cache.Get(key)
	}
	if !found {
		definition, err = r.delegate.ReadSchemaDefinition(ctx, entityType, version)
		if err != nil {
			return definition, err
		}
		r.cache.Set(key, definition, 1)
		return definition, nil
	}
	def, ok := s.(*base.EntityDefinition)
	if !ok {
		return nil, errors.New(base.ErrorCode_ERROR_CODE_SCAN.String())
	}
	return def, err
}
