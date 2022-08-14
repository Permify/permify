package services

import (
	"context"
	"errors"

	"github.com/Permify/permify/internal/commands"
	"github.com/Permify/permify/internal/entities"
	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/pkg/dsl/schema"
	"github.com/Permify/permify/pkg/tuple"
)

var ActionCannotFoundError = errors.New("action cannot found")

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
func (service *PermissionService) Check(ctx context.Context, subject tuple.Subject, action string, entity tuple.Entity, d int) (response commands.CheckResponse) {
	var cnf entities.EntityConfig
	cnf, response.Error = service.entityConfigRepository.Read(ctx, entity.Type)
	if response.Error != nil {
		return
	}

	var sch schema.Schema
	sch, response.Error = cnf.ToSchema()
	if response.Error != nil {
		return
	}

	var child schema.Child
	child = sch.GetEntityByName(entity.Type).GetAction(action).Child
	if child == nil {
		return
	}

	re := &commands.CheckQuery{
		Entity:  entity,
		Subject: subject,
	}

	re.SetDepth(d)

	return service.check.Execute(ctx, re, child)
}

// Expand -
func (service *PermissionService) Expand(ctx context.Context, entity tuple.Entity, action string, d int) (response commands.ExpandResponse) {
	var cnf entities.EntityConfig
	cnf, response.Error = service.entityConfigRepository.Read(ctx, entity.Type)
	if response.Error != nil {
		return
	}

	var sch schema.Schema
	sch, response.Error = cnf.ToSchema()
	if response.Error != nil {
		return
	}

	var child schema.Child
	child = sch.GetEntityByName(entity.Type).GetAction(action).Child
	if child == nil {
		return
	}

	re := &commands.ExpandQuery{
		Entity: entity,
	}

	re.SetDepth(d)

	return service.expand.Execute(ctx, re, child)
}
