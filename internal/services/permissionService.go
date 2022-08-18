package services

import (
	"context"

	"github.com/Permify/permify/internal/commands"
	"github.com/Permify/permify/internal/entities"
	internalErrors "github.com/Permify/permify/internal/internal-errors"
	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/pkg/dsl/schema"
	"github.com/Permify/permify/pkg/tuple"
)

// PermissionService -
type PermissionService struct {
	entityConfigRepository repositories.IEntityConfigRepository

	// commands
	check  commands.ICheckCommand
	expand commands.IExpandCommand
}

// NewPermissionService -
func NewPermissionService(cc commands.ICheckCommand, ec commands.IExpandCommand, en repositories.IEntityConfigRepository) *PermissionService {
	return &PermissionService{
		entityConfigRepository: en,
		check:                  cc,
		expand:                 ec,
	}
}

// Check -
func (service *PermissionService) Check(ctx context.Context, subject tuple.Subject, action string, entity tuple.Entity, version string, d int32) (response commands.CheckResponse, err error) {
	var cnf entities.EntityConfig
	cnf, err = service.entityConfigRepository.Read(ctx, entity.Type, version)
	if err != nil {
		return
	}

	var sch schema.Schema
	sch, err = cnf.ToSchema()
	if err != nil {
		return
	}

	var child schema.Child

	var en schema.Entity
	en = sch.GetEntityByName(entity.Type)
	for _, e := range en.Actions {
		if e.Name == action {
			child = e.Child
			goto check
		}
	}

	return response, internalErrors.ActionCannotFoundError

check:

	q := &commands.CheckQuery{
		Entity:  entity,
		Subject: subject,
	}

	q.SetDepth(d)

	return service.check.Execute(ctx, q, child)
}

// Expand -
func (service *PermissionService) Expand(ctx context.Context, entity tuple.Entity, action string, version string) (response commands.ExpandResponse, err error) {
	var cnf entities.EntityConfig
	cnf, err = service.entityConfigRepository.Read(ctx, entity.Type, version)
	if err != nil {
		return
	}

	var sch schema.Schema
	sch, err = cnf.ToSchema()
	if err != nil {
		return
	}

	var child schema.Child

	var en schema.Entity
	en = sch.GetEntityByName(entity.Type)
	for _, e := range en.Actions {
		if e.Name == action {
			child = e.Child
			goto expand
		}
	}

	return response, internalErrors.ActionCannotFoundError

expand:

	q := &commands.ExpandQuery{
		Entity: entity,
	}

	return service.expand.Execute(ctx, q, child)
}
