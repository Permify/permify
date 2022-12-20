package commands

import (
	"context"
	"errors"
	"sync"

	"go.opentelemetry.io/otel"
	otelCodes "go.opentelemetry.io/otel/codes"
	"github.com/Permify/permify/internal/keys"
	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/pkg/database"
	"github.com/Permify/permify/pkg/dsl/schema"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/token"
	"github.com/Permify/permify/pkg/tuple"
)

var tracer = otel.Tracer("commands")

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
// there are two option for request's permission field.
// - relation
// - action
func (command *CheckCommand) Execute(ctx context.Context, request *base.PermissionCheckRequest) (response *base.PermissionCheckResponse, err error) {
	ctx, span := tracer.Start(ctx, "permissions.check.execute")
	defer span.End()

	emptyResp := denied(&base.PermissionCheckResponseMetadata{
		CheckCount: 0,
	})

	if request.GetMetadata().GetSnapToken() == "" {
		var st token.SnapToken
		st, err = command.relationshipReader.HeadSnapshot(ctx)
		if err != nil {
			return emptyResp, err
		}
		request.Metadata.SnapToken = st.Encode().String()
	}

	if request.GetMetadata().GetSchemaVersion() == "" {
		request.Metadata.SchemaVersion, err = command.schemaReader.HeadVersion(ctx)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())
      return emptyResp, err
		}
	}

	err = checkDepth(request)
	if err != nil {
		return emptyResp, err
	}

	var en *base.EntityDefinition
	en, _, err = command.schemaReader.ReadSchemaDefinition(ctx, request.GetEntity().GetType(), request.GetMetadata().GetSchemaVersion())
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		return emptyResp, err
	}

	var tor base.EntityDefinition_RelationalReference
	tor, err = schema.GetTypeOfRelationalReferenceByNameInEntityDefinition(en, request.GetPermission())
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		return emptyResp, err
	}

	if tor != base.EntityDefinition_RELATIONAL_REFERENCE_ACTION {
		res, found := command.commandKeyManager.GetCheckKey(request)
		if found {
			if request.GetMetadata().GetExclusion() {
				if res.GetCan() == base.PermissionCheckResponse_RESULT_ALLOWED {
					return denied(&base.PermissionCheckResponseMetadata{}), nil
				}
				return allowed(&base.PermissionCheckResponseMetadata{}), nil
			}
			return &base.PermissionCheckResponse{
				Can:      res.GetCan(),
				Metadata: &base.PermissionCheckResponseMetadata{},
			}, nil
		}
	}

	var res *base.PermissionCheckResponse
	res, err = command.check(ctx, request, tor, en)(ctx)
	if err != nil {
		return emptyResp, err
	}

	if tor != base.EntityDefinition_RELATIONAL_REFERENCE_ACTION {
		res.Metadata = increaseCheckCount(res.Metadata)
		command.commandKeyManager.SetCheckKey(request, &base.PermissionCheckResponse{
			Can:      res.GetCan(),
			Metadata: &base.PermissionCheckResponseMetadata{},
		})
		if request.GetMetadata().GetExclusion() {
			if res.GetCan() == base.PermissionCheckResponse_RESULT_ALLOWED {
				return denied(res.Metadata), nil
			}
			return allowed(res.Metadata), nil
		}
	}

	return &base.PermissionCheckResponse{
		Can:      res.Can,
		Metadata: res.Metadata,
	}, nil
}

// CheckFunction -
type CheckFunction func(ctx context.Context) (*base.PermissionCheckResponse, error)

// CheckCombiner .
type CheckCombiner func(ctx context.Context, functions []CheckFunction, limit int) (*base.PermissionCheckResponse, error)

// execute -
func (command *CheckCommand) execute(ctx context.Context, request *base.PermissionCheckRequest) CheckFunction {
	return func(ctx context.Context) (*base.PermissionCheckResponse, error) {
		return command.Execute(ctx, request)
	}
}

// check -
func (command *CheckCommand) check(ctx context.Context, request *base.PermissionCheckRequest, tor base.EntityDefinition_RelationalReference, en *base.EntityDefinition) CheckFunction {
	var err error
	var fn CheckFunction
	if tor == base.EntityDefinition_RELATIONAL_REFERENCE_ACTION {
		var child *base.Child
		var action *base.ActionDefinition
		action, err = schema.GetActionByNameInEntityDefinition(en, request.GetPermission())
		if err != nil {
			return checkFail(err)
		}
		child = action.GetChild()
		if child.GetRewrite() != nil {
			fn = command.checkRewrite(ctx, request, child.GetRewrite())
		} else {
			fn = command.checkLeaf(ctx, request, child.GetLeaf())
		}
	} else {
		fn = command.checkDirect(ctx, request)
	}

	if fn == nil {
		return checkFail(errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_CHILD_KIND.String()))
	}

	return func(ctx context.Context) (*base.PermissionCheckResponse, error) {
		return checkUnion(ctx, []CheckFunction{fn}, command.concurrencyLimit)
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
		return checkFail(errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_CHILD_TYPE.String()))
	}
}

// checkLeaf -
func (command *CheckCommand) checkLeaf(ctx context.Context, request *base.PermissionCheckRequest, leaf *base.Leaf) CheckFunction {
	request.Metadata.Exclusion = leaf.GetExclusion()
	switch op := leaf.GetType().(type) {
	case *base.Leaf_TupleToUserSet:
		return command.checkTupleToUserSet(ctx, request, op.TupleToUserSet)
	case *base.Leaf_ComputedUserSet:
		return command.checkComputedUserSet(ctx, request, op.ComputedUserSet)
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
		return combiner(ctx, functions, command.concurrencyLimit)
	}
}

