package commands

import (
	"context"
	"errors"

	otelCodes "go.opentelemetry.io/otel/codes"

	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/pkg/database"
	"github.com/Permify/permify/pkg/dsl/schema"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/token"
	"github.com/Permify/permify/pkg/tuple"
)

// ExpandCommand -
type ExpandCommand struct {
	// repositories
	schemaReader       repositories.SchemaReader
	relationshipReader repositories.RelationshipReader
}

// NewExpandCommand -
func NewExpandCommand(sr repositories.SchemaReader, rr repositories.RelationshipReader) *ExpandCommand {
	return &ExpandCommand{
		schemaReader:       sr,
		relationshipReader: rr,
	}
}

// Execute -
func (command *ExpandCommand) Execute(ctx context.Context, request *base.PermissionExpandRequest) (response *base.PermissionExpandResponse, err error) {
	ctx, span := tracer.Start(ctx, "permissions.expand.execute")
	defer span.End()

	if request.GetMetadata().GetSnapToken() == "" {
		var st token.SnapToken
		st, err = command.relationshipReader.HeadSnapshot(ctx)
		if err != nil {
			return response, err
		}
		request.Metadata.SnapToken = st.Encode().String()
	}

	if request.GetMetadata().GetSchemaVersion() == "" {
		request.Metadata.SchemaVersion, err = command.schemaReader.HeadVersion(ctx)
		if err != nil {
			return response, err
		}
	}

	resp := command.expand(ctx, request, false)
	if resp.Err != nil {
		span.RecordError(resp.Err)
		span.SetStatus(otelCodes.Error, resp.Err.Error())
	}
	return resp.Response, resp.Err
}

// ExpandResponse -
type ExpandResponse struct {
	Response *base.PermissionExpandResponse
	Err      error
}

// ExpandFunction -
type ExpandFunction func(ctx context.Context, expandChain chan<- ExpandResponse)

// ExpandCombiner .
type ExpandCombiner func(ctx context.Context, functions []ExpandFunction) ExpandResponse

// e -
func (command *ExpandCommand) expand(ctx context.Context, request *base.PermissionExpandRequest, exclusion bool) ExpandResponse {
	en, _, err := command.schemaReader.ReadSchemaDefinition(ctx, request.GetEntity().GetType(), request.GetMetadata().GetSchemaVersion())
	if err != nil {
		return ExpandResponse{Err: err}
	}

	var typeOfRelation base.EntityDefinition_RelationalReference
	typeOfRelation, err = schema.GetTypeOfRelationalReferenceByNameInEntityDefinition(en, request.GetPermission())
	if err != nil {
		return ExpandResponse{Err: err}
	}

	var fn ExpandFunction
	if typeOfRelation == base.EntityDefinition_RELATIONAL_REFERENCE_ACTION {
		var child *base.Child
		var action *base.ActionDefinition
		action, err = schema.GetActionByNameInEntityDefinition(en, request.GetPermission())
		if err != nil {
			return ExpandResponse{Err: err}
		}
		child = action.GetChild()
		if child.GetRewrite() != nil {
			fn = command.expandRewrite(ctx, request, child.GetRewrite())
		} else {
			fn = command.expandLeaf(ctx, request, child.GetLeaf())
		}
	} else {
		fn = command.expandDirect(ctx, request, exclusion)
	}

	if fn == nil {
		return ExpandResponse{Err: errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_CHILD_KIND.String())}
	}

	return expandRoot(ctx, fn)
}

// expandRewrite -
func (command *ExpandCommand) expandRewrite(ctx context.Context, request *base.PermissionExpandRequest, rewrite *base.Rewrite) ExpandFunction {
	switch rewrite.GetRewriteOperation() {
	case *base.Rewrite_OPERATION_UNION.Enum():
		return command.setChild(ctx, request, rewrite.GetChildren(), expandUnion)
	case *base.Rewrite_OPERATION_INTERSECTION.Enum():
		return command.setChild(ctx, request, rewrite.GetChildren(), expandIntersection)
	default:
		return expandFail(errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_CHILD_TYPE.String()))
	}
}

