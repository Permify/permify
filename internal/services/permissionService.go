package services

import (
	"context"
	"errors"
	"strings"

	"github.com/Permify/permify/internal/commands"
	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/pkg/dsl/schema"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/token"
)

// PermissionService -
type PermissionService struct {
	// repositories
	sr repositories.SchemaReader
	rr repositories.RelationshipReader
	// commands
	cc  commands.ICheckCommand
	ec  commands.IExpandCommand
	lqc commands.ILookupQueryCommand
}

// NewPermissionService -
func NewPermissionService(cc commands.ICheckCommand, ec commands.IExpandCommand, lqc commands.ILookupQueryCommand, sr repositories.SchemaReader, rr repositories.RelationshipReader) *PermissionService {
	return &PermissionService{
		rr:  rr,
		sr:  sr,
		cc:  cc,
		ec:  ec,
		lqc: lqc,
	}
}

// CheckPermissions -
func (service *PermissionService) CheckPermissions(ctx context.Context, request *base.PermissionCheckRequest) (response *base.PermissionCheckResponse, err error) {
	if request.GetSnapToken() == "" {
		var hs token.SnapToken
		hs, err = service.rr.HeadSnapshot(ctx)
		if err != nil {
			return response, err
		}
		request.SnapToken = hs.Encode().String()
	}

	if request.GetSchemaVersion() == "" {
		var v string
		v, err = service.sr.HeadVersion(ctx)
		if err != nil {
			return response, err
		}
		request.SchemaVersion = v
	}

	var en *base.EntityDefinition
	en, _, err = service.sr.ReadSchemaDefinition(ctx, request.GetEntity().GetType(), request.GetSchemaVersion())
	if err != nil {
		return response, err
	}

	var typeOfRelation base.EntityDefinition_RelationalReference
	typeOfRelation, err = schema.GetTypeOfRelationalReferenceByNameInEntityDefinition(en, request.GetAction())
	if err != nil {
		return response, err
	}

	var child *base.Child
	switch typeOfRelation {
	case base.EntityDefinition_RELATIONAL_REFERENCE_ACTION:
		var a *base.ActionDefinition
		a, err = schema.GetActionByNameInEntityDefinition(en, request.GetAction())
		if err != nil {
			return response, err
		}
		child = a.Child
		break
	case base.EntityDefinition_RELATIONAL_REFERENCE_RELATION:
		var leaf *base.Leaf
		sp := strings.Split(request.GetAction(), ".")
		if len(sp) == 1 {
			computedUserSet := &base.ComputedUserSet{Relation: request.GetAction()}
			leaf = &base.Leaf{
				Type:      &base.Leaf_ComputedUserSet{ComputedUserSet: computedUserSet},
				Exclusion: false,
			}
		} else if len(sp) == 2 {
			tupleToUserSet := &base.TupleToUserSet{Relation: request.GetAction()}
			leaf = &base.Leaf{
				Type:      &base.Leaf_TupleToUserSet{TupleToUserSet: tupleToUserSet},
				Exclusion: false,
			}
		} else {
			return response, errors.New(base.ErrorCode_ERROR_CODE_ACTION_DEFINITION_NOT_FOUND.String())
		}
		child = &base.Child{Type: &base.Child_Leaf{Leaf: leaf}}
		break
	default:
		return response, errors.New(base.ErrorCode_ERROR_CODE_ACTION_DEFINITION_NOT_FOUND.String())
	}

	return service.cc.Execute(ctx, request, child)
}

// ExpandPermissions -
func (service *PermissionService) ExpandPermissions(ctx context.Context, entity *base.Entity, action string, version string, snapToken string) (response commands.ExpandResponse, err error) {
	var en *base.EntityDefinition
	en, _, err = service.sr.ReadSchemaDefinition(ctx, entity.GetType(), version)
	if err != nil {
		return response, err
	}

	var a *base.ActionDefinition
	a, err = schema.GetActionByNameInEntityDefinition(en, action)
	if err != nil {
		return response, err
	}

	child := a.Child

	q := &commands.ExpandQuery{
		Entity:    entity,
		SnapToken: snapToken,
	}

	return service.ec.Execute(ctx, q, child)
}

// LookupQueryPermissions -
func (service *PermissionService) LookupQueryPermissions(ctx context.Context, entityType string, subject *base.Subject, action string, version string) (response commands.LookupQueryResponse, err error) {
	var sch *base.IndexedSchema
	sch, err = service.sr.ReadSchema(ctx, version)
	if err != nil {
		return response, err
	}

	// entityType
	var en *base.EntityDefinition
	en, err = schema.GetEntityByName(sch, entityType)
	if err != nil {
		return response, err
	}

	var a *base.ActionDefinition
	a, err = schema.GetActionByNameInEntityDefinition(en, action)
	if err != nil {
		return response, err
	}

	child := a.Child

	q := &commands.LookupQueryQuery{
		EntityType: entityType,
		Action:     action,
		Subject:    subject,
	}

	q.SetSchema(sch)

	return service.lqc.Execute(ctx, q, child)
}
