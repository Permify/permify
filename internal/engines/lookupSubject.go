package engines

import (
	"context"
	"errors"
	"sync"

	"golang.org/x/exp/slices"

	"github.com/Permify/permify/internal/schema"
	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/pkg/database"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
)

type LookupSubjectEngine struct {
	// schemaReader is responsible for reading schema information
	schemaReader storage.SchemaReader
	// relationshipReader is responsible for reading relationship information
	relationshipReader storage.RelationshipReader
	// concurrencyLimit is the maximum number of concurrent permission checks allowed
	concurrencyLimit int
}

// NewLookupSubjectEngine is a constructor function for a LookupSubjectEngine.
// It initializes a new LookupSubjectEngine with a schema reader, a relationship reader,
// and an optional set of LookupSubjectOptions.
func NewLookupSubjectEngine(sr storage.SchemaReader, rr storage.RelationshipReader, opts ...LookupSubjectOption) *LookupSubjectEngine {
	// Creating a new LookupSubjectEngine and providing it with the schema reader and relationship reader.
	// By default, the concurrency limit is set to a predefined constant _defaultConcurrencyLimit.
	engine := &LookupSubjectEngine{
		schemaReader:       sr,
		relationshipReader: rr,
		concurrencyLimit:   _defaultConcurrencyLimit,
	}

	// opts is a variadic function argument, which means it can contain zero or more LookupSubjectOptions.
	// We loop through each option in opts and apply it to the engine.
	// This is a common way in Go to allow optional configuration of a new object.
	for _, opt := range opts {
		opt(engine)
	}

	// Finally, we return the newly configured LookupSubjectEngine.
	return engine
}

// LookupSubjectFunction defines the type for a function that takes a context and
// returns a pointer to a PermissionLookupSubjectResponse and an error.
// This type is often used when you want to pass around functions with this specific signature.
type LookupSubjectFunction func(ctx context.Context) (*base.PermissionLookupSubjectResponse, error)

// LookupSubjectCombiner defines the type for a function that takes a context, a slice of LookupSubjectFunctions,
// an integer as a limit and returns a pointer to a PermissionLookupSubjectResponse and an error.
// This type is useful when you want to define a function that can execute multiple LookupSubjectFunctions in a specific way
// (like concurrently with a limit or sequentially) and combine their results into a single PermissionLookupSubjectResponse.
type LookupSubjectCombiner func(ctx context.Context, functions []LookupSubjectFunction, limit int) (*base.PermissionLookupSubjectResponse, error)

// LookupSubject is a method for the LookupSubjectEngine struct.
// It takes a context and a pointer to a PermissionLookupSubjectRequest
// and returns a pointer to a PermissionLookupSubjectResponse and an error.
func (engine *LookupSubjectEngine) LookupSubject(ctx context.Context, request *base.PermissionLookupSubjectRequest) (response *base.PermissionLookupSubjectResponse, err error) {
	// ReadSchemaDefinition method of the SchemaReader interface is used to retrieve the entity's schema definition.
	// GetTenantId, GetType and GetSchemaVersion methods are used to provide necessary arguments to ReadSchemaDefinition.
	var en *base.EntityDefinition
	en, _, err = engine.schemaReader.ReadSchemaDefinition(ctx, request.GetTenantId(), request.GetEntity().GetType(), request.GetMetadata().GetSchemaVersion())
	if err != nil {
		// If an error is encountered while reading the schema definition, return an empty response and the error.
		return lookupSubjectEmpty(), err
	}

	var res *base.PermissionLookupSubjectResponse
	// Call the lookupSubject method of the engine, which returns a function.
	// That function is then immediately called with the context to perform the actual subject lookup.
	res, err = engine.lookupSubject(ctx, request, en)(ctx)
	if err != nil {
		// If an error is encountered during the lookup, return an empty response and the error.
		return lookupSubjectEmpty(), err
	}

	// If everything went smoothly, return the lookup result and nil error.
	return res, nil
}

