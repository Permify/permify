package development

import (
	"context"

	"github.com/Permify/permify/internal/commands"
	"github.com/Permify/permify/internal/managers"
	"github.com/Permify/permify/internal/repositories/filters"
	"github.com/Permify/permify/internal/services"
	"github.com/Permify/permify/pkg/dsl/schema"
	"github.com/Permify/permify/pkg/errors"
	"github.com/Permify/permify/pkg/tuple"
)

// CheckQuery -
type CheckQuery struct {
	Subject tuple.Subject
	Action  string
	Entity  tuple.Entity
}

// Check -
func Check(ctx context.Context, service services.IPermissionService, subject tuple.Subject, action string, entity tuple.Entity, version string) (res commands.CheckResponse, err errors.Error) {
	return service.Check(ctx, subject, action, entity, version, 20)
}

// LookupQueryQuery -
type LookupQueryQuery struct {
	EntityType string
	Action     string
	Subject    tuple.Subject
}

// LookupQuery -
func LookupQuery(ctx context.Context, service services.IPermissionService, entityType string, action string, subject tuple.Subject, version string) (res commands.LookupQueryResponse, err errors.Error) {
	return service.LookupQuery(ctx, entityType, subject, action, version)
}

// WriteTuple -
func WriteTuple(ctx context.Context, service services.IRelationshipService, tuple tuple.Tuple, version string) (err errors.Error) {
	return service.WriteRelationship(ctx, tuple, version)
}

// ReadTuple -
func ReadTuple(ctx context.Context, service services.IRelationshipService, filter filters.RelationTupleFilter) (tuples []tuple.Tuple, err errors.Error) {
	return service.ReadRelationships(ctx, filter)
}

// DeleteTuple -
func DeleteTuple(ctx context.Context, service services.IRelationshipService, tuple tuple.Tuple) (err errors.Error) {
	return service.DeleteRelationship(ctx, tuple)
}

// WriteSchema -
func WriteSchema(ctx context.Context, manager managers.IEntityConfigManager, configs string) (version string, err errors.Error) {
	return manager.Write(ctx, configs)
}

// ReadSchema -
func ReadSchema(ctx context.Context, manager managers.IEntityConfigManager, version string) (sch schema.Schema, err errors.Error) {
	return manager.All(ctx, version)
}
