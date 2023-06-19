package engines

import (
	"context"
	"errors"
	"sync"

	"github.com/Permify/permify/internal/invoke"
	"github.com/Permify/permify/internal/schema"
	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/pkg/database"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
)

// CheckEngine is a core component responsible for performing permission checks.
// It reads schema and relationship information, and uses the engine key manager
// to validate permission requests.
type CheckEngine struct {
	// delegate is responsible for performing permission checks
	invoker invoke.Check
	// schemaReader is responsible for reading schema information
	schemaReader storage.SchemaReader
	// relationshipReader is responsible for reading relationship information
	relationshipReader storage.RelationshipReader
	// concurrencyLimit is the maximum number of concurrent permission checks allowed
	concurrencyLimit int
}

// NewCheckEngine creates a new CheckEngine instance for performing permission checks.
// It takes a key manager, schema reader, and relationship reader as parameters.
// Additionally, it allows for optional configuration through CheckOption function arguments.
func NewCheckEngine(sr storage.SchemaReader, rr storage.RelationshipReader, opts ...CheckOption) *CheckEngine {
	// Initialize a CheckEngine with default concurrency limit and provided parameters
	engine := &CheckEngine{
		schemaReader:       sr,
		relationshipReader: rr,
		concurrencyLimit:   _defaultConcurrencyLimit,
	}

	// Apply provided options to configure the CheckEngine
	for _, opt := range opts {
		opt(engine)
	}

	return engine
}

// SetInvoker sets the delegate for the CheckEngine.
func (engine *CheckEngine) SetInvoker(invoker invoke.Check) {
	engine.invoker = invoker
}

// Check executes a permission check based on the provided request.
// The permission field in the request can either be a relation or an permission.
// This function performs various checks and returns the permission check response
// along with any errors that may have occurred.
func (engine *CheckEngine) Check(ctx context.Context, request *base.PermissionCheckRequest) (response *base.PermissionCheckResponse, err error) {
	emptyResp := denied(&base.PermissionCheckResponseMetadata{
		CheckCount: 0,
	})

	// Retrieve entity definition
	var en *base.EntityDefinition
	en, _, err = engine.schemaReader.ReadSchemaDefinition(ctx, request.GetTenantId(), request.GetEntity().GetType(), request.GetMetadata().GetSchemaVersion())
	if err != nil {
		return emptyResp, err
	}

	// Perform permission check
	var res *base.PermissionCheckResponse
	res, err = engine.check(ctx, request, en)(ctx)
	if err != nil {
		return emptyResp, err
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
func (engine *CheckEngine) invoke(ctx context.Context, request *base.PermissionCheckRequest) CheckFunction {
	return func(ctx context.Context) (*base.PermissionCheckResponse, error) {
		return engine.invoker.Check(ctx, request)
	}
}

// check is a method for the CheckEngine struct.
// It determines the type of relational reference for the permission field in the request,
// prepares and returns a CheckFunction accordingly.
func (engine *CheckEngine) check(
	ctx context.Context,
	request *base.PermissionCheckRequest,
	en *base.EntityDefinition,
) CheckFunction {
	// The GetTypeOfRelationalReferenceByNameInEntityDefinition method retrieves the type of relational reference for the permission field.
	tor, err := schema.GetTypeOfRelationalReferenceByNameInEntityDefinition(en, request.GetPermission())
	if err != nil {
		// If an error is encountered while determining the type, a CheckFunction is returned that always fails with this error.
		return checkFail(err)
	}

	var fn CheckFunction
	// Depending on the type of the relational reference, we handle the permission differently.
	if tor == base.EntityDefinition_RELATIONAL_REFERENCE_PERMISSION {
		// Get the permission by its name in the entity definition.
		permission, err := schema.GetPermissionByNameInEntityDefinition(en, request.GetPermission())
		if err != nil {
			// If an error is encountered while getting the permission, a CheckFunction is returned that always fails with this error.
			return checkFail(err)
		}
		child := permission.GetChild()

		// Depending on whether the child permission has a rewrite rule,
		// we prepare a CheckFunction to handle it accordingly.
		if child.GetRewrite() != nil {
			fn = engine.checkRewrite(ctx, request, child.GetRewrite())
		} else {
			fn = engine.checkLeaf(ctx, request, child.GetLeaf())
		}
	} else {
		// If the relational reference is not a permission, we directly check the permission.
		fn = engine.checkDirect(ctx, request)
	}

	// If we could not prepare a CheckFunction, we return a function that always fails with an error indicating undefined child kind.
	if fn == nil {
		return checkFail(errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_CHILD_KIND.String()))
	}

	// Finally, we return a function that combines results from the prepared function.
	return func(ctx context.Context) (*base.PermissionCheckResponse, error) {
		return checkUnion(ctx, []CheckFunction{fn}, engine.concurrencyLimit)
	}
}

// checkRewrite prepares a CheckFunction according to the provided Rewrite operation.
// It uses a Rewrite object that describes how to combine the results of multiple CheckFunctions.
func (engine *CheckEngine) checkRewrite(ctx context.Context, request *base.PermissionCheckRequest, rewrite *base.Rewrite) CheckFunction {
	// Switch statement depending on the Rewrite operation
	switch rewrite.GetRewriteOperation() {
	// In case of UNION operation, set the children CheckFunctions to be run concurrently
	// and return the permission if any of the CheckFunctions succeeds (union).
	case *base.Rewrite_OPERATION_UNION.Enum():
		return engine.setChild(ctx, request, rewrite.GetChildren(), checkUnion)
	// In case of INTERSECTION operation, set the children CheckFunctions to be run concurrently
	// and return the permission if all the CheckFunctions succeed (intersection).
	case *base.Rewrite_OPERATION_INTERSECTION.Enum():
		return engine.setChild(ctx, request, rewrite.GetChildren(), checkIntersection)
	// In case of EXCLUSION operation, set the children CheckFunctions to be run concurrently
	// and return the permission if the first CheckFunction succeeds and all others fail (exclusion).
	case *base.Rewrite_OPERATION_EXCLUSION.Enum():
		return engine.setChild(ctx, request, rewrite.GetChildren(), checkExclusion)
	// In case of an undefined child type, return a CheckFunction that always fails.
	default:
		return checkFail(errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_CHILD_TYPE.String()))
	}
}

