package services

import (
	"context"

	"github.com/Permify/permify/internal/commands"
	"github.com/Permify/permify/pkg/database"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/token"
)

// IPermissionService -
type IPermissionService interface {
	CheckPermissions(ctx context.Context, subject *base.Subject, action string, entity *base.Entity, version string, d int32) (response commands.CheckResponse, err error)
	ExpandPermissions(ctx context.Context, entity *base.Entity, action string, version string) (response commands.ExpandResponse, err error)
	LookupQueryPermissions(ctx context.Context, entityType string, subject *base.Subject, action string, version string) (response commands.LookupQueryResponse, err error)
}

// IRelationshipService -
type IRelationshipService interface {
	ReadRelationships(ctx context.Context, filter *base.TupleFilter, token token.SnapToken) (database.ITupleCollection, error)
	WriteRelationships(ctx context.Context, tuples []*base.Tuple, version string) (token.SnapToken, error)
	DeleteRelationships(ctx context.Context, filter *base.TupleFilter) (token.SnapToken, error)
}

// ISchemaService -
type ISchemaService interface {
	ReadSchema(ctx context.Context, version string) (response *base.IndexedSchema, err error)
	WriteSchema(ctx context.Context, schema string) (version string, err error)
	LookupSchema(ctx context.Context, entityType string, relationNames []string, version string) (response commands.SchemaLookupResponse, err error)
}
