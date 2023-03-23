package engines

import (
	"context"
	"errors"

	otelCodes "go.opentelemetry.io/otel/codes"

	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/internal/schema"
	"github.com/Permify/permify/pkg/database"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/token"
	"github.com/Permify/permify/pkg/tuple"
)

// ExpandEngine - This comment is describing a type called ExpandEngine. The ExpandEngine type contains two fields: schemaReader,
// which is a repositories.SchemaReader object, and relationshipReader, which is a repositories.RelationshipReader object.
// The ExpandEngine type is used to expand permission scopes based on a given user ID and a set of permission requirements.
type ExpandEngine struct {
	schemaReader       repositories.SchemaReader
	relationshipReader repositories.RelationshipReader
}

// NewExpandEngine - This function creates a new instance of ExpandEngine by taking a SchemaReader and a RelationshipReader as
// parameters and returning a pointer to the created instance. The SchemaReader is used to read schema definitions, while the
// RelationshipReader is used to read relationship definitions.
func NewExpandEngine(sr repositories.SchemaReader, rr repositories.RelationshipReader) *ExpandEngine {
	return &ExpandEngine{
		schemaReader:       sr,
		relationshipReader: rr,
	}
}

// Run - This is the Run function of the ExpandEngine type, which takes a context, a PermissionExpandRequest,
// and returns a PermissionExpandResponse and an error.
// The function begins by starting a new OpenTelemetry span, with the name "permissions.expand.execute".
// It then checks if a snap token and schema version are included in the request. If not, it retrieves the head
// snapshot and head schema version, respectively, from the appropriate repository.
//
// Finally, the function calls the expand function of the ExpandEngine type with the context, PermissionExpandRequest,
// and false value, and returns the resulting PermissionExpandResponse and error. If there is an error, the span records
// the error and sets the status to indicate an error.
func (command *ExpandEngine) Run(ctx context.Context, request *base.PermissionExpandRequest) (response *base.PermissionExpandResponse, err error) {
	ctx, span := tracer.Start(ctx, "permissions.expand.execute")
	defer span.End()

	if request.GetMetadata().GetSnapToken() == "" {
		var st token.SnapToken
		st, err = command.relationshipReader.HeadSnapshot(ctx, request.GetTenantId())
		if err != nil {
			return response, err
		}
		request.Metadata.SnapToken = st.Encode().String()
	}

	if request.GetMetadata().GetSchemaVersion() == "" {
		request.Metadata.SchemaVersion, err = command.schemaReader.HeadVersion(ctx, request.GetTenantId())
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

// ExpandResponse is a struct that contains the response and error returned from
// the expand function in the ExpandEngine. It is used to return the response and
// error together as a single object.
type ExpandResponse struct {
	Response *base.PermissionExpandResponse
	Err      error
}

// ExpandFunction represents a function that expands the schema and relationships
// of a request and sends the response through the provided channel.
type ExpandFunction func(ctx context.Context, expandChain chan<- ExpandResponse)

// ExpandCombiner represents a function that combines the results of multiple
// ExpandFunction calls into a single ExpandResponse.
type ExpandCombiner func(ctx context.Context, functions []ExpandFunction) ExpandResponse

// expand is a helper function that determines the type of relational reference being requested in the expand request
// and selects the appropriate expand function to execute on it. It returns an ExpandResponse that contains the result of
// the selected expand function.
func (command *ExpandEngine) expand(ctx context.Context, request *base.PermissionExpandRequest, exclusion bool) ExpandResponse {
	en, _, err := command.schemaReader.ReadSchemaDefinition(ctx, request.GetTenantId(), request.GetEntity().GetType(), request.GetMetadata().GetSchemaVersion())
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

// expandRewrite is a function that returns an ExpandFunction. It takes the current context, a PermissionExpandRequest, and a Rewrite as input parameters.
// It selects the appropriate expansion function based on the given operation of the Rewrite, and returns that expansion function as an ExpandFunction.
// If the operation is not recognized, it returns an ExpandFunction that indicates an error.
func (command *ExpandEngine) expandRewrite(ctx context.Context, request *base.PermissionExpandRequest, rewrite *base.Rewrite) ExpandFunction {
	switch rewrite.GetRewriteOperation() {
	case *base.Rewrite_OPERATION_UNION.Enum():
		return command.setChild(ctx, request, rewrite.GetChildren(), expandUnion)
	case *base.Rewrite_OPERATION_INTERSECTION.Enum():
		return command.setChild(ctx, request, rewrite.GetChildren(), expandIntersection)
	default:
		return expandFail(errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_CHILD_TYPE.String()))
	}
}

// expandLeaf takes in the context, the expand request and a leaf and returns an ExpandFunction.
// It determines the type of the leaf and returns the appropriate function to expand it. If the leaf is undefined,
// it returns an error using expandFail.
func (command *ExpandEngine) expandLeaf(ctx context.Context, request *base.PermissionExpandRequest, leaf *base.Leaf) ExpandFunction {
	switch op := leaf.GetType().(type) {
	case *base.Leaf_TupleToUserSet:
		return command.expandTupleToUserSet(ctx, request, op.TupleToUserSet, leaf.GetExclusion())
	case *base.Leaf_ComputedUserSet:
		return command.expandComputedUserSet(ctx, request, op.ComputedUserSet, leaf.GetExclusion())
	default:
		return expandFail(errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_CHILD_TYPE.String()))
	}
}

// setChild is a helper function that takes a list of Child objects and combines them into a single ExpandFunction.
// It does this by iterating through the Child objects and converting each one into an ExpandFunction using either expandRewrite
// or expandLeaf, depending on the type of the Child. These individual ExpandFunctions are then combined using the provided
// combiner function to produce a single ExpandFunction. This resulting ExpandFunction can then be used to recursively expand
// a permission by evaluating the permission's children.
//
// If a Child object with an undefined type is encountered, expandFail is used to create an error response.
func (command *ExpandEngine) setChild(ctx context.Context, request *base.PermissionExpandRequest, children []*base.Child, combiner ExpandCombiner) ExpandFunction {
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

// expandDirect Expands the target permission for direct subjects, i.e., those whose subject entity has the same type as the target
// entity type and whose subject relation is not an ellipsis. It queries the relationship store for relationships that
// match the target permission, and for each matching relationship, calls Expand on the corresponding subject entity
// and relation. If there are no matching relationships, it returns a leaf node of the expand tree containing the direct
// subjects. If there are matching relationships, it computes the union of the results of calling Expand on each matching
// relationship's subject entity and relation, and attaches the resulting expand nodes as children of a union node in
// the expand tree. Finally, it returns the top-level expand node.
func (command *ExpandEngine) expandDirect(ctx context.Context, request *base.PermissionExpandRequest, exclusion bool) ExpandFunction {
	return func(ctx context.Context, expandChan chan<- ExpandResponse) {
		var err error
		var it *database.TupleIterator
		it, err = command.relationshipReader.QueryRelationships(ctx, request.GetTenantId(), &base.TupleFilter{
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

		directUserCollection := database.NewSubjectCollection()

		for it.HasNext() {
			subject := it.GetNext().GetSubject()
			if !tuple.IsSubjectUser(subject) && subject.GetRelation() != tuple.ELLIPSIS {
				expandFunctions = append(expandFunctions, func(ctx context.Context, resultChan chan<- ExpandResponse) {
					result := command.expand(ctx, &base.PermissionExpandRequest{
						TenantId: request.GetTenantId(),
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

// expandTupleToUserSet is an ExpandFunction that retrieves relationships matching the given entity and relation filter,
// and expands each relationship into a set of users that have the corresponding tuple values. If the relationship subject
// contains an ellipsis (i.e. "..."), the function will recursively expand the computed user set for that entity. The
// exclusion parameter determines whether the resulting user set should be included or excluded from the final permission set.
// The function returns an ExpandFunction that sends the expanded user set to the provided channel.
//
// Parameters:
//   - ctx: context.Context for the request
//   - request: base.PermissionExpandRequest containing the request parameters
//   - ttu: base.TupleToUserSet containing the tuple filter and computed user set
//   - exclusion: bool indicating whether to exclude or include the resulting user set in the final permission set
//
// Returns:
//   - ExpandFunction that sends the expanded user set to the provided channel
func (command *ExpandEngine) expandTupleToUserSet(ctx context.Context, request *base.PermissionExpandRequest, ttu *base.TupleToUserSet, exclusion bool) ExpandFunction {
	return func(ctx context.Context, expandChan chan<- ExpandResponse) {
		var err error

		var it *database.TupleIterator
		it, err = command.relationshipReader.QueryRelationships(ctx, request.GetTenantId(), &base.TupleFilter{
			Entity: &base.EntityFilter{
				Type: request.GetEntity().GetType(),
				Ids:  []string{request.GetEntity().GetId()},
			},
			Relation: ttu.GetTupleSet().GetRelation(),
		}, request.GetMetadata().GetSnapToken())
		if err != nil {
			expandChan <- expandFailResponse(err)
		}

		var expandFunctions []ExpandFunction
		for it.HasNext() {
			subject := it.GetNext().GetSubject()
			if subject.GetRelation() == tuple.ELLIPSIS {
				expandFunctions = append(expandFunctions, command.expandComputedUserSet(ctx, &base.PermissionExpandRequest{
					TenantId: request.GetTenantId(),
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

// expandComputedUserSet is an ExpandFunction that expands the computed user set for the given entity and relation filter.
// The function first retrieves the set of tuples that match the filter, and then expands each tuple into a set of users based
// on the values in the computed user set. The exclusion parameter determines whether the resulting user set should be included
// or excluded from the final permission set. The function returns an ExpandFunction that sends the expanded user set to the
// provided channel.
//
// Parameters:
//   - ctx: context.Context for the request
//   - request: base.PermissionExpandRequest containing the request parameters
//   - cu: base.ComputedUserSet containing the computed user set to be expanded
//   - exclusion: bool indicating whether to exclude or include the resulting user set in the final permission set
//
// Returns:
//   - ExpandFunction that sends the expanded user set to the provided channel
func (command *ExpandEngine) expandComputedUserSet(ctx context.Context, request *base.PermissionExpandRequest, cu *base.ComputedUserSet, exclusion bool) ExpandFunction {
	return func(ctx context.Context, resultChan chan<- ExpandResponse) {
		result := command.expand(ctx, &base.PermissionExpandRequest{
			TenantId: request.GetTenantId(),
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

// expandOperation is a helper function that executes multiple ExpandFunctions in parallel and combines their results into
// a single ExpandResponse containing an ExpandTreeNode with the specified operation and child nodes. The function creates a
// new context and goroutine for each ExpandFunction to allow for cancellation and concurrent execution. If any of the
// ExpandFunctions return an error, the function returns an ExpandResponse with the error. If the context is cancelled before
// all ExpandFunctions complete, the function returns an ExpandResponse with an error indicating that the operation was cancelled.
//
// Parameters:
//   - ctx: context.Context for the request
//   - functions: slice of ExpandFunctions to execute in parallel
//   - op: base.ExpandTreeNode_Operation indicating the operation to perform on the child nodes
//
// Returns:
//   - ExpandResponse containing an ExpandTreeNode with the specified operation and child nodes, or an error if any of the ExpandFunctions failed
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

// expandRoot is a helper function that executes an ExpandFunction and returns the resulting ExpandResponse. The function
// creates a goroutine for the ExpandFunction to allow for cancellation and concurrent execution. If the ExpandFunction
// returns an error, the function returns an ExpandResponse with the error. If the context is cancelled before the
// ExpandFunction completes, the function returns an ExpandResponse with an error indicating that the operation was cancelled.
//
// Parameters:
//   - ctx: context.Context for the request
//   - fn: ExpandFunction to execute
//
// Returns:
//   - ExpandResponse containing the expanded user set or an error if the ExpandFunction failed
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

// expandUnion is a helper function that executes multiple ExpandFunctions in parallel and returns an ExpandResponse containing
// the union of their expanded user sets. The function delegates to expandOperation with the UNION operation. If any of the
// ExpandFunctions return an error, the function returns an ExpandResponse with the error. If the context is cancelled before
// all ExpandFunctions complete, the function returns an ExpandResponse with an error indicating that the operation was cancelled.
//
// Parameters:
//   - ctx: context.Context for the request
//   - functions: slice of ExpandFunctions to execute in parallel
//
// Returns:
//   - ExpandResponse containing the union of the expanded user sets, or an error if any of the ExpandFunctions failed
func expandUnion(ctx context.Context, functions []ExpandFunction) ExpandResponse {
	return expandOperation(ctx, functions, base.ExpandTreeNode_OPERATION_UNION)
}

// expandIntersection is a helper function that executes multiple ExpandFunctions in parallel and returns an ExpandResponse
// containing the intersection of their expanded user sets. The function delegates to expandOperation with the INTERSECTION
// operation. If any of the ExpandFunctions return an error, the function returns an ExpandResponse with the error. If the
// context is cancelled before all ExpandFunctions complete, the function returns an ExpandResponse with an error indicating
// that the operation was cancelled.
//
// Parameters:
//   - ctx: context.Context for the request
//   - functions: slice of ExpandFunctions to execute in parallel
//
// Returns:
//   - ExpandResponse containing the intersection of the expanded user sets, or an error if any of the ExpandFunctions failed
func expandIntersection(ctx context.Context, functions []ExpandFunction) ExpandResponse {
	return expandOperation(ctx, functions, base.ExpandTreeNode_OPERATION_INTERSECTION)
}

// expandFail is a helper function that returns an ExpandFunction that immediately sends an ExpandResponse with the specified error
// to the provided channel. The resulting ExpandResponse contains an empty ExpandTreeNode and the specified error.
//
// Parameters:
//   - err: error to include in the resulting ExpandResponse
//
// Returns:
//   - ExpandFunction that sends an ExpandResponse with the specified error to the provided channel
func expandFail(err error) ExpandFunction {
	return func(ctx context.Context, expandChan chan<- ExpandResponse) {
		expandChan <- expandFailResponse(err)
	}
}

// expandFailResponse is a helper function that returns an ExpandResponse with the specified error and an empty ExpandTreeNode.
//
// Parameters:
//   - err: error to include in the resulting ExpandResponse
//
// Returns:
//   - ExpandResponse with the specified error and an empty ExpandTreeNode
func expandFailResponse(err error) ExpandResponse {
	return ExpandResponse{
		Response: &base.PermissionExpandResponse{
			Tree: &base.Expand{},
		},
		Err: err,
	}
}
