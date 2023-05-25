package engines

import (
	"context"
	"errors"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/Permify/permify/internal/schema"
	"github.com/Permify/permify/internal/storage"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
)

// EntityFilterEngine is responsible for executing linked entity operations
type EntityFilterEngine struct {
	// schemaReader is responsible for reading schema information
	schemaReader storage.SchemaReader
	// relationshipReader is responsible for reading relationship information
	relationshipReader storage.RelationshipReader
	// schemaMap is a map that keeps track of schema versions
	schemaMap sync.Map
}

// NewEntityFilterEngine creates a new EntityFilter engine
func NewEntityFilterEngine(schemaReader storage.SchemaReader, relationshipReader storage.RelationshipReader) *EntityFilterEngine {
	return &EntityFilterEngine{
		schemaReader:       schemaReader,
		relationshipReader: relationshipReader,
		schemaMap:          sync.Map{},
	}
}

// EntityFilter is a method of the EntityFilterEngine struct. It executes a permission request for linked entities.
func (engine *EntityFilterEngine) EntityFilter(
	ctx context.Context, // A context used for tracing and cancellation.
	request *base.PermissionEntityFilterRequest, // A permission request for linked entities.
	visits *ERMap, // A map that keeps track of visited entities to avoid infinite loops.
	publisher *BulkPublisher, // A custom publisher that publishes results in bulk.
) (err error) { // Returns an error if one occurs during execution.
	// Check if direct result
	if request.GetEntityReference().GetType() == request.GetSubject().GetType() && request.GetEntityReference().GetRelation() == request.GetSubject().GetRelation() {
		// TODO: Implement direct result and exclusion logic.

		found := &base.Entity{
			Type: request.GetSubject().GetType(),
			Id:   request.GetSubject().GetId(),
		}

		// If the entity reference is the same as the subject, publish the result directly and return.
		publisher.Publish(found, &base.PermissionCheckRequestMetadata{
			SnapToken:     request.GetMetadata().GetSnapToken(),
			SchemaVersion: request.GetMetadata().GetSchemaVersion(),
			Depth:         request.GetMetadata().GetDepth(),
		}, base.PermissionCheckResponse_RESULT_UNKNOWN)
	}

	// Retrieve entity definition
	var sc *base.SchemaDefinition
	sc, err = engine.readSchema(ctx, request.GetTenantId(), request.GetMetadata().GetSchemaVersion())
	if err != nil {
		return err
	}

	// Retrieve linked entrances
	cn := schema.NewLinkedGraph(sc) // Create a new linked graph from the schema definition.
	var entrances []*schema.LinkedEntrance
	entrances, err = cn.RelationshipLinkedEntrances(
		&base.RelationReference{
			Type:     request.GetEntityReference().GetType(),
			Relation: request.GetEntityReference().GetRelation(),
		},
		&base.RelationReference{
			Type:     request.GetSubject().GetType(),
			Relation: request.GetSubject().GetRelation(),
		},
	) // Retrieve the linked entrances between the entity reference and subject.

	// Create a new context for executing goroutines and a cancel function.
	cctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Create a new errgroup and a new context that inherits the original context.
	g, cont := errgroup.WithContext(cctx)

	// Loop over each linked entrance.
	for _, entrance := range entrances {
		// Switch on the kind of linked entrance.
		switch entrance.LinkedEntranceKind() {
		case schema.RelationLinkedEntrance: // If the linked entrance is a relation entrance.
			err = engine.relationEntrance(cont, request, entrance, visits, g, publisher) // Call the relation entrance method.
			if err != nil {
				return err
			}
		case schema.ComputedUserSetLinkedEntrance: // If the linked entrance is a computed user set entrance.
			err = engine.l(ctx, request, &base.EntityAndRelation{ // Call the run method with a new entity and relation.
				Entity: &base.Entity{
					Type: entrance.TargetEntrance.GetType(),
					Id:   request.GetSubject().GetId(),
				},
				Relation: entrance.TargetEntrance.GetRelation(),
			}, visits, g, publisher)
			if err != nil {
				return err
			}
		case schema.TupleToUserSetLinkedEntrance: // If the linked entrance is a tuple to user set entrance.
			err = engine.tupleToUserSetEntrance(cont, request, entrance, visits, g, publisher) // Call the tuple to user set entrance method.
			if err != nil {
				return err
			}
		default:
			return errors.New("unknown linked entrance type") // Return an error if the linked entrance is of an unknown type.
		}
	}

	return g.Wait() // Wait for all goroutines in the errgroup to complete and return any errors that occur.
}