// expandLeaf -
func (command *ExpandCommand) expandLeaf(ctx context.Context, request *base.PermissionExpandRequest, leaf *base.Leaf) ExpandFunction {
	switch op := leaf.GetType().(type) {
	case *base.Leaf_TupleToUserSet:
		return command.expandTupleToUserSet(ctx, request, op.TupleToUserSet, leaf.GetExclusion())
	case *base.Leaf_ComputedUserSet:
		return command.expandComputedUserSet(ctx, request, op.ComputedUserSet, leaf.GetExclusion())
	default:
		return expandFail(errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_CHILD_TYPE.String()))
	}
}

// set -
func (command *ExpandCommand) setChild(ctx context.Context, request *base.PermissionExpandRequest, children []*base.Child, combiner ExpandCombiner) ExpandFunction {
	var functions []ExpandFunction
	for _, child := range children {
		switch child.GetType().(type) {
		case *base.Child_Rewrite:
			functions = append(functions, command.expandRewrite(ctx, request, child.GetRewrite()))
		case *base.Child_Leaf:
			functions = append(functions, command.expandLeaf(ctx, request, child.GetLeaf()))
		default:
			return expandFail(errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_CHILD_KIND.String()))
		}
	}

	return func(ctx context.Context, resultChan chan<- ExpandResponse) {
		resultChan <- combiner(ctx, functions)
	}
}

// expandDirect -
func (command *ExpandCommand) expandDirect(ctx context.Context, request *base.PermissionExpandRequest, exclusion bool) ExpandFunction {
	return func(ctx context.Context, expandChan chan<- ExpandResponse) {
		var err error
		var tupleCollection database.ITupleCollection
		tupleCollection, err = command.relationshipReader.QueryRelationships(ctx, &base.TupleFilter{
			Entity: &base.EntityFilter{
				Type: request.GetEntity().GetType(),
				Ids:  []string{request.GetEntity().GetId()},
			},
			Relation: request.GetPermission(),
		}, request.GetMetadata().GetSnapToken())
		if err != nil {
			expandChan <- expandFailResponse(err)
			return
		}

		var expandFunctions []ExpandFunction
		it := tupleCollection.ToSubjectCollection().CreateSubjectIterator()

		directUserCollection := database.NewSubjectCollection()

		for it.HasNext() {
			subject := it.GetNext()
			if !tuple.IsSubjectUser(subject) && subject.GetRelation() != tuple.ELLIPSIS {
				expandFunctions = append(expandFunctions, func(ctx context.Context, resultChan chan<- ExpandResponse) {
					result := command.expand(ctx, &base.PermissionExpandRequest{
						Entity: &base.Entity{
							Type: subject.GetType(),
							Id:   subject.GetId(),
						},
						Permission: subject.GetRelation(),
						Metadata:   request.GetMetadata(),
					}, exclusion)
					resultChan <- result
				})
			} else {
				directUserCollection.Add(subject)
			}
		}

		target := &base.EntityAndRelation{
			Entity:   request.GetEntity(),
			Relation: request.GetPermission(),
		}

		if len(expandFunctions) == 0 {
			expandChan <- ExpandResponse{
				Response: &base.PermissionExpandResponse{
					Tree: &base.Expand{
						Node: &base.Expand_Leaf{
							Leaf: &base.Result{
								Target:    target,
								Exclusion: exclusion,
								Subjects:  directUserCollection.GetSubjects(),
							},
						},
					},
				},
			}
			return
		}

		result := expandUnion(ctx, expandFunctions)
		if result.Err != nil {
			expandChan <- expandFailResponse(result.Err)
			return
		}

		var ex []*base.Expand
		ex = append(ex, &base.Expand{
			Node: &base.Expand_Leaf{
				Leaf: &base.Result{
					Target:    target,
					Exclusion: exclusion,
					Subjects:  directUserCollection.GetSubjects(),
				},
			},
		})

		result.Response.Tree.GetExpand().Children = ex
		expandChan <- result
	}
}

