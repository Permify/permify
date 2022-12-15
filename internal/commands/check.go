package commands

import (
	"context"
	"errors"
	"sync"

	"github.com/Permify/permify/internal/keys"
	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/pkg/database"
	"github.com/Permify/permify/pkg/dsl/schema"
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
	// options
	concurrencyLimit int
}

// NewCheckCommand -
func NewCheckCommand(km keys.CommandKeyManager, sr repositories.SchemaReader, rr repositories.RelationshipReader, opts ...CheckOption) *CheckCommand {
	command := &CheckCommand{
		schemaReader:       sr,
		commandKeyManager:  km,
		relationshipReader: rr,
		concurrencyLimit:   _defaultConcurrencyLimit,
	}

	// options
	for _, opt := range opts {
		opt(command)
	}

	return command
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

	meta := NewCheckMetadata()
	var res *base.PermissionCheckResponse
	res, err = command.check(ctx, request, meta)(ctx)
	return &base.PermissionCheckResponse{
		Can:            res.GetCan(),
		RemainingDepth: request.GetDepth() - meta.GetCallCount(),
	}, err
}

// CheckFunction -
type CheckFunction func(ctx context.Context) (*base.PermissionCheckResponse, error)

// CheckCombiner .
type CheckCombiner func(ctx context.Context, functions []CheckFunction, limit int) (*base.PermissionCheckResponse, error)

// checkResult -
type checkResult struct {
	resp *base.PermissionCheckResponse
	err  error
}

// check -
func (command *CheckCommand) check(ctx context.Context, request *base.PermissionCheckRequest, meta *CheckMetadata) CheckFunction {

	if meta.AddCall() >= request.GetDepth() {
		return checkFail(errors.New(base.ErrorCode_ERROR_CODE_DEPTH_NOT_ENOUGH.String()))
	}

	en, _, err := command.schemaReader.ReadSchemaDefinition(ctx, request.GetEntity().GetType(), request.GetSchemaVersion())
	if err != nil {
		return checkFail(err)
	}

	var typeOfRelation base.EntityDefinition_RelationalReference
	typeOfRelation, err = schema.GetTypeOfRelationalReferenceByNameInEntityDefinition(en, request.GetPermission())
	if err != nil {
		return checkFail(err)
	}

	var fn CheckFunction
	if typeOfRelation == base.EntityDefinition_RELATIONAL_REFERENCE_ACTION {
		var child *base.Child
		var action *base.ActionDefinition
		action, err = schema.GetActionByNameInEntityDefinition(en, request.GetPermission())
		if err != nil {
			return checkFail(err)
		}
		child = action.GetChild()
		if action.Child.GetRewrite() != nil {
			fn = command.checkRewrite(ctx, request, child.GetRewrite(), meta)
		} else {
			fn = command.checkLeaf(ctx, request, child.GetLeaf(), meta)
		}
	} else {
		fn = command.checkDirect(ctx, request, meta)
	}

	if fn == nil {
		return checkFail(errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_CHILD_KIND.String()))
	}

	return func(ctx context.Context) (*base.PermissionCheckResponse, error) {
		return checkUnion(ctx, []CheckFunction{fn}, command.concurrencyLimit)
	}
}

// checkRewrite -
func (command *CheckCommand) checkRewrite(ctx context.Context, request *base.PermissionCheckRequest, rewrite *base.Rewrite, meta *CheckMetadata) CheckFunction {
	switch rewrite.GetRewriteOperation() {
	case *base.Rewrite_OPERATION_UNION.Enum():
		return command.setChild(ctx, request, rewrite.GetChildren(), checkUnion, meta)
	case *base.Rewrite_OPERATION_INTERSECTION.Enum():
		return command.setChild(ctx, request, rewrite.GetChildren(), checkIntersection, meta)
	default:
		return checkFail(errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_CHILD_TYPE.String()))
	}
}

