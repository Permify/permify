package services

import (
	"context"
	"errors"

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
func (service *RelationshipService) ReadRelationships(ctx context.Context, filter *base.TupleFilter, token token.SnapToken) (tuples database.ITupleCollection, err error) {
	return service.rr.QueryRelationships(ctx, filter, token)
}

// WriteRelationships -
func (service *RelationshipService) WriteRelationships(ctx context.Context, tuples []*base.Tuple, version string) (token token.SnapToken, err error) {
	for _, tup := range tuples {
		var entity *base.EntityDefinition
		entity, err = service.sr.ReadSchemaDefinition(ctx, tup.GetEntity().GetType(), version)
		if err != nil {
			return token, err
		}

		if tuple.IsEntityAndSubjectEquals(tup) {
			return token, errors.New(base.ErrorCode_ERROR_CODE_ENTITY_AND_SUBJECT_CANNOT_BE_EQUAL.String())
		}

		var rel *base.RelationDefinition
		var vt []string
		rel, err = schema.GetRelationByNameInEntityDefinition(entity, tup.GetRelation())
		for _, t := range rel.GetRelationReferences() {
			vt = append(vt, t.GetName())
		}

		err = tuple.ValidateSubjectType(tup.Subject, vt)
		if err != nil {
			return token, err
		}
	}

	return service.rw.WriteRelationships(ctx, database.NewTupleCollection(tuples...))
}

// DeleteRelationships -
func (service *RelationshipService) DeleteRelationships(ctx context.Context, filter *base.TupleFilter) (token.SnapToken, error) {
	return service.rw.DeleteRelationships(ctx, filter)
}
