package services

import (
	"context"

	"github.com/Permify/permify/internal/commands"
	"github.com/Permify/permify/internal/managers"
	"github.com/Permify/permify/pkg/errors"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// SchemaService -
type SchemaService struct {
	manager managers.IEntityConfigManager

	// commands
	schemaLookup commands.ISchemaLookupCommand
}

// NewSchemaService -
func NewSchemaService(sc commands.ISchemaLookupCommand, en managers.IEntityConfigManager) *SchemaService {
	return &SchemaService{
		manager:      en,
		schemaLookup: sc,
	}
}

// Lookup -
func (service *SchemaService) Lookup(ctx context.Context, entityType string, relations []string, version string) (response commands.SchemaLookupResponse, err errors.Error) {
	var en *base.EntityDefinition
	en, err = service.manager.Read(ctx, entityType, version)
	if err != nil {
		return
	}

	q := &commands.SchemaLookupQuery{
		Relations: relations,
	}

	return service.schemaLookup.Execute(ctx, q, en.Actions)
}
