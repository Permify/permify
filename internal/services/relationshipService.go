package services

import (
	"context"

	e "github.com/Permify/permify/internal/entities"
	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/internal/repositories/filters"
)

// IRelationshipService -
type IRelationshipService interface {
	ReadRelationships(ctx context.Context, filter filters.RelationTupleFilter) ([]e.RelationTuple, error)
	WriteRelationship(ctx context.Context, entities []e.RelationTuple) error
	DeleteRelationship(ctx context.Context, entities []e.RelationTuple) error
}

// RelationshipService -
type RelationshipService struct {
	repository repositories.IRelationTupleRepository
}

// NewRelationshipService -
func NewRelationshipService(repo repositories.IRelationTupleRepository) *RelationshipService {
	return &RelationshipService{
		repository: repo,
	}
}

// ReadRelationships -
func (service *RelationshipService) ReadRelationships(ctx context.Context, filter filters.RelationTupleFilter) ([]e.RelationTuple, error) {
	return service.repository.Read(ctx, filter)
}

// WriteRelationship -
func (service *RelationshipService) WriteRelationship(ctx context.Context, entities []e.RelationTuple) error {
	return service.repository.Write(ctx, entities)
}

// DeleteRelationship -
func (service *RelationshipService) DeleteRelationship(ctx context.Context, entities []e.RelationTuple) error {
	return service.repository.Delete(ctx, entities)
}
