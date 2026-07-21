package engines

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/sync/errgroup"

	"github.com/Permify/permify/internal/schema"
	"github.com/Permify/permify/internal/storage"
	storageContext "github.com/Permify/permify/internal/storage/context"
	tokenutils "github.com/Permify/permify/internal/storage/context/utils"
	"github.com/Permify/permify/pkg/database"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// _maxBFSDepth bounds same-type recursive relation expansion in entity lookup.
const _maxBFSDepth = 100

// EntityFilter is a struct that performs permission checks on a set of entities
type EntityFilter struct {
	// dataReader is responsible for reading relationship information
	dataReader storage.DataReader

	graph *schema.LinkedSchemaGraph

	// maxBatchSize limits how many found entities are processed per batch in recursive calls
	maxBatchSize int
}

// NewEntityFilter creates a new EntityFilter engine
func NewEntityFilter(dataReader storage.DataReader, sch *base.SchemaDefinition, maxBatchSize int) *EntityFilter {
	if maxBatchSize <= 0 {
		maxBatchSize = _defaultMaxBatchSize
	}
	return &EntityFilter{
		dataReader:   dataReader,
		graph:        schema.NewLinkedGraph(sch),
		maxBatchSize: maxBatchSize,
	}
}

// EntityFilter is a method of the EntityFilterEngine struct. It executes a permission request for linked entities.
// subjectIDs allows batching: multiple subject IDs with the same (Type, Relation) are queried in one DB call.
// Pass nil to use request.GetSubject().GetId() as the single subject ID.
func (engine *EntityFilter) EntityFilter(
	ctx context.Context, // A context used for tracing and cancellation.
	request *base.PermissionEntityFilterRequest, // A permission request for linked entities.
	subjectIDs []string, // Batch subject IDs (nil = use request.Subject.Id).
	visits *VisitsMap, // A map that keeps track of visited entities to avoid infinite loops.
	publisher *BulkEntityPublisher, // A custom publisher that publishes results in bulk.
) (err error) { // Returns an error if one occurs during execution.
	if len(subjectIDs) == 0 {
		subjectIDs = []string{request.GetSubject().GetId()}
	}

	// Check if direct result
	if request.GetEntrance().GetType() == request.GetSubject().GetType() && request.GetEntrance().GetValue() == request.GetSubject().GetRelation() {
		for _, id := range subjectIDs {
			found := &base.Entity{
				Type: request.GetSubject().GetType(),
				Id:   id,
			}

			if !visits.AddPublished(found) { // If the entity and relation has already been visited.
				continue
			}

			// If the entity reference is the same as the subject, publish the result directly and return.
			publisher.Publish(found, &base.PermissionCheckRequestMetadata{
				SnapToken:     request.GetMetadata().GetSnapToken(),
				SchemaVersion: request.GetMetadata().GetSchemaVersion(),
				Depth:         request.GetMetadata().GetDepth(),
			}, request.GetContext(), base.CheckResult_CHECK_RESULT_UNSPECIFIED)
		}
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
			err = engine.relationEntrance(cont, request, entrance, subjectIDs, visits, g, publisher) // Call the relation entrance method.
			if err != nil {
				return err
			}
		case schema.ComputedUserSetLinkedEntrance: // If the linked entrance is a computed user set entrance.
			err = engine.processFoundEntities(cont, request, entrance.TargetEntrance.GetType(), entrance.TargetEntrance.GetValue(), subjectIDs, visits, g, publisher)
			if err != nil {
				return err
			}
		case schema.AttributeLinkedEntrance: // If the linked entrance is a computed user set entrance.
			err = engine.attributeEntrance(cont, request, entrance, visits, publisher) // Call the tuple to user set entrance method.
			if err != nil {
				return err
			}
		case schema.TupleToUserSetLinkedEntrance: // If the linked entrance is a tuple to user set entrance.
			err = engine.tupleToUserSetEntrance(cont, request, entrance, subjectIDs, visits, g, publisher) // Call the tuple to user set entrance method.
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

	selfCycleRelations := engine.graph.SelfCycleRelationsForPermission(
		request.GetEntrance().GetType(),
		request.GetEntrance().GetValue(),
	)

	expandRecursive := request.GetEntrance().GetType() == entrance.TargetEntrance.GetType() &&
		len(selfCycleRelations) > 0

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

	// Only publish entities of the target type (the type we're looking up).
	// Attribute entrances on intermediate types are not candidates.
	if entrance.TargetEntrance.GetType() != request.GetEntrance().GetType() {
		return nil
	}

	var attributeEntityIDs []string
	attributeEntityIDSet := make(map[string]struct{})

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

		if expandRecursive {
			if _, ok := attributeEntityIDSet[entity.GetId()]; !ok {
				attributeEntityIDSet[entity.GetId()] = struct{}{}
				attributeEntityIDs = append(attributeEntityIDs, entity.GetId())
			}
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

	// For same-type recursive permissions, collect recursion seeds
	// without cursor filtering so descendant entities remain reachable on later pages.
	if expandRecursive && request.GetCursor() != "" {
		seedPagination := database.NewCursorPagination(database.Sort("entity_id"))
		seedCTI, err := storageContext.NewContextualAttributes(request.GetContext().GetAttributes()...).QueryAttributes(filter, seedPagination)
		if err != nil {
			return err
		}

		seedRIT, err := engine.dataReader.QueryAttributes(ctx, request.GetTenantId(), filter, request.GetMetadata().GetSnapToken(), seedPagination)
		if err != nil {
			return err
		}

		seedIt := database.NewUniqueAttributeIterator(seedRIT, seedCTI)
		for seedIt.HasNext() {
			current, ok := seedIt.GetNext()
			if !ok {
				break
			}

			id := current.GetEntity().GetId()
			if _, ok := attributeEntityIDSet[id]; ok {
				continue
			}
			attributeEntityIDSet[id] = struct{}{}
			attributeEntityIDs = append(attributeEntityIDs, id)
		}
	}

	// Expand recursive relations for same-type attribute permissions
	if expandRecursive && len(attributeEntityIDs) > 0 {
		for _, relation := range selfCycleRelations {
			err := engine.expandRecursiveRelation(ctx, request, entrance.TargetEntrance.GetType(), relation, attributeEntityIDs, visits, publisher)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// decodeCursorValue decodes a cursor token and returns its underlying value.
// It returns an empty string when the cursor is empty.
// If decoding fails, the error is returned.
// If the decoded token is not a ContinuousToken, it returns an empty string.
func decodeCursorValue(cursor string) (string, error) {
	if cursor == "" {
		return "", nil
	}
	t, err := tokenutils.EncodedContinuousToken{Value: cursor}.Decode()
	if err != nil {
		return "", err
	}
	decoded, ok := t.(tokenutils.ContinuousToken)
	if !ok {
		return "", nil
	}
	return decoded.Value, nil
}

// expandRecursiveRelation publishes all entities reachable from seed subjects via a relation,
// walking the relation transitively (self-recursive permissions).
func (engine *EntityFilter) expandRecursiveRelation(
	ctx context.Context,
	request *base.PermissionEntityFilterRequest,
	entityType string,
	relation string,
	seedSubjectIDs []string,
	visits *VisitsMap,
	publisher *BulkEntityPublisher,
) error {
	if len(seedSubjectIDs) == 0 {
		return nil
	}

	cursorValue := ""
	if request.GetEntrance().GetType() == entityType && request.GetCursor() != "" {
		var err error
		cursorValue, err = decodeCursorValue(request.GetCursor())
		if err != nil {
			return err
		}
	}

	scope, exists := request.GetScope()[entityType]
	var data []string
	if exists {
		data = scope.GetData()
	}

	seen := make(map[string]struct{}, len(seedSubjectIDs))
	queue := make([]string, 0, len(seedSubjectIDs))
	for _, id := range seedSubjectIDs {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		queue = append(queue, id)
	}

	hops := 0
	for len(queue) > 0 {
		if hops >= _maxBFSDepth {
			return fmt.Errorf("recursive relation expansion exceeded maximum depth (%d)", _maxBFSDepth)
		}
		hops++

		currentIDs := queue
		queue = nil

		filter := &base.TupleFilter{
			Entity: &base.EntityFilter{
				Type: entityType,
				Ids:  data,
			},
			Relation: relation,
			Subject: &base.SubjectFilter{
				Type:     entityType,
				Ids:      currentIDs,
				Relation: "",
			},
		}

		pagination := database.NewCursorPagination()
		cti, err := storageContext.NewContextualTuples(request.GetContext().GetTuples()...).QueryRelationships(filter, pagination)
		if err != nil {
			return err
		}

		rit, err := engine.dataReader.QueryRelationships(ctx, request.GetTenantId(), filter, request.GetMetadata().GetSnapToken(), pagination)
		if err != nil {
			return err
		}

		it := database.NewUniqueTupleIterator(rit, cti)
		for it.HasNext() {
			current, ok := it.GetNext()
			if !ok {
				break
			}

			entity := &base.Entity{
				Type: current.GetEntity().GetType(),
				Id:   current.GetEntity().GetId(),
			}

			if cursorValue == "" || entity.GetId() >= cursorValue {
				if visits.AddPublished(entity) {
					publisher.Publish(entity, &base.PermissionCheckRequestMetadata{
						SnapToken:     request.GetMetadata().GetSnapToken(),
						SchemaVersion: request.GetMetadata().GetSchemaVersion(),
						Depth:         request.GetMetadata().GetDepth(),
					}, request.GetContext(), base.CheckResult_CHECK_RESULT_UNSPECIFIED)
				}
			}

			if _, ok := seen[entity.GetId()]; ok {
				continue
			}
			seen[entity.GetId()] = struct{}{}
			queue = append(queue, entity.GetId())
		}
	}

	return nil
}

// relationEntrance is a method of the EntityFilterEngine struct. It handles relation entrances.
// Uses batch subject IDs in SubjectFilter for a single DB query instead of per-entity queries.
func (engine *EntityFilter) relationEntrance(
	ctx context.Context, // A context used for tracing and cancellation.
	request *base.PermissionEntityFilterRequest, // A permission request for linked entities.
	entrance *schema.LinkedEntrance, // A linked entrance.
	subjectIDs []string, // Batch subject IDs for SubjectFilter.
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
			Ids:      subjectIDs, // Batch: all subject IDs in one query
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

	// Process results in chunks to avoid large intermediate slices.
	// All results share the same entity type and relation (from the filter).
	for it.HasNext() {
		var chunkType, chunkRelation string
		chunk := make([]string, 0, engine.maxBatchSize)
		for it.HasNext() && len(chunk) < engine.maxBatchSize {
			current, ok := it.GetNext()
			if !ok {
				break
			}
			chunkType = current.GetEntity().GetType()
			chunkRelation = current.GetRelation()
			chunk = append(chunk, current.GetEntity().GetId())
		}
		if err := engine.processFoundEntities(ctx, request, chunkType, chunkRelation, chunk, visits, g, publisher); err != nil {
			return err
		}
	}
	return nil
}

// tupleToUserSetEntrance is a method of the EntityFilterEngine struct. It handles tuple to user set entrances.
// Uses batch subject IDs in SubjectFilter for a single DB query instead of per-entity queries.
func (engine *EntityFilter) tupleToUserSetEntrance(
	// A context used for tracing and cancellation.
	ctx context.Context,
	// A permission request for linked entities.
	request *base.PermissionEntityFilterRequest,
	// A linked entrance.
	entrance *schema.LinkedEntrance,
	// Batch subject IDs for SubjectFilter.
	subjectIDs []string,
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
			Ids:      subjectIDs, // Batch: all subject IDs in one query
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

	// Process results in chunks to avoid large intermediate slices.
	// All results share the same entity type and relation (from the entrance).
	entType := entrance.TargetEntrance.GetType()
	entRel := entrance.TargetEntrance.GetValue()
	for it.HasNext() {
		chunk := make([]string, 0, engine.maxBatchSize)
		for it.HasNext() && len(chunk) < engine.maxBatchSize {
			current, ok := it.GetNext()
			if !ok {
				break
			}
			chunk = append(chunk, current.GetEntity().GetId())
		}
		if err := engine.processFoundEntities(ctx, request, entType, entRel, chunk, visits, g, publisher); err != nil {
			return err
		}
	}
	return nil
}

// processFoundEntities handles batch processing of found entities at a single graph level.
// All founds share the same (entityType, relation). It checks LinkedEntrances once,
// publishes direct matches, and recursively calls EntityFilter with batch subject IDs.
func (engine *EntityFilter) processFoundEntities(
	ctx context.Context,
	request *base.PermissionEntityFilterRequest,
	entityType string,
	relation string,
	entityIds []string,
	visits *VisitsMap,
	g *errgroup.Group,
	publisher *BulkEntityPublisher,
) error {
	// Visit check each entity.
	var filtered []string
	for _, id := range entityIds {
		if visits.AddER(&base.Entity{Type: entityType, Id: id}, relation) {
			filtered = append(filtered, id)
		}
	}
	if len(filtered) == 0 {
		return nil
	}

	// Compute LinkedEntrances once (depends on Type+Relation, not Id).
	entrances, err := engine.graph.LinkedEntrances(
		request.GetEntrance(),
		&base.Entrance{Type: entityType, Value: relation},
	)
	if err != nil {
		return err
	}

	if entrances == nil {
		// Direct match: publish all entities.
		if entityType == request.GetEntrance().GetType() && relation == request.GetEntrance().GetValue() {
			for _, id := range filtered {
				entity := &base.Entity{Type: entityType, Id: id}
				if !visits.AddPublished(entity) {
					continue
				}
				publisher.Publish(entity, &base.PermissionCheckRequestMetadata{
					SnapToken:     request.GetMetadata().GetSnapToken(),
					SchemaVersion: request.GetMetadata().GetSchemaVersion(),
					Depth:         request.GetMetadata().GetDepth(),
				}, request.GetContext(), base.CheckResult_CHECK_RESULT_UNSPECIFIED)
			}
		}
		return nil
	}

	// Needs recursion: ONE EntityFilter call with all IDs as batch subject IDs.
	// Input is already chunked by the caller (relationEntrance/tupleToUserSetEntrance).
	g.Go(func() error {
		return engine.EntityFilter(ctx, &base.PermissionEntityFilterRequest{
			TenantId: request.GetTenantId(),
			Entrance: request.GetEntrance(),
			Subject: &base.Subject{
				Type:     entityType,
				Id:       filtered[0], // Representative ID for Subject field
				Relation: relation,
			},
			Scope:    request.GetScope(),
			Metadata: request.GetMetadata(),
			Context:  request.GetContext(),
			Cursor:   request.GetCursor(),
		}, filtered, visits, publisher)
	})
	return nil
}

// pathChainEntrance handles multi-hop relation chain traversal for nested attributes
//
// TODO: Optimize performance with smart batching:
// - Extract unique attributes from path chain entrances to avoid duplicate queries
// - Use batch processing when scope limits entity IDs or when attribute count is low (<=1)
// - Use individual processing when no scope exists and multiple attributes are present
// - Consider refactoring into helper functions for improved maintainability
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
				Relation: subjectRelation, // Preserve subject relation for references like @group#member.
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
