package development

import (
	"context"

	"github.com/Permify/permify/internal/commands"
	"github.com/Permify/permify/internal/managers"
	"github.com/Permify/permify/internal/services"
	v1 "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
)

// Check -
func Check(ctx context.Context, service services.IPermissionService, subject *v1.Subject, action string, entity *v1.Entity, version string) (res commands.CheckResponse, err error) {
	return service.Check(ctx, subject, action, entity, version, 20)
}

// LookupQuery -
func LookupQuery(ctx context.Context, service services.IPermissionService, entityType string, action string, subject *v1.Subject, version string) (res commands.LookupQueryResponse, err error) {
	return service.LookupQuery(ctx, entityType, subject, action, version)
}

// WriteTuple -
func WriteTuple(ctx context.Context, service services.IRelationshipService, tuple *v1.Tuple, version string) (err error) {
	return service.WriteRelationship(ctx, tuple, version)
}

// ReadTuple -
func ReadTuple(ctx context.Context, service services.IRelationshipService, filter *v1.TupleFilter) (tuples tuple.ITupleCollection, err error) {
	return service.ReadRelationships(ctx, filter)
}

// DeleteTuple -
func DeleteTuple(ctx context.Context, service services.IRelationshipService, tuple *v1.Tuple) (err error) {
	return service.DeleteRelationship(ctx, tuple)
}

// WriteSchema -
func WriteSchema(ctx context.Context, manager managers.IEntityConfigManager, configs string) (version string, err error) {
	return manager.Write(ctx, configs)
}

// ReadSchema -
func ReadSchema(ctx context.Context, manager managers.IEntityConfigManager, version string) (sch *v1.Schema, err error) {
	return manager.All(ctx, version)
}
