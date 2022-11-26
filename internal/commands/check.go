package commands

import (
	"context"
	"errors"
	`google.golang.org/protobuf/types/known/wrapperspb`

	"github.com/Permify/permify/internal/keys"
	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/pkg/database"
	`github.com/Permify/permify/pkg/dsl/schema`
	"github.com/Permify/permify/pkg/logger"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	`github.com/Permify/permify/pkg/token`
	"github.com/Permify/permify/pkg/tuple"
)

// CheckCommand -
type CheckCommand struct {
	// repositories
	schemaReader       repositories.SchemaReader
	relationshipReader repositories.RelationshipReader
	// key manager
	commandKeyManager keys.CommandKeyManager
	// logger
	logger logger.Interface
}

// NewCheckCommand -
func NewCheckCommand(km keys.CommandKeyManager, sr repositories.SchemaReader, rr repositories.RelationshipReader, l logger.Interface) *CheckCommand {
	return &CheckCommand{
		schemaReader:       sr,
		commandKeyManager:  km,
		relationshipReader: rr,
		logger:             l,
	}
}

// Execute -
func (command *CheckCommand) Execute(ctx context.Context, request *base.PermissionCheckRequest) (response *base.PermissionCheckResponse, err error) {

	if request.Depth == nil {
		request.Depth = &wrapperspb.Int32Value{Value: 20}
	}

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
	en, _, err = command.schemaReader.ReadSchemaDefinition(ctx, request.GetEntity().GetType(), request.GetSchemaVersion())
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

	response, err = command.c(ctx, request, child)
	response.RemainingDepth = request.GetDepth().GetValue()
	return
}

// CheckFunction -
type CheckFunction func(ctx context.Context) (*base.PermissionCheckResponse, error)

// CheckCombiner .
type CheckCombiner func(ctx context.Context, functions []CheckFunction) (*base.PermissionCheckResponse, error)

// c -
func (command *CheckCommand) c(ctx context.Context, request *base.PermissionCheckRequest, child *base.Child) (*base.PermissionCheckResponse, error) {
	var fn CheckFunction
	switch ch := child.GetType().(type) {
	case *base.Child_Rewrite:
		fn = command.checkRewrite(ctx, request, ch.Rewrite)
	case *base.Child_Leaf:
		fn = command.checkLeaf(ctx, request, ch.Leaf)
	}

	if fn == nil {
		return &base.PermissionCheckResponse{
			Can: base.PermissionCheckResponse_RESULT_DENIED,
		}, errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_CHILD_KIND.String())
	}

	return checkUnion(ctx, []CheckFunction{fn})
}

// checkRewrite -
func (command *CheckCommand) checkRewrite(ctx context.Context, request *base.PermissionCheckRequest, rewrite *base.Rewrite) CheckFunction {
	switch rewrite.GetRewriteOperation() {
	case *base.Rewrite_OPERATION_UNION.Enum():
		return command.setChild(ctx, request, rewrite.GetChildren(), checkUnion)
	case *base.Rewrite_OPERATION_INTERSECTION.Enum():
		return command.setChild(ctx, request, rewrite.GetChildren(), checkIntersection)
	default:
		return checkFail(errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_CHILD_TYPE.String()))
	}
}

// checkLeaf -
func (command *CheckCommand) checkLeaf(ctx context.Context, request *base.PermissionCheckRequest, leaf *base.Leaf) CheckFunction {
	switch op := leaf.GetType().(type) {
	case *base.Leaf_TupleToUserSet:
		return command.checkTupleToUserSet(ctx, request, op.TupleToUserSet, leaf.GetExclusion())
	case *base.Leaf_ComputedUserSet:
		return command.checkComputedUserSet(ctx, request, op.ComputedUserSet, leaf.GetExclusion())
	default:
		return checkFail(errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_CHILD_TYPE.String()))
	}
}

// setChild -
func (command *CheckCommand) setChild(ctx context.Context, request *base.PermissionCheckRequest, children []*base.Child, combiner CheckCombiner) CheckFunction {
	var functions []CheckFunction
	for _, child := range children {
		switch child.GetType().(type) {
		case *base.Child_Rewrite:
			functions = append(functions, command.checkRewrite(ctx, request, child.GetRewrite()))
		case *base.Child_Leaf:
			functions = append(functions, command.checkLeaf(ctx, request, child.GetLeaf()))
		default:
			return checkFail(errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_CHILD_TYPE.String()))
		}
	}

	return func(ctx context.Context) (*base.PermissionCheckResponse, error) {
		return combiner(ctx, functions)
	}
}

