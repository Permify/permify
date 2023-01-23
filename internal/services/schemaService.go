package services

import (
	"context"

	"github.com/rs/xid"

	otelCodes "go.opentelemetry.io/otel/codes"

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
func (service *SchemaService) ReadSchema(ctx context.Context, tenantID uint64, version string) (response *base.IndexedSchema, err error) {
	ctx, span := tracer.Start(ctx, "schemas.read")
	defer span.End()

	if version == "" {
		var ver string
		ver, err = service.sr.HeadVersion(ctx, tenantID)
		if err != nil {
			return response, err
		}
		version = ver
	}

	return service.sr.ReadSchema(ctx, tenantID, version)
}

// WriteSchema -
func (service *SchemaService) WriteSchema(ctx context.Context, tenantID uint64, schema string) (response string, err error) {
	ctx, span := tracer.Start(ctx, "schemas.write")
	defer span.End()

	sch, err := parser.NewParser(schema).Parse()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		return "", err
	}

	_, err = compiler.NewCompiler(false, sch).Compile()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		return "", err
	}

	version := xid.New().String()

	cnf := make([]repositories.SchemaDefinition, 0, len(sch.Statements))
	for _, st := range sch.Statements {
		cnf = append(cnf, repositories.SchemaDefinition{
			TenantID:             tenantID,
			Version:              version,
			EntityType:           st.(*ast.EntityStatement).Name.Literal,
			SerializedDefinition: []byte(st.String()),
		})
	}

	err = service.sw.WriteSchema(ctx, cnf)
	if err != nil {
		return "", err
	}
	return version, nil
}