// checkLeaf prepares a CheckFunction according to the provided Leaf operation.
// It uses a Leaf object that describes how to check a permission request.
func (engine *CheckEngine) checkLeaf(ctx context.Context, request *base.PermissionCheckRequest, leaf *base.Leaf) CheckFunction {
	// Switch statement depending on the Leaf type
	switch op := leaf.GetType().(type) {
	// In case of TupleToUserSet operation, prepare a CheckFunction that checks
	// if the request's user is in the UserSet referenced by the tuple.
	case *base.Leaf_TupleToUserSet:
		return engine.checkTupleToUserSet(ctx, request, op.TupleToUserSet)
	// In case of ComputedUserSet operation, prepare a CheckFunction that checks
	// if the request's user is in the computed UserSet.
	case *base.Leaf_ComputedUserSet:
		return engine.checkComputedUserSet(ctx, request, op.ComputedUserSet)
	// In case of an undefined type, return a CheckFunction that always fails.
	default:
		return checkFail(errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_CHILD_TYPE.String()))
	}
}

// setChild prepares a CheckFunction according to the provided combiner function
// and children. It uses the Child object which contains the information about the child
// nodes and can be either a Rewrite or a Leaf.
func (engine *CheckEngine) setChild(
	ctx context.Context,
	request *base.PermissionCheckRequest,
	children []*base.Child,
	combiner CheckCombiner,
) CheckFunction {
	// Create a slice to store the CheckFunctions
	var functions []CheckFunction
	// Loop over each child node
	for _, child := range children {
		// Switch on the type of the child node
		switch child.GetType().(type) {
		// In case of a Rewrite node, create a CheckFunction for the Rewrite and append it
		case *base.Child_Rewrite:
			functions = append(functions, engine.checkRewrite(ctx, request, child.GetRewrite()))
		// In case of a Leaf node, create a CheckFunction for the Leaf and append it
		case *base.Child_Leaf:
			functions = append(functions, engine.checkLeaf(ctx, request, child.GetLeaf()))
		// In case of an undefined type, return a CheckFunction that always fails
		default:
			return checkFail(errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_CHILD_TYPE.String()))
		}
	}

	// Return a function that when called, runs the appropriate combiner function
	// (union, intersection, exclusion) on the prepared CheckFunctions with the provided concurrency limit
	return func(ctx context.Context) (*base.PermissionCheckResponse, error) {
		return combiner(ctx, functions, engine.concurrencyLimit)
	}
}

