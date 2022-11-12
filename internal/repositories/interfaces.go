package repositories

import (
	"context"

	"github.com/Permify/permify/pkg/database"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/token"
)

// RelationshipReader -
type RelationshipReader interface {
	// QueryRelationships reads relation tuples from the repository.
	QueryRelationships(ctx context.Context, filter *base.TupleFilter, token string) (collection database.ITupleCollection, err error)
	// HeadSnapshot reads the latest version of the snapshot from the repository.
	HeadSnapshot(ctx context.Context) (token.SnapToken, error)
}

// RelationshipWriter -
type RelationshipWriter interface {
	// WriteRelationships writes relation tuples to the repository.
	WriteRelationships(ctx context.Context, collection database.ITupleCollection) (token token.EncodedSnapToken, err error)
	// DeleteRelationships deletes relation tuples from the repository.
	DeleteRelationships(ctx context.Context, filter *base.TupleFilter) (token token.EncodedSnapToken, err error)
}

// SchemaReader -
type SchemaReader interface {
	// ReadSchema reads entity config from the repository.
	ReadSchema(ctx context.Context, version string) (schema *base.IndexedSchema, err error)
	// ReadSchemaDefinition reads entity config from the repository.
	ReadSchemaDefinition(ctx context.Context, entityType string, version string) (definition *base.EntityDefinition, v string, err error)
	// HeadVersion reads the latest version of the schema from the repository.
	HeadVersion(ctx context.Context) (version string, err error)
}

// SchemaWriter -
type SchemaWriter interface {
	// WriteSchema writes schema to the repository.
	WriteSchema(ctx context.Context, definitions []SchemaDefinition) (version string, err error)
}