// lookupSubject is a method for the LookupSubjectEngine struct.
// It determines the type of relational reference for the permission field in the request,
// prepares and returns a LookupSubjectFunction accordingly.
func (engine *LookupSubjectEngine) lookupSubject(
	ctx context.Context,
	request *base.PermissionLookupSubjectRequest,
	en *base.EntityDefinition,
) LookupSubjectFunction {
	// The GetTypeOfRelationalReferenceByNameInEntityDefinition method retrieves the type of relational reference for the permission field.
	tor, err := schema.GetTypeOfRelationalReferenceByNameInEntityDefinition(en, request.GetPermission())
	if err != nil {
		// If an error is encountered while determining the type, a LookupSubjectFunction is returned that always fails with this error.
		return lookupSubjectFail(err)
	}

	var fn LookupSubjectFunction
	// Depending on the type of the relational reference, we handle the permission differently.
	if tor == base.EntityDefinition_RELATIONAL_REFERENCE_PERMISSION {
		// Get the permission by its name in the entity definition.
		permission, err := schema.GetPermissionByNameInEntityDefinition(en, request.GetPermission())
		if err != nil {
			// If an error is encountered while getting the permission, a LookupSubjectFunction is returned that always fails with this error.
			return lookupSubjectFail(err)
		}
		child := permission.GetChild()

		// Depending on whether the child permission has a rewrite rule,
		// we prepare a LookupSubjectFunction to handle it accordingly.
		if child.GetRewrite() != nil {
			fn = engine.lookupSubjectRewrite(ctx, request, child.GetRewrite())
		} else {
			fn = engine.lookupSubjectLeaf(request, child.GetLeaf())
		}
	} else {
		// If the relational reference is not a permission, we directly lookup the subject.
		return engine.lookupSubjectDirect(request)
	}

	// If we could not prepare a LookupSubjectFunction, we return a function that always fails with an error indicating undefined child kind.
	if fn == nil {
		return lookupSubjectFail(errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_CHILD_KIND.String()))
	}

	// Finally, we return a function that combines results from the prepared function.
	return func(ctx context.Context) (*base.PermissionLookupSubjectResponse, error) {
		return lookupSubjectUnion(ctx, []LookupSubjectFunction{fn}, engine.concurrencyLimit)
	}
}

// lookupSubjectRewrite is a method for the LookupSubjectEngine struct.
// It generates a LookupSubjectFunction based on the rewrite operation type present in the provided request.
func (engine *LookupSubjectEngine) lookupSubjectRewrite(
	ctx context.Context,
	request *base.PermissionLookupSubjectRequest,
	rewrite *base.Rewrite,
) LookupSubjectFunction {
	// Check the type of the rewrite operation in the request
	switch rewrite.GetRewriteOperation() {
	// If the operation type is UNION, we prepare a function using lookupSubjectUnion
	// If the request has an exclusion, we use lookupSubjectIntersection instead
	case *base.Rewrite_OPERATION_UNION.Enum():
		// Use the chosen function to set the child entities for lookup
		return engine.setChild(ctx, request, rewrite.GetChildren(), lookupSubjectUnion)

	// If the operation type is INTERSECTION, we prepare a function using lookupSubjectIntersection
	// If the request has an exclusion, we use lookupSubjectUnion instead
	case *base.Rewrite_OPERATION_INTERSECTION.Enum():

		// Use the chosen function to set the child entities for lookup
		return engine.setChild(ctx, request, rewrite.GetChildren(), lookupSubjectIntersection)

	// If the operation type is not recognized, we return a function that always fails with an error indicating undefined child type.
	case *base.Rewrite_OPERATION_EXCLUSION.Enum():

		// implement exclusion
		return engine.setChild(ctx, request, rewrite.GetChildren(), lookupSubjectExclusion)
	default:
		return lookupSubjectFail(errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_CHILD_TYPE.String()))
	}
}

