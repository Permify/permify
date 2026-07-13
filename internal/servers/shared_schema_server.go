package servers

import (
	"context"
	"log/slog"

	"github.com/rs/xid"
	api "go.opentelemetry.io/otel/metric"
	"google.golang.org/grpc/status"

	otelCodes "go.opentelemetry.io/otel/codes"

	"github.com/Permify/permify/internal"
	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/pkg/database"
	"github.com/Permify/permify/pkg/dsl/compiler"
	"github.com/Permify/permify/pkg/dsl/parser"
	v1 "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/telemetry"
)

// SharedSchemaServer - Structure for Shared Schema Server
type SharedSchemaServer struct {
	v1.UnimplementedSharedSchemaServer

	ssw                         storage.SharedSchemaWriter
	ssr                         storage.SharedSchemaReader
	writeSharedSchemaHistogram  api.Int64Histogram
	readSharedSchemaHistogram   api.Int64Histogram
	listSharedSchemaHistogram   api.Int64Histogram
	assignSharedSchemaHistogram api.Int64Histogram
}

// NewSharedSchemaServer - Creates new Shared Schema Server
func NewSharedSchemaServer(ssw storage.SharedSchemaWriter, ssr storage.SharedSchemaReader) *SharedSchemaServer {
	return &SharedSchemaServer{
		ssw:                         ssw,
		ssr:                         ssr,
		writeSharedSchemaHistogram:  telemetry.NewHistogram(internal.Meter, "write_shared_schema", "amount", "Number of writing shared schema"),
		readSharedSchemaHistogram:   telemetry.NewHistogram(internal.Meter, "read_shared_schema", "amount", "Number of reading shared schema"),
		listSharedSchemaHistogram:   telemetry.NewHistogram(internal.Meter, "list_shared_schema", "amount", "Number of listing shared schema"),
		assignSharedSchemaHistogram: telemetry.NewHistogram(internal.Meter, "assign_shared_schema", "amount", "Number of assigning shared schema"),
	}
}

// Write - Create or update a shared schema
func (r *SharedSchemaServer) Write(ctx context.Context, request *v1.SharedSchemaWriteRequest) (*v1.SharedSchemaWriteResponse, error) {
	ctx, span := internal.Tracer.Start(ctx, "shared-schemas.write")
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

	cnf := make([]storage.SharedSchemaDefinition, 0, len(sch.Statements))
	for _, st := range sch.Statements {
		cnf = append(cnf, storage.SharedSchemaDefinition{
			SharedSchemaID:       request.GetSharedSchemaId(),
			Version:              version,
			Name:                 st.GetName(),
			SerializedDefinition: []byte(st.String()),
		})
	}

	err = r.ssw.WriteSharedSchema(ctx, cnf)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		slog.ErrorContext(ctx, err.Error())
		return nil, status.Error(GetStatus(err), err.Error())
	}

	r.writeSharedSchemaHistogram.Record(ctx, 1)

	return &v1.SharedSchemaWriteResponse{
		SchemaVersion: version,
	}, nil
}

// Read - Read a shared schema
func (r *SharedSchemaServer) Read(ctx context.Context, request *v1.SharedSchemaReadRequest) (*v1.SharedSchemaReadResponse, error) {
	ctx, span := internal.Tracer.Start(ctx, "shared-schemas.read")
	defer span.End()

	version := request.GetMetadata().GetSchemaVersion()
	if version == "" {
		ver, err := r.ssr.SharedHeadVersion(ctx, request.GetSharedSchemaId())
		if err != nil {
			return nil, status.Error(GetStatus(err), err.Error())
		}
		version = ver
	}

	response, err := r.ssr.ReadSharedSchema(ctx, request.GetSharedSchemaId(), version)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		slog.ErrorContext(ctx, err.Error())
		return nil, status.Error(GetStatus(err), err.Error())
	}

	r.readSharedSchemaHistogram.Record(ctx, 1)

	return &v1.SharedSchemaReadResponse{
		Schema: response,
	}, nil
}

// Assign - Assign a shared schema to tenants
func (r *SharedSchemaServer) Assign(ctx context.Context, request *v1.SharedSchemaAssignRequest) (*v1.SharedSchemaAssignResponse, error) {
	ctx, span := internal.Tracer.Start(ctx, "shared-schemas.assign")
	defer span.End()

	// Validate shared schema exists
	_, err := r.ssr.SharedHeadVersion(ctx, request.GetSharedSchemaId())
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, status.Error(GetStatus(err), err.Error())
	}

	err = r.ssw.AssignSharedSchema(ctx, request.GetSharedSchemaId(), request.GetTenantIds())
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		slog.ErrorContext(ctx, err.Error())
		return nil, status.Error(GetStatus(err), err.Error())
	}

	r.assignSharedSchemaHistogram.Record(ctx, 1)

	return &v1.SharedSchemaAssignResponse{
		TenantIds: request.GetTenantIds(),
	}, nil
}

// List - List all shared schemas
func (r *SharedSchemaServer) List(ctx context.Context, request *v1.SharedSchemaListRequest) (*v1.SharedSchemaListResponse, error) {
	ctx, span := internal.Tracer.Start(ctx, "shared-schemas.list")
	defer span.End()

	schemas, ct, err := r.ssr.ListSharedSchemas(ctx, database.NewPagination(database.Size(request.GetPageSize()), database.Token(request.GetContinuousToken())))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		slog.ErrorContext(ctx, err.Error())
		return nil, status.Error(GetStatus(err), err.Error())
	}

	r.listSharedSchemaHistogram.Record(ctx, 1)

	return &v1.SharedSchemaListResponse{
		Schemas:         schemas,
		ContinuousToken: ct.String(),
	}, nil
}
