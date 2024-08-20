package engines

import (
	"context"
	"errors"

	"golang.org/x/sync/errgroup"

	"github.com/Permify/permify/internal/schema"
	"github.com/Permify/permify/internal/storage"
	storageContext "github.com/Permify/permify/internal/storage/context"
	"github.com/Permify/permify/pkg/database"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
)

// SchemaBasedEntityFilter is a struct that performs permission checks on a set of entities
type SchemaBasedEntityFilter struct {
	// dataReader is responsible for reading relationship information
	dataReader storage.DataReader

	schema *base.SchemaDefinition
}

// NewSchemaBasedEntityFilter creates a new EntityFilter engine
func NewSchemaBasedEntityFilter(dataReader storage.DataReader, sch *base.SchemaDefinition) *SchemaBasedEntityFilter {
	return &SchemaBasedEntityFilter{
		dataReader: dataReader,
		schema:     sch,
	}
}

// EntityFilter is a method of the EntityFilterEngine struct. It executes a permission request for linked entities.
func (engine *SchemaBasedEntityFilter) EntityFilter(
	ctx context.Context, // A context used for tracing and cancellation.
	request *base.PermissionEntityFilterRequest, // A permission request for linked entities.
	visits *ERMap, // A map that keeps track of visited entities to avoid infinite loops.
	publisher *BulkEntityPublisher, // A custom publisher that publishes results in bulk.
) (err error) { // Returns an error if one occurs during execution.
	// Check if direct result
	if request.GetEntityReference().GetType() == request.GetSubject().GetType() && request.GetEntityReference().GetRelation() == request.GetSubject().GetRelation() {
		found := &base.Entity{
			Type: request.GetSubject().GetType(),
			Id:   request.GetSubject().GetId(),
		}

		// If the entity reference is the same as the subject, publish the result directly and return.
		publisher.Publish(found, &base.PermissionCheckRequestMetadata{
			SnapToken:     request.GetMetadata().GetSnapToken(),
			SchemaVersion: request.GetMetadata().GetSchemaVersion(),
			Depth:         request.GetMetadata().GetDepth(),
		}, request.GetContext(), base.CheckResult_CHECK_RESULT_UNSPECIFIED)
	}

	// Retrieve linked entrances
	cn := schema.NewLinkedGraph(engine.schema) // Create a new linked graph from the schema definition.
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
func (engine *SchemaBasedEntityFilter) relationEntrance(
	ctx context.Context, // A context used for tracing and cancellation.
	request *base.PermissionEntityFilterRequest, // A permission request for linked entities.
	entrance *schema.LinkedEntrance, // A linked entrance.
	visits *ERMap, // A map that keeps track of visited entities to avoid infinite loops.
	g *errgroup.Group, // An errgroup used for executing goroutines.
	publisher *BulkEntityPublisher, // A custom publisher that publishes results in bulk.
) error { // Returns an error if one occurs during execution.
	// Define a TupleFilter. This specifies which tuples we're interested in.
	// We want tuples that match the entity type and ID from the request, and have a specific relation.
	filter := &base.TupleFilter{
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
	}

	var (
		cti, rit   *database.TupleIterator
		err        error
		pagination database.CursorPagination
	)

	// Determine the pagination settings based on the entity type in the request.
	// If the entity type matches the target entrance, use cursor pagination with sorting by "entity_id".
	// Otherwise, use the default pagination settings.
	if request.GetEntityReference().GetType() == entrance.TargetEntrance.GetType() {
		pagination = database.NewCursorPagination(database.Cursor(request.GetCursor()), database.Sort("entity_id"))
	} else {
		pagination = database.NewCursorPagination()
	}

	// Query the relationships using the specified pagination settings.
	// The context tuples are filtered based on the provided filter.
	cti, err = storageContext.NewContextualTuples(request.GetContext().GetTuples()...).QueryRelationships(filter, pagination)
	if err != nil {
		return err
	}

	// Query the relationships for the entity in the request.
	// The results are filtered based on the provided filter and pagination settings.
	rit, err = engine.dataReader.QueryRelationships(ctx, request.GetTenantId(), filter, request.GetMetadata().GetSnapToken(), pagination)
	if err != nil {
		return err
	}

	// Create a new UniqueTupleIterator from the two TupleIterators.
	// NewUniqueTupleIterator() ensures that the iterator only returns unique tuples.
	it := database.NewUniqueTupleIterator(rit, cti)

	for it.HasNext() { // Loop over each relationship.
		// Get the next tuple's subject.
		current, ok := it.GetNext()
		if !ok {
			break
		}
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
func (engine *SchemaBasedEntityFilter) tupleToUserSetEntrance(
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
	publisher *BulkEntityPublisher,
) error { // Returns an error if one occurs during execution.
	for _, relation := range []string{"", tuple.ELLIPSIS} {
		// Define a TupleFilter. This specifies which tuples we're interested in.
		// We want tuples that match the entity type and ID from the request, and have a specific relation.
		filter := &base.TupleFilter{
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
		}

		var (
			cti, rit   *database.TupleIterator
			err        error
			pagination database.CursorPagination
		)

		// Determine the pagination settings based on the entity type in the request.
		// If the entity type matches the target entrance, use cursor pagination with sorting by "entity_id".
		// Otherwise, use the default pagination settings.
		if request.GetEntityReference().GetType() == entrance.TargetEntrance.GetType() {
			pagination = database.NewCursorPagination(database.Cursor(request.GetCursor()), database.Sort("entity_id"))
		} else {
			pagination = database.NewCursorPagination()
		}

		// Query the relationships using the specified pagination settings.
		// The context tuples are filtered based on the provided filter.
		cti, err = storageContext.NewContextualTuples(request.GetContext().GetTuples()...).QueryRelationships(filter, pagination)
		if err != nil {
			return err
		}

		// Query the relationships for the entity in the request.
		// The results are filtered based on the provided filter and pagination settings.
		rit, err = engine.dataReader.QueryRelationships(ctx, request.GetTenantId(), filter, request.GetMetadata().GetSnapToken(), pagination)
		if err != nil {
			return err
		}

		// Create a new UniqueTupleIterator from the two TupleIterators.
		// NewUniqueTupleIterator() ensures that the iterator only returns unique tuples.
		it := database.NewUniqueTupleIterator(rit, cti)

		for it.HasNext() { // Loop over each relationship.
			// Get the next tuple's subject.
			current, ok := it.GetNext()
			if !ok {
				break
			}
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
func (engine *SchemaBasedEntityFilter) l(
	ctx context.Context, // A context used for tracing and cancellation.
	request *base.PermissionEntityFilterRequest, // A permission request for linked entities.
	found *base.EntityAndRelation, // An entity and relation that was previously found.
	visits *ERMap, // A map that keeps track of visited entities to avoid infinite loops.
	g *errgroup.Group, // An errgroup used for executing goroutines.
	publisher *BulkEntityPublisher, // A custom publisher that publishes results in bulk.
) error { // Returns an error if one occurs during execution.
	if !visits.Add(found.GetEntity(), found.GetRelation()) { // If the entity and relation has already been visited.
		return nil
	}

	var err error

	// Retrieve linked entrances
	cn := schema.NewLinkedGraph(engine.schema)
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
	if err != nil {
		return err
	}

	if entrances == nil { // If there are no linked entrances for the request.
		if found.GetEntity().GetType() == request.GetEntityReference().GetType() && found.GetRelation() == request.GetEntityReference().GetRelation() { // Check if the found entity matches the requested entity reference.
			publisher.Publish(found.GetEntity(), &base.PermissionCheckRequestMetadata{ // Publish the found entity with the permission check metadata.
				SnapToken:     request.GetMetadata().GetSnapToken(),
				SchemaVersion: request.GetMetadata().GetSchemaVersion(),
				Depth:         request.GetMetadata().GetDepth(),
			}, request.GetContext(), base.CheckResult_CHECK_RESULT_UNSPECIFIED)
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
			Context:  request.GetContext(),
			Cursor:   request.GetCursor(),
		}, visits, publisher)
	})
	return nil
}
