package services

import (
	"context"

	"github.com/Permify/permify/internal/managers"
	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/pkg/dsl/schema"
	"github.com/Permify/permify/pkg/errors"
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
func (service *RelationshipService) ReadRelationships(ctx context.Context, filter *base.TupleFilter) (tuples tuple.ITupleCollection, err errors.Error) {
	return service.rt.Read(ctx, filter)
}

// WriteRelationship -
func (service *RelationshipService) WriteRelationship(ctx context.Context, tup *base.Tuple, version string) (err errors.Error) {
	var entity *base.EntityDefinition
	entity, err = service.mn.Read(ctx, tup.GetEntity().GetType(), version)
	if err != nil {
		return err
	}

	if tuple.IsEntityAndSubjectEquals(tup) {
		return errors.ValidationError.SetParams(map[string]interface{}{
			"subject": "entity and subject cannot be equal",
		})
	}

	var rel *base.RelationDefinition
	var vt []string
	rel, err = schema.GetRelation(entity, tup.GetRelation())
	for _, t := range rel.GetTypes() {
		vt = append(vt, t.GetName())
	}

	err = tuple.ValidateSubjectType(tup.Subject, vt)
	if err != nil {
		return errors.ValidationError.SetParams(map[string]interface{}{
			"relation": err.Error(),
		})
	}

	return service.rt.Write(ctx, tuple.NewTupleCollection(tup).CreateTupleIterator())
}

// DeleteRelationship -
func (service *RelationshipService) DeleteRelationship(ctx context.Context, tup *base.Tuple) errors.Error {
	return service.rt.Delete(ctx, tuple.NewTupleCollection(tup).CreateTupleIterator())
}
