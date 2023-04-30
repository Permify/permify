package engines

import (
	"context"
	"errors"

	otelCodes "go.opentelemetry.io/otel/codes"
	"golang.org/x/sync/errgroup"

	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/internal/schema"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/token"
	"github.com/Permify/permify/pkg/tuple"
)

// LinkedEntityEngine is responsible for executing linked entity operations
type LinkedEntityEngine struct {
	// schemaReader is responsible for reading schema information
	schemaReader repositories.SchemaReader
	// relationshipReader is responsible for reading relationship information
	relationshipReader repositories.RelationshipReader
}

// NewLinkedEntityEngine creates a new LinkedEntity engine
func NewLinkedEntityEngine(schemaReader repositories.SchemaReader, relationshipReader repositories.RelationshipReader) *LinkedEntityEngine {
	return &LinkedEntityEngine{
		schemaReader:       schemaReader,
		relationshipReader: relationshipReader,
	}
}

// Run is a method of the LinkedEntityEngine struct. It executes a permission request for linked entities.
func (engine *LinkedEntityEngine) Run(
	ctx context.Context, // A context used for tracing and cancellation.
	request *base.PermissionLinkedEntityRequest, // A permission request for linked entities.
	visits *ERMap, // A map that keeps track of visited entities to avoid infinite loops.
	publisher *BulkPublisher, // A custom publisher that publishes results in bulk.
) (err error) { // Returns an error if one occurs during execution.
	ctx, span := tracer.Start(ctx, "permissions.linked-entity.execute") // Start a new span for tracing purposes.
	defer span.End()

	// Set SnapToken if not provided
	if request.GetMetadata().GetSnapToken() == "" { // Check if the request has a SnapToken.
		var st token.SnapToken
		st, err = engine.relationshipReader.HeadSnapshot(ctx, request.GetTenantId()) // Retrieve the head snapshot from the relationship reader.
		if err != nil {
			return err
		}
		request.Metadata.SnapToken = st.Encode().String() // Set the SnapToken in the request metadata.
	}

	// Set SchemaVersion if not provided
	if request.GetMetadata().GetSchemaVersion() == "" { // Check if the request has a SchemaVersion.
		request.Metadata.SchemaVersion, err = engine.schemaReader.HeadVersion(ctx, request.GetTenantId()) // Retrieve the head schema version from the schema reader.
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())
			return err
		}
	}

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
			Exclusion:     false,
		}, base.PermissionCheckResponse_RESULT_UNKNOWN)
	}

	// Retrieve entity definition
	var sc *base.SchemaDefinition
	sc, err = engine.schemaReader.ReadSchema(ctx, request.GetTenantId(), request.GetMetadata().GetSchemaVersion())
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
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
			err = engine.run(ctx, request, &base.EntityAndRelation{ // Call the run method with a new entity and relation.
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

// relationEntrance is a method of the LinkedEntityEngine struct. It handles relation entrances.
func (engine *LinkedEntityEngine) relationEntrance(
	ctx context.Context, // A context used for tracing and cancellation.
	request *base.PermissionLinkedEntityRequest, // A permission request for linked entities.
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
			return engine.run(ctx, request, &base.EntityAndRelation{ // Call the run method with a new entity and relation.
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

// tupleToUserSetEntrance is a method of the LinkedEntityEngine struct. It handles tuple to user set entrances.
func (engine *LinkedEntityEngine) tupleToUserSetEntrance(
	// A context used for tracing and cancellation.
	ctx context.Context,
	// A permission request for linked entities.
	request *base.PermissionLinkedEntityRequest,
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
				return engine.run(ctx, request, &base.EntityAndRelation{ // Call the run method with a new entity and relation.
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

// run is a method of the LinkedEntityEngine struct. It executes the linked entity engine for a given request.
func (engine *LinkedEntityEngine) run(
	ctx context.Context, // A context used for tracing and cancellation.
	request *base.PermissionLinkedEntityRequest, // A permission request for linked entities.
	found *base.EntityAndRelation, // An entity and relation that was previously found.
	visits *ERMap, // A map that keeps track of visited entities to avoid infinite loops.
	g *errgroup.Group, // An errgroup used for executing goroutines.
	publisher *BulkPublisher, // A custom publisher that publishes results in bulk.
) error { // Returns an error if one occurs during execution.

	if !visits.Add(found) { // If the entity and relation has already been visited.
		return nil
	}

	sc, err := engine.schemaReader.ReadSchema(ctx, request.GetTenantId(), request.GetMetadata().GetSchemaVersion()) // Retrieve the entity definition for the request.
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
				Exclusion:     false,
			}, base.PermissionCheckResponse_RESULT_UNKNOWN)
			return nil
		}
		return nil // Otherwise, return without publishing any results.
	}

	g.Go(func() error {
		return engine.Run(ctx, &base.PermissionLinkedEntityRequest{ // Call the Run method recursively with a new permission request.
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
