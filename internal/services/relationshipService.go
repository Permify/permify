package services

import (
	"context"
	"errors"

	otelCodes "go.opentelemetry.io/otel/codes"

	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/pkg/database"
	"github.com/Permify/permify/pkg/dsl/schema"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/token"
	"github.com/Permify/permify/pkg/tuple"
)

// RelationshipService -
type RelationshipService struct {
	sr repositories.SchemaReader
	rr repositories.RelationshipReader
	rw repositories.RelationshipWriter
}

// NewRelationshipService -
func NewRelationshipService(rr repositories.RelationshipReader, rw repositories.RelationshipWriter, sr repositories.SchemaReader) *RelationshipService {
	return &RelationshipService{
		sr: sr,
		rr: rr,
		rw: rw,
	}
}

// ReadRelationships -
func (service *RelationshipService) ReadRelationships(ctx context.Context, tenantID string, filter *base.TupleFilter, token string) (tuples database.ITupleCollection, err error) {
	ctx, span := tracer.Start(ctx, "relationships.read")
	defer span.End()

	return service.rr.QueryRelationships(ctx, tenantID, filter, token)
}

// WriteRelationships -
func (service *RelationshipService) WriteRelationships(ctx context.Context, tenantID string, tuples []*base.Tuple, version string) (token token.EncodedSnapToken, err error) {
	ctx, span := tracer.Start(ctx, "relationships.write")
	defer span.End()

	if version == "" {
		var v string
		v, err = service.sr.HeadVersion(ctx, tenantID)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())
			return token, err
		}
		version = v
	}

	for _, tup := range tuples {
		var entity *base.EntityDefinition
		entity, _, err = service.sr.ReadSchemaDefinition(ctx, tenantID, tup.GetEntity().GetType(), version)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())
			return token, err
		}

		if tuple.IsEntityAndSubjectEquals(tup) {
			return token, errors.New(base.ErrorCode_ERROR_CODE_ENTITY_AND_SUBJECT_CANNOT_BE_EQUAL.String())
		}

		var rel *base.RelationDefinition
		var vt []string
		rel, err = schema.GetRelationByNameInEntityDefinition(entity, tup.GetRelation())
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())
			return token, err
		}
		for _, t := range rel.GetRelationReferences() {
			vt = append(vt, t.GetName())
		}

		err = tuple.ValidateSubjectType(tup.Subject, vt)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())
			return token, err
		}
	}

	return service.rw.WriteRelationships(ctx, tenantID, database.NewTupleCollection(tuples...))
}

// DeleteRelationships -
func (service *RelationshipService) DeleteRelationships(ctx context.Context, tenantID string, filter *base.TupleFilter) (token.EncodedSnapToken, error) {
	ctx, span := tracer.Start(ctx, "relationships.delete")
	defer span.End()

	return service.rw.DeleteRelationships(ctx, tenantID, filter)
}