// checkDirect is a method of CheckEngine struct that returns a CheckFunction.
// It's responsible for directly checking the permissions on an entity
func (engine *CheckEngine) checkDirect(ctx context.Context, request *base.PermissionCheckRequest) CheckFunction {
	// The returned CheckFunction is a closure over the provided context and request
	return func(ctx context.Context) (result *base.PermissionCheckResponse, err error) {
		// Define a TupleFilter. This specifies which tuples we're interested in.
		// We want tuples that match the entity type and ID from the request, and have a specific relation.
		filter := &base.TupleFilter{
			Entity: &base.EntityFilter{
				Type: request.GetEntity().GetType(),
				Ids:  []string{request.GetEntity().GetId()},
			},
			Relation: request.GetPermission(),
		}

		// Use the filter to query for relationships in the given context.
		// NewContextualRelationships() creates a ContextualRelationships instance from tuples in the request.
		// QueryRelationships() then uses the filter to find and return matching relationships.
		var cti *database.TupleIterator
		cti, err = storage.NewContextualTuples(request.GetContextualTuples()...).QueryRelationships(filter)
		if err != nil {
			// If an error occurred while querying, return a "denied" response and the error.
			return denied(&base.PermissionCheckResponseMetadata{}), err
		}

		// Query the relationships for the entity in the request.
		// TupleFilter helps in filtering out the relationships for a specific entity and a permission.
		var rit *database.TupleIterator
		rit, err = engine.relationshipReader.QueryRelationships(ctx, request.GetTenantId(), filter, request.GetMetadata().GetSnapToken())

		// If there's an error in querying, return a denied permission response along with the error.
		if err != nil {
			return denied(&base.PermissionCheckResponseMetadata{}), err
		}

		// Create a new UniqueTupleIterator from the two TupleIterators.
		// NewUniqueTupleIterator() ensures that the iterator only returns unique tuples.
		it := database.NewUniqueTupleIterator(rit, cti)

		// Define a slice of CheckFunctions to hold the check functions for each subject.
		var checkFunctions []CheckFunction
		// Iterate over all tuples returned by the iterator.
		for it.HasNext() {
			// Get the next tuple's subject.
			next, ok := it.GetNext()
			if !ok {
				break
			}
			subject := next.GetSubject()

			// If the subject of the tuple is the same as the subject in the request, permission is allowed.
			if tuple.AreSubjectsEqual(subject, request.GetSubject()) {
				return allowed(&base.PermissionCheckResponseMetadata{}), nil
			}
			// If the subject is not a user and the relation is not ELLIPSIS, append a check function to the list.
			if !tuple.IsDirectSubject(subject) && subject.GetRelation() != tuple.ELLIPSIS {
				checkFunctions = append(checkFunctions, engine.invoke(ctx, &base.PermissionCheckRequest{
					TenantId: request.GetTenantId(),
					Entity: &base.Entity{
						Type: subject.GetType(),
						Id:   subject.GetId(),
					},
					Permission:       subject.GetRelation(),
					Subject:          request.GetSubject(),
					Metadata:         request.GetMetadata(),
					ContextualTuples: request.GetContextualTuples(),
				}))
			}
		}

		// If there's any CheckFunction in the list, return the union of all CheckFunctions
		if len(checkFunctions) > 0 {
			return checkUnion(ctx, checkFunctions, engine.concurrencyLimit)
		}

		// If there's no CheckFunction, return a denied permission response.
		return denied(&base.PermissionCheckResponseMetadata{}), nil
	}
}

