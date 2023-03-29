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

// Run returns a list of linked entities
func (engine *LinkedEntityEngine) Run(ctx context.Context, request *base.PermissionLookupEntityRequest, visits *ERMap, publisher *BulkPublisher) (err error) {
	ctx, span := tracer.Start(ctx, "permissions.linked-entity.execute")
	defer span.End()

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

// relationEntrance is responsible for executing a relation linked entrance
func (engine *LinkedEntityEngine) relationEntrance(
	ctx context.Context,
	request LinkedEntityRequest,
	g *errgroup.Group,
	visited *ERMap,
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

		if current.GetEntity().GetType() == request.Target.GetType() {
			result := base.PermissionCheckResponse_RESULT_UNKNOWN
			if request.Entrance.IsDirect {
				result = base.PermissionCheckResponse_RESULT_ALLOWED
			}
			publisher.Publish(current.GetEntity(), &base.PermissionCheckRequestMetadata{
				SnapToken:     request.Metadata.GetSnapToken(),
				SchemaVersion: request.Metadata.GetSchemaVersion(),
				Depth:         request.Metadata.GetDepth(),
				Exclusion:     false,
			}, result)
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
			}, visited, publisher)
		})
	}
	return nil
}

// tupleToUserSetEntrance is responsible for executing a tuple to user set linked entrance
func (engine *LinkedEntityEngine) tupleToUserSetEntrance(
	ctx context.Context,
	request LinkedEntityRequest,
	g *errgroup.Group,
	visited *ERMap,
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

		if current.GetEntity().GetType() == request.Target.GetType() {
			result := base.PermissionCheckResponse_RESULT_UNKNOWN
			if request.Entrance.IsDirect {
				result = base.PermissionCheckResponse_RESULT_ALLOWED
			}
			publisher.Publish(current.GetEntity(), &base.PermissionCheckRequestMetadata{
				SnapToken:     request.Metadata.GetSnapToken(),
				SchemaVersion: request.Metadata.GetSchemaVersion(),
				Depth:         request.Metadata.GetDepth(),
				Exclusion:     false,
			}, result)
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
			}, visited, publisher)
		})
	}
	return nil
}
