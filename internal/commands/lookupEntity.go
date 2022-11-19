package commands

import (
	`context`
	`errors`
	`fmt`

	`github.com/Permify/permify/internal/repositories`
	`github.com/Permify/permify/pkg/dsl/schema`
	`github.com/Permify/permify/pkg/logger`
	base `github.com/Permify/permify/pkg/pb/base/v1`
	`github.com/Permify/permify/pkg/token`
)

// LookupEntityCommand -
type LookupEntityCommand struct {
	// repositories
	schemaReader       repositories.SchemaReader
	relationshipReader repositories.RelationshipReader
	// logger
	logger logger.Interface
}

// NewLookupEntityCommand -
func NewLookupEntityCommand(sr repositories.SchemaReader, rr repositories.RelationshipReader, l logger.Interface) *LookupEntityCommand {
	return &LookupEntityCommand{
		relationshipReader: rr,
		logger:             l,
	}
}

// Execute -
func (command *LookupEntityCommand) Execute(ctx context.Context, request *base.PermissionLookupEntityRequest) (response *base.PermissionLookupEntityResponse, err error) {

	if request.GetSnapToken() == "" {
		var st token.SnapToken
		st, err = command.relationshipReader.HeadSnapshot(ctx)
		if err != nil {
			return response, err
		}
		request.SnapToken = st.Encode().String()
	}

	if request.GetSchemaVersion() == "" {
		var ver string
		ver, err = command.schemaReader.HeadVersion(ctx)
		if err != nil {
			return response, err
		}
		request.SchemaVersion = ver
	}

	var en *base.EntityDefinition
	en, _, err = command.schemaReader.ReadSchemaDefinition(ctx, request.GetEntityType(), request.GetSchemaVersion())
	if err != nil {
		return response, err
	}

	var typeOfRelation base.EntityDefinition_RelationalReference
	typeOfRelation, err = schema.GetTypeOfRelationalReferenceByNameInEntityDefinition(en, request.GetPermission())
	if err != nil {
		return response, err
	}

	var child *base.Child
	switch typeOfRelation {
	case base.EntityDefinition_RELATIONAL_REFERENCE_ACTION:
		var action *base.ActionDefinition
		action, err = schema.GetActionByNameInEntityDefinition(en, request.GetPermission())
		if err != nil {
			return response, err
		}
		child = action.Child
		break
	case base.EntityDefinition_RELATIONAL_REFERENCE_RELATION:
		var leaf *base.Leaf
		computedUserSet := &base.ComputedUserSet{Relation: request.GetPermission()}
		leaf = &base.Leaf{
			Type:      &base.Leaf_ComputedUserSet{ComputedUserSet: computedUserSet},
			Exclusion: false,
		}
		child = &base.Child{Type: &base.Child_Leaf{Leaf: leaf}}
		break
	default:
		return response, errors.New(base.ErrorCode_ERROR_CODE_ACTION_DEFINITION_NOT_FOUND.String())
	}

	fmt.Println(child)
	//response, err = command.c(ctx, request, child)
	//response.RemainingDepth = request.GetDepth().Value
	return
}