// checkTupleToUserSet is a method of CheckEngine that checks permissions using the
// TupleToUserSet data structure. It returns a CheckFunction closure that does the check.
func (engine *CheckEngine) checkTupleToUserSet(
	ctx context.Context,
	request *base.PermissionCheckRequest,
	ttu *base.TupleToUserSet,
) CheckFunction {
	// The returned CheckFunction is a closure over the provided context, request, and ttu.
	return func(ctx context.Context) (*base.PermissionCheckResponse, error) {
		// Define a TupleFilter. This specifies which tuples we're interested in.
		// We want tuples that match the entity type and ID from the request, and have a specific relation.
		filter := &base.TupleFilter{
			Entity: &base.EntityFilter{
				Type: request.GetEntity().GetType(),         // Filter by entity type from request
				Ids:  []string{request.GetEntity().GetId()}, // Filter by entity ID from request
			},
			Relation: ttu.GetTupleSet().GetRelation(), // Filter by relation from tuple set
		}

		// Use the filter to query for relationships in the given context.
		// NewContextualRelationships() creates a ContextualRelationships instance from tuples in the request.
		// QueryRelationships() then uses the filter to find and return matching relationships.
		cti, err := storage.NewContextualTuples(request.GetContextualTuples()...).QueryRelationships(filter)
		if err != nil {
			// If an error occurred while querying, return a "denied" response and the error.
			return denied(&base.PermissionCheckResponseMetadata{}), err
		}

		// Use the filter to query for relationships in the database.
		// relationshipReader.QueryRelationships() uses the filter to find and return matching relationships.
		rit, err := engine.relationshipReader.QueryRelationships(ctx, request.GetTenantId(), filter, request.GetMetadata().GetSnapToken())
		if err != nil {
			// If an error occurred while querying, return a "denied" response and the error.
			return denied(&base.PermissionCheckResponseMetadata{}), err
		}

		// Create a new UniqueTupleIterator from the two TupleIterators.
		// NewUniqueTupleIterator() ensures that the iterator only returns unique tuples.
		it := database.NewUniqueTupleIterator(rit, cti)

		// Define a slice of CheckFunctions to hold the check functions for each subject.
		var checkFunctions []CheckFunction
		// Iterate over all tuples returned by the iterator.
		for it.HasNext() {
			// Get the next tuple's subject.
			next, ok := it.GetNext()
			if !ok {
				break
			}
			subject := next.GetSubject()

			// For each subject, generate a check function for its computed user set and append it to the list.
			checkFunctions = append(checkFunctions, engine.checkComputedUserSet(ctx, &base.PermissionCheckRequest{
				TenantId: request.GetTenantId(),
				Entity: &base.Entity{
					Type: subject.GetType(),
					Id:   subject.GetId(),
				},
				Permission:       subject.GetRelation(),
				Subject:          request.GetSubject(),
				Metadata:         request.GetMetadata(),
				ContextualTuples: request.GetContextualTuples(),
			}, ttu.GetComputed()))
		}

		// Return the union of all CheckFunctions
		// If any one of the check functions allows the action, the permission is granted.
		return checkUnion(ctx, checkFunctions, engine.concurrencyLimit)
	}
}

// metadata to determine if the computed user set should be excluded from the result.
// checkComputedUserSet is a method of CheckEngine that checks permissions using the
// ComputedUserSet data structure. It returns a CheckFunction closure that performs the check.
func (engine *CheckEngine) checkComputedUserSet(
	ctx context.Context, // The context carrying deadline and cancellation signal
	request *base.PermissionCheckRequest, // The request containing details about the permission to be checked
	cu *base.ComputedUserSet, // The computed user set containing user set information
) CheckFunction {
	// The returned CheckFunction invokes a permission check with a new request that is almost the same
	// as the incoming request, but changes the Permission to be the relation defined in the computed user set.
	// This is how the check "descends" into the computed user set to check permissions there.
	return engine.invoke(ctx, &base.PermissionCheckRequest{
		TenantId:         request.GetTenantId(), // Tenant ID from the incoming request
		Entity:           request.GetEntity(),
		Permission:       cu.GetRelation(),      // Permission is set to the relation defined in the computed user set
		Subject:          request.GetSubject(),  // The subject from the incoming request
		Metadata:         request.GetMetadata(), // Metadata from the incoming request
		ContextualTuples: request.GetContextualTuples(),
	})
}

