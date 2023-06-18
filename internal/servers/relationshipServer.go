package servers

import (
	"google.golang.org/grpc/status"

	otelCodes "go.opentelemetry.io/otel/codes"
	"golang.org/x/net/context"

	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/internal/validation"
	"github.com/Permify/permify/pkg/database"
	"github.com/Permify/permify/pkg/logger"
	v1 "github.com/Permify/permify/pkg/pb/base/v1"
)

// RelationshipServer - Structure for Relationship Server
type RelationshipServer struct {
	v1.UnimplementedRelationshipServer

	sr     storage.SchemaReader
	rr     storage.RelationshipReader
	rw     storage.RelationshipWriter
	logger logger.Interface
}

// NewRelationshipServer - Creates new Relationship Server
func NewRelationshipServer(
	rr storage.RelationshipReader,
	rw storage.RelationshipWriter,
	sr storage.SchemaReader,
	l logger.Interface,
) *RelationshipServer {
	return &RelationshipServer{
		rr:     rr,
		rw:     rw,
		sr:     sr,
		logger: l,
	}
}

// Read - Allows directly querying the stored engines data to display and filter stored relational tuples
func (r *RelationshipServer) Read(ctx context.Context, request *v1.RelationshipReadRequest) (*v1.RelationshipReadResponse, error) {
	ctx, span := tracer.Start(ctx, "relationships.read")
	defer span.End()

	v := request.Validate()
	if v != nil {
		return nil, v
	}

	snap := request.GetMetadata().GetSnapToken()
	if snap == "" {
		st, err := r.rr.HeadSnapshot(ctx, request.GetTenantId())
		if err != nil {
			return nil, err
		}
		snap = st.Encode().String()
	}

	collection, ct, err := r.rr.ReadRelationships(
		ctx,
		request.GetTenantId(),
		request.GetFilter(),
		snap,
		database.NewPagination(
			database.Size(request.GetPageSize()),
			database.Token(request.GetContinuousToken()),
		),
	)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		r.logger.Error(err.Error())
		return nil, status.Error(GetStatus(err), err.Error())
	}

	return &v1.RelationshipReadResponse{
		Tuples:          collection.GetTuples(),
		ContinuousToken: ct.String(),
	}, nil
}

// Write - Write relation tuples to writeDB
func (r *RelationshipServer) Write(ctx context.Context, request *v1.RelationshipWriteRequest) (*v1.RelationshipWriteResponse, error) {
	ctx, span := tracer.Start(ctx, "relationships.write")
	defer span.End()

	v := request.Validate()
	if v != nil {
		return nil, v
	}

	version := request.GetMetadata().GetSchemaVersion()
	if version == "" {
		v, err := r.sr.HeadVersion(ctx, request.GetTenantId())
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())
			return nil, err
		}
		version = v
	}

	relationships := make([]*v1.Tuple, 0, len(request.GetTuples()))

	for _, tup := range request.GetTuples() {

		definition, _, err := r.sr.ReadSchemaDefinition(ctx, request.GetTenantId(), tup.GetEntity().GetType(), version)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())
			return nil, err
		}

		err = validation.ValidateTuple(definition, tup)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())
			return nil, err
		}

		relationships = append(relationships, tup)
	}

	snap, err := r.rw.WriteRelationships(ctx, request.GetTenantId(), database.NewTupleCollection(relationships...))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		r.logger.Error(err.Error())
		return nil, status.Error(GetStatus(err), err.Error())
	}

	return &v1.RelationshipWriteResponse{
		SnapToken: snap.String(),
	}, nil
}

// Delete - Delete relation tuples to writeDB
func (r *RelationshipServer) Delete(ctx context.Context, request *v1.RelationshipDeleteRequest) (*v1.RelationshipDeleteResponse, error) {
	ctx, span := tracer.Start(ctx, "relationships.delete")
	defer span.End()

	v := request.Validate()
	if v != nil {
		return nil, v
	}

	err := validation.ValidateFilter(request.GetFilter())
	if err != nil {
		return nil, v
	}

	snap, err := r.rw.DeleteRelationships(ctx, request.GetTenantId(), request.GetFilter())
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		r.logger.Error(err.Error())
		return nil, status.Error(GetStatus(err), err.Error())
	}

	return &v1.RelationshipDeleteResponse{
		SnapToken: snap.String(),
	}, nil
}
