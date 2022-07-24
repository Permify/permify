package services

import (
	"strings"

	"golang.org/x/net/context"

	e "github.com/Permify/permify/internal/entities"
	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/pkg/dsl/parser"
	"github.com/Permify/permify/pkg/dsl/schema"
)

// ISchemaService -
type ISchemaService interface {
	Schema(ctx context.Context) (sch schema.Schema, err error)
	Replace(ctx context.Context, configs []e.EntityConfig) (err error)
}

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

// Schema -
func (service *SchemaService) Schema(ctx context.Context) (sch schema.Schema, err error) {
	if service.cache != nil {
		return *service.cache, nil
	}

	var cn []e.EntityConfig
	cn, err = service.repository.All(ctx)
	if err != nil {
		return schema.Schema{}, err
	}

	var ecs []string
	for _, c := range cn {
		ecs = append(ecs, string(c.SerializedConfig))
	}

	pr := parser.NewParser(strings.Join(ecs, "\n"))
	parsed := pr.Parse()
	if pr.Error() != nil {
		return schema.Schema{}, pr.Error()
	}

	var s *parser.SchemaTranslator
	s, err = parser.NewSchemaTranslator(parsed)
	if err != nil {
		return schema.Schema{}, err
	}

	return s.Translate(), err
}

// Replace -
func (service *SchemaService) Replace(ctx context.Context, configs []e.EntityConfig) (err error) {
	service.cache = nil
	return service.repository.Replace(ctx, configs)
}