// lookupSubjectLeaf is a method for the LookupSubjectEngine struct.
// It generates a LookupSubjectFunction based on the type of the leaf node in the provided request.
func (engine *LookupSubjectEngine) lookupSubjectLeaf(
	request *base.PermissionLookupSubjectRequest,
	leaf *base.Leaf,
) LookupSubjectFunction {
	// Check the type of the leaf node in the request
	switch op := leaf.GetType().(type) {
	// If the type is TupleToUserSet, we prepare a function using lookupSubjectTupleToUserSet
	case *base.Leaf_TupleToUserSet:
		return engine.lookupSubjectTupleToUserSet(request, op.TupleToUserSet)

	// If the type is ComputedUserSet, we prepare a function using lookupSubjectComputedUserSet
	case *base.Leaf_ComputedUserSet:
		return engine.lookupSubjectComputedUserSet(request, op.ComputedUserSet)

	// If the leaf type is not recognized, we return a function that always fails with an error indicating undefined child type.
	default:
		return lookupSubjectFail(errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_CHILD_TYPE.String()))
	}
}

// lookupSubjectDirect is a method of the LookupSubjectEngine struct.
// It creates a LookupSubjectFunction for a direct lookup of a subject.
func (engine *LookupSubjectEngine) lookupSubjectDirect(
	request *base.PermissionLookupSubjectRequest,
) LookupSubjectFunction {
	// The returned LookupSubjectFunction first queries relationships of the entity in the request using the request's permission.
	// Then it separates the subjects into foundedUsers and foundedUserSets depending on the relation and exclusion flag.
	// Finally, it adds the ids of all foundedUsers and foundedUserSets to the response.
	return func(ctx context.Context) (result *base.PermissionLookupSubjectResponse, err error) {
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
			return lookupSubjectEmpty(), err
		}

		// Query the relationships for the entity in the request.
		// TupleFilter helps in filtering out the relationships for a specific entity and a permission.
		var rit *database.TupleIterator
		rit, err = engine.relationshipReader.QueryRelationships(ctx, request.GetTenantId(), filter, request.GetMetadata().GetSnapToken())
		if err != nil {
			return lookupSubjectEmpty(), err
		}

		// Create a new UniqueTupleIterator from the two TupleIterators.
		// NewUniqueTupleIterator() ensures that the iterator only returns unique tuples.
		it := database.NewUniqueTupleIterator(rit, cti)

		// Initialize the collections for storing found user sets and users.
		foundedUserSets := database.NewSubjectCollection()
		foundedUsers := database.NewSubjectCollection()

		for it.HasNext() {
			// For each subject in the relationships, categorize it into foundedUsers or foundedUserSets.
			// Get the next tuple's subject.
			next, ok := it.GetNext()
			if !ok {
				break
			}
			subject := next.GetSubject()

			if tuple.AreRelationReferencesEqual(
				&base.RelationReference{
					Type:     subject.GetType(),
					Relation: subject.GetRelation(),
				}, &base.RelationReference{
					Type:     request.GetSubjectReference().GetType(),
					Relation: request.GetSubjectReference().GetRelation(),
				}) {
				foundedUsers.Add(subject)
				continue
			}

			if !tuple.IsDirectSubject(subject) && subject.GetRelation() != tuple.ELLIPSIS {
				foundedUserSets.Add(subject)
			}
		}

		// Initialize the response.
		re := &base.PermissionLookupSubjectResponse{
			SubjectIds: make([]string, 0, len(foundedUsers.GetSubjects())),
		}

		// Add the ids of all foundedUsers to the response.
		for _, s := range foundedUsers.GetSubjects() {
			re.SubjectIds = append(re.SubjectIds, s.GetId())
		}

		// Iterate over the foundedUserSets.
		fi := foundedUserSets.CreateSubjectIterator()
		for fi.HasNext() {
			// For each user set, make a LookupSubject request.
			subject := fi.GetNext()
			resp, err := engine.LookupSubject(ctx, &base.PermissionLookupSubjectRequest{
				TenantId: request.GetTenantId(),
				Entity: &base.Entity{
					Type: subject.GetType(),
					Id:   subject.GetId(),
				},
				Permission:       subject.GetRelation(),
				SubjectReference: request.GetSubjectReference(),
				Metadata:         request.GetMetadata(),
				ContextualTuples: request.GetContextualTuples(),
			})
			if err != nil {
				return lookupSubjectEmpty(), err
			}

			// Add the subject ids from the response to the final response.
			re.SubjectIds = append(re.SubjectIds, resp.GetSubjectIds()...)
		}

		// Return the final response with all subject ids and nil error.
		return re, nil
	}
}

