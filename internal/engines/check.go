package engines

import (
	"context"
	"errors"
	"sync"

	otelCodes "go.opentelemetry.io/otel/codes"

	"github.com/Permify/permify/internal/keys"
	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/internal/schema"
	"github.com/Permify/permify/pkg/database"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/token"
	"github.com/Permify/permify/pkg/tuple"
)

// CheckEngine is a core component responsible for performing permission checks.
// It reads schema and relationship information, and uses the engine key manager
// to validate permission requests.
type CheckEngine struct {
	// schemaReader is responsible for reading schema information
	schemaReader repositories.SchemaReader
	// relationshipReader is responsible for reading relationship information
	relationshipReader repositories.RelationshipReader
	// engineKeyManager manages keys for the permission check engine
	engineKeyManager keys.EngineKeyManager
	// concurrencyLimit is the maximum number of concurrent permission checks allowed
	concurrencyLimit int
}

// NewCheckEngine creates a new CheckEngine instance for performing permission checks.
// It takes a key manager, schema reader, and relationship reader as parameters.
// Additionally, it allows for optional configuration through CheckOption function arguments.
func NewCheckEngine(km keys.EngineKeyManager, sr repositories.SchemaReader, rr repositories.RelationshipReader, opts ...CheckOption) *CheckEngine {
	// Initialize a CheckEngine with default concurrency limit and provided parameters
	engine := &CheckEngine{
		schemaReader:       sr,
		engineKeyManager:   km,
		relationshipReader: rr,
		concurrencyLimit:   _defaultConcurrencyLimit,
	}

	// Apply provided options to configure the CheckEngine
	for _, opt := range opts {
		opt(engine)
	}

	return engine
}

