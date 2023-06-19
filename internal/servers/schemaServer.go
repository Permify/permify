package servers

import (
	"github.com/rs/xid"
	"google.golang.org/grpc/status"

	otelCodes "go.opentelemetry.io/otel/codes"
	"golang.org/x/net/context"

	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/pkg/dsl/ast"
	"github.com/Permify/permify/pkg/dsl/compiler"
	"github.com/Permify/permify/pkg/dsl/parser"
	"github.com/Permify/permify/pkg/logger"
	v1 "github.com/Permify/permify/pkg/pb/base/v1"
)

// SchemaServer - Structure for Schema Server
type SchemaServer struct {
	v1.UnimplementedSchemaServer

	sw     storage.SchemaWriter
	sr     storage.SchemaReader
	logger logger.Interface
}

// NewSchemaServer - Creates new Schema Server
func NewSchemaServer(sw storage.SchemaWriter, sr storage.SchemaReader, l logger.Interface) *SchemaServer {
	return &SchemaServer{
		sw:     sw,
		sr:     sr,
		logger: l,
	}
}

// Write - Configure new Permify Schema to Permify
func (r *SchemaServer) Write(ctx context.Context, request *v1.SchemaWriteRequest) (*v1.SchemaWriteResponse, error) {
	ctx, span := tracer.Start(ctx, "schemas.write")
	defer span.End()

	sch, err := parser.NewParser(request.GetSchema()).Parse()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, err
	}

	_, err = compiler.NewCompiler(true, sch).Compile()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, err
	}

	version := xid.New().String()

	cnf := make([]storage.SchemaDefinition, 0, len(sch.Statements))
	for _, st := range sch.Statements {
		cnf = append(cnf, storage.SchemaDefinition{
			TenantID:             request.GetTenantId(),
			Version:              version,
			EntityType:           st.(*ast.EntityStatement).Name.Literal,
			SerializedDefinition: []byte(st.String()),
		})
	}

	err = r.sw.WriteSchema(ctx, cnf)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		r.logger.Error(err.Error())
		return nil, status.Error(GetStatus(err), err.Error())
	}

	return &v1.SchemaWriteResponse{
		SchemaVersion: version,
	}, nil
}

// Read - Read created Schema
func (r *SchemaServer) Read(ctx context.Context, request *v1.SchemaReadRequest) (*v1.SchemaReadResponse, error) {
	ctx, span := tracer.Start(ctx, "schemas.read")
	defer span.End()

	version := request.GetMetadata().GetSchemaVersion()
	if version == "" {
		ver, err := r.sr.HeadVersion(ctx, request.GetTenantId())
		if err != nil {
			return nil, err
		}
		version = ver
	}

	response, err := r.sr.ReadSchema(ctx, request.GetTenantId(), version)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		r.logger.Error(err.Error())
		return nil, status.Error(GetStatus(err), err.Error())
	}

	return &v1.SchemaReadResponse{
		Schema: response,
	}, nil
}
