package development

import (
	"context"

	"github.com/Permify/permify/internal/services"
	"github.com/Permify/permify/pkg/database"
	v1 "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/token"
)

// Check - Creates new permission check request
func Check(ctx context.Context, service services.IPermissionService, subject *v1.Subject, action string, entity *v1.Entity, version, snapToken string) (res *v1.PermissionCheckResponse, err error) {
	req := &v1.PermissionCheckRequest{
		Entity:     entity,
		Subject:    subject,
		Permission: action,
		Metadata: &v1.PermissionCheckRequestMetadata{
			SchemaVersion: version,
			SnapToken:     snapToken,
			Depth:         20,
			Exclusion:     false,
		},
	}
	return service.CheckPermissions(ctx, req)
}

// LookupEntity -
func LookupEntity(ctx context.Context, service services.IPermissionService, subject *v1.Subject, permission, entityType, version, snapToken string) (res *v1.PermissionLookupEntityResponse, err error) {
	req := &v1.PermissionLookupEntityRequest{
		EntityType: entityType,
		Subject:    subject,
		Permission: permission,
		Metadata: &v1.PermissionLookupEntityRequestMetadata{
			SchemaVersion: version,
			SnapToken:     snapToken,
			Depth:         20,
		},
	}
	return service.LookupEntity(ctx, req)
}

// ReadTuple - Creates new read API request
func ReadTuple(ctx context.Context, service services.IRelationshipService, filter *v1.TupleFilter, token string) (tuples database.ITupleCollection, err error) {
	return service.ReadRelationships(ctx, "noop", filter, token)
}

// WriteTuple - Creates new write API request
func WriteTuple(ctx context.Context, service services.IRelationshipService, tuples []*v1.Tuple, version string) (token token.EncodedSnapToken, err error) {
	return service.WriteRelationships(ctx, "noop", tuples, version)
}

// DeleteTuple - Creates new delete relation tuple request
func DeleteTuple(ctx context.Context, service services.IRelationshipService, filter *v1.TupleFilter) (token token.EncodedSnapToken, err error) {
	return service.DeleteRelationships(ctx, "noop", filter)
}

// WriteSchema - Creates new write schema request
func WriteSchema(ctx context.Context, service services.ISchemaService, schema string) (version string, err error) {
	return service.WriteSchema(ctx, "noop", schema)
}

// ReadSchema - Creates new read schema request
func ReadSchema(ctx context.Context, service services.ISchemaService, version string) (sch *v1.IndexedSchema, err error) {
	return service.ReadSchema(ctx, "noop", version)
}
