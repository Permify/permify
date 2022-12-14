package commands

import (
	"context"
	"errors"
	"sync"

	"github.com/Permify/permify/internal/keys"
	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/pkg/database"
	"github.com/Permify/permify/pkg/dsl/schema"
	"github.com/Permify/permify/pkg/logger"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/token"
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
	if request.Depth == 0 {
		request.Depth = 20
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

	return command.check(ctx, request)(ctx)
}

// CheckFunction -
type CheckFunction func(ctx context.Context) (*base.PermissionCheckResponse, error)

// CheckCombiner .
type CheckCombiner func(ctx context.Context, request *base.PermissionCheckRequest, functions []CheckFunction) (*base.PermissionCheckResponse, error)

// checkResult -
type checkResult struct {
	resp *base.PermissionCheckResponse
	err  error
}

// decrease -
func decrease(request *base.PermissionCheckRequest, v int32) int32 {
	return request.Depth - v
}

// isDepthFinish -
func checkDepth(request *base.PermissionCheckRequest) bool {
	return request.GetDepth() <= 0
}

// check -
func (command *CheckCommand) check(ctx context.Context, request *base.PermissionCheckRequest) CheckFunction {

	if checkDepth(request) {
		return checkFail(request.GetDepth(), errors.New(base.ErrorCode_ERROR_CODE_DEPTH_NOT_ENOUGH.String()))
	}

	en, _, err := command.schemaReader.ReadSchemaDefinition(ctx, request.GetEntity().GetType(), request.GetSchemaVersion())
	if err != nil {
		return checkFail(request.GetDepth(), err)
	}

	var typeOfRelation base.EntityDefinition_RelationalReference
	typeOfRelation, err = schema.GetTypeOfRelationalReferenceByNameInEntityDefinition(en, request.GetPermission())
	if err != nil {
		return checkFail(request.GetDepth(), err)
	}

	var fn CheckFunction
	if typeOfRelation == base.EntityDefinition_RELATIONAL_REFERENCE_ACTION {
		var child *base.Child
		var action *base.ActionDefinition
		action, err = schema.GetActionByNameInEntityDefinition(en, request.GetPermission())
		if err != nil {
			return checkFail(request.GetDepth(), err)
		}
		child = action.GetChild()
		if action.Child.GetRewrite() != nil {
			fn = command.checkRewrite(ctx, request, child.GetRewrite())
		} else {
			fn = command.checkLeaf(ctx, request, child.GetLeaf())
		}
	} else {
		fn = command.checkDirect(ctx, request)
	}

	if fn == nil {
		return checkFail(request.GetDepth(), errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_CHILD_KIND.String()))
	}

	return func(ctx context.Context) (*base.PermissionCheckResponse, error) {
		return checkUnion(ctx, request, []CheckFunction{fn})
	}
}

// checkRewrite -
func (command *CheckCommand) checkRewrite(ctx context.Context, request *base.PermissionCheckRequest, rewrite *base.Rewrite) CheckFunction {
	switch rewrite.GetRewriteOperation() {
	case *base.Rewrite_OPERATION_UNION.Enum():
		return command.setChild(ctx, request, rewrite.GetChildren(), checkUnion)
	case *base.Rewrite_OPERATION_INTERSECTION.Enum():
		return command.setChild(ctx, request, rewrite.GetChildren(), checkIntersection)
	default:
		return checkFail(request.GetDepth(), errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_CHILD_TYPE.String()))
	}
}

// checkLeaf -
func (command *CheckCommand) checkLeaf(ctx context.Context, request *base.PermissionCheckRequest, leaf *base.Leaf) CheckFunction {
	request.Exclusion = leaf.GetExclusion()
	switch op := leaf.GetType().(type) {
	case *base.Leaf_TupleToUserSet:
		return command.checkTupleToUserSet(ctx, request, op.TupleToUserSet)
	case *base.Leaf_ComputedUserSet:
		return command.checkComputedUserSet(ctx, request, op.ComputedUserSet)
	default:
		return checkFail(request.GetDepth(), errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_CHILD_TYPE.String()))
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
			return checkFail(request.GetDepth(), errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_CHILD_TYPE.String()))
		}
	}

	return func(ctx context.Context) (*base.PermissionCheckResponse, error) {
		return combiner(ctx, request, functions)
	}
}

