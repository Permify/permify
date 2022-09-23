package services

import (
	"context"

	"github.com/Permify/permify/internal/managers"
	"github.com/Permify/permify/internal/repositories"
	e "github.com/Permify/permify/internal/repositories/entities"
	"github.com/Permify/permify/internal/repositories/filters"
	"github.com/Permify/permify/pkg/dsl/schema"
	"github.com/Permify/permify/pkg/errors"
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
func (service *RelationshipService) ReadRelationships(ctx context.Context, filter filters.RelationTupleFilter) (tuples []tuple.Tuple, err errors.Error) {
	var t []e.RelationTuple
	t, err = service.rt.Read(ctx, filter)
	for _, tup := range t {
		tuples = append(tuples, tup.ToTuple())
	}
	return tuples, err
}

// WriteRelationship -
func (service *RelationshipService) WriteRelationship(ctx context.Context, tup tuple.Tuple, version string) (err errors.Error) {
	var entity schema.Entity
	entity, err = service.mn.Read(ctx, tup.Entity.Type, version)
	if err != nil {
		return err
	}

	var vt []string
	for _, rel := range entity.Relations {
		if rel.Name == tup.Relation.String() {
			vt = rel.Types
			break
		}
	}

	err = tup.Subject.ValidateSubjectType(vt)
	if err != nil {
		return errors.ValidationError.SetParams(map[string]interface{}{
			"relation": err.Error(),
		})
	}

	return service.rt.Write(ctx, e.RelationTuples{e.RelationTuple{
		Entity:          tup.Entity.Type,
		ObjectID:        tup.Entity.ID,
		Relation:        tup.Relation.String(),
		UsersetEntity:   tup.Subject.Type,
		UsersetObjectID: tup.Subject.ID,
		UsersetRelation: tup.Subject.Relation.String(),
	}})
}

// DeleteRelationship -
func (service *RelationshipService) DeleteRelationship(ctx context.Context, tup tuple.Tuple) errors.Error {
	return service.rt.Delete(ctx, e.RelationTuples{e.RelationTuple{
		Entity:          tup.Entity.Type,
		ObjectID:        tup.Entity.ID,
		Relation:        tup.Relation.String(),
		UsersetEntity:   tup.Subject.Type,
		UsersetObjectID: tup.Subject.ID,
		UsersetRelation: tup.Subject.Relation.String(),
	}})
}