// expandTupleToUserSet -
func (command *ExpandCommand) expandTupleToUserSet(ctx context.Context, request *base.PermissionExpandRequest, ttu *base.TupleToUserSet, exclusion bool) ExpandFunction {
	return func(ctx context.Context, expandChan chan<- ExpandResponse) {
		var err error

		var tupleCollection database.ITupleCollection
		tupleCollection, err = command.relationshipReader.QueryRelationships(ctx, &base.TupleFilter{
			Entity: &base.EntityFilter{
				Type: request.GetEntity().GetType(),
				Ids:  []string{request.GetEntity().GetId()},
			},
			Relation: ttu.GetTupleSet().GetRelation(),
		}, request.GetMetadata().GetSnapToken())
		if err != nil {
			expandChan <- expandFailResponse(err)
		}

		it := tupleCollection.ToSubjectCollection().CreateSubjectIterator()
		var expandFunctions []ExpandFunction
		for it.HasNext() {
			subject := it.GetNext()
			if subject.GetRelation() == tuple.ELLIPSIS {
				expandFunctions = append(expandFunctions, command.expandComputedUserSet(ctx, &base.PermissionExpandRequest{
					Entity: &base.Entity{
						Type: subject.GetType(),
						Id:   subject.GetId(),
					},
					Permission: subject.GetRelation(),
					Metadata:   request.GetMetadata(),
				}, ttu.GetComputed(), exclusion))
			}
		}

		expandChan <- expandUnion(ctx, expandFunctions)
	}
}

// expandComputedUserSet -
func (command *ExpandCommand) expandComputedUserSet(ctx context.Context, request *base.PermissionExpandRequest, cu *base.ComputedUserSet, exclusion bool) ExpandFunction {
	return func(ctx context.Context, resultChan chan<- ExpandResponse) {
		result := command.expand(ctx, &base.PermissionExpandRequest{
			Entity: &base.Entity{
				Type: request.GetEntity().GetType(),
				Id:   request.GetEntity().GetId(),
			},
			Permission: cu.GetRelation(),
			Metadata:   request.GetMetadata(),
		}, exclusion)
		resultChan <- result
	}
}

// expandOperation -
func expandOperation(
	ctx context.Context,
	functions []ExpandFunction,
	op base.ExpandTreeNode_Operation,
) ExpandResponse {
	children := make([]*base.Expand, 0, len(functions))

	if len(functions) == 0 {
		return ExpandResponse{
			Response: &base.PermissionExpandResponse{
				Tree: &base.Expand{
					Node: &base.Expand_Expand{
						Expand: &base.ExpandTreeNode{
							Operation: op,
							Children:  children,
						},
					},
				},
			},
		}
	}

	c, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()
	}()

	results := make([]chan ExpandResponse, 0, len(functions))
	for _, fn := range functions {
		fc := make(chan ExpandResponse, 1)
		results = append(results, fc)
		go fn(c, fc)
	}

	for _, result := range results {
		select {
		case resp := <-result:
			if resp.Err != nil {
				return expandFailResponse(resp.Err)
			}
			children = append(children, resp.Response.GetTree())
		case <-ctx.Done():
			return expandFailResponse(errors.New(base.ErrorCode_ERROR_CODE_CANCELLED.String()))
		}
	}

	return ExpandResponse{
		Response: &base.PermissionExpandResponse{
			Tree: &base.Expand{
				Node: &base.Expand_Expand{
					Expand: &base.ExpandTreeNode{
						Operation: op,
						Children:  children,
					},
				},
			},
		},
	}
}

// expandRoot -
func expandRoot(ctx context.Context, fn ExpandFunction) ExpandResponse {
	r := make(chan ExpandResponse, 1)
	go fn(ctx, r)
	select {
	case result := <-r:
		if result.Err == nil {
			return result
		}
		return expandFailResponse(result.Err)
	case <-ctx.Done():
		return expandFailResponse(errors.New(base.ErrorCode_ERROR_CODE_CANCELLED.String()))
	}
}

// expandUnion -
func expandUnion(ctx context.Context, functions []ExpandFunction) ExpandResponse {
	return expandOperation(ctx, functions, base.ExpandTreeNode_OPERATION_UNION)
}

// expandIntersection -
func expandIntersection(ctx context.Context, functions []ExpandFunction) ExpandResponse {
	return expandOperation(ctx, functions, base.ExpandTreeNode_OPERATION_INTERSECTION)
}

// expandFail -
func expandFail(err error) ExpandFunction {
	return func(ctx context.Context, expandChan chan<- ExpandResponse) {
		expandChan <- expandFailResponse(err)
	}
}

// expandFailResponse -
func expandFailResponse(err error) ExpandResponse {
	return ExpandResponse{
		Response: &base.PermissionExpandResponse{
			Tree: &base.Expand{},
		},
		Err: err,
	}
}
