package development

import (
	"context"

	"github.com/Permify/permify/internal/commands"
	"github.com/Permify/permify/internal/managers"
	"github.com/Permify/permify/internal/services"
	"github.com/Permify/permify/pkg/dsl/schema"
	"github.com/Permify/permify/pkg/tuple"
)

// CheckQuery -
type CheckQuery struct {
	Subject tuple.Subject
	Action  string
	Entity  tuple.Entity
}

// Check -
func Check(ctx context.Context, service services.IPermissionService, subject tuple.Subject, action string, entity tuple.Entity) (res commands.CheckResponse, err error) {
	return service.Check(ctx, subject, action, entity, "", 20)
}

// WriteTuple -
func WriteTuple(ctx context.Context, service services.IRelationshipService, tuple tuple.Tuple) (err error) {
	return service.WriteRelationship(ctx, tuple, "")
}

// DeleteTuple -
func DeleteTuple(ctx context.Context, service services.IRelationshipService, tuple tuple.Tuple) (err error) {
	return service.DeleteRelationship(ctx, tuple)
}

// WriteSchema -
func WriteSchema(ctx context.Context, manager managers.IEntityConfigManager, configs string) (version string, err error) {
	return manager.Write(ctx, configs)
}

// ReadSchema -
func ReadSchema(ctx context.Context, manager managers.IEntityConfigManager) (sch schema.Schema, err error) {
	return manager.All(ctx, "")
}
