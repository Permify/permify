package services

import (
	"github.com/rs/xid"
	"golang.org/x/net/context"

	e "github.com/Permify/permify/internal/entities"
	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/pkg/dsl/schema"
)

// SchemaService -
type SchemaService struct {
	repository repositories.IEntityConfigRepository
}

// NewSchemaService -
func NewSchemaService(repo repositories.IEntityConfigRepository) *SchemaService {
	return &SchemaService{
		repository: repo,
	}
}

// All -
func (service *SchemaService) All(ctx context.Context, version string) (sch schema.Schema, err error) {
	var cn e.EntityConfigs
	cn, err = service.repository.All(ctx, version)
	if err != nil {
		return schema.Schema{}, err
	}
	sch, err = cn.ToSchema()
	if err != nil {
		return schema.Schema{}, err
	}
	return
}

// Read -
func (service *SchemaService) Read(ctx context.Context, name string, version string) (sch schema.Schema, err error) {
	var cn e.EntityConfig
	cn, err = service.repository.Read(ctx, name, version)
	if err != nil {
		return schema.Schema{}, err
	}
	sch, err = cn.ToSchema()
	if err != nil {
		return schema.Schema{}, err
	}
	return
}

// Replace -
func (service *SchemaService) Write(ctx context.Context, configs e.EntityConfigs) (version string, err error) {
	version = xid.New().String()
	return version, service.repository.Write(ctx, configs, version)
}
