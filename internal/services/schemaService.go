package services

import (
	"context"

	"github.com/Permify/permify/internal/commands"
	"github.com/Permify/permify/internal/managers"
	"github.com/Permify/permify/pkg/dsl/schema"
)

// SchemaService -
type SchemaService struct {
	manager managers.IEntityConfigManager

	// commands
	lookup commands.ISchemaLookupCommand
}

// NewSchemaService -
func NewSchemaService(sc commands.ISchemaLookupCommand, en managers.IEntityConfigManager) *SchemaService {
	return &SchemaService{
		manager: en,
		lookup:  sc,
	}
}

// Lookup -
func (service *SchemaService) Lookup(ctx context.Context, entityType string, relations []string, version string) (response commands.SchemaLookupResponse, err error) {
	var en schema.Entity
	en, err = service.manager.Read(ctx, entityType, version)
	if err != nil {
		return
	}

	q := &commands.SchemaLookupQuery{
		Relations: relations,
	}

	return service.lookup.Execute(ctx, q, en.Actions)
}
