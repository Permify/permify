package development

import (
	"context"

	"github.com/Permify/permify/internal/commands"
	"github.com/Permify/permify/internal/services"
	"github.com/Permify/permify/pkg/database"
	v1 "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/token"
)

// Check -
func Check(ctx context.Context, service services.IPermissionService, subject *v1.Subject, action string, entity *v1.Entity, version string) (res commands.CheckResponse, err error) {
	return service.CheckPermissions(ctx, subject, action, entity, version, 20)
}

// LookupQuery -
func LookupQuery(ctx context.Context, service services.IPermissionService, entityType string, action string, subject *v1.Subject, version string) (res commands.LookupQueryResponse, err error) {
	return service.LookupQueryPermissions(ctx, entityType, subject, action, version)
}

// ReadTuple -
func ReadTuple(ctx context.Context, service services.IRelationshipService, filter *v1.TupleFilter, token token.SnapToken) (tuples database.ITupleCollection, err error) {
	return service.ReadRelationships(ctx, filter, token)
}

// WriteTuple -
func WriteTuple(ctx context.Context, service services.IRelationshipService, tuples []*v1.Tuple, version string) (token token.SnapToken, err error) {
	return service.WriteRelationships(ctx, tuples, version)
}

// DeleteTuple -
func DeleteTuple(ctx context.Context, service services.IRelationshipService, filter *v1.TupleFilter) (token token.SnapToken, err error) {
	return service.DeleteRelationships(ctx, filter)
}

// WriteSchema -
func WriteSchema(ctx context.Context, service services.ISchemaService, schema string) (version string, err error) {
	return service.WriteSchema(ctx, schema)
}

// ReadSchema -
func ReadSchema(ctx context.Context, service services.ISchemaService, version string) (sch *v1.IndexedSchema, err error) {
	return service.ReadSchema(ctx, version)
}
