package services

import (
	"golang.org/x/net/context"

	e "github.com/Permify/permify/internal/entities"
	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/pkg/dsl/schema"
)

// SchemaService -
type SchemaService struct {
	repository repositories.IEntityConfigRepository
	cache      *schema.Schema
}

// NewSchemaService -
func NewSchemaService(repo repositories.IEntityConfigRepository) *SchemaService {
	return &SchemaService{
		repository: repo,
	}
}

// All -
func (service *SchemaService) All(ctx context.Context) (sch schema.Schema, err error) {
	var cn e.EntityConfigs
	cn, err = service.repository.All(ctx)
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
func (service *SchemaService) Read(ctx context.Context, name string) (sch schema.Schema, err error) {
	if service.cache != nil {
		return *service.cache, nil
	}
	var cn e.EntityConfig
	cn, err = service.repository.Read(ctx, name)
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
func (service *SchemaService) Replace(ctx context.Context, configs e.EntityConfigs) (err error) {
	service.cache = nil
	return service.repository.Replace(ctx, configs)
}
