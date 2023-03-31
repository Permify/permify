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

// Run is a function that processes PermissionLookupEntityRequest and returns linked entities based on the given request.
// It finds the relationships and associated permissions between entities, and then publishes the results using a BulkPublisher.
// This function handles different types of LinkedEntrance, such as RelationLinkedEntrance, ComputedUserSetLinkedEntrance, and TupleToUserSetLinkedEntrance.
//
// Parameters:
// - ctx: A context that carries deadlines, cancellations, and other request-scoped values.
// - request: A PermissionLookupEntityRequest object that contains information about the entities, permissions, and metadata.
// - visits: An ERMap used to store the visited entity relationships to avoid cyclic dependencies and duplicate processing.
// - publisher: A BulkPublisher used to publish the results.
//
// Returns:
// - An error if any issues are encountered during the process, otherwise nil.
//
// The function starts by checking if the subject type and relation match the requested entity type and permission.
// If they match, it publishes an entity with an unknown permission result and returns.
//
// It then sets the SnapToken and SchemaVersion in the request metadata if they are not provided.
//
// The function retrieves the schema definition and linked entrances based on the entity type, permission, subject type, and subject relation.
//
// It iterates through the entrances, handling each entrance type differently.
// For RelationLinkedEntrance, it calls the relationEntrance function.
// For ComputedUserSetLinkedEntrance, it recursively calls the Run function with an updated subject relation.
// For TupleToUserSetLinkedEntrance, it calls the tupleToUserSetEntrance function.
//
// In case of an unknown entrance type, it returns an error.
//
// Finally, the function waits for all the goroutines to finish and returns any errors encountered.
func (engine *LinkedEntityEngine) Run(ctx context.Context, request *base.PermissionLookupEntityRequest, visits *ERMap, publisher *BulkPublisher) (err error) {
	ctx, span := tracer.Start(ctx, "permissions.linked-entity.execute")
	defer span.End()

	if request.GetSubject().GetType() == request.GetEntityType() && request.GetSubject().GetRelation() == request.GetPermission() {

		// (TODO) direct result and exclusion impl.

		publisher.Publish(&base.Entity{
			Type: request.GetSubject().GetType(),
			Id:   request.GetSubject().GetId(),
		}, &base.PermissionCheckRequestMetadata{
			SnapToken:     request.Metadata.GetSnapToken(),
			SchemaVersion: request.Metadata.GetSchemaVersion(),
			Depth:         request.Metadata.GetDepth(),
			Exclusion:     false,
		}, base.PermissionCheckResponse_RESULT_UNKNOWN)
		return nil
	}

	// Set SnapToken if not provided
	if request.GetMetadata().GetSnapToken() == "" {
		var st token.SnapToken
		st, err = engine.relationshipReader.HeadSnapshot(ctx, request.GetTenantId())
		if err != nil {
			return err
		}
		request.Metadata.SnapToken = st.Encode().String()
	}

	// Set SchemaVersion if not provided
	if request.GetMetadata().GetSchemaVersion() == "" {
		request.Metadata.SchemaVersion, err = engine.schemaReader.HeadVersion(ctx, request.GetTenantId())
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())
			return err
		}
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
	cn := schema.NewLinkedGraph(sc)
	var entrances []schema.LinkedEntrance
	entrances, err = cn.RelationshipLinkedEntrances(
		&base.RelationReference{
			Type:     request.GetEntityType(),
			Relation: request.GetPermission(),
		},
		&base.RelationReference{
			Type:     request.GetSubject().GetType(),
			Relation: request.GetSubject().GetRelation(),
		},
	)

	cctx, cancel := context.WithCancel(ctx)
	defer cancel()

	g, cont := errgroup.WithContext(cctx)

	for _, entrance := range entrances {
		switch entrance.LinkedEntranceKind() {
		case schema.RelationLinkedEntrance:
			err = engine.relationEntrance(cont, LinkedEntityRequest{
				Metadata: request.GetMetadata(),
				TenantID: request.GetTenantId(),
				Entrance: entrance,
				Target: &base.RelationReference{
					Type:     request.GetEntityType(),
					Relation: request.GetPermission(),
				},
				Subject: request.GetSubject(),
			}, g, visits, publisher)
			if err != nil {
				return err
			}
		case schema.ComputedUserSetLinkedEntrance:
			return engine.Run(ctx, &base.PermissionLookupEntityRequest{
				TenantId:   request.GetTenantId(),
				Metadata:   request.GetMetadata(),
				EntityType: request.GetEntityType(),
				Permission: request.GetPermission(),
				Subject: &base.Subject{
					Type:     request.GetSubject().GetType(),
					Id:       request.GetSubject().GetId(),
					Relation: entrance.LinkedEntrance.GetRelation(),
				},
			}, visits, publisher)
		case schema.TupleToUserSetLinkedEntrance:
			err = engine.tupleToUserSetEntrance(cont, LinkedEntityRequest{
				Metadata: request.GetMetadata(),
				TenantID: request.GetTenantId(),
				Entrance: entrance,
				Target: &base.RelationReference{
					Type:     request.GetEntityType(),
					Relation: request.GetPermission(),
				},
				Subject: request.GetSubject(),
			}, g, visits, publisher)
			if err != nil {
				return err
			}
		default:
			return errors.New("unknown linked entrance type")
		}
	}

	return g.Wait()
}