// Run executes a permission check based on the provided request.
// The permission field in the request can either be a relation or an permission.
// This function performs various checks and returns the permission check response
// along with any errors that may have occurred.
func (engine *CheckEngine) Run(ctx context.Context, request *base.PermissionCheckRequest) (response *base.PermissionCheckResponse, err error) {
	ctx, span := tracer.Start(ctx, "permissions.check.execute")
	defer span.End()

	emptyResp := denied(&base.PermissionCheckResponseMetadata{
		CheckCount: 0,
	})

	// Set SnapToken if not provided
	if request.GetMetadata().GetSnapToken() == "" {
		var st token.SnapToken
		st, err = engine.relationshipReader.HeadSnapshot(ctx, request.GetTenantId())
		if err != nil {
			return emptyResp, err
		}
		request.Metadata.SnapToken = st.Encode().String()
	}

	// Set SchemaVersion if not provided
	if request.GetMetadata().GetSchemaVersion() == "" {
		request.Metadata.SchemaVersion, err = engine.schemaReader.HeadVersion(ctx, request.GetTenantId())
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())
			return emptyResp, err
		}
	}

	// Validate depth of request
	err = checkDepth(request)
	if err != nil {
		return emptyResp, err
	}

	// Retrieve entity definition
	var en *base.EntityDefinition
	en, _, err = engine.schemaReader.ReadSchemaDefinition(ctx, request.GetTenantId(), request.GetEntity().GetType(), request.GetMetadata().GetSchemaVersion())
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		return emptyResp, err
	}

	// Determine type of relational reference for the permission field
	var tor base.EntityDefinition_RelationalReference
	tor, err = schema.GetTypeOfRelationalReferenceByNameInEntityDefinition(en, request.GetPermission())
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		return emptyResp, err
	}

	// If permission field is not an permission, try getting cached check result
	if tor != base.EntityDefinition_RELATIONAL_REFERENCE_PERMISSION {
		res, found := engine.engineKeyManager.GetCheckKey(request)
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

	// Perform permission check
	var res *base.PermissionCheckResponse
	res, err = engine.check(ctx, request, tor, en)(ctx)
	if err != nil {
		return emptyResp, err
	}

	// Handle caching and exclusion logic for non-permission permissions
	if tor != base.EntityDefinition_RELATIONAL_REFERENCE_PERMISSION {
		res.Metadata = increaseCheckCount(res.Metadata)
		engine.engineKeyManager.SetCheckKey(request, &base.PermissionCheckResponse{
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

// CheckFunction is a type that represents a function that takes a context
// and returns a PermissionCheckResponse along with an error. It is used
// to perform individual permission checks within the CheckEngine.
type CheckFunction func(ctx context.Context) (*base.PermissionCheckResponse, error)

// CheckCombiner is a type that represents a function which takes a context,
// a slice of CheckFunctions, and a limit. It combines the results of
// multiple CheckFunctions according to a specific strategy and returns
// a PermissionCheckResponse along with an error.
type CheckCombiner func(ctx context.Context, functions []CheckFunction, limit int) (*base.PermissionCheckResponse, error)

// run is a helper function that takes a context and a PermissionCheckRequest,
// and returns a CheckFunction. The returned CheckFunction, when called with
// a context, executes the Run method of the CheckEngine with the given
// request, and returns the resulting PermissionCheckResponse and error.
func (engine *CheckEngine) run(ctx context.Context, request *base.PermissionCheckRequest) CheckFunction {
	return func(ctx context.Context) (*base.PermissionCheckResponse, error) {
		return engine.Run(ctx, request)
	}
}

// check is a function that takes a context, a PermissionCheckRequest, a
// RelationalReference, and an EntityDefinition. It determines the appropriate
// CheckFunction based on the provided parameters and returns a wrapped
// CheckFunction. The returned CheckFunction, when called with a context,
// executes the appropriate checking method and returns the resulting
// PermissionCheckResponse and error.
func (engine *CheckEngine) check(ctx context.Context, request *base.PermissionCheckRequest, tor base.EntityDefinition_RelationalReference, en *base.EntityDefinition) CheckFunction {
	var err error
	var fn CheckFunction
	if tor == base.EntityDefinition_RELATIONAL_REFERENCE_PERMISSION {
		var child *base.Child
		var permission *base.PermissionDefinition
		permission, err = schema.GetPermissionByNameInEntityDefinition(en, request.GetPermission())
		if err != nil {
			return checkFail(err)
		}
		child = permission.GetChild()
		if child.GetRewrite() != nil {
			fn = engine.checkRewrite(ctx, request, child.GetRewrite())
		} else {
			fn = engine.checkLeaf(ctx, request, child.GetLeaf())
		}
	} else {
		fn = engine.checkDirect(ctx, request)
	}

	if fn == nil {
		return checkFail(errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_CHILD_KIND.String()))
	}

	return func(ctx context.Context) (*base.PermissionCheckResponse, error) {
		return checkUnion(ctx, []CheckFunction{fn}, engine.concurrencyLimit)
	}
}

// checkRewrite is a function that takes a context, a PermissionCheckRequest,
// and a Rewrite object. It returns a CheckFunction based on the Rewrite
// operation type (union or intersection). The returned CheckFunction, when
// called with a context, executes the appropriate rewrite operation and
// returns the resulting PermissionCheckResponse and error.
func (engine *CheckEngine) checkRewrite(ctx context.Context, request *base.PermissionCheckRequest, rewrite *base.Rewrite) CheckFunction {
	switch rewrite.GetRewriteOperation() {
	case *base.Rewrite_OPERATION_UNION.Enum():
		return engine.setChild(ctx, request, rewrite.GetChildren(), checkUnion)
	case *base.Rewrite_OPERATION_INTERSECTION.Enum():
		return engine.setChild(ctx, request, rewrite.GetChildren(), checkIntersection)
	default:
		return checkFail(errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_CHILD_TYPE.String()))
	}
}

// checkLeaf is a function that takes a context, a PermissionCheckRequest, and
// a Leaf object. It returns a CheckFunction based on the Leaf type
// (TupleToUserSet or ComputedUserSet). The returned CheckFunction, when called
// with a context, executes the appropriate leaf operation and returns the
// resulting PermissionCheckResponse and error.
func (engine *CheckEngine) checkLeaf(ctx context.Context, request *base.PermissionCheckRequest, leaf *base.Leaf) CheckFunction {
	switch op := leaf.GetType().(type) {
	case *base.Leaf_TupleToUserSet:
		return engine.checkTupleToUserSet(ctx, request, op.TupleToUserSet, leaf.GetExclusion())
	case *base.Leaf_ComputedUserSet:
		return engine.checkComputedUserSet(ctx, request, op.ComputedUserSet, leaf.GetExclusion())
	default:
		return checkFail(errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_CHILD_TYPE.String()))
	}
}

// setChild is a function that takes a context, a PermissionCheckRequest, a
// slice of Child objects, and a CheckCombiner function. It constructs a
// CheckFunction for each child based on the child type (either Rewrite or Leaf)
// and returns a new CheckFunction that, when called with a context, combines
// the results of the child functions using the provided CheckCombiner function,
// returning the resulting PermissionCheckResponse and error.
func (engine *CheckEngine) setChild(ctx context.Context, request *base.PermissionCheckRequest, children []*base.Child, combiner CheckCombiner) CheckFunction {
	var functions []CheckFunction
	for _, child := range children {
		switch child.GetType().(type) {
		case *base.Child_Rewrite:
			functions = append(functions, engine.checkRewrite(ctx, request, child.GetRewrite()))
		case *base.Child_Leaf:
			functions = append(functions, engine.checkLeaf(ctx, request, child.GetLeaf()))
		default:
			return checkFail(errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_CHILD_TYPE.String()))
		}
	}

	return func(ctx context.Context) (*base.PermissionCheckResponse, error) {
		return combiner(ctx, functions, engine.concurrencyLimit)
	}
}

