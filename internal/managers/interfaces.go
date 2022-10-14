package managers

import (
	"context"

	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// IEntityConfigManager -
type IEntityConfigManager interface {
	All(ctx context.Context, version string) (sch *base.Schema, err error)
	Read(ctx context.Context, name string, version string) (entity *base.EntityDefinition, err error)
	Write(ctx context.Context, configs string) (version string, err error)
}
