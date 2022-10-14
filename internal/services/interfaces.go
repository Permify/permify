package services

import (
	"context"

	"github.com/Permify/permify/internal/commands"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
)

// IPermissionService -
type IPermissionService interface {
	Check(ctx context.Context, subject *base.Subject, action string, entity *base.Entity, version string, d int32) (response commands.CheckResponse, err error)
	Expand(ctx context.Context, entity *base.Entity, action string, version string) (response commands.ExpandResponse, err error)
	LookupQuery(ctx context.Context, entityType string, subject *base.Subject, action string, version string) (response commands.LookupQueryResponse, err error)
}

// IRelationshipService -
type IRelationshipService interface {
	ReadRelationships(ctx context.Context, filter *base.TupleFilter) (tuple.ITupleCollection, error)
	WriteRelationship(ctx context.Context, tuple *base.Tuple, version string) error
	DeleteRelationship(ctx context.Context, tuple *base.Tuple) error
}

// ISchemaService -
type ISchemaService interface {
	Lookup(ctx context.Context, entityType string, relationNames []string, version string) (response commands.SchemaLookupResponse, err error)
}
