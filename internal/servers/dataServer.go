package servers

import (
	"log/slog"

	"google.golang.org/grpc/status"

	otelCodes "go.opentelemetry.io/otel/codes"
	"golang.org/x/net/context"

	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/internal/validation"
	"github.com/Permify/permify/pkg/database"
	v1 "github.com/Permify/permify/pkg/pb/base/v1"
)

// DataServer - Structure for Data Server
type DataServer struct {
	v1.UnimplementedDataServer

	sr storage.SchemaReader
	dr storage.DataReader
	br storage.BundleReader
	dw storage.DataWriter
}

// NewDataServer - Creates new Data Server
func NewDataServer(
	dr storage.DataReader,
	dw storage.DataWriter,
	br storage.BundleReader,
	sr storage.SchemaReader,
) *DataServer {
	return &DataServer{
		dr: dr,
		dw: dw,
		br: br,
		sr: sr,
	}
}

// ReadRelationships - Allows directly querying the stored engines data to display and filter stored relational tuples
func (r *DataServer) ReadRelationships(ctx context.Context, request *v1.RelationshipReadRequest) (*v1.RelationshipReadResponse, error) {
	ctx, span := tracer.Start(ctx, "data.read.relationships")
	defer span.End()

	v := request.Validate()
	if v != nil {
		return nil, v
	}

	snap := request.GetMetadata().GetSnapToken()
	if snap == "" {
		st, err := r.dr.HeadSnapshot(ctx, request.GetTenantId())
		if err != nil {
			return nil, err
		}
		snap = st.Encode().String()
	}

	collection, ct, err := r.dr.ReadRelationships(
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
		slog.Error(err.Error())
		return nil, status.Error(GetStatus(err), err.Error())
	}

	return &v1.RelationshipReadResponse{
		Tuples:          collection.GetTuples(),
		ContinuousToken: ct.String(),
	}, nil
}

// ReadAttributes - Allows directly querying the stored engines data to display and filter stored attribute tuples
func (r *DataServer) ReadAttributes(ctx context.Context, request *v1.AttributeReadRequest) (*v1.AttributeReadResponse, error) {
	ctx, span := tracer.Start(ctx, "data.read.attributes")
	defer span.End()

	v := request.Validate()
	if v != nil {
		return nil, v
	}

	snap := request.GetMetadata().GetSnapToken()
	if snap == "" {
		st, err := r.dr.HeadSnapshot(ctx, request.GetTenantId())
		if err != nil {
			return nil, err
		}
		snap = st.Encode().String()
	}

	collection, ct, err := r.dr.ReadAttributes(
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
		slog.Error(err.Error())
		return nil, status.Error(GetStatus(err), err.Error())
	}

	return &v1.AttributeReadResponse{
		Attributes:      collection.GetAttributes(),
		ContinuousToken: ct.String(),
	}, nil
}

// Write - Write relationships and attributes to writeDB
func (r *DataServer) Write(ctx context.Context, request *v1.DataWriteRequest) (*v1.DataWriteResponse, error) {
	ctx, span := tracer.Start(ctx, "data.write")
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
		definition, _, err := r.sr.ReadEntityDefinition(ctx, request.GetTenantId(), tup.GetEntity().GetType(), version)
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

	attributes := make([]*v1.Attribute, 0, len(request.GetAttributes()))

	for _, attribute := range request.GetAttributes() {
		definition, _, err := r.sr.ReadEntityDefinition(ctx, request.GetTenantId(), attribute.GetEntity().GetType(), version)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())
			return nil, err
		}

		err = validation.ValidateAttribute(definition, attribute)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())
			return nil, err
		}

		attributes = append(attributes, attribute)
	}

	snap, err := r.dw.Write(ctx, request.GetTenantId(), database.NewTupleCollection(relationships...), database.NewAttributeCollection(attributes...))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		slog.Error(err.Error())
		return nil, status.Error(GetStatus(err), err.Error())
	}

	return &v1.DataWriteResponse{
		SnapToken: snap.String(),
	}, nil
}

// WriteRelationships - Write relation tuples to writeDB
func (r *DataServer) WriteRelationships(ctx context.Context, request *v1.RelationshipWriteRequest) (*v1.RelationshipWriteResponse, error) {
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
		definition, _, err := r.sr.ReadEntityDefinition(ctx, request.GetTenantId(), tup.GetEntity().GetType(), version)
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

	snap, err := r.dw.Write(ctx, request.GetTenantId(), database.NewTupleCollection(relationships...), database.NewAttributeCollection())
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		slog.Error(err.Error())
		return nil, status.Error(GetStatus(err), err.Error())
	}

	return &v1.RelationshipWriteResponse{
		SnapToken: snap.String(),
	}, nil
}

// Delete - Delete relationships and attributes from writeDB
func (r *DataServer) Delete(ctx context.Context, request *v1.DataDeleteRequest) (*v1.DataDeleteResponse, error) {
	ctx, span := tracer.Start(ctx, "data.delete")
	defer span.End()

	v := request.Validate()
	if v != nil {
		return nil, v
	}

	err := validation.ValidateFilters(request.GetTupleFilter(), request.GetAttributeFilter())
	if err != nil {
		return nil, v
	}

	snap, err := r.dw.Delete(ctx, request.GetTenantId(), request.GetTupleFilter(), request.GetAttributeFilter())
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		slog.Error(err.Error())
		return nil, status.Error(GetStatus(err), err.Error())
	}

	return &v1.DataDeleteResponse{
		SnapToken: snap.String(),
	}, nil
}

// DeleteRelationships - Delete relationships from writeDB
func (r *DataServer) DeleteRelationships(ctx context.Context, request *v1.RelationshipDeleteRequest) (*v1.RelationshipDeleteResponse, error) {
	ctx, span := tracer.Start(ctx, "relationships.delete")
	defer span.End()

	v := request.Validate()
	if v != nil {
		return nil, v
	}

	err := validation.ValidateTupleFilter(request.GetFilter())
	if err != nil {
		return nil, v
	}

	snap, err := r.dw.Delete(ctx, request.GetTenantId(), request.GetFilter(), &v1.AttributeFilter{})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		slog.Error(err.Error())
		return nil, status.Error(GetStatus(err), err.Error())
	}

	return &v1.RelationshipDeleteResponse{
		SnapToken: snap.String(),
	}, nil
}

// RunBundle executes a bundle and returns its snapshot token.
func (r *DataServer) RunBundle(ctx context.Context, request *v1.BundleRunRequest) (*v1.BundleRunResponse, error) {
	ctx, span := tracer.Start(ctx, "bundle.run")
	defer span.End()

	v := request.Validate()
	if v != nil {
		return nil, v
	}

	bundle, err := r.br.Read(ctx, request.GetTenantId(), request.GetName())
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		slog.Error(err.Error())
		return nil, status.Error(GetStatus(err), err.Error())
	}

	err = validation.ValidateBundleArguments(bundle.GetArguments(), request.GetArguments())
	if err != nil {
		return nil, err
	}

	snap, err := r.dw.RunBundle(ctx, request.GetTenantId(), request.GetArguments(), bundle)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		slog.Error(err.Error())
		return nil, status.Error(GetStatus(err), err.Error())
	}

	return &v1.BundleRunResponse{
		SnapToken: snap.String(),
	}, nil
}
