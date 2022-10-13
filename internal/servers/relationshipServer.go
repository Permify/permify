package servers

import (
	"fmt"

	otelCodes "go.opentelemetry.io/otel/codes"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Permify/permify/internal/services"
	"github.com/Permify/permify/pkg/database"
	"github.com/Permify/permify/pkg/errors"
	"github.com/Permify/permify/pkg/logger"
	v1 "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
)

// RelationshipServer -
type RelationshipServer struct {
	v1.UnimplementedRelationshipAPIServer

	relationshipService services.IRelationshipService
	l                   logger.Interface
}

// NewRelationshipServer -
func NewRelationshipServer(r services.IRelationshipService, l logger.Interface) *RelationshipServer {
	return &RelationshipServer{
		relationshipService: r,
		l:                   l,
	}
}

// Read -
func (r *RelationshipServer) Read(ctx context.Context, request *v1.RelationshipReadRequest) (*v1.RelationshipReadResponse, error) {
	ctx, span := tracer.Start(ctx, "relationships.read")
	defer span.End()

	v := request.Validate()
	if v != nil {
		return nil, v
	}

	var err errors.Error

	var collection tuple.ITupleCollection
	collection, err = r.relationshipService.ReadRelationships(ctx, request.GetFilter())
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

	return &v1.RelationshipReadResponse{
		Tuples: collection.GetTuples(),
	}, nil
}

// Write -
func (r *RelationshipServer) Write(ctx context.Context, request *v1.RelationshipWriteRequest) (*v1.RelationshipWriteResponse, error) {
	ctx, span := tracer.Start(ctx, "relationships.write")
	defer span.End()

	v := request.Validate()
	if v != nil {
		return nil, v
	}

	var err errors.Error

	t := &v1.Tuple{Entity: request.GetEntity(), Relation: request.GetRelation(), Subject: request.GetSubject()}
	err = r.relationshipService.WriteRelationship(ctx, t, request.SchemaVersion)
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

	return &v1.RelationshipWriteResponse{
		Tuple: t,
	}, nil
}

// Delete -
func (r *RelationshipServer) Delete(ctx context.Context, request *v1.RelationshipDeleteRequest) (*v1.RelationshipDeleteResponse, error) {
	ctx, span := tracer.Start(ctx, "relationships.delete")
	defer span.End()

	v := request.Validate()
	if v != nil {
		return nil, v
	}

	var err errors.Error

	t := &v1.Tuple{Entity: request.Entity, Relation: request.Relation, Subject: request.Subject}
	err = r.relationshipService.DeleteRelationship(ctx, t)
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

	return &v1.RelationshipDeleteResponse{
		Tuple: t,
	}, nil
}