// setChild generates a LookupSubjectFunction by applying a LookupSubjectCombiner
// to a set of child permission lookups, given a request and a list of Child objects.
func (engine *LookupSubjectEngine) setChild(
	ctx context.Context, // The context for carrying out the operation
	request *base.PermissionLookupSubjectRequest, // The request containing parameters for lookup
	children []*base.Child, // The children of a particular node in the permission schema
	combiner LookupSubjectCombiner, // A function to combine the results from multiple lookup functions
) LookupSubjectFunction {
	var functions []LookupSubjectFunction // Array of functions to store lookup functions for each child

	// Iterating over the list of child nodes
	for _, child := range children {
		// Check the type of the child node and generate the appropriate lookup function
		switch child.GetType().(type) {
		case *base.Child_Rewrite:
			// If child is of type Rewrite, generate a lookup function using lookupSubjectRewrite
			functions = append(functions, engine.lookupSubjectRewrite(ctx, request, child.GetRewrite()))
		case *base.Child_Leaf:
			// If child is of type Leaf, generate a lookup function using lookupSubjectLeaf
			functions = append(functions, engine.lookupSubjectLeaf(request, child.GetLeaf()))
		default:
			// If child type is not recognised, return a failed lookup function with error
			return lookupSubjectFail(errors.New("set child error"))
		}
	}

	// Return a function that when executed, applies the LookupSubjectCombiner on the generated lookup functions
	return func(ctx context.Context) (*base.PermissionLookupSubjectResponse, error) {
		return combiner(ctx, functions, engine.concurrencyLimit)
	}
}

// lookupSubjectComputedUserSet is a method for the LookupSubjectEngine struct.
// It creates a LookupSubjectFunction for a specific ComputedUserSet.
func (engine *LookupSubjectEngine) lookupSubjectComputedUserSet(
	request *base.PermissionLookupSubjectRequest,
	cu *base.ComputedUserSet,
) LookupSubjectFunction {
	// This function creates a new LookupSubjectFunction. This function is defined to call LookupSubject on the engine,
	// with a new PermissionLookupSubjectRequest based on the current request and the ComputedUserSet.
	return func(ctx context.Context) (*base.PermissionLookupSubjectResponse, error) {
		return engine.LookupSubject(ctx, &base.PermissionLookupSubjectRequest{
			// The tenant ID is preserved from the original request.
			TenantId: request.GetTenantId(),

			// The entity type and ID are preserved from the original request.
			Entity: &base.Entity{
				Type: request.GetEntity().GetType(),
				Id:   request.GetEntity().GetId(),
			},

			// The permission field is replaced with the relation from the ComputedUserSet.
			Permission: cu.GetRelation(),

			// The subject reference is preserved from the original request.
			SubjectReference: request.GetSubjectReference(),

			// The metadata is preserved from the original request.
			Metadata: request.GetMetadata(),

			// The contextual tuples are preserved from the original request.
			ContextualTuples: request.GetContextualTuples(),
		})
	}
}