// checkDirect -
func (command *CheckCommand) checkDirect(ctx context.Context, request *base.PermissionCheckRequest) CheckFunction {
	return func(ctx context.Context) (result *base.PermissionCheckResponse, err error) {
		var tupleCollection database.ITupleCollection
		tupleCollection, err = command.relationshipReader.QueryRelationships(ctx, &base.TupleFilter{
			Entity: &base.EntityFilter{
				Type: request.GetEntity().GetType(),
				Ids:  []string{request.GetEntity().GetId()},
			},
			Relation: request.GetPermission(),
		}, request.GetMetadata().GetSnapToken())
		if err != nil {
			return denied(&base.PermissionCheckResponseMetadata{}), err
		}

		it := tupleCollection.CreateTupleIterator()
		var checkFunctions []CheckFunction
		for it.HasNext() {
			t := it.GetNext()
			if tuple.AreSubjectsEqual(t.GetSubject(), request.GetSubject()) {
				result = allowed(&base.PermissionCheckResponseMetadata{})
				command.commandKeyManager.SetCheckKey(request, result)
				return result, nil
			}
			if !tuple.IsSubjectUser(t.GetSubject()) && t.GetSubject().GetRelation() != tuple.ELLIPSIS {
				checkFunctions = append(checkFunctions, command.execute(ctx, &base.PermissionCheckRequest{
					Entity: &base.Entity{
						Type: t.GetSubject().GetType(),
						Id:   t.GetSubject().GetId(),
					},
					Permission: t.GetSubject().GetRelation(),
					Subject:    request.GetSubject(),
					Metadata:   decreaseDepth(request.Metadata),
				}))
			}
		}

		if len(checkFunctions) > 0 {
			return checkUnion(ctx, checkFunctions, command.concurrencyLimit)
		}

		result = denied(&base.PermissionCheckResponseMetadata{})
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
		}, request.GetMetadata().GetSnapToken())
		if err != nil {
			return denied(&base.PermissionCheckResponseMetadata{}), err
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
				Permission: subject.GetRelation(),
				Subject:    request.GetSubject(),
				Metadata:   request.GetMetadata(),
			}, ttu.GetComputed()))
		}

		return checkUnion(ctx, checkFunctions, command.concurrencyLimit)
	}
}

// checkComputedUserSet -
func (command *CheckCommand) checkComputedUserSet(ctx context.Context, request *base.PermissionCheckRequest, cu *base.ComputedUserSet) CheckFunction {
	return command.execute(ctx, &base.PermissionCheckRequest{
		Entity: &base.Entity{
			Type: request.GetEntity().GetType(),
			Id:   request.GetEntity().GetId(),
		},
		Permission: cu.GetRelation(),
		Subject:    request.GetSubject(),
		Metadata:   decreaseDepth(request.Metadata),
	})
}

// checkUnion -
func checkUnion(ctx context.Context, functions []CheckFunction, limit int) (*base.PermissionCheckResponse, error) {
	responseMetadata := &base.PermissionCheckResponseMetadata{}

	if len(functions) == 0 {
		return &base.PermissionCheckResponse{
			Can:      base.PermissionCheckResponse_RESULT_DENIED,
			Metadata: responseMetadata,
		}, nil
	}

	decisionChan := make(chan CheckResponse, len(functions))
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
			responseMetadata = joinResponseMetas(responseMetadata, d.resp.Metadata)
			if d.err != nil {
				return denied(responseMetadata), d.err
			}
			if d.resp.GetCan() == base.PermissionCheckResponse_RESULT_ALLOWED {
				return allowed(responseMetadata), nil
			}
		case <-ctx.Done():
			return denied(responseMetadata), errors.New(base.ErrorCode_ERROR_CODE_CANCELLED.String())
		}
	}

	return denied(responseMetadata), nil
}

// checkIntersection -
func checkIntersection(ctx context.Context, functions []CheckFunction, limit int) (*base.PermissionCheckResponse, error) {
	responseMetadata := &base.PermissionCheckResponseMetadata{}

	if len(functions) == 0 {
		return denied(responseMetadata), nil
	}

	decisionChan := make(chan CheckResponse, len(functions))
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
			responseMetadata = joinResponseMetas(responseMetadata, d.resp.Metadata)
			if d.err != nil {
				return denied(responseMetadata), d.err
			}
			if d.resp.GetCan() == base.PermissionCheckResponse_RESULT_DENIED {
				return denied(responseMetadata), nil
			}
		case <-ctx.Done():
			return denied(responseMetadata), errors.New(base.ErrorCode_ERROR_CODE_CANCELLED.String())
		}
	}

	return allowed(responseMetadata), nil
}

// run -
func run(ctx context.Context, functions []CheckFunction, decisionChan chan<- CheckResponse, limit int) func() {
	cl := make(chan struct{}, limit)
	var wg sync.WaitGroup

	check := func(child CheckFunction) {
		result, err := child(ctx)
		decisionChan <- CheckResponse{
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
		return denied(&base.PermissionCheckResponseMetadata{}), err
	}
}

// denied -
func denied(meta *base.PermissionCheckResponseMetadata) *base.PermissionCheckResponse {
	return &base.PermissionCheckResponse{
		Can:      base.PermissionCheckResponse_RESULT_DENIED,
		Metadata: meta,
	}
}

// allowed
func allowed(meta *base.PermissionCheckResponseMetadata) *base.PermissionCheckResponse {
	return &base.PermissionCheckResponse{
		Can:      base.PermissionCheckResponse_RESULT_ALLOWED,
		Metadata: meta,
	}
}
