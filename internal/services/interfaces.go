package services

import (
	"context"

	"github.com/Permify/permify/internal/commands"
	"github.com/Permify/permify/pkg/errors"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
)

// IPermissionService -
type IPermissionService interface {
	Check(ctx context.Context, subject *base.Subject, action string, entity *base.Entity, version string, d int32) (response commands.CheckResponse, err errors.Error)
	Expand(ctx context.Context, entity *base.Entity, action string, version string) (response commands.ExpandResponse, err errors.Error)
	LookupQuery(ctx context.Context, entityType string, subject *base.Subject, action string, version string) (response commands.LookupQueryResponse, err errors.Error)
}

// IRelationshipService -
type IRelationshipService interface {
	ReadRelationships(ctx context.Context, filter *base.TupleFilter) (tuple.ITupleCollection, errors.Error)
	WriteRelationship(ctx context.Context, tuple *base.Tuple, version string) errors.Error
	DeleteRelationship(ctx context.Context, tuple *base.Tuple) errors.Error
}

// ISchemaService -
type ISchemaService interface {
	Lookup(ctx context.Context, entityType string, relationNames []string, version string) (response commands.SchemaLookupResponse, err errors.Error)
}