// lookupSubjectTupleToUserSet is a method of the LookupSubjectEngine struct.
// It creates a LookupSubjectFunction for a specific TupleToUserSet.
func (engine *LookupSubjectEngine) lookupSubjectTupleToUserSet(
	request *base.PermissionLookupSubjectRequest,
	ttu *base.TupleToUserSet,
) LookupSubjectFunction {
	// The returned LookupSubjectFunction first queries relationships of the entity in the request using the relation from the TupleToUserSet.
	// For each subject in the relationships, it generates a LookupSubjectFunction by treating it as a ComputedUserSet.
	// Finally, it combines all these functions into a single response.
	return func(ctx context.Context) (*base.PermissionLookupSubjectResponse, error) {
		// Define a TupleFilter. This specifies which tuples we're interested in.
		// We want tuples that match the entity type and ID from the request, and have a specific relation.
		filter := &base.TupleFilter{
			Entity: &base.EntityFilter{
				Type: request.GetEntity().GetType(),
				Ids:  []string{request.GetEntity().GetId()},
			},
			Relation: ttu.GetTupleSet().GetRelation(),
		}

		// Use the filter to query for relationships in the given context.
		// NewContextualRelationships() creates a ContextualRelationships instance from tuples in the request.
		// QueryRelationships() then uses the filter to find and return matching relationships.
		cti, err := storage.NewContextualTuples(request.GetContextualTuples()...).QueryRelationships(filter)
		if err != nil {
			// If an error occurs during querying, an empty response with the error is returned.
			return lookupSubjectEmpty(), err
		}

		// Query the relationships for the entity in the request.
		// TupleFilter helps in filtering out the relationships for a specific entity and a permission.
		var rit *database.TupleIterator
		rit, err = engine.relationshipReader.QueryRelationships(ctx, request.GetTenantId(), filter, request.GetMetadata().GetSnapToken())
		if err != nil {
			// If an error occurs during querying, an empty response with the error is returned.
			return lookupSubjectEmpty(), err
		}

		// Create a new UniqueTupleIterator from the two TupleIterators.
		// NewUniqueTupleIterator() ensures that the iterator only returns unique tuples.
		it := database.NewUniqueTupleIterator(rit, cti)

		// Initialize the slice for storing LookupSubjectFunction instances.
		var lookupSubjectFunctions []LookupSubjectFunction
		for it.HasNext() {
			// For each subject in the relationships, create a LookupSubjectFunction by treating it as a ComputedUserSet.
			// Get the next tuple's subject.
			next, ok := it.GetNext()
			if !ok {
				break
			}
			subject := next.GetSubject()

			lookupSubjectFunctions = append(lookupSubjectFunctions, engine.lookupSubjectComputedUserSet(&base.PermissionLookupSubjectRequest{
				TenantId: request.GetTenantId(),
				Entity: &base.Entity{
					Type: subject.GetType(),
					Id:   subject.GetId(),
				},
				Permission:       subject.GetRelation(),
				SubjectReference: request.GetSubjectReference(),
				Metadata:         request.GetMetadata(),
				ContextualTuples: request.GetContextualTuples(),
			}, ttu.GetComputed()))
		}

		// Combine all the lookupSubjectFunctions into a single response using the lookupSubjectUnion method.
		return lookupSubjectUnion(ctx, lookupSubjectFunctions, engine.concurrencyLimit)
	}
}

// lookupSubjectUnion function is used to find the union of subjects
// returned by executing multiple lookup subject functions concurrently.
func lookupSubjectUnion(ctx context.Context, functions []LookupSubjectFunction, limit int) (*base.PermissionLookupSubjectResponse, error) {
	// If there are no functions to be executed, return an empty response
	if len(functions) == 0 {
		return lookupSubjectEmpty(), nil
	}

	// Create a buffered channel to collect the results of the concurrently executing functions
	decisionChan := make(chan LookupSubjectResponse, len(functions))

	// Create a context that can be cancelled
	cancelCtx, cancel := context.WithCancel(ctx)

	// Execute the functions concurrently, passing the cancel context and decision channel
	clean := lookupSubjectsRun(cancelCtx, functions, decisionChan, limit)

	// Ensure resources are cleaned up correctly after functions are done executing
	defer func() {
		// Cancel the context
		cancel()
		// Clean up any remaining resources
		clean()
		// Close the decision channel
		close(decisionChan)
	}()

	// Initialize the response which will hold the union of subjects
	res := &base.PermissionLookupSubjectResponse{}

	// For each function that was executed, collect its result from the decision channel
	for i := 0; i < len(functions); i++ {
		select {
		case d := <-decisionChan:
			// If an error occurred executing the function, return the error
			if d.err != nil {
				return lookupSubjectEmpty(), d.err
			}
			// If no error occurred, append the unique subject ids from the function result to the response
			for _, id := range d.resp.GetSubjectIds() {
				if !slices.Contains(res.GetSubjectIds(), id) {
					res.SubjectIds = append(res.SubjectIds, id)
				}
			}
		// If the context is cancelled before all results are collected, return an error
		case <-ctx.Done():
			return lookupSubjectEmpty(), errors.New(base.ErrorCode_ERROR_CODE_CANCELLED.String())
		}
	}

	// Return the response containing the union of subject IDs
	return res, nil
}

