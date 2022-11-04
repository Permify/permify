package servers

import (
	"fmt"

	"google.golang.org/grpc/status"

	otelCodes "go.opentelemetry.io/otel/codes"
	"golang.org/x/net/context"

	"github.com/Permify/permify/internal/services"
	"github.com/Permify/permify/pkg/logger"
	v1 "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/token"
	"github.com/Permify/permify/pkg/tuple"
)

// RelationshipServer -
type RelationshipServer struct {
	v1.UnimplementedRelationshipServer

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

	collection, err := r.relationshipService.ReadRelationships(ctx, request.GetFilter(), token.StringToSnapToken(request.SnapToken))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		r.l.Error(fmt.Sprintf(err.Error()))
		return nil, status.Error(GetStatus(err), err.Error())
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

	for _, tup := range request.GetTuples() {
		v = tuple.ValidateSubject(tup.GetSubject())
		if v != nil {
			return nil, v
		}
	}

	snap, err := r.relationshipService.WriteRelationships(ctx, request.GetTuples(), request.SchemaVersion)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		r.l.Error(fmt.Sprintf(err.Error()))
		return nil, status.Error(GetStatus(err), err.Error())
	}

	return &v1.RelationshipWriteResponse{
		SnapToken: snap.String(),
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

	snap, err := r.relationshipService.DeleteRelationships(ctx, request.GetFilter())
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		r.l.Error(fmt.Sprintf(err.Error()))
		return nil, status.Error(GetStatus(err), err.Error())
	}

	return &v1.RelationshipDeleteResponse{
		SnapToken: snap.String(),
	}, nil
}
