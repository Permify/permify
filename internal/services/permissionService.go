package services

import (
	"context"

	"github.com/Permify/permify/internal/commands"
	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/pkg/dsl/schema"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// PermissionService -
type PermissionService struct {
	// repositories
	sr repositories.SchemaReader
	// commands
	cc  commands.ICheckCommand
	ec  commands.IExpandCommand
	lqc commands.ILookupQueryCommand
}

// NewPermissionService -
func NewPermissionService(cc commands.ICheckCommand, ec commands.IExpandCommand, lqc commands.ILookupQueryCommand, sr repositories.SchemaReader) *PermissionService {
	return &PermissionService{
		sr:  sr,
		cc:  cc,
		ec:  ec,
		lqc: lqc,
	}
}

// CheckPermissions -
func (service *PermissionService) CheckPermissions(ctx context.Context, subject *base.Subject, action string, entity *base.Entity, version string, snapToken string, d int32) (response commands.CheckResponse, err error) {
	var en *base.EntityDefinition
	en, _, err = service.sr.ReadSchemaDefinition(ctx, entity.GetType(), version)
	if err != nil {
		return response, err
	}

	var a *base.ActionDefinition
	a, err = schema.GetActionByNameInEntityDefinition(en, action)
	if err != nil {
		return response, err
	}

	child := a.Child

	q := &commands.CheckQuery{
		Entity:    entity,
		Subject:   subject,
		SnapToken: snapToken,
	}

	q.SetDepth(d)

	return service.cc.Execute(ctx, q, child)
}

// ExpandPermissions -
func (service *PermissionService) ExpandPermissions(ctx context.Context, entity *base.Entity, action string, version string, snapToken string) (response commands.ExpandResponse, err error) {
	var en *base.EntityDefinition
	en, _, err = service.sr.ReadSchemaDefinition(ctx, entity.GetType(), version)
	if err != nil {
		return response, err
	}

	var a *base.ActionDefinition
	a, err = schema.GetActionByNameInEntityDefinition(en, action)
	if err != nil {
		return response, err
	}

	child := a.Child

	q := &commands.ExpandQuery{
		Entity:    entity,
		SnapToken: snapToken,
	}

	return service.ec.Execute(ctx, q, child)
}

// LookupQueryPermissions -
func (service *PermissionService) LookupQueryPermissions(ctx context.Context, entityType string, subject *base.Subject, action string, version string) (response commands.LookupQueryResponse, err error) {
	var sch *base.IndexedSchema
	sch, err = service.sr.ReadSchema(ctx, version)
	if err != nil {
		return response, err
	}

	// entityType
	var en *base.EntityDefinition
	en, err = schema.GetEntityByName(sch, entityType)
	if err != nil {
		return response, err
	}

	var a *base.ActionDefinition
	a, err = schema.GetActionByNameInEntityDefinition(en, action)
	if err != nil {
		return response, err
	}

	child := a.Child

	q := &commands.LookupQueryQuery{
		EntityType: entityType,
		Action:     action,
		Subject:    subject,
	}

	q.SetSchema(sch)

	return service.lqc.Execute(ctx, q, child)
}