// lookupSubjectIntersection function is used to find the intersection of subjects
// returned by executing multiple lookup subject functions concurrently.
func lookupSubjectIntersection(ctx context.Context, functions []LookupSubjectFunction, limit int) (*base.PermissionLookupSubjectResponse, error) {
	// If there are no functions to be executed, return an empty response
	if len(functions) == 0 {
		return lookupSubjectEmpty(), nil
	}

	// Create a buffered channel to collect the results of the concurrently executing functions
	decisionChan := make(chan LookupSubjectResponse, len(functions))

	// Create a context that can be cancelled
	cancelCtx, cancel := context.WithCancel(ctx)

	// Execute the functions concurrently, passing the cancel context and decision channel
	clean := lookupSubjectsRun(cancelCtx, functions, decisionChan, limit)

	// Ensure resources are cleaned up correctly after functions are done executing
	defer func() {
		// Cancel the context
		cancel()
		// Clean up any remaining resources
		clean()
		// Close the decision channel
		close(decisionChan)
	}()

	// ids is a slice to collect the subject ids returned by the functions
	var ids []string
	res := &base.PermissionLookupSubjectResponse{}

	// For each function that was executed, collect its result from the decision channel
	for i := 0; i < len(functions); i++ {
		select {
		case d := <-decisionChan:
			// If an error occurred executing the function, return the error
			if d.err != nil {
				return lookupSubjectEmpty(), d.err
			}
			// If no error occurred, append the subject ids from the function result to the ids slice
			for _, id := range d.resp.GetSubjectIds() {
				if !slices.Contains(res.GetSubjectIds(), id) {
					ids = append(ids, id)
				}
			}
		// If the context is cancelled before all results are collected, return an error
		case <-ctx.Done():
			return lookupSubjectEmpty(), errors.New(base.ErrorCode_ERROR_CODE_CANCELLED.String())
		}
	}

	// Filter the ids to only include duplicates (the intersection)
	duplicates := getDuplicates(ids)

	// Set the subject ids in the response to the duplicates and return the response
	res.SubjectIds = duplicates
	return res, nil
}

// lookupSubjectExclusion is a function that checks for a subject's exclusion
// among different lookup subject functions and returns the subjects not found
// in the excluded list.
func lookupSubjectExclusion(ctx context.Context, functions []LookupSubjectFunction, limit int) (*base.PermissionLookupSubjectResponse, error) {
	// If there are not more than one lookup functions, it returns an error because
	// exclusion requires at least two functions for comparison.
	if len(functions) <= 1 {
		return lookupSubjectEmpty(), errors.New(base.ErrorCode_ERROR_CODE_EXCLUSION_REQUIRES_MORE_THAN_ONE_FUNCTION.String())
	}

	// Create channels to handle asynchronous responses from lookup functions.
	leftDecisionChan := make(chan LookupSubjectResponse, 1)
	decisionChan := make(chan LookupSubjectResponse, len(functions)-1)

	// Create a cancellable context to be able to stop the function execution prematurely.
	cancelCtx, cancel := context.WithCancel(ctx)

	// Use a WaitGroup to ensure all goroutines have completed.
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		// Run the first lookup function in a goroutine.
		result, err := functions[0](cancelCtx)
		// Send the result to the channel.
		leftDecisionChan <- LookupSubjectResponse{
			resp: result,
			err:  err,
		}
		wg.Done()
	}()

	// Run the rest of the lookup functions in parallel.
	clean := lookupSubjectsRun(cancelCtx, functions[1:], decisionChan, limit-1)

	// Defer a function to cancel the context, clean up the resources, close the channels, and wait for all goroutines to finish.
	defer func() {
		cancel()
		clean()
		close(decisionChan)
		wg.Wait()
		close(leftDecisionChan)
	}()

	var leftIds []string

	// Retrieve the result of the first lookup function.
	select {
	case left := <-leftDecisionChan:
		// If there's an error, return it.
		if left.err != nil {
			return lookupSubjectEmpty(), left.err
		}
		// Otherwise, get the list of subject IDs.
		leftIds = left.resp.GetSubjectIds()

	// If the context is cancelled, return a cancellation error.
	case <-ctx.Done():
		return lookupSubjectEmpty(), errors.New(base.ErrorCode_ERROR_CODE_CANCELLED.String())
	}

	// Initialize the response.
	res := &base.PermissionLookupSubjectResponse{}

	var exIds []string

	// Retrieve the results of the remaining lookup functions.
	for i := 0; i < len(functions)-1; i++ {
		select {
		case d := <-decisionChan:
			// If there's an error, return it.
			if d.err != nil {
				return lookupSubjectEmpty(), d.err
			}
			// Otherwise, append the IDs to the list of exclusion IDs.
			exIds = append(exIds, d.resp.GetSubjectIds()...)
		// If the context is cancelled, return a cancellation error.
		case <-ctx.Done():
			return lookupSubjectEmpty(), errors.New(base.ErrorCode_ERROR_CODE_CANCELLED.String())
		}
	}

	// For each ID from the first lookup function, if it's not in the list of exclusion IDs,
	// add it to the response.
	for _, id := range leftIds {
		if !slices.Contains(exIds, id) {
			res.SubjectIds = append(res.SubjectIds, id)
		}
	}

	// Return the response and no error.
	return res, nil
}

