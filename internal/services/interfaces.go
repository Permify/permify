package services

import (
	"context"

	"github.com/Permify/permify/internal/commands"
	"github.com/Permify/permify/internal/entities"
	"github.com/Permify/permify/pkg/dsl/schema"
	"github.com/Permify/permify/pkg/tuple"
)

// IPermissionService -
type IPermissionService interface {
	Check(ctx context.Context, subject tuple.Subject, action string, entity tuple.Entity, d int) (response commands.CheckResponse)
	Expand(ctx context.Context, entity tuple.Entity, action string, d int) (response commands.ExpandResponse)
}

// ISchemaService -
type ISchemaService interface {
	All(ctx context.Context) (sch schema.Schema, err error)
	Read(ctx context.Context, name string) (sch schema.Schema, err error)
	Replace(ctx context.Context, configs entities.EntityConfigs) (err error)
}