// checkLeaf -
func (command *CheckCommand) checkLeaf(ctx context.Context, request *base.PermissionCheckRequest, leaf *base.Leaf, meta *CheckMetadata) CheckFunction {
	request.Exclusion = leaf.GetExclusion()
	switch op := leaf.GetType().(type) {
	case *base.Leaf_TupleToUserSet:
		return command.checkTupleToUserSet(ctx, request, op.TupleToUserSet, meta)
	case *base.Leaf_ComputedUserSet:
		return command.checkComputedUserSet(ctx, request, op.ComputedUserSet, meta)
	default:
		return checkFail(errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_CHILD_TYPE.String()))
	}
}

// setChild -
func (command *CheckCommand) setChild(ctx context.Context, request *base.PermissionCheckRequest, children []*base.Child, combiner CheckCombiner, meta *CheckMetadata) CheckFunction {
	var functions []CheckFunction
	for _, child := range children {
		switch child.GetType().(type) {
		case *base.Child_Rewrite:
			functions = append(functions, command.checkRewrite(ctx, request, child.GetRewrite(), meta))
		case *base.Child_Leaf:
			functions = append(functions, command.checkLeaf(ctx, request, child.GetLeaf(), meta))
		default:
			return checkFail(errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_CHILD_TYPE.String()))
		}
	}

	return func(ctx context.Context) (*base.PermissionCheckResponse, error) {
		return combiner(ctx, functions, command.concurrencyLimit)
	}
}

// checkDirect -
func (command *CheckCommand) checkDirect(ctx context.Context, request *base.PermissionCheckRequest, meta *CheckMetadata) CheckFunction {
	return func(ctx context.Context) (result *base.PermissionCheckResponse, err error) {
		resp, found := command.commandKeyManager.GetCheckKey(request)
		if found {
			if request.GetExclusion() {
				if resp.GetCan() == base.PermissionCheckResponse_RESULT_ALLOWED {
					return denied()
				}
				return allowed()
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
					return denied()
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
					Depth:         request.GetDepth(),
				}, meta))
			}
		}

		if len(checkFunctions) > 0 {
			return checkUnion(ctx, checkFunctions, command.concurrencyLimit)
		}

		if request.GetExclusion() {
			command.commandKeyManager.SetCheckKey(request, &base.PermissionCheckResponse{
				Can:            base.PermissionCheckResponse_RESULT_DENIED,
				RemainingDepth: request.GetDepth(),
			})
			return allowed()
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
func (command *CheckCommand) checkTupleToUserSet(ctx context.Context, request *base.PermissionCheckRequest, ttu *base.TupleToUserSet, meta *CheckMetadata) CheckFunction {
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
			}, ttu.GetComputed(), meta))
		}

		return checkUnion(ctx, checkFunctions, command.concurrencyLimit)
	}
}

// checkComputedUserSet -
func (command *CheckCommand) checkComputedUserSet(ctx context.Context, request *base.PermissionCheckRequest, cu *base.ComputedUserSet, meta *CheckMetadata) CheckFunction {
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
		Depth:         request.GetDepth(),
	}, meta)
}

// checkUnion -
func checkUnion(ctx context.Context, functions []CheckFunction, limit int) (*base.PermissionCheckResponse, error) {
	if len(functions) == 0 {
		return denied()
	}

	decisionChan := make(chan checkResult, len(functions))
	cancelCtx, cancel := context.WithCancel(ctx)

	clean := run(cancelCtx, functions, decisionChan, limit)

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
					Can: base.PermissionCheckResponse_RESULT_DENIED,
				}, d.err
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
func checkIntersection(ctx context.Context, functions []CheckFunction, limit int) (*base.PermissionCheckResponse, error) {
	if len(functions) == 0 {
		return denied()
	}

	decisionChan := make(chan checkResult, len(functions))
	cancelCtx, cancel := context.WithCancel(ctx)

	clean := run(cancelCtx, functions, decisionChan, limit)

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
					Can: base.PermissionCheckResponse_RESULT_DENIED,
				}, d.err
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

// run -
func run(ctx context.Context, functions []CheckFunction, decisionChan chan<- checkResult, limit int) func() {
	cl := make(chan struct{}, limit)
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
	run:
		for _, fun := range functions {
			child := fun
			select {
			case cl <- struct{}{}:
				wg.Add(1)
				go check(child)
			case <-ctx.Done():
				break run
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