// checkDirect is a function that takes a context and a PermissionCheckRequest.
// It returns a CheckFunction that, when called with a context, performs a
// direct permission check on the request. The function iterates through
// relationships and checks if the subject matches the request's subject. If a
// match is found, it returns an allowed PermissionCheckResponse. If not, it
// adds the necessary checks to a list of CheckFunctions to be combined later.
// The final result is determined by combining the check results using the
// checkUnion function.
func (engine *CheckEngine) checkDirect(ctx context.Context, request *base.PermissionCheckRequest) CheckFunction {
	return func(ctx context.Context) (result *base.PermissionCheckResponse, err error) {
		var it *database.TupleIterator
		it, err = engine.relationshipReader.QueryRelationships(ctx, request.GetTenantId(), &base.TupleFilter{
			Entity: &base.EntityFilter{
				Type: request.GetEntity().GetType(),
				Ids:  []string{request.GetEntity().GetId()},
			},
			Relation: request.GetPermission(),
		}, request.GetMetadata().GetSnapToken())
		if err != nil {
			return denied(&base.PermissionCheckResponseMetadata{}), err
		}

		var checkFunctions []CheckFunction
		for it.HasNext() {
			subject := it.GetNext().GetSubject()
			if tuple.AreSubjectsEqual(subject, request.GetSubject()) {
				result = allowed(&base.PermissionCheckResponseMetadata{})
				engine.engineKeyManager.SetCheckKey(request, result)
				return result, nil
			}
			if !tuple.IsSubjectUser(subject) && subject.GetRelation() != tuple.ELLIPSIS {
				checkFunctions = append(checkFunctions, engine.run(ctx, &base.PermissionCheckRequest{
					TenantId: request.GetTenantId(),
					Entity: &base.Entity{
						Type: subject.GetType(),
						Id:   subject.GetId(),
					},
					Permission: subject.GetRelation(),
					Subject:    request.GetSubject(),
					Metadata: &base.PermissionCheckRequestMetadata{
						SchemaVersion: request.Metadata.GetSchemaVersion(),
						Exclusion:     request.Metadata.GetExclusion(),
						SnapToken:     request.Metadata.GetSnapToken(),
						Depth:         request.Metadata.Depth - 1,
					},
				}))
			}
		}

		if len(checkFunctions) > 0 {
			return checkUnion(ctx, checkFunctions, engine.concurrencyLimit)
		}

		result = denied(&base.PermissionCheckResponseMetadata{})
		engine.engineKeyManager.SetCheckKey(request, result)
		return
	}
}

