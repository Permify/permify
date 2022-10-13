package development

import (
	"context"

	"github.com/Permify/permify/internal/commands"
	"github.com/Permify/permify/internal/managers"
	"github.com/Permify/permify/internal/services"
	"github.com/Permify/permify/pkg/errors"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
)

// CheckQuery -
type CheckQuery struct {
	Subject *base.Subject
	Action  string
	Entity  *base.Entity
}

// Check -
func Check(ctx context.Context, service services.IPermissionService, subject *base.Subject, action string, entity *base.Entity, version string) (res commands.CheckResponse, err errors.Error) {
	return service.Check(ctx, subject, action, entity, version, 20)
}

// LookupQueryQuery -
type LookupQueryQuery struct {
	EntityType string
	Action     string
	Subject    *base.Subject
}

// LookupQuery -
func LookupQuery(ctx context.Context, service services.IPermissionService, entityType string, action string, subject *base.Subject, version string) (res commands.LookupQueryResponse, err errors.Error) {
	return service.LookupQuery(ctx, entityType, subject, action, version)
}

// WriteTuple -
func WriteTuple(ctx context.Context, service services.IRelationshipService, tuple *base.Tuple, version string) (err errors.Error) {
	return service.WriteRelationship(ctx, tuple, version)
}

// ReadTuple -
func ReadTuple(ctx context.Context, service services.IRelationshipService, filter *base.TupleFilter) (tuples tuple.ITupleCollection, err errors.Error) {
	return service.ReadRelationships(ctx, filter)
}

// DeleteTuple -
func DeleteTuple(ctx context.Context, service services.IRelationshipService, tuple *base.Tuple) (err errors.Error) {
	return service.DeleteRelationship(ctx, tuple)
}

// WriteSchema -
func WriteSchema(ctx context.Context, manager managers.IEntityConfigManager, configs string) (version string, err errors.Error) {
	return manager.Write(ctx, configs)
}

// ReadSchema -
func ReadSchema(ctx context.Context, manager managers.IEntityConfigManager, version string) (sch *base.Schema, err errors.Error) {
	return manager.All(ctx, version)
}