// checkUnion checks if the subject has permission by running multiple CheckFunctions concurrently,
// the permission check is successful if any one of the CheckFunctions succeeds (union).
func checkUnion(ctx context.Context, functions []CheckFunction, limit int) (*base.PermissionCheckResponse, error) {
	// Initialize the response metadata
	responseMetadata := &base.PermissionCheckResponseMetadata{}

	// If there are no functions, deny the permission and return
	if len(functions) == 0 {
		return &base.PermissionCheckResponse{
			Can:      base.PermissionCheckResponse_RESULT_DENIED,
			Metadata: responseMetadata,
		}, nil
	}

	// Create a channel to receive the results of the CheckFunctions
	decisionChan := make(chan CheckResponse, len(functions))
	// Create a context that can be cancelled
	cancelCtx, cancel := context.WithCancel(ctx)

	// Run the CheckFunctions concurrently
	clean := checkRun(cancelCtx, functions, decisionChan, limit)

	// When the function returns, ensure to cancel the context and clean up the resources
	defer func() {
		cancel()
		clean()
		close(decisionChan)
	}()

	// Iterate over the results of the CheckFunctions
	for i := 0; i < len(functions); i++ {
		select {
		// If a result is received
		case d := <-decisionChan:
			// Merge the response metadata with the received metadata
			responseMetadata = joinResponseMetas(responseMetadata, d.resp.Metadata)
			// If there was an error, deny the permission and return the error
			if d.err != nil {
				return denied(responseMetadata), d.err
			}
			// If the CheckFunction allowed the permission, allow the permission and return
			if d.resp.GetCan() == base.PermissionCheckResponse_RESULT_ALLOWED {
				return allowed(responseMetadata), nil
			}
		// If the context is done, deny the permission and return a cancellation error
		case <-ctx.Done():
			return denied(responseMetadata), errors.New(base.ErrorCode_ERROR_CODE_CANCELLED.String())
		}
	}

	// If all CheckFunctions are done and none have allowed the permission, deny the permission and return
	return denied(responseMetadata), nil
}

// checkIntersection checks if the subject has permission by running multiple CheckFunctions concurrently,
// the permission check is successful only when all CheckFunctions succeed (intersection).
func checkIntersection(ctx context.Context, functions []CheckFunction, limit int) (*base.PermissionCheckResponse, error) {
	// Initialize the response metadata
	responseMetadata := &base.PermissionCheckResponseMetadata{}

	// If there are no functions, deny the permission and return
	if len(functions) == 0 {
		return denied(responseMetadata), nil
	}

	// Create a channel to receive the results of the CheckFunctions
	decisionChan := make(chan CheckResponse, len(functions))
	// Create a context that can be cancelled
	cancelCtx, cancel := context.WithCancel(ctx)

	// Run the CheckFunctions concurrently
	clean := checkRun(cancelCtx, functions, decisionChan, limit)

	// When the function returns, ensure to cancel the context and clean up the resources
	defer func() {
		cancel()
		clean()
		close(decisionChan)
	}()

	// Iterate over the results of the CheckFunctions
	for i := 0; i < len(functions); i++ {
		select {
		// If a result is received
		case d := <-decisionChan:
			// Merge the response metadata with the received metadata
			responseMetadata = joinResponseMetas(responseMetadata, d.resp.Metadata)
			// If there was an error, deny the permission and return the error
			if d.err != nil {
				return denied(responseMetadata), d.err
			}
			// If the CheckFunction denied the permission, deny the permission and return
			if d.resp.GetCan() == base.PermissionCheckResponse_RESULT_DENIED {
				return denied(responseMetadata), nil
			}
		// If the context is done, deny the permission and return a cancellation error
		case <-ctx.Done():
			return denied(responseMetadata), errors.New(base.ErrorCode_ERROR_CODE_CANCELLED.String())
		}
	}

	// If all CheckFunctions allowed the permission, allow the permission and return
	return allowed(responseMetadata), nil
}