// checkTupleToUserSet is a function that takes a context, a PermissionCheckRequest,
// a TupleToUserSet object, and an exclusion flag. It returns a CheckFunction that,
// when called with a context, performs a permission check by querying relationships
// based on the TupleToUserSet. For each tuple found, it adds a check function for
// the computed user set to a list of CheckFunctions. The final result is determined
// by combining the check results using the checkUnion function.
func (engine *CheckEngine) checkTupleToUserSet(ctx context.Context, request *base.PermissionCheckRequest, ttu *base.TupleToUserSet, exclusion bool) CheckFunction {
	return func(ctx context.Context) (*base.PermissionCheckResponse, error) {
		var err error
		var it *database.TupleIterator
		it, err = engine.relationshipReader.QueryRelationships(ctx, request.GetTenantId(), &base.TupleFilter{
			Entity: &base.EntityFilter{
				Type: request.GetEntity().GetType(),
				Ids:  []string{request.GetEntity().GetId()},
			},
			Relation: ttu.GetTupleSet().GetRelation(),
		}, request.GetMetadata().GetSnapToken())
		if err != nil {
			return denied(&base.PermissionCheckResponseMetadata{}), err
		}

		var checkFunctions []CheckFunction
		for it.HasNext() {
			subject := it.GetNext().GetSubject()
			checkFunctions = append(checkFunctions, engine.checkComputedUserSet(ctx, &base.PermissionCheckRequest{
				TenantId: request.GetTenantId(),
				Entity: &base.Entity{
					Type: subject.GetType(),
					Id:   subject.GetId(),
				},
				Permission: subject.GetRelation(),
				Subject:    request.GetSubject(),
				Metadata:   request.GetMetadata(),
			}, ttu.GetComputed(), exclusion))
		}

		return checkUnion(ctx, checkFunctions, engine.concurrencyLimit)
	}
}

// checkComputedUserSet is a function that takes a context, a PermissionCheckRequest,
// a ComputedUserSet object, and an exclusion flag. It returns a CheckFunction that,
// when called with a context, performs a permission check using the ComputedUserSet
// by creating a new PermissionCheckRequest with the relation from the ComputedUserSet
// and the adjusted depth. The exclusion flag is passed through to the new request's
// metadata to determine if the computed user set should be excluded from the result.
func (engine *CheckEngine) checkComputedUserSet(ctx context.Context, request *base.PermissionCheckRequest, cu *base.ComputedUserSet, exclusion bool) CheckFunction {
	return engine.run(ctx, &base.PermissionCheckRequest{
		TenantId: request.GetTenantId(),
		Entity: &base.Entity{
			Type: request.GetEntity().GetType(),
			Id:   request.GetEntity().GetId(),
		},
		Permission: cu.GetRelation(),
		Subject:    request.GetSubject(),
		Metadata: &base.PermissionCheckRequestMetadata{
			SchemaVersion: request.Metadata.GetSchemaVersion(),
			Exclusion:     exclusion,
			SnapToken:     request.Metadata.GetSnapToken(),
			Depth:         request.Metadata.Depth - 1,
		},
	})
}

// checkUnion is a function that evaluates a set of CheckFunctions concurrently
// to determine if access should be allowed or denied based on the union of the functions' results.
// It takes a context, a slice of CheckFunctions, and a limit for concurrent execution as input parameters.
// The function returns a PermissionCheckResponse and an error.
//
// The function works as follows:
// 1. If there are no CheckFunctions provided, access is denied.
// 2. A decision channel (decisionChan) is created to collect the results of the CheckFunctions.
// 3. A new context (cancelCtx) is created to enable cancellation of CheckFunctions.
// 4. The CheckFunctions are run concurrently using the run function, passing the cancelCtx, functions, decisionChan, and limit.
// 5. After all CheckFunctions are executed, the results are processed.
//   - If an error is encountered, access is denied and the error is returned.
//   - If a response indicates access should be allowed, access is allowed and no error is returned.
//
// 6. If the context is done (e.g., due to a timeout), access is denied and a cancellation error is returned.
// 7. If none of the above conditions are met, access is denied and no error is returned.
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

