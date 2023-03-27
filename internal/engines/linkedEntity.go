package engines

import (
	`context`
	`errors`
	otelCodes `go.opentelemetry.io/otel/codes`
	`golang.org/x/sync/errgroup`

	`github.com/Permify/permify/internal/repositories`
	`github.com/Permify/permify/internal/schema`
	base `github.com/Permify/permify/pkg/pb/base/v1`
	`github.com/Permify/permify/pkg/token`
	`github.com/Permify/permify/pkg/tuple`
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
func (engine *LinkedEntityEngine) Run(ctx context.Context, request *base.PermissionLookupEntityRequest, publisher *BulkPublisher) (err error) {
	ctx, span := tracer.Start(ctx, "permissions.linked-entity.execute")
	defer span.End()

	if request.GetSubject().GetType() == request.GetEntityType() && request.GetSubject().GetRelation() == request.GetPermission() {
		publisher.Publish(convertToCheck(request, request.GetSubject().GetId()), base.PermissionCheckResponse_RESULT_ALLOWED)
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
			err = engine.relationEntrance(cont, request, entrance, g, publisher)
			if err != nil {
				return err
			}
		case schema.ComputedUserSetLinkedEntrance:
			err = engine.computedUserSetEntrance(cont, request, entrance, g, publisher)
			if err != nil {
				return err
			}
		case schema.TupleToUserSetLinkedEntrance:
			err = engine.tupleToUserSetEntrance(cont, request, entrance, g, publisher)
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
func (engine *LinkedEntityEngine) relationEntrance(ctx context.Context, request *base.PermissionLookupEntityRequest, entrance schema.LinkedEntrance, g *errgroup.Group, publisher *BulkPublisher) error {
	it, err := engine.relationshipReader.QueryRelationships(ctx, request.GetTenantId(), &base.TupleFilter{
		Entity: &base.EntityFilter{
			Type: entrance.LinkedEntrance.GetType(),
			Ids:  []string{},
		},
		Relation: entrance.LinkedEntrance.GetRelation(),
		Subject: &base.SubjectFilter{
			Type:     request.GetSubject().GetType(),
			Ids:      []string{request.GetSubject().GetId()},
			Relation: request.GetSubject().GetRelation(),
		},
	}, request.GetMetadata().GetSnapToken())
	if err != nil {
		return err
	}

	for it.HasNext() {
		current := it.GetNext()
		g.Go(func() error {
			return engine.resolve(ctx, &base.EntityAndRelation{
				Entity:   current.GetEntity(),
				Relation: current.GetRelation(),
			}, request, g, publisher)
		})
	}
	return nil
}

// computedUserSetEntrance is responsible for executing a computed user set linked entrance
func (engine *LinkedEntityEngine) computedUserSetEntrance(ctx context.Context, request *base.PermissionLookupEntityRequest, entrance schema.LinkedEntrance, g *errgroup.Group, publisher *BulkPublisher) error {
	return engine.resolve(ctx, &base.EntityAndRelation{
		Entity: &base.Entity{
			Type: entrance.LinkedEntrance.GetType(),
			Id:   request.GetSubject().GetId(),
		},
		Relation: entrance.LinkedEntrance.GetRelation(),
	}, request, g, publisher)
}

// tupleToUserSetEntrance is responsible for executing a tuple to user set linked entrance
func (engine *LinkedEntityEngine) tupleToUserSetEntrance(ctx context.Context, request *base.PermissionLookupEntityRequest, entrance schema.LinkedEntrance, g *errgroup.Group, publisher *BulkPublisher) error {
	relations := []string{tuple.ELLIPSIS, request.GetSubject().GetRelation()}
	for _, relation := range relations {
		it, err := engine.relationshipReader.QueryRelationships(ctx, request.GetTenantId(), &base.TupleFilter{
			Entity: &base.EntityFilter{
				Type: entrance.LinkedEntrance.GetType(),
				Ids:  []string{},
			},
			Relation: entrance.TupleSetRelation.GetRelation(),
			Subject: &base.SubjectFilter{
				Type:     request.GetSubject().GetType(),
				Ids:      []string{request.GetSubject().GetId()},
				Relation: relation,
			},
		}, request.GetMetadata().GetSnapToken())
		if err != nil {
			return err
		}

		for it.HasNext() {
			current := it.GetNext()
			g.Go(func() error {
				return engine.resolve(ctx, &base.EntityAndRelation{
					Entity: &base.Entity{
						Type: entrance.TupleSetRelation.GetType(),
						Id:   current.GetEntity().GetId(),
					},
					Relation: entrance.TupleSetRelation.GetRelation(),
				}, request, g, publisher)
			})
		}
	}
	return nil
}

// resolve is responsible for resolving a linked entity
func (engine *LinkedEntityEngine) resolve(ctx context.Context, found *base.EntityAndRelation, request *base.PermissionLookupEntityRequest, g *errgroup.Group, publisher *BulkPublisher) error {

	// Retrieve entity definition
	sc, err := engine.schemaReader.ReadSchema(ctx, request.GetTenantId(), request.GetMetadata().GetSchemaVersion())
	if err != nil {
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
			Type:     found.GetEntity().GetType(),
			Relation: found.GetRelation(),
		},
	)

	// If no linked entrances are found, we can stop here
	if len(entrances) == 0 {
		if found.GetEntity().GetType() == request.GetEntityType() && found.GetRelation() == request.GetPermission() {
			publisher.Publish(convertToCheck(request, found.GetEntity().GetId()), base.PermissionCheckResponse_RESULT_UNKNOWN)
			return nil
		}
		return nil
	}

	// If the linked entrance is a relation, we need to resolve it
	g.Go(func() error {
		return engine.Run(ctx, &base.PermissionLookupEntityRequest{
			TenantId:   request.GetTenantId(),
			Metadata:   request.GetMetadata(),
			EntityType: request.GetEntityType(),
			Permission: request.GetPermission(),
			Subject: &base.Subject{
				Type:     found.GetEntity().GetType(),
				Id:       found.GetEntity().GetId(),
				Relation: found.GetRelation(),
			},
		}, publisher)
	})

	return nil
}

// convertToCheck a PermissionLookupEntityRequest to a PermissionCheckRequest
func convertToCheck(req *base.PermissionLookupEntityRequest, entityID string) *base.PermissionCheckRequest {
	return &base.PermissionCheckRequest{
		TenantId: req.GetTenantId(),
		Metadata: &base.PermissionCheckRequestMetadata{
			SnapToken:     req.GetMetadata().GetSnapToken(),
			SchemaVersion: req.GetMetadata().GetSchemaVersion(),
			Depth:         req.GetMetadata().GetDepth(),
			Exclusion:     false,
		},
		Entity: &base.Entity{
			Type: req.GetEntityType(),
			Id:   entityID,
		},
		Permission: req.GetPermission(),
		Subject:    req.GetSubject(),
	}
}
