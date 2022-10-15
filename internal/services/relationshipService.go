package services

import (
	"context"
	"errors"

	"github.com/Permify/permify/internal/managers"
	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/pkg/dsl/schema"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
)

// RelationshipService -
type RelationshipService struct {
	mn managers.IEntityConfigManager
	rt repositories.IRelationTupleRepository
}

// NewRelationshipService -
func NewRelationshipService(rt repositories.IRelationTupleRepository, mn managers.IEntityConfigManager) *RelationshipService {
	return &RelationshipService{
		mn: mn,
		rt: rt,
	}
}

// ReadRelationships -
func (service *RelationshipService) ReadRelationships(ctx context.Context, filter *base.TupleFilter) (tuples tuple.ITupleCollection, err error) {
	return service.rt.Read(ctx, filter)
}

// WriteRelationship -
func (service *RelationshipService) WriteRelationship(ctx context.Context, tup *base.Tuple, version string) (err error) {
	var entity *base.EntityDefinition
	entity, err = service.mn.Read(ctx, tup.GetEntity().GetType(), version)
	if err != nil {
		return err
	}

	if tuple.IsEntityAndSubjectEquals(tup) {
		return errors.New(base.ErrorCode_entity_and_subject_cannot_be_equal.String())
	}

	var rel *base.RelationDefinition
	var vt []string
	rel, err = schema.GetRelationByNameInEntityDefinition(entity, tup.GetRelation())
	for _, t := range rel.GetRelationReferences() {
		vt = append(vt, t.GetName())
	}

	err = tuple.ValidateSubjectType(tup.Subject, vt)
	if err != nil {
		return err
	}

	return service.rt.Write(ctx, tuple.NewTupleCollection(tup).CreateTupleIterator())
}

// DeleteRelationship -
func (service *RelationshipService) DeleteRelationship(ctx context.Context, tup *base.Tuple) error {
	return service.rt.Delete(ctx, tuple.NewTupleCollection(tup).CreateTupleIterator())
}