// checkIntersection is a function that evaluates a set of CheckFunctions concurrently
// to determine if access should be allowed or denied based on the intersection of the functions' results.
// It takes a context, a slice of CheckFunctions, and a limit for concurrent execution as input parameters.
// The function returns a PermissionCheckResponse and an error.
//
// The function works as follows:
// 1. If there are no CheckFunctions provided, access is denied.
// 2. A decision channel (decisionChan) is created to collect the results of the CheckFunctions.
// 3. A new context (cancelCtx) is created to enable cancellation of CheckFunctions.
// 4. The CheckFunctions are run concurrently using the run function, passing the cancelCtx, functions, decisionChan, and limit.
// 5. After all CheckFunctions are executed, the results are processed.
//   - If an error is encountered, access is denied and the error is returned.
//   - If a response indicates access should be denied, access is denied and no error is returned.
//
// 6. If the context is done (e.g., due to a timeout), access is denied and a cancellation error is returned.
// 7. If none of the above conditions are met, access is allowed and no error is returned.
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

// run is a function that concurrently executes a set of CheckFunctions within a context,
// with a specified concurrency limit, and writes their results to a decision channel.
// The function returns a cleanup function that waits for all CheckFunctions to complete
// and closes the concurrency limit channel.
//
// The function works as follows:
//  1. A concurrency limit channel (cl) is created to control the number of concurrently running CheckFunctions.
//  2. A WaitGroup (wg) is created to ensure that all CheckFunctions have completed before cleanup.
//  3. A helper function (check) is defined that executes a single CheckFunction, writes its result to the decision channel,
//     releases a slot from the concurrency limit channel, and signals the WaitGroup that it has completed.
//  4. The WaitGroup is incremented and a goroutine is started to loop through the CheckFunctions.
//  5. For each CheckFunction, the function waits for a slot in the concurrency limit channel,
//     increments the WaitGroup, and starts a new goroutine to execute the check helper function.
//  6. If the context is done, the loop is broken and the outer goroutine signals the WaitGroup that it has completed.
//  7. The cleanup function returned by run waits for the WaitGroup to complete and closes the concurrency limit channel.
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

// checkFail is a helper function that returns a CheckFunction that always returns a denied PermissionCheckResponse
// with the provided error and an empty PermissionCheckResponseMetadata.
//
// The function works as follows:
//  1. The function takes an error as input parameter.
//  2. The function returns a CheckFunction that takes a context as input parameter and always returns a denied
//     PermissionCheckResponse with the provided error and an empty PermissionCheckResponseMetadata.
func checkFail(err error) CheckFunction {
	return func(ctx context.Context) (*base.PermissionCheckResponse, error) {
		return denied(&base.PermissionCheckResponseMetadata{}), err
	}
}

// denied is a helper function that returns a denied PermissionCheckResponse with the provided PermissionCheckResponseMetadata.
//
// The function works as follows:
// 1. The function takes a PermissionCheckResponseMetadata as input parameter.
// 2. The function returns a denied PermissionCheckResponse with a RESULT_DENIED Can value and the provided metadata.
func denied(meta *base.PermissionCheckResponseMetadata) *base.PermissionCheckResponse {
	return &base.PermissionCheckResponse{
		Can:      base.PermissionCheckResponse_RESULT_DENIED,
		Metadata: meta,
	}
}

// allowed is a helper function that returns an allowed PermissionCheckResponse with the provided PermissionCheckResponseMetadata.
//
// The function works as follows:
// 1. The function takes a PermissionCheckResponseMetadata as input parameter.
// 2. The function returns an allowed PermissionCheckResponse with a RESULT_ALLOWED Can value and the provided metadata.
func allowed(meta *base.PermissionCheckResponseMetadata) *base.PermissionCheckResponse {
	return &base.PermissionCheckResponse{
		Can:      base.PermissionCheckResponse_RESULT_ALLOWED,
		Metadata: meta,
	}
}
