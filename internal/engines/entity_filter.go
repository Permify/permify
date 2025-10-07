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
)

// EntityFilter is a struct that performs permission checks on a set of entities
type EntityFilter struct {
	// dataReader is responsible for reading relationship information
	dataReader storage.DataReader

	schema *base.SchemaDefinition
}

// NewEntityFilter creates a new EntityFilter engine
func NewEntityFilter(dataReader storage.DataReader, sch *base.SchemaDefinition) *EntityFilter {
	return &EntityFilter{
		dataReader: dataReader,
		schema:     sch,
	}
}

// EntityFilter is a method of the EntityFilterEngine struct. It executes a permission request for linked entities.
func (engine *EntityFilter) EntityFilter(
	ctx context.Context, // A context used for tracing and cancellation.
	request *base.PermissionEntityFilterRequest, // A permission request for linked entities.
	visits *VisitsMap, // A map that keeps track of visited entities to avoid infinite loops.
	publisher *BulkEntityPublisher, // A custom publisher that publishes results in bulk.
) (err error) { // Returns an error if one occurs during execution.
	// Check if direct result
	if request.GetEntrance().GetType() == request.GetSubject().GetType() && request.GetEntrance().GetValue() == request.GetSubject().GetRelation() {
		found := &base.Entity{
			Type: request.GetSubject().GetType(),
			Id:   request.GetSubject().GetId(),
		}

		if !visits.AddPublished(found) { // If the entity and relation has already been visited.
			return nil
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
	entrances, err = cn.LinkedEntrances(
		request.GetEntrance(),
		&base.Entrance{
			Type:  request.GetSubject().GetType(),
			Value: request.GetSubject().GetRelation(),
		},
	) // Retrieve the linked entrances between the entity reference and subject.

	if entrances == nil {
		return nil
	}

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
			err = engine.lt(cont, request, &base.EntityAndRelation{ // Call the run method with a new entity and relation.
				Entity: &base.Entity{
					Type: entrance.TargetEntrance.GetType(),
					Id:   request.GetSubject().GetId(),
				},
				Relation: entrance.TargetEntrance.GetValue(),
			}, visits, g, publisher)
			if err != nil {
				return err
			}
		case schema.AttributeLinkedEntrance: // If the linked entrance is a computed user set entrance.
			err = engine.attributeEntrance(cont, request, entrance, visits, publisher) // Call the tuple to user set entrance method.
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
func (engine *EntityFilter) attributeEntrance(
	ctx context.Context, // A context used for tracing and cancellation.
	request *base.PermissionEntityFilterRequest, // A permission request for linked entities.
	entrance *schema.LinkedEntrance, // A linked entrance.
	visits *VisitsMap, // A map that keeps track of visited entities to avoid infinite loops.
	publisher *BulkEntityPublisher, // A custom publisher that publishes results in bulk.
) error { // Returns an error if one occurs during execution.
	if request.GetEntrance().GetType() != entrance.TargetEntrance.GetType() {
		return nil
	}

	if !visits.AddEA(entrance.TargetEntrance.GetType(), entrance.TargetEntrance.GetValue()) { // If the entity and relation has already been visited.
		return nil
	}

	// Retrieve the scope associated with the target entrance type.
	// Check if it exists to avoid accessing a nil map entry.
	scope, exists := request.GetScope()[entrance.TargetEntrance.GetType()]

	// Initialize data as an empty slice of strings.
	var data []string

	// If the scope exists, assign its Data field to the data slice.
	if exists {
		data = scope.GetData()
	}

	// Define a TupleFilter. This specifies which tuples we're interested in.
	// We want tuples that match the entity type and ID from the request, and have a specific relation.
	filter := &base.AttributeFilter{
		Entity: &base.EntityFilter{
			Type: entrance.TargetEntrance.GetType(),
			Ids:  data,
		},
		Attributes: []string{entrance.TargetEntrance.GetValue()},
	}

	var (
		cti, rit   *database.AttributeIterator
		err        error
		pagination database.CursorPagination
	)

	pagination = database.NewCursorPagination(database.Cursor(request.GetCursor()), database.Sort("entity_id"))

	// Query the relationships using the specified pagination settings.
	// The context tuples are filtered based on the provided filter.
	cti, err = storageContext.NewContextualAttributes(request.GetContext().GetAttributes()...).QueryAttributes(filter, pagination)
	if err != nil {
		return err
	}

	// Query the relationships for the entity in the request.
	// The results are filtered based on the provided filter and pagination settings.
	rit, err = engine.dataReader.QueryAttributes(ctx, request.GetTenantId(), filter, request.GetMetadata().GetSnapToken(), pagination)
	if err != nil {
		return err
	}

	// Create a new UniqueTupleIterator from the two TupleIterators.
	// NewUniqueTupleIterator() ensures that the iterator only returns unique tuples.
	it := database.NewUniqueAttributeIterator(rit, cti)

	// Iterate over the relationships.
	for it.HasNext() {
		// Get the next attribute's entity.
		current, ok := it.GetNext()
		if !ok {
			break
		}

		// Extract the entity details.
		entity := &base.Entity{
			Type: entrance.TargetEntrance.GetType(), // Example: using the type from a previous variable 'entrance'
			Id:   current.GetEntity().GetId(),
		}

		// Check if the entity has already been visited to prevent processing it again.
		if !visits.AddPublished(entity) {
			continue // Skip this entity if it has already been visited.
		}

		// Publish the entity with its metadata.
		publisher.Publish(entity, &base.PermissionCheckRequestMetadata{
			SnapToken:     request.GetMetadata().GetSnapToken(),
			SchemaVersion: request.GetMetadata().GetSchemaVersion(),
			Depth:         request.GetMetadata().GetDepth(),
		}, request.GetContext(), base.CheckResult_CHECK_RESULT_UNSPECIFIED)
	}

	return nil
}

// relationEntrance is a method of the EntityFilterEngine struct. It handles relation entrances.
func (engine *EntityFilter) relationEntrance(
	ctx context.Context, // A context used for tracing and cancellation.
	request *base.PermissionEntityFilterRequest, // A permission request for linked entities.
	entrance *schema.LinkedEntrance, // A linked entrance.
	visits *VisitsMap, // A map that keeps track of visited entities to avoid infinite loops.
	g *errgroup.Group, // An errgroup used for executing goroutines.
	publisher *BulkEntityPublisher, // A custom publisher that publishes results in bulk.
) error { // Returns an error if one occurs during execution.
	// Retrieve the scope associated with the target entrance type.
	// Check if it exists to avoid accessing a nil map entry.
	scope, exists := request.GetScope()[entrance.TargetEntrance.GetType()]

	// Initialize data as an empty slice of strings.
	var data []string

	// If the scope exists, assign its Data field to the data slice.
	if exists {
		data = scope.GetData()
	}

	// Define a TupleFilter. This specifies which tuples we're interested in.
	// We want tuples that match the entity type and ID from the request, and have a specific relation.
	filter := &base.TupleFilter{
		Entity: &base.EntityFilter{
			Type: entrance.TargetEntrance.GetType(),
			Ids:  data,
		},
		Relation: entrance.TargetEntrance.GetValue(),
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
	if request.GetEntrance().GetType() == entrance.TargetEntrance.GetType() {
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
			return engine.lt(ctx, request, &base.EntityAndRelation{ // Call the run method with a new entity and relation.
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
func (engine *EntityFilter) tupleToUserSetEntrance(
	// A context used for tracing and cancellation.
	ctx context.Context,
	// A permission request for linked entities.
	request *base.PermissionEntityFilterRequest,
	// A linked entrance.
	entrance *schema.LinkedEntrance,
	// A map that keeps track of visited entities to avoid infinite loops.
	visits *VisitsMap,
	// An errgroup used for executing goroutines.
	g *errgroup.Group,
	// A custom publisher that publishes results in bulk.
	publisher *BulkEntityPublisher,
) error { // Returns an error if one occurs during execution.
	// Retrieve the scope associated with the target entrance type.
	// Check if it exists to avoid accessing a nil map entry.
	scope, exists := request.GetScope()[entrance.TargetEntrance.GetType()]

	// Initialize data as an empty slice of strings.
	var data []string

	// If the scope exists, assign its Data field to the data slice.
	if exists {
		data = scope.GetData()
	}

	// Define a TupleFilter. This specifies which tuples we're interested in.
	// We want tuples that match the entity type and ID from the request, and have a specific relation.
	filter := &base.TupleFilter{
		Entity: &base.EntityFilter{
			Type: entrance.TargetEntrance.GetType(),
			Ids:  data,
		},
		Relation: entrance.TupleSetRelation, // Query for relationships that match the tuple set relation.
		Subject: &base.SubjectFilter{
			Type:     request.GetSubject().GetType(),
			Ids:      []string{request.GetSubject().GetId()},
			Relation: "",
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
	if request.GetEntrance().GetType() == entrance.TargetEntrance.GetType() {
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
			return engine.lt(ctx, request, &base.EntityAndRelation{ // Call the run method with a new entity and relation.
				Entity: &base.Entity{
					Type: entrance.TargetEntrance.GetType(),
					Id:   current.GetEntity().GetId(),
				},
				Relation: entrance.TargetEntrance.GetValue(),
			}, visits, g, publisher)
		})
	}
	return nil
}

// run is a method of the EntityFilterEngine struct. It executes the linked entity engine for a given request.
func (engine *EntityFilter) lt(
	ctx context.Context, // A context used for tracing and cancellation.
	request *base.PermissionEntityFilterRequest, // A permission request for linked entities.
	found *base.EntityAndRelation, // An entity and relation that was previously found.
	visits *VisitsMap, // A map that keeps track of visited entities to avoid infinite loops.
	g *errgroup.Group, // An errgroup used for executing goroutines.
	publisher *BulkEntityPublisher, // A custom publisher that publishes results in bulk.
) error { // Returns an error if one occurs during execution.
	if !visits.AddER(found.GetEntity(), found.GetRelation()) { // If the entity and relation has already been visited.
		return nil
	}

	var err error

	// Retrieve linked entrances
	cn := schema.NewLinkedGraph(engine.schema)
	var entrances []*schema.LinkedEntrance
	entrances, err = cn.LinkedEntrances(
		request.GetEntrance(),
		&base.Entrance{
			Type:  request.GetSubject().GetType(),
			Value: request.GetSubject().GetRelation(),
		},
	) // Retrieve the linked entrances for the request.
	if err != nil {
		return err
	}

	if entrances == nil { // If there are no linked entrances for the request.
		if found.GetEntity().GetType() == request.GetEntrance().GetType() && found.GetRelation() == request.GetEntrance().GetValue() { // Check if the found entity matches the requested entity reference.
			if !visits.AddPublished(found.GetEntity()) { // If the entity and relation has already been visited.
				return nil
			}
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
			TenantId: request.GetTenantId(),
			Entrance: request.GetEntrance(),
			Subject: &base.Subject{
				Type:     found.GetEntity().GetType(),
				Id:       found.GetEntity().GetId(),
				Relation: found.GetRelation(),
			},
			Scope:    request.GetScope(),
			Metadata: request.GetMetadata(),
			Context:  request.GetContext(),
			Cursor:   request.GetCursor(),
		}, visits, publisher)
	})
	return nil
}
