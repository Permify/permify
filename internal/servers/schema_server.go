package servers

import (
	"log/slog"
	"strings"

	"github.com/rs/xid"
	api "go.opentelemetry.io/otel/metric"
	"google.golang.org/grpc/status"

	otelCodes "go.opentelemetry.io/otel/codes"
	"golang.org/x/net/context"

	"github.com/Permify/permify/internal"
	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/pkg/database"
	"github.com/Permify/permify/pkg/dsl/compiler"
	"github.com/Permify/permify/pkg/dsl/parser"
	v1 "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/telemetry"
)

// SchemaServer - Structure for Schema Server
type SchemaServer struct {
	v1.UnimplementedSchemaServer

	sw                   storage.SchemaWriter
	sr                   storage.SchemaReader
	writeSchemaHistogram api.Int64Histogram
	readSchemaHistogram  api.Int64Histogram
	listSchemaHistogram  api.Int64Histogram
}

// NewSchemaServer - Creates new Schema Server
func NewSchemaServer(sw storage.SchemaWriter, sr storage.SchemaReader) *SchemaServer {
	return &SchemaServer{
		sw:                   sw,
		sr:                   sr,
		writeSchemaHistogram: telemetry.NewHistogram(internal.Meter, "write_schema", "amount", "Number of writing schema in"),
		readSchemaHistogram:  telemetry.NewHistogram(internal.Meter, "read_schema", "amount", "Number of reading schema"),
		listSchemaHistogram:  telemetry.NewHistogram(internal.Meter, "list_schema", "amount", "Number of listing schema"),
	}
}

// Write - Configure new Permify Schema to Permify
func (r *SchemaServer) Write(ctx context.Context, request *v1.SchemaWriteRequest) (*v1.SchemaWriteResponse, error) {
	ctx, span := internal.Tracer.Start(ctx, "schemas.write")
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
		slog.ErrorContext(ctx, err.Error())
		return nil, status.Error(GetStatus(err), err.Error())
	}

	r.writeSchemaHistogram.Record(ctx, 1)

	return &v1.SchemaWriteResponse{
		SchemaVersion: version,
	}, nil
}

// PartialWrite applies incremental updates to the schema of a specific tenant based on the provided request.
func (r *SchemaServer) PartialWrite(ctx context.Context, request *v1.SchemaPartialWriteRequest) (*v1.SchemaPartialWriteResponse, error) {
	// Start a new tracing span for monitoring and observability.
	ctx, span := internal.Tracer.Start(ctx, "schemas.partial-write")
	defer span.End() // Ensure the span is closed at the end of the function.

	// Retrieve or default the schema version from the request.
	version := request.GetMetadata().GetSchemaVersion()
	if version == "" { // If not provided, fetch the latest version.
		ver, err := r.sr.HeadVersion(ctx, request.GetTenantId())
		if err != nil {
			return nil, status.Error(GetStatus(err), err.Error()) // Return gRPC status error on failure.
		}
		version = ver
	}

	// Fetch the current schema definition as a string.
	definitions, err := r.sr.ReadSchemaString(ctx, request.GetTenantId(), version)
	if err != nil {
		span.RecordError(err) // Log and record the error.
		return nil, status.Error(GetStatus(err), err.Error())
	}

	// Parse the schema definitions into a structured format.
	p := parser.NewParser(strings.Join(definitions, "\n"))
	schema, err := p.Parse()
	if err != nil {
		span.RecordError(err) // Log and record the error.
		return nil, status.Error(GetStatus(err), err.Error())
	}

	// Iterate through each partial update in the request and apply changes.
	for entityName, partials := range request.GetPartials() {
		for _, write := range partials.GetWrite() { // Handle new schema statements.
			pr := parser.NewParser(write)
			stmt, err := pr.ParsePartial(entityName)
			if err != nil {
				span.RecordError(err)
				return nil, status.Error(GetStatus(err), err.Error())
			}
			err = schema.AddStatement(entityName, stmt)
			if err != nil {
				span.RecordError(err)
				return nil, status.Error(GetStatus(err), err.Error())
			}
		}

		for _, update := range partials.GetUpdate() { // Handle schema updates.
			pr := parser.NewParser(update)
			stmt, err := pr.ParsePartial(entityName)
			if err != nil {
				span.RecordError(err)
				return nil, status.Error(GetStatus(err), err.Error())
			}
			err = schema.UpdateStatement(entityName, stmt)
			if err != nil {
				span.RecordError(err)
				return nil, status.Error(GetStatus(err), err.Error())
			}
		}

		for _, del := range partials.GetDelete() { // Handle schema deletions.
			err = schema.DeleteStatement(entityName, del)
			if err != nil {
				span.RecordError(err)
				return nil, status.Error(GetStatus(err), err.Error())
			}
		}
	}

	// Re-parse the updated schema to ensure consistency.
	sch, err := parser.NewParser(schema.String()).Parse()
	if err != nil {
		span.RecordError(err)
		return nil, status.Error(GetStatus(err), err.Error())
	}

	// Compile the new schema to validate its correctness.
	_, _, err = compiler.NewCompiler(true, sch).Compile()
	if err != nil {
		span.RecordError(err)
		return nil, status.Error(GetStatus(err), err.Error())
	}

	// Generate a new version ID for the updated schema.
	newVersion := xid.New().String()

	// Prepare the new schema definition for storage.
	cnf := make([]storage.SchemaDefinition, 0, len(sch.Statements))
	for _, st := range sch.Statements {
		cnf = append(cnf, storage.SchemaDefinition{
			TenantID:             request.GetTenantId(),
			Version:              newVersion,
			Name:                 st.GetName(),
			SerializedDefinition: []byte(st.String()),
		})
	}

	// Write the updated schema to storage.
	err = r.sw.WriteSchema(ctx, cnf)
	if err != nil {
		span.RecordError(err)
		return nil, status.Error(GetStatus(err), err.Error())
	}

	// Return the response with the new schema version.
	return &v1.SchemaPartialWriteResponse{
		SchemaVersion: newVersion,
	}, nil
}

// Read - Read created Schema
func (r *SchemaServer) Read(ctx context.Context, request *v1.SchemaReadRequest) (*v1.SchemaReadResponse, error) {
	ctx, span := internal.Tracer.Start(ctx, "schemas.read")
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
		slog.ErrorContext(ctx, err.Error())
		return nil, status.Error(GetStatus(err), err.Error())
	}

	r.readSchemaHistogram.Record(ctx, 1)

	return &v1.SchemaReadResponse{
		Schema: response,
	}, nil
}

// List - List Schemas
func (r *SchemaServer) List(ctx context.Context, request *v1.SchemaListRequest) (*v1.SchemaListResponse, error) {
	ctx, span := internal.Tracer.Start(ctx, "schemas.list")
	defer span.End()

	schemas, ct, err := r.sr.ListSchemas(ctx, request.GetTenantId(), database.NewPagination(database.Size(request.GetPageSize()), database.Token(request.GetContinuousToken())))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		slog.ErrorContext(ctx, err.Error())
		return nil, status.Error(GetStatus(err), err.Error())
	}

	head, err := r.sr.HeadVersion(ctx, request.GetTenantId())
	if err != nil {
		return nil, status.Error(GetStatus(err), err.Error())
	}

	r.listSchemaHistogram.Record(ctx, 1)

	return &v1.SchemaListResponse{
		Head:            head,
		Schemas:         schemas,
		ContinuousToken: ct.String(),
	}, nil
}