// lookupSubjectsRun executes a list of LookupSubjectFunction concurrently, with a limit on the maximum number of concurrent executions.
// It sends the responses from these functions to the provided decisionChan.
// It returns a function that blocks until all LookupSubjectFunctions have completed.
func lookupSubjectsRun(ctx context.Context, functions []LookupSubjectFunction, decisionChan chan<- LookupSubjectResponse, limit int) func() {
	// cl is a channel used to control the concurrency level.
	// Its capacity is set to the limit argument to restrict the maximum number of concurrent goroutines.
	cl := make(chan struct{}, limit)

	// wg is a WaitGroup that is used to wait until all LookupSubjectFunctions have completed.
	var wg sync.WaitGroup

	// lookupSubject is a function that executes a LookupSubjectFunction and sends its result to the decisionChan.
	lookupSubject := func(child LookupSubjectFunction) {
		result, err := child(ctx)
		decisionChan <- LookupSubjectResponse{
			resp: result,
			err:  err,
		}
		// Release a spot in the concurrency limit channel and notify the WaitGroup that a function has completed.
		<-cl
		wg.Done()
	}

	// Start a goroutine that loops through all the LookupSubjectFunctions and starts a new goroutine for each one, up to the concurrency limit.
	wg.Add(1)
	go func() {
	run:
		for _, fun := range functions {
			child := fun
			select {
			// If there is room in the concurrency limit channel, start a new goroutine.
			case cl <- struct{}{}:
				wg.Add(1)
				go lookupSubject(child)
			// If the context is done, break out of the loop.
			case <-ctx.Done():
				break run
			}
		}
		// Notify the WaitGroup that this function is done.
		wg.Done()
	}()

	// Return a function that waits until all LookupSubjectFunctions have completed, then closes the concurrency limit channel.
	return func() {
		wg.Wait()
		close(cl)
	}
}

// lookupSubjectFail returns a function conforming to the LookupSubjectFunction signature.
// The returned function, when invoked, will always return an error that is provided as an argument to lookupSubjectFail.
// The LookupSubjectFunction could be used as a mock function in unit tests to simulate failure scenarios.
func lookupSubjectFail(err error) LookupSubjectFunction {
	return func(ctx context.Context) (*base.PermissionLookupSubjectResponse, error) {
		// We return a default PermissionLookupSubjectResponse with an empty slice of subject IDs
		// and the error that was passed into the lookupSubjectFail function.
		return &base.PermissionLookupSubjectResponse{
			SubjectIds: []string{},
		}, err
	}
}

// empty is a helper function that returns a pointer to a PermissionLookupSubjectResponse
// with an empty SubjectIds slice.
// This function could be used in tests where a default or "empty" PermissionLookupSubjectResponse is needed.
func lookupSubjectEmpty() *base.PermissionLookupSubjectResponse {
	return &base.PermissionLookupSubjectResponse{
		SubjectIds: []string{},
	}
}
