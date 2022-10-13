package servers

import (
	"fmt"

	otelCodes "go.opentelemetry.io/otel/codes"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Permify/permify/internal/managers"
	"github.com/Permify/permify/internal/services"
	"github.com/Permify/permify/pkg/database"
	"github.com/Permify/permify/pkg/errors"
	"github.com/Permify/permify/pkg/logger"
	v1 "github.com/Permify/permify/pkg/pb/base/v1"
)

// SchemaServer -
type SchemaServer struct {
	v1.UnimplementedSchemaAPIServer

	schemaManager managers.IEntityConfigManager
	schemaService services.ISchemaService
	l             logger.Interface
}

// NewSchemaServer -
func NewSchemaServer(m managers.IEntityConfigManager, s services.ISchemaService, l logger.Interface) *SchemaServer {
	return &SchemaServer{
		schemaManager: m,
		schemaService: s,
		l:             l,
	}
}

// Write -
func (r *SchemaServer) Write(ctx context.Context, request *v1.SchemaWriteRequest) (*v1.SchemaWriteResponse, error) {
	ctx, span := tracer.Start(ctx, "schemas.write")
	defer span.End()

	var err errors.Error
	var version string
	version, err = r.schemaManager.Write(ctx, request.Schema)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		r.l.Error(fmt.Sprintf(err.Error()))
		switch err.Kind() {
		case errors.Database:
			return nil, status.Error(database.GetKindToGRPCStatus(err.SubKind()), err.Error())
		case errors.Validation:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case errors.Service:
			return nil, status.Error(codes.Internal, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &v1.SchemaWriteResponse{
		SchemaVersion: version,
	}, nil
}

// Write -
func (r *SchemaServer) Read(ctx context.Context, request *v1.SchemaReadRequest) (*v1.SchemaReadResponse, error) {
	ctx, span := tracer.Start(ctx, "schemas.write")
	defer span.End()

	var err errors.Error
	var response *v1.Schema
	response, err = r.schemaManager.All(ctx, request.SchemaVersion)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		r.l.Error(fmt.Sprintf(err.Error()))
		switch err.Kind() {
		case errors.Database:
			return nil, status.Error(database.GetKindToGRPCStatus(err.SubKind()), err.Error())
		case errors.Validation:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case errors.Service:
			return nil, status.Error(codes.Internal, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &v1.SchemaReadResponse{
		Schema: response,
	}, nil
}
