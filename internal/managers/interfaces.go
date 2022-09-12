package managers

import (
	"context"

	"github.com/Permify/permify/pkg/dsl/schema"
	"github.com/Permify/permify/pkg/errors"
)

// IEntityConfigManager -
type IEntityConfigManager interface {
	All(ctx context.Context, version string) (sch schema.Schema, err errors.Error)
	Read(ctx context.Context, name string, version string) (entity schema.Entity, err errors.Error)
	Write(ctx context.Context, configs string) (version string, err errors.Error)
}