// checkDirect -
func (command *CheckCommand) checkDirect(ctx context.Context, request *base.PermissionCheckRequest, exclusion bool) CheckFunction {
	return func(ctx context.Context) (result *base.PermissionCheckResponse, err error) {
		resp, found := command.commandKeyManager.GetCheckKey(request)
		if found {
			if exclusion {
				if resp.GetCan() == base.PermissionCheckResponse_RESULT_ALLOWED {
					return denied()
				} else {
					return allowed()
				}
			} else {
				return &base.PermissionCheckResponse{
					Can: resp.GetCan(),
				}, nil
			}
		}

		if request.Depth.Value <= 0 {
			return &base.PermissionCheckResponse{
				Can: base.PermissionCheckResponse_RESULT_DENIED,
			}, errors.New(base.ErrorCode_ERROR_CODE_DEPTH_NOT_ENOUGH.String())
		}

		var tupleCollection database.ITupleCollection
		tupleCollection, err = command.relationshipReader.QueryRelationships(ctx, &base.TupleFilter{
			Entity: &base.EntityFilter{
				Type: request.GetEntity().GetType(),
				Ids:  []string{request.GetEntity().GetId()},
			},
			Relation: request.GetPermission(),
		}, request.GetSnapToken())
		if err != nil {
			return &base.PermissionCheckResponse{
				Can: base.PermissionCheckResponse_RESULT_DENIED,
			}, err
		}

		it := tupleCollection.CreateTupleIterator()
		var checkFunctions []CheckFunction
		for it.HasNext() {
			t := it.GetNext()
			if tuple.AreSubjectsEqual(t.GetSubject(), request.GetSubject()) {
				if exclusion {
					command.commandKeyManager.SetCheckKey(request, &base.PermissionCheckResponse{
						Can: base.PermissionCheckResponse_RESULT_ALLOWED,
					})
					return denied()
				}

				result = &base.PermissionCheckResponse{
					Can: base.PermissionCheckResponse_RESULT_ALLOWED,
				}
				command.commandKeyManager.SetCheckKey(request, result)
				return result, nil
			} else {
				if !tuple.IsSubjectUser(t.GetSubject()) && t.GetSubject().GetRelation() != tuple.ELLIPSIS {
					checkFunctions = append(checkFunctions, command.checkDirect(ctx, &base.PermissionCheckRequest{
						Entity: &base.Entity{
							Type: t.GetSubject().GetType(),
							Id:   t.GetSubject().GetId(),
						},
						Permission:    t.GetSubject().GetRelation(),
						Subject:       request.GetSubject(),
						SnapToken:     request.GetSnapToken(),
						SchemaVersion: request.GetSchemaVersion(),
						Depth:         &wrapperspb.Int32Value{Value: request.Depth.Value - 1},
					}, exclusion))
				}
			}
		}

		if len(checkFunctions) > 0 {
			return checkUnion(ctx, checkFunctions)
		}

		if exclusion {
			command.commandKeyManager.SetCheckKey(request, &base.PermissionCheckResponse{
				Can: base.PermissionCheckResponse_RESULT_DENIED,
			})
			return allowed()
		}
		result = &base.PermissionCheckResponse{
			Can: base.PermissionCheckResponse_RESULT_DENIED,
		}
		command.commandKeyManager.SetCheckKey(request, result)
		return
	}
}

// checkTupleToUserSet -
func (command *CheckCommand) checkTupleToUserSet(ctx context.Context, request *base.PermissionCheckRequest, ttu *base.TupleToUserSet, exclusion bool) CheckFunction {
	return func(ctx context.Context) (*base.PermissionCheckResponse, error) {

		if request.Depth.Value <= 0 {
			return &base.PermissionCheckResponse{
				Can: base.PermissionCheckResponse_RESULT_DENIED,
			}, errors.New(base.ErrorCode_ERROR_CODE_DEPTH_NOT_ENOUGH.String())
		}

		var err error
		var tupleCollection database.ITupleCollection
		tupleCollection, err = command.relationshipReader.QueryRelationships(ctx, &base.TupleFilter{
			Entity: &base.EntityFilter{
				Type: request.GetEntity().GetType(),
				Ids:  []string{request.GetEntity().GetId()},
			},
			Relation: ttu.GetTupleSet().GetRelation(),
		}, request.GetSnapToken())
		if err != nil {
			return &base.PermissionCheckResponse{
				Can: base.PermissionCheckResponse_RESULT_DENIED,
			}, err
		}

		it := tupleCollection.ToSubjectCollection().CreateSubjectIterator()
		var checkFunctions []CheckFunction
		for it.HasNext() {
			subject := it.GetNext()
			checkFunctions = append(checkFunctions, command.checkComputedUserSet(ctx, &base.PermissionCheckRequest{
				Entity: &base.Entity{
					Type: subject.GetType(),
					Id:   subject.GetId(),
				},
				Permission:    subject.GetRelation(),
				Subject:       request.GetSubject(),
				SnapToken:     request.GetSnapToken(),
				SchemaVersion: request.GetSchemaVersion(),
				Depth:         &wrapperspb.Int32Value{Value: request.Depth.Value - 1},
			}, ttu.GetComputed(), exclusion))
		}

		return checkUnion(ctx, checkFunctions)
	}
}