// checkExclusion is a function that checks if there are any exclusions for given CheckFunctions
func checkExclusion(ctx context.Context, functions []CheckFunction, limit int) (*base.PermissionCheckResponse, error) {
	// Initialize the response metadata
	responseMetadata := &base.PermissionCheckResponseMetadata{}

	// Check if there are at least 2 functions, otherwise return an error indicating that exclusion requires more than one function
	if len(functions) <= 1 {
		return denied(responseMetadata), errors.New(base.ErrorCode_ERROR_CODE_EXCLUSION_REQUIRES_MORE_THAN_ONE_FUNCTION.String())
	}

	// Initialize channels to handle the result of the first function and the remaining functions separately
	leftDecisionChan := make(chan CheckResponse, 1)
	decisionChan := make(chan CheckResponse, len(functions)-1)

	// Create a new context that can be cancelled
	cancelCtx, cancel := context.WithCancel(ctx)

	// Start the first function in a separate goroutine
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		result, err := functions[0](cancelCtx)
		leftDecisionChan <- CheckResponse{
			resp: result,
			err:  err,
		}
		wg.Done()
	}()

	// Run the remaining functions concurrently with a limit
	clean := checkRun(cancelCtx, functions[1:], decisionChan, limit-1)

	// Ensure that all resources are properly cleaned up when the function exits
	defer func() {
		cancel()
		clean()
		close(decisionChan)
		wg.Wait()
		close(leftDecisionChan)
	}()

	// Process the result from the first function
	select {
	case left := <-leftDecisionChan:
		responseMetadata = joinResponseMetas(responseMetadata, left.resp.Metadata)

		if left.err != nil {
			return denied(responseMetadata), left.err
		}

		if left.resp.GetCan() == base.PermissionCheckResponse_RESULT_DENIED {
			return denied(responseMetadata), nil
		}

	case <-ctx.Done():
		return denied(responseMetadata), errors.New(base.ErrorCode_ERROR_CODE_CANCELLED.String())
	}

	// Process the results from the remaining functions
	for i := 0; i < len(functions)-1; i++ {
		select {
		case d := <-decisionChan:
			responseMetadata = joinResponseMetas(responseMetadata, d.resp.Metadata)

			if d.err != nil {
				return denied(responseMetadata), d.err
			}

			if d.resp.GetCan() == base.PermissionCheckResponse_RESULT_ALLOWED {
				return denied(responseMetadata), nil
			}

		case <-ctx.Done():
			return denied(responseMetadata), errors.New(base.ErrorCode_ERROR_CODE_CANCELLED.String())
		}
	}

	// If none of the functions allowed the action, then it's allowed by exclusion
	return allowed(responseMetadata), nil
}

// checkRun is a function that executes a list of CheckFunctions concurrently with a specified limit.
func checkRun(ctx context.Context, functions []CheckFunction, decisionChan chan<- CheckResponse, limit int) func() {
	// Create a channel that enforces the concurrency limit
	cl := make(chan struct{}, limit)
	var wg sync.WaitGroup

	// Define a helper function that calls a CheckFunction and sends the result to the decisionChan
	check := func(child CheckFunction) {
		result, err := child(ctx)
		decisionChan <- CheckResponse{
			resp: result,
			err:  err,
		}
		// Once the CheckFunction is done, release the concurrency limit
		<-cl
		wg.Done()
	}

	// Start a goroutine that iterates over the functions
	wg.Add(1)
	go func() {
	run:
		// Iterate over the functions
		for _, fun := range functions {
			child := fun
			select {
			// If the concurrency limit allows it, start the function in a new goroutine
			case cl <- struct{}{}:
				wg.Add(1)
				go check(child)
			// If the context is done, break the loop
			case <-ctx.Done():
				break run
			}
		}
		wg.Done()
	}()

	// Return a cleanup function that waits for all goroutines to finish and then closes the concurrency limit channel
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