// checkDirect -
func (command *CheckCommand) checkDirect(ctx context.Context, request *base.PermissionCheckRequest) CheckFunction {
	return func(ctx context.Context) (result *base.PermissionCheckResponse, err error) {
		resp, found := command.commandKeyManager.GetCheckKey(request)
		if found {
			if request.GetExclusion() {
				if resp.GetCan() == base.PermissionCheckResponse_RESULT_ALLOWED {
					return denied(request.GetDepth())
				}
				return allowed(request.GetDepth())
			}
			return &base.PermissionCheckResponse{
				Can:            resp.GetCan(),
				RemainingDepth: request.GetDepth(),
			}, nil
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
				Can:            base.PermissionCheckResponse_RESULT_DENIED,
				RemainingDepth: request.GetDepth(),
			}, err
		}

		it := tupleCollection.CreateTupleIterator()
		var checkFunctions []CheckFunction
		for it.HasNext() {
			t := it.GetNext()
			if tuple.AreSubjectsEqual(t.GetSubject(), request.GetSubject()) {
				if request.GetExclusion() {
					command.commandKeyManager.SetCheckKey(request, &base.PermissionCheckResponse{
						Can:            base.PermissionCheckResponse_RESULT_ALLOWED,
						RemainingDepth: request.GetDepth(),
					})
					return denied(request.GetDepth())
				}

				result = &base.PermissionCheckResponse{
					Can:            base.PermissionCheckResponse_RESULT_ALLOWED,
					RemainingDepth: request.GetDepth(),
				}
				command.commandKeyManager.SetCheckKey(request, result)
				return result, nil
			}
			if !tuple.IsSubjectUser(t.GetSubject()) && t.GetSubject().GetRelation() != tuple.ELLIPSIS {
				checkFunctions = append(checkFunctions, command.check(ctx, &base.PermissionCheckRequest{
					Exclusion: request.GetExclusion(),
					Entity: &base.Entity{
						Type: t.GetSubject().GetType(),
						Id:   t.GetSubject().GetId(),
					},
					Permission:    t.GetSubject().GetRelation(),
					Subject:       request.GetSubject(),
					SnapToken:     request.GetSnapToken(),
					SchemaVersion: request.GetSchemaVersion(),
					Depth:         decrease(request, 1),
				}))
			}
		}

		if len(checkFunctions) > 0 {
			return checkUnion(ctx, request, checkFunctions)
		}

		if request.GetExclusion() {
			command.commandKeyManager.SetCheckKey(request, &base.PermissionCheckResponse{
				Can:            base.PermissionCheckResponse_RESULT_DENIED,
				RemainingDepth: request.GetDepth(),
			})
			return allowed(request.GetDepth())
		}
		result = &base.PermissionCheckResponse{
			Can:            base.PermissionCheckResponse_RESULT_DENIED,
			RemainingDepth: request.GetDepth(),
		}
		command.commandKeyManager.SetCheckKey(request, result)
		return
	}
}

// checkTupleToUserSet -
func (command *CheckCommand) checkTupleToUserSet(ctx context.Context, request *base.PermissionCheckRequest, ttu *base.TupleToUserSet) CheckFunction {
	return func(ctx context.Context) (*base.PermissionCheckResponse, error) {
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
				Can:            base.PermissionCheckResponse_RESULT_DENIED,
				RemainingDepth: request.GetDepth(),
			}, err
		}

		it := tupleCollection.ToSubjectCollection().CreateSubjectIterator()
		var checkFunctions []CheckFunction
		for it.HasNext() {
			subject := it.GetNext()
			checkFunctions = append(checkFunctions, command.checkComputedUserSet(ctx, &base.PermissionCheckRequest{
				Exclusion: request.GetExclusion(),
				Entity: &base.Entity{
					Type: subject.GetType(),
					Id:   subject.GetId(),
				},
				Permission:    subject.GetRelation(),
				Subject:       request.GetSubject(),
				SnapToken:     request.GetSnapToken(),
				SchemaVersion: request.GetSchemaVersion(),
				Depth:         request.GetDepth(),
			}, ttu.GetComputed()))
		}

		return checkUnion(ctx, request, checkFunctions)
	}
}

