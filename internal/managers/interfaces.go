package managers

import (
	"context"

	"github.com/Permify/permify/internal/entities"
	"github.com/Permify/permify/pkg/dsl/schema"
)

// IEntityConfigManager -
type IEntityConfigManager interface {
	All(ctx context.Context, version string) (sch schema.Schema, err error)
	Read(ctx context.Context, name string, version string) (entity schema.Entity, err error)
	Write(ctx context.Context, configs entities.EntityConfigs) (version string, err error)
}
