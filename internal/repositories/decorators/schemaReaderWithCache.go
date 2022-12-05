package decorators

import (
	"context"
	"errors"
	"fmt"
	`reflect`

	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/pkg/cache"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// SchemaReaderWithCache - Add cache behaviour to schema reader
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

// ReadSchema  - Read schema from the repository
func (r *SchemaReaderWithCache) ReadSchema(ctx context.Context, version string) (schema *base.IndexedSchema, err error) {
	return r.delegate.ReadSchema(ctx, version)
}

// ReadSchemaDefinition - Read schema definition from the repository
func (r *SchemaReaderWithCache) ReadSchemaDefinition(ctx context.Context, entityType string, version string) (definition *base.EntityDefinition, v string, err error) {
	var s interface{}
	found := false
	if version != "" {
		s, found = r.cache.Get(fmt.Sprintf("%s|%s", entityType, version))
	}
	if !found {
		definition, version, err = r.delegate.ReadSchemaDefinition(ctx, entityType, version)
		if err != nil {
			return nil, "", err
		}
		size := reflect.TypeOf(definition).Size()
		r.cache.Set(fmt.Sprintf("%s|%s", entityType, version), definition, int64(size))
		return definition, version, nil
	}
	def, ok := s.(*base.EntityDefinition)
	if !ok {
		return nil, "", errors.New(base.ErrorCode_ERROR_CODE_SCAN.String())
	}
	return def, "", err
}

// HeadVersion - Finds the latest version of the schema.
func (r *SchemaReaderWithCache) HeadVersion(ctx context.Context) (version string, err error) {
	return r.delegate.HeadVersion(ctx)
}