// checkComputedUserSet -
func (command *CheckCommand) checkComputedUserSet(ctx context.Context, request *base.PermissionCheckRequest, cu *base.ComputedUserSet) CheckFunction {
	return command.check(ctx, &base.PermissionCheckRequest{
		Exclusion: request.GetExclusion(),
		Entity: &base.Entity{
			Type: request.GetEntity().GetType(),
			Id:   request.GetEntity().GetId(),
		},
		Permission:    cu.GetRelation(),
		Subject:       request.GetSubject(),
		SnapToken:     request.GetSnapToken(),
		SchemaVersion: request.GetSchemaVersion(),
		Depth:         decrease(request, 1),
	})
}

// checkUnion -
func checkUnion(ctx context.Context, request *base.PermissionCheckRequest, functions []CheckFunction) (*base.PermissionCheckResponse, error) {
	if len(functions) == 0 {
		return denied(request.GetDepth())
	}

	decisionChan := make(chan checkResult, len(functions))
	cancelCtx, cancel := context.WithCancel(ctx)

	clean := handler(cancelCtx, functions, decisionChan)

	defer func() {
		cancel()
		clean()
		close(decisionChan)
	}()

	for i := 0; i < len(functions); i++ {
		select {
		case d := <-decisionChan:
			if d.err != nil {
				return &base.PermissionCheckResponse{
					Can:            base.PermissionCheckResponse_RESULT_DENIED,
					RemainingDepth: request.GetDepth(),
				}, d.err
			}
			if d.resp.GetCan() == base.PermissionCheckResponse_RESULT_ALLOWED {
				return allowed(request.GetDepth())
			}
		case <-ctx.Done():
			return nil, errors.New(base.ErrorCode_ERROR_CODE_CANCELLED.String())
		}
	}

	return denied(request.GetDepth())
}

// checkIntersection -
func checkIntersection(ctx context.Context, request *base.PermissionCheckRequest, functions []CheckFunction) (*base.PermissionCheckResponse, error) {
	if len(functions) == 0 {
		return denied(request.GetDepth())
	}

	decisionChan := make(chan checkResult, len(functions))
	cancelCtx, cancel := context.WithCancel(ctx)

	clean := handler(cancelCtx, functions, decisionChan)

	defer func() {
		cancel()
		clean()
		close(decisionChan)
	}()

	for i := 0; i < len(functions); i++ {
		select {
		case d := <-decisionChan:
			if d.err != nil {
				return &base.PermissionCheckResponse{
					Can:            base.PermissionCheckResponse_RESULT_DENIED,
					RemainingDepth: request.GetDepth(),
				}, d.err
			}
			if d.resp.GetCan() == base.PermissionCheckResponse_RESULT_DENIED {
				return denied(request.GetDepth())
			}
		case <-ctx.Done():
			return nil, errors.New(base.ErrorCode_ERROR_CODE_CANCELLED.String())
		}
	}

	return allowed(request.GetDepth())
}

// handler -
func handler(ctx context.Context, functions []CheckFunction, decisionChan chan<- checkResult) func() {
	cl := make(chan struct{}, 100)
	var wg sync.WaitGroup

	check := func(child CheckFunction) {
		result, err := child(ctx)
		decisionChan <- checkResult{
			resp: result,
			err:  err,
		}
		<-cl
		wg.Done()
	}

	wg.Add(1)
	go func() {
	handler:
		for _, fun := range functions {
			child := fun
			select {
			case cl <- struct{}{}:
				wg.Add(1)
				go check(child)
			case <-ctx.Done():
				break handler
			}
		}
		wg.Done()
	}()

	return func() {
		wg.Wait()
		close(cl)
	}
}

// checkFail -
func checkFail(depth int32, err error) CheckFunction {
	return func(ctx context.Context) (*base.PermissionCheckResponse, error) {
		return &base.PermissionCheckResponse{
			Can:            base.PermissionCheckResponse_RESULT_DENIED,
			RemainingDepth: depth,
		}, err
	}
}

// denied -
func denied(depth int32) (*base.PermissionCheckResponse, error) {
	return &base.PermissionCheckResponse{
		Can:            base.PermissionCheckResponse_RESULT_DENIED,
		RemainingDepth: depth,
	}, nil
}

// allowed
func allowed(depth int32) (*base.PermissionCheckResponse, error) {
	return &base.PermissionCheckResponse{
		Can:            base.PermissionCheckResponse_RESULT_ALLOWED,
		RemainingDepth: depth,
	}, nil
}