// relationEntrance is a method of the EntityFilterEngine struct. It handles relation entrances.
func (engine *EntityFilterEngine) relationEntrance(
	ctx context.Context, // A context used for tracing and cancellation.
	request *base.PermissionEntityFilterRequest, // A permission request for linked entities.
	entrance *schema.LinkedEntrance, // A linked entrance.
	visits *ERMap, // A map that keeps track of visited entities to avoid infinite loops.
	g *errgroup.Group, // An errgroup used for executing goroutines.
	publisher *BulkPublisher, // A custom publisher that publishes results in bulk.
) error { // Returns an error if one occurs during execution.
	it, err := engine.relationshipReader.QueryRelationships(ctx, request.GetTenantId(), &base.TupleFilter{
		Entity: &base.EntityFilter{
			Type: entrance.TargetEntrance.GetType(),
			Ids:  []string{},
		},
		Relation: entrance.TargetEntrance.GetRelation(),
		Subject: &base.SubjectFilter{
			Type:     request.GetSubject().GetType(),
			Ids:      []string{request.GetSubject().GetId()},
			Relation: request.GetSubject().GetRelation(),
		},
	}, request.GetMetadata().GetSnapToken()) // Query the relationship reader for relationships that match the linked entrance and the request metadata.
	if err != nil {
		return err
	}

	for it.HasNext() { // Loop over each relationship.
		current := it.GetNext()
		g.Go(func() error {
			return engine.l(ctx, request, &base.EntityAndRelation{ // Call the run method with a new entity and relation.
				Entity: &base.Entity{
					Type: current.GetEntity().GetType(),
					Id:   current.GetEntity().GetId(),
				},
				Relation: current.GetRelation(),
			}, visits, g, publisher)
		})
	}
	return nil
}

// tupleToUserSetEntrance is a method of the EntityFilterEngine struct. It handles tuple to user set entrances.
func (engine *EntityFilterEngine) tupleToUserSetEntrance(
	// A context used for tracing and cancellation.
	ctx context.Context,
	// A permission request for linked entities.
	request *base.PermissionEntityFilterRequest,
	// A linked entrance.
	entrance *schema.LinkedEntrance,
	// A map that keeps track of visited entities to avoid infinite loops.
	visits *ERMap,
	// An errgroup used for executing goroutines.
	g *errgroup.Group,
	// A custom publisher that publishes results in bulk.
	publisher *BulkPublisher,
) error { // Returns an error if one occurs during execution.
	for _, relation := range []string{tuple.ELLIPSIS, request.GetSubject().GetRelation()} {
		it, err := engine.relationshipReader.QueryRelationships(ctx, request.GetTenantId(), &base.TupleFilter{
			Entity: &base.EntityFilter{
				Type: entrance.TargetEntrance.GetType(),
				Ids:  []string{},
			},
			Relation: entrance.TupleSetRelation, // Query for relationships that match the tuple set relation.
			Subject: &base.SubjectFilter{
				Type:     request.GetSubject().GetType(),
				Ids:      []string{request.GetSubject().GetId()},
				Relation: relation,
			},
		}, request.GetMetadata().GetSnapToken())
		if err != nil {
			return err
		}

		for it.HasNext() { // Loop over each relationship.
			current := it.GetNext()
			g.Go(func() error {
				return engine.l(ctx, request, &base.EntityAndRelation{ // Call the run method with a new entity and relation.
					Entity: &base.Entity{
						Type: entrance.TargetEntrance.GetType(),
						Id:   current.GetEntity().GetId(),
					},
					Relation: entrance.TargetEntrance.GetRelation(),
				}, visits, g, publisher)
			})
		}
	}
	return nil
}

