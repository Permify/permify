package servers

import (
	"log/slog"

	"github.com/rs/xid"
	"google.golang.org/grpc/status"

	otelCodes "go.opentelemetry.io/otel/codes"
	"golang.org/x/net/context"

	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/pkg/database"
	"github.com/Permify/permify/pkg/dsl/compiler"
	"github.com/Permify/permify/pkg/dsl/parser"
	v1 "github.com/Permify/permify/pkg/pb/base/v1"
)

// SchemaServer - Structure for Schema Server
type SchemaServer struct {
	v1.UnimplementedSchemaServer

	sw storage.SchemaWriter
	sr storage.SchemaReader
}

// NewSchemaServer - Creates new Schema Server
func NewSchemaServer(sw storage.SchemaWriter, sr storage.SchemaReader) *SchemaServer {
	return &SchemaServer{
		sw: sw,
		sr: sr,
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
		return nil, status.Error(GetStatus(err), err.Error())
	}

	_, _, err = compiler.NewCompiler(true, sch).Compile()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, status.Error(GetStatus(err), err.Error())
	}

	version := xid.New().String()

	cnf := make([]storage.SchemaDefinition, 0, len(sch.Statements))
	for _, st := range sch.Statements {
		cnf = append(cnf, storage.SchemaDefinition{
			TenantID:             request.GetTenantId(),
			Version:              version,
			Name:                 st.GetName(),
			SerializedDefinition: []byte(st.String()),
		})
	}

	err = r.sw.WriteSchema(ctx, cnf)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		slog.Error(err.Error())
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
			return nil, status.Error(GetStatus(err), err.Error())
		}
		version = ver
	}

	response, err := r.sr.ReadSchema(ctx, request.GetTenantId(), version)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		slog.Error(err.Error())
		return nil, status.Error(GetStatus(err), err.Error())
	}

	return &v1.SchemaReadResponse{
		Schema: response,
	}, nil
}

// List - List Schemas
func (r *SchemaServer) List(ctx context.Context, request *v1.SchemaListRequest) (*v1.SchemaListResponse, error) {
	ctx, span := tracer.Start(ctx, "schemas.list")
	defer span.End()

	schemas, ct, err := r.sr.ListSchemas(ctx, request.GetTenantId(), database.NewPagination(database.Size(request.GetPageSize()), database.Token(request.GetContinuousToken())))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		slog.Error(err.Error())
		return nil, status.Error(GetStatus(err), err.Error())
	}

	head, err := r.sr.HeadVersion(ctx, request.GetTenantId())
	if err != nil {
		return nil, status.Error(GetStatus(err), err.Error())
	}

	return &v1.SchemaListResponse{
		Head:            head,
		Schemas:         schemas,
		ContinuousToken: ct.String(),
	}, nil
}