// checkComputedUserSet -
func (command *CheckCommand) checkComputedUserSet(ctx context.Context, request *base.PermissionCheckRequest, cu *base.ComputedUserSet, exclusion bool) CheckFunction {
	return command.checkDirect(ctx, &base.PermissionCheckRequest{
		Entity: &base.Entity{
			Type: request.GetEntity().GetType(),
			Id:   request.GetEntity().GetId(),
		},
		Permission:    cu.GetRelation(),
		Subject:       request.GetSubject(),
		SnapToken:     request.GetSnapToken(),
		SchemaVersion: request.GetSchemaVersion(),
		Depth:         request.GetDepth(),
	}, exclusion)
}

type checkResult struct {
	resp *base.PermissionCheckResponse
	err  error
}

// checkUnion -
func checkUnion(ctx context.Context, functions []CheckFunction) (*base.PermissionCheckResponse, error) {
	if len(functions) == 0 {
		return &base.PermissionCheckResponse{
			Can: base.PermissionCheckResponse_RESULT_DENIED,
		}, nil
	}

	decisionChan := make(chan checkResult, len(functions))
	cancelCtx, cancel := context.WithCancel(ctx)

	go handler(cancelCtx, functions, decisionChan)

	defer func() {
		cancel()
	}()

	for i := 0; i < len(functions); i++ {
		select {
		case d := <-decisionChan:
			if d.err != nil {
				return nil, d.err
			}
			if d.resp.GetCan() == base.PermissionCheckResponse_RESULT_ALLOWED {
				return allowed()
			}
		case <-ctx.Done():
			return nil, errors.New(base.ErrorCode_ERROR_CODE_CANCELLED.String())
		}
	}

	return denied()
}

// checkIntersection -
func checkIntersection(ctx context.Context, functions []CheckFunction) (*base.PermissionCheckResponse, error) {
	if len(functions) == 0 {
		return denied()
	}

	decisionChan := make(chan checkResult, len(functions))
	cancelCtx, cancel := context.WithCancel(ctx)

	go handler(cancelCtx, functions, decisionChan)

	defer func() {
		cancel()
	}()

	for i := 0; i < len(functions); i++ {
		select {
		case d := <-decisionChan:
			if d.err != nil {
				return nil, d.err
			}
			if d.resp.GetCan() == base.PermissionCheckResponse_RESULT_DENIED {
				return denied()
			}
		case <-ctx.Done():
			return nil, errors.New(base.ErrorCode_ERROR_CODE_CANCELLED.String())
		}
	}

	return allowed()
}

// handler -
func handler(ctx context.Context, functions []CheckFunction, decisionChan chan<- checkResult) {
	for _, f := range functions {
		f := f
		go func() {
			decision, err := f(ctx)
			decisionChan <- checkResult{
				resp: decision,
				err:  err,
			}
			return
		}()
	}
}

// checkFail -
func checkFail(err error) CheckFunction {
	return func(ctx context.Context) (*base.PermissionCheckResponse, error) {
		return &base.PermissionCheckResponse{
			Can: base.PermissionCheckResponse_RESULT_DENIED,
		}, err
	}
}

// denied -
func denied() (*base.PermissionCheckResponse, error) {
	return &base.PermissionCheckResponse{
		Can: base.PermissionCheckResponse_RESULT_DENIED,
	}, nil
}

// allowed
func allowed() (*base.PermissionCheckResponse, error) {
	return &base.PermissionCheckResponse{
		Can: base.PermissionCheckResponse_RESULT_ALLOWED,
	}, nil
}
