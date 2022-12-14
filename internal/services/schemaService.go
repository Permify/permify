package services

import (
	"context"

	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/pkg/dsl/ast"
	"github.com/Permify/permify/pkg/dsl/compiler"
	"github.com/Permify/permify/pkg/dsl/parser"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// SchemaService -
type SchemaService struct {
	// repositories
	sw repositories.SchemaWriter
	sr repositories.SchemaReader
}

// NewSchemaService -
func NewSchemaService(sw repositories.SchemaWriter, sr repositories.SchemaReader) *SchemaService {
	return &SchemaService{
		sw: sw,
		sr: sr,
	}
}

// ReadSchema -
func (service *SchemaService) ReadSchema(ctx context.Context, version string) (response *base.IndexedSchema, err error) {
	if version == "" {
		var ver string
		ver, err = service.sr.HeadVersion(ctx)
		if err != nil {
			return response, err
		}
		version = ver
	}

	return service.sr.ReadSchema(ctx, version)
}

// WriteSchema -
func (service *SchemaService) WriteSchema(ctx context.Context, schema string) (response string, err error) {
	sch, err := parser.NewParser(schema).Parse()
	if err != nil {
		return "", err
	}

	_, err = compiler.NewCompiler(false, sch).Compile()
	if err != nil {
		return "", err
	}

	cnf := make([]repositories.SchemaDefinition, 0, len(sch.Statements))
	for _, st := range sch.Statements {
		cnf = append(cnf, repositories.SchemaDefinition{
			EntityType:           st.(*ast.EntityStatement).Name.Literal,
			SerializedDefinition: []byte(st.String()),
		})
	}

	return service.sw.WriteSchema(ctx, cnf)
}
