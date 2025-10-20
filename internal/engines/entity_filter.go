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

	graph *schema.LinkedSchemaGraph
}

// NewEntityFilter creates a new EntityFilter engine
func NewEntityFilter(dataReader storage.DataReader, sch *base.SchemaDefinition) *EntityFilter {
	return &EntityFilter{
		dataReader: dataReader,
		graph:      schema.NewLinkedGraph(sch),
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
	var entrances []*schema.LinkedEntrance
	entrances, err = engine.graph.LinkedEntrances(
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
		case schema.PathChainLinkedEntrance: // If the linked entrance is a path chain entrance.
			err = engine.pathChainEntrance(cont, request, entrance, visits, publisher) // Call the path chain entrance method.
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
	// attributeEntrance only handles direct attribute access
	if !visits.AddEA(entrance.TargetEntrance.GetType(), entrance.TargetEntrance.GetValue()) {
		return nil
	}

	// Retrieve the scope associated with the target entrance type
	scope, exists := request.GetScope()[entrance.TargetEntrance.GetType()]
	var data []string
	if exists {
		data = scope.GetData()
	}

	// Query attributes directly
	filter := &base.AttributeFilter{
		Entity: &base.EntityFilter{
			Type: entrance.TargetEntrance.GetType(),
			Ids:  data,
		},
		Attributes: []string{entrance.TargetEntrance.GetValue()},
	}

	pagination := database.NewCursorPagination(database.Cursor(request.GetCursor()), database.Sort("entity_id"))

	cti, err := storageContext.NewContextualAttributes(request.GetContext().GetAttributes()...).QueryAttributes(filter, pagination)
	if err != nil {
		return err
	}

	rit, err := engine.dataReader.QueryAttributes(ctx, request.GetTenantId(), filter, request.GetMetadata().GetSnapToken(), pagination)
	if err != nil {
		return err
	}

	it := database.NewUniqueAttributeIterator(rit, cti)

	// Publish entities directly for regular case
	for it.HasNext() {
		current, ok := it.GetNext()
		if !ok {
			break
		}

		entity := &base.Entity{
			Type: entrance.TargetEntrance.GetType(),
			Id:   current.GetEntity().GetId(),
		}

		if !visits.AddPublished(entity) {
			continue
		}

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
	var entrances []*schema.LinkedEntrance
	entrances, err = engine.graph.LinkedEntrances(
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

// pathChainEntrance handles multi-hop relation chain traversal for nested attributes
//
// TODO: This function can be optimized for better performance by implementing smart batching logic:
// - Extract unique attributes from path chain entrances to avoid duplicate queries
// - Implement batch vs individual processing based on scope and attribute count:
//   - Use batch mode when we have scope (limited entity IDs) or few attributes (<=1)
//   - Use individual mode when no scope and multiple attributes to avoid loading large result sets
//   - Refactor into smaller helper functions: extractUniqueAttributes, getScopeIds, shouldUseBatchMode,
//     processBatchMode, processIndividualMode, queryAttributesBatch, processEntranceWithResults
//   - Remove debug statements after optimization is tested
func (engine *EntityFilter) pathChainEntrance(
	ctx context.Context,
	request *base.PermissionEntityFilterRequest,
	entrance *schema.LinkedEntrance,
	visits *VisitsMap,
	publisher *BulkEntityPublisher,
) error {
	if !visits.AddEA(entrance.TargetEntrance.GetType(), entrance.TargetEntrance.GetValue()) {
		return nil
	}

	// 1. Query attributes of the target type with scope optimization
	scope, exists := request.GetScope()[entrance.TargetEntrance.GetType()]
	var data []string
	if exists {
		data = scope.GetData()
	}

	filter := &base.AttributeFilter{
		Entity: &base.EntityFilter{
			Type: entrance.TargetEntrance.GetType(),
			Ids:  data,
		},
		Attributes: []string{entrance.TargetEntrance.GetValue()},
	}

	pagination := database.NewCursorPagination()
	cti, err := storageContext.NewContextualAttributes(request.GetContext().GetAttributes()...).QueryAttributes(filter, pagination)
	if err != nil {
		return err
	}

	rit, err := engine.dataReader.QueryAttributes(ctx, request.GetTenantId(), filter, request.GetMetadata().GetSnapToken(), pagination)
	if err != nil {
		return err
	}

	it := database.NewUniqueAttributeIterator(rit, cti)

	// 2. Collect all attribute entity IDs first (batch approach)
	var attributeEntityIds []string
	sourceType := request.GetEntrance().GetType()
	targetType := entrance.TargetEntrance.GetType()

	// Collect all entity IDs that have the attribute
	for it.HasNext() {
		current, ok := it.GetNext()
		if !ok {
			break
		}
		attributeEntityIds = append(attributeEntityIds, current.GetEntity().GetId())
	}

	if len(attributeEntityIds) == 0 {
		return nil
	}

	// 3. Use the PathChain from entrance to traverse relation chain
	chain := entrance.PathChain
	if len(chain) == 0 {
		return errors.New("PathChainLinkedEntrance missing PathChain")
	}

	// 4. Fold IDs across the relation chain from attribute type back to source type
	currentType := targetType
	currentIds := attributeEntityIds

	for i := len(chain) - 1; i >= 0; i-- {
		walk := chain[i] // walk.Type == left entity type; walk.Relation relates walk.Type -> currentType

		// Apply scope optimization only on the final walk (source type)
		var entIds []string
		if i == 0 {
			if s, exists := request.GetScope()[sourceType]; exists {
				entIds = s.GetData()
			}
		}

		// Determine correct subject relation for complex cases like @group#member
		subjectRelation := engine.graph.GetSubjectRelationForPathWalk(walk.GetType(), walk.GetRelation(), currentType)

		relationFilter := &base.TupleFilter{
			Entity: &base.EntityFilter{
				Type: walk.GetType(),
				Ids:  entIds,
			},
			Relation: walk.GetRelation(),
			Subject: &base.SubjectFilter{
				Type:     currentType,
				Ids:      currentIds,
				Relation: subjectRelation, // Fixed: Use correct subject relation for complex cases
			},
		}

		pagination := database.NewCursorPagination()
		ctiR, err := storageContext.NewContextualTuples(request.GetContext().GetTuples()...).QueryRelationships(relationFilter, pagination)
		if err != nil {
			return err
		}

		ritR, err := engine.dataReader.QueryRelationships(ctx, request.GetTenantId(), relationFilter, request.GetMetadata().GetSnapToken(), pagination)
		if err != nil {
			return err
		}

		relationIt := database.NewUniqueTupleIterator(ritR, ctiR)

		// collect next frontier (left entity IDs)
		nextIdsSet := make(map[string]struct{})
		for relationIt.HasNext() {
			tuple, ok := relationIt.GetNext()
			if !ok {
				break
			}
			nextIdsSet[tuple.GetEntity().GetId()] = struct{}{}
		}

		var nextIds []string
		for id := range nextIdsSet {
			nextIds = append(nextIds, id)
		}

		if len(nextIdsSet) == 0 {
			return nil // No path found through this walk
		}

		// prepare for next walk
		currentType = walk.GetType()
		currentIds = nextIds
	}

	// 5. Publish all resolved source entities
	for _, id := range currentIds {
		entity := &base.Entity{Type: sourceType, Id: id}
		if !visits.AddPublished(entity) {
			continue
		}

		publisher.Publish(entity, &base.PermissionCheckRequestMetadata{
			SnapToken:     request.GetMetadata().GetSnapToken(),
			SchemaVersion: request.GetMetadata().GetSchemaVersion(),
			Depth:         request.GetMetadata().GetDepth(),
		}, request.GetContext(), base.CheckResult_CHECK_RESULT_UNSPECIFIED)
	}

	return nil
}