// relationEntrance is a function that handles the execution of a RelationLinkedEntrance.
// It queries relationships based on the provided LinkedEntityRequest and publishes the results using a BulkPublisher.
//
// Parameters:
// - ctx: A context that carries deadlines, cancellations, and other request-scoped values.
// - request: A LinkedEntityRequest object containing the metadata, tenant ID, entrance, target, and subject information.
// - g: An errgroup.Group to manage concurrent goroutines and collect errors.
// - visits: An ERMap used to store the visited entity relationships to avoid cyclic dependencies and duplicate processing.
// - publisher: A BulkPublisher used to publish the results.
//
// Returns:
// - An error if any issues are encountered during the process, otherwise nil.
//
// The function starts by querying relationships using the relationshipReader, based on the entrance type, relation, and subject filter.
// The results are obtained through an iterator.
//
// It iterates through the relationships returned by the iterator.
// If the current entity relationship has not been visited, it adds the entity to the visits ERMap.
//
// The function then spawns a new goroutine, calling the Run function with a PermissionLookupEntityRequest built from the current entity relationship.
//
// The Run function handles the next steps of the process and publishes the results using the provided BulkPublisher.
//
// The relationEntrance function returns nil once all relationships have been processed.
func (engine *LinkedEntityEngine) relationEntrance(
	ctx context.Context,
	request LinkedEntityRequest,
	g *errgroup.Group,
	visits *ERMap,
	publisher *BulkPublisher,
) error {
	it, err := engine.relationshipReader.QueryRelationships(ctx, request.TenantID, &base.TupleFilter{
		Entity: &base.EntityFilter{
			Type: request.Entrance.LinkedEntrance.GetType(),
			Ids:  []string{},
		},
		Relation: request.Entrance.LinkedEntrance.GetRelation(),
		Subject: &base.SubjectFilter{
			Type:     request.Subject.GetType(),
			Ids:      []string{request.Subject.GetId()},
			Relation: request.Subject.Relation,
		},
	}, request.Metadata.GetSnapToken())
	if err != nil {
		return err
	}

	for it.HasNext() {
		current := it.GetNext()

		if !visits.Add(current.GetEntity()) {
			continue
		}

		g.Go(func() error {
			return engine.Run(ctx, &base.PermissionLookupEntityRequest{
				TenantId:   request.TenantID,
				Metadata:   request.Metadata,
				EntityType: request.Target.GetType(),
				Permission: request.Target.GetRelation(),
				Subject: &base.Subject{
					Type:     current.GetEntity().GetType(),
					Id:       current.GetEntity().GetId(),
					Relation: current.GetRelation(),
				},
			}, visits, publisher)
		})
	}
	return nil
}

// tupleToUserSetEntrance is a function that handles the execution of a TupleToUserSetLinkedEntrance.
// It queries relationships based on the provided LinkedEntityRequest and publishes the results using a BulkPublisher.
//
// Parameters:
// - ctx: A context that carries deadlines, cancellations, and other request-scoped values.
// - request: A LinkedEntityRequest object containing the metadata, tenant ID, entrance, target, and subject information.
// - g: An errgroup.Group to manage concurrent goroutines and collect errors.
// - visits: An ERMap used to store the visited entity relationships to avoid cyclic dependencies and duplicate processing.
// - publisher: A BulkPublisher used to publish the results.
//
// Returns:
// - An error if any issues are encountered during the process, otherwise nil.
//
// The function starts by querying relationships using the relationshipReader, based on the entrance type, tuple set relation, and subject filter.
// The results are obtained through an iterator.
//
// It iterates through the relationships returned by the iterator.
// If the current entity relationship has not been visited, it adds the entity to the visits ERMap.
//
// The function then spawns a new goroutine, calling the Run function with a PermissionLookupEntityRequest built from the current entity relationship,
// and the relation from the LinkedEntrance instead of the tuple set relation.
//
// The Run function handles the next steps of the process and publishes the results using the provided BulkPublisher.
//
// The tupleToUserSetEntrance function returns nil once all relationships have been processed.
func (engine *LinkedEntityEngine) tupleToUserSetEntrance(
	ctx context.Context,
	request LinkedEntityRequest,
	g *errgroup.Group,
	visits *ERMap,
	publisher *BulkPublisher,
) error {
	it, err := engine.relationshipReader.QueryRelationships(ctx, request.TenantID, &base.TupleFilter{
		Entity: &base.EntityFilter{
			Type: request.Entrance.LinkedEntrance.GetType(),
			Ids:  []string{},
		},
		Relation: request.Entrance.TupleSetRelation.GetRelation(),
		Subject: &base.SubjectFilter{
			Type:     request.Subject.GetType(),
			Ids:      []string{request.Subject.GetId()},
			Relation: tuple.ELLIPSIS,
		},
	}, request.Metadata.GetSnapToken())
	if err != nil {
		return err
	}

	for it.HasNext() {
		current := it.GetNext()

		if !visits.Add(current.GetEntity()) {
			continue
		}

		g.Go(func() error {
			return engine.Run(ctx, &base.PermissionLookupEntityRequest{
				TenantId:   request.TenantID,
				Metadata:   request.Metadata,
				EntityType: request.Target.GetType(),
				Permission: request.Target.GetRelation(),
				Subject: &base.Subject{
					Type:     current.GetEntity().GetType(),
					Id:       current.GetEntity().GetId(),
					Relation: request.Entrance.LinkedEntrance.GetRelation(),
				},
			}, visits, publisher)
		})
	}
	return nil
}
