package servers

import (
	"google.golang.org/grpc/status"

	otelCodes "go.opentelemetry.io/otel/codes"
	"golang.org/x/net/context"

	"github.com/Permify/permify/internal/services"
	"github.com/Permify/permify/pkg/logger"
	v1 "github.com/Permify/permify/pkg/pb/base/v1"
)

// SchemaServer - Structure for Schema Server
type SchemaServer struct {
	v1.UnimplementedSchemaServer

	schemaService services.ISchemaService
	logger        logger.Interface
}

// NewSchemaServer - Creates new Schema Server
func NewSchemaServer(s services.ISchemaService, l logger.Interface) *SchemaServer {
	return &SchemaServer{
		schemaService: s,
		logger:        l,
	}
}

// Write - Configure new Permify Schema to Permify
func (r *SchemaServer) Write(ctx context.Context, request *v1.SchemaWriteRequest) (*v1.SchemaWriteResponse, error) {
	ctx, span := tracer.Start(ctx, "schemas.write")
	defer span.End()

	version, err := r.schemaService.WriteSchema(ctx, request.GetTenantId(), request.GetSchema())
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

	var err error
	var response *v1.SchemaDefinition
	response, err = r.schemaService.ReadSchema(ctx, request.GetTenantId(), request.GetMetadata().GetSchemaVersion())
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
