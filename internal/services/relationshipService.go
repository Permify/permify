package services

import (
	"context"

	e "github.com/Permify/permify/internal/entities"
	"github.com/Permify/permify/internal/managers"
	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/internal/repositories/filters"
	"github.com/Permify/permify/pkg/dsl/schema"
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
func (service *RelationshipService) ReadRelationships(ctx context.Context, filter filters.RelationTupleFilter) ([]e.RelationTuple, error) {
	return service.rt.Read(ctx, filter)
}

// WriteRelationship -
func (service *RelationshipService) WriteRelationship(ctx context.Context, en e.RelationTuple, version string) (err error) {
	ct := en.ToTuple()
	var entity schema.Entity
	entity, err = service.mn.Read(ctx, ct.Entity.Type, version)
	if err != nil {
		return err
	}

	var vt []string
	for _, rel := range entity.Relations {
		if rel.Name == en.Relation {
			vt = rel.Types
			break
		}
	}

	err = ct.Subject.ValidateSubjectType(vt)
	if err != nil {
		return err
	}

	return service.rt.Write(ctx, e.RelationTuples{en})
}

// DeleteRelationship -
func (service *RelationshipService) DeleteRelationship(ctx context.Context, tuple e.RelationTuple) error {
	return service.rt.Delete(ctx, e.RelationTuples{tuple})
}
