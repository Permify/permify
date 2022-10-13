package managers

import (
	"context"

	"github.com/Permify/permify/pkg/errors"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// IEntityConfigManager -
type IEntityConfigManager interface {
	All(ctx context.Context, version string) (sch *base.Schema, err errors.Error)
	Read(ctx context.Context, name string, version string) (entity *base.EntityDefinition, err errors.Error)
	Write(ctx context.Context, configs string) (version string, err errors.Error)
}