// run is a method of the EntityFilterEngine struct. It executes the linked entity engine for a given request.
func (engine *EntityFilterEngine) l(
	ctx context.Context, // A context used for tracing and cancellation.
	request *base.PermissionEntityFilterRequest, // A permission request for linked entities.
	found *base.EntityAndRelation, // An entity and relation that was previously found.
	visits *ERMap, // A map that keeps track of visited entities to avoid infinite loops.
	g *errgroup.Group, // An errgroup used for executing goroutines.
	publisher *BulkPublisher, // A custom publisher that publishes results in bulk.
) error { // Returns an error if one occurs during execution.

	if !visits.Add(found) { // If the entity and relation has already been visited.
		return nil
	}

	// Retrieve entity definition
	sc, err := engine.readSchema(ctx, request.GetTenantId(), request.GetMetadata().GetSchemaVersion())
	if err != nil {
		return err
	}

	// Retrieve linked entrances
	cn := schema.NewLinkedGraph(sc)
	var entrances []*schema.LinkedEntrance
	entrances, err = cn.RelationshipLinkedEntrances(
		&base.RelationReference{
			Type:     request.GetEntityReference().GetType(),
			Relation: request.GetEntityReference().GetRelation(),
		},
		&base.RelationReference{
			Type:     request.GetSubject().GetType(),
			Relation: request.GetSubject().GetRelation(),
		},
	) // Retrieve the linked entrances for the request.

	if entrances == nil { // If there are no linked entrances for the request.
		if found.GetEntity().GetType() == request.GetEntityReference().GetType() && found.GetRelation() == request.GetEntityReference().GetRelation() { // Check if the found entity matches the requested entity reference.
			publisher.Publish(found.GetEntity(), &base.PermissionCheckRequestMetadata{ // Publish the found entity with the permission check metadata.
				SnapToken:     request.GetMetadata().GetSnapToken(),
				SchemaVersion: request.GetMetadata().GetSchemaVersion(),
				Depth:         request.GetMetadata().GetDepth(),
			}, base.PermissionCheckResponse_RESULT_UNKNOWN)
			return nil
		}
		return nil // Otherwise, return without publishing any results.
	}

	g.Go(func() error {
		return engine.EntityFilter(ctx, &base.PermissionEntityFilterRequest{ // Call the Run method recursively with a new permission request.
			TenantId:        request.GetTenantId(),
			EntityReference: request.GetEntityReference(),
			Subject: &base.Subject{
				Type:     found.GetEntity().GetType(),
				Id:       found.GetEntity().GetId(),
				Relation: found.GetRelation(),
			},
			Metadata: request.GetMetadata(),
		}, visits, publisher)
	})
	return nil
}

// getSchema is a method of the EntityFilterEngine struct. It retrieves the schema for a given tenant and schema version.
func (engine *EntityFilterEngine) readSchema(ctx context.Context, tenantID string, schemaVersion string) (*base.SchemaDefinition, error) {
	// Create a cache key by concatenating the tenantID and schemaVersion with a separator.
	cacheKey := tenantID + "|" + schemaVersion

	// Check if the schema is present in the cache by trying to load it using the cacheKey.
	if sch, ok := engine.schemaMap.Load(cacheKey); ok {
		// If the schema is found in the cache, cast it to the appropriate type and return it.
		return sch.(*base.SchemaDefinition), nil
	}

	// If the schema is not found in the cache, read it using the schemaReader.
	sch, err := engine.schemaReader.ReadSchema(ctx, tenantID, schemaVersion)
	if err != nil {
		// If there's an error while reading the schema, return the error.
		return nil, err
	}

	// If the schema is successfully read, store it in the cache using the cacheKey.
	engine.schemaMap.Store(cacheKey, sch)

	// Return the schema.
	return sch, nil
}
