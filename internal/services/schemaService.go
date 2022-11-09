package services

import (
	"context"

	"github.com/Permify/permify/internal/commands"
	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/pkg/dsl/ast"
	"github.com/Permify/permify/pkg/dsl/compiler"
	"github.com/Permify/permify/pkg/dsl/parser"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// SchemaService -
type SchemaService struct {
	sw repositories.SchemaWriter
	sr repositories.SchemaReader

	// commands
	schemaLookup commands.ISchemaLookupCommand
}

// NewSchemaService -
func NewSchemaService(sc commands.ISchemaLookupCommand, sw repositories.SchemaWriter, sr repositories.SchemaReader) *SchemaService {
	return &SchemaService{
		sw:           sw,
		sr:           sr,
		schemaLookup: sc,
	}
}

// ReadSchema -
func (service *SchemaService) ReadSchema(ctx context.Context, version string) (response *base.IndexedSchema, err error) {
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

	var cnf []repositories.SchemaDefinition
	for _, st := range sch.Statements {
		cnf = append(cnf, repositories.SchemaDefinition{
			EntityType:           st.(*ast.EntityStatement).Name.Literal,
			SerializedDefinition: []byte(st.String()),
		})
	}

	return service.sw.WriteSchema(ctx, cnf)
}

// LookupSchema -
func (service *SchemaService) LookupSchema(ctx context.Context, entityType string, relations []string, version string) (response commands.SchemaLookupResponse, err error) {
	var en *base.EntityDefinition
	en, _, err = service.sr.ReadSchemaDefinition(ctx, entityType, version)
	if err != nil {
		return
	}

	q := &commands.SchemaLookupQuery{
		Relations: relations,
	}

	return service.schemaLookup.Execute(ctx, q, en.Actions)
}
