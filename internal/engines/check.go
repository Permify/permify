package engines

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/google/cel-go/cel"

	"github.com/Permify/permify/internal/invoke"
	"github.com/Permify/permify/internal/schema"
	"github.com/Permify/permify/internal/storage"
	storageContext "github.com/Permify/permify/internal/storage/context"
	"github.com/Permify/permify/pkg/database"
	"github.com/Permify/permify/pkg/dsl/utils"
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
	dataReader storage.DataReader
	// concurrencyLimit is the maximum number of concurrent permission checks allowed
	concurrencyLimit int
	// maxBatchSize is the maximum number of entity IDs per batch SQL query (IN clause)
	maxBatchSize int
}

// NewCheckEngine creates a new CheckEngine instance for performing permission checks.
// It takes a key manager, schema reader, and relationship reader as parameters.
// Additionally, it allows for optional configuration through CheckOption function arguments.
func NewCheckEngine(sr storage.SchemaReader, rr storage.DataReader, opts ...CheckOption) *CheckEngine {
	// Initialize a CheckEngine with default concurrency limit and provided parameters
	engine := &CheckEngine{
		schemaReader:     sr,
		dataReader:       rr,
		concurrencyLimit: _defaultConcurrencyLimit,
		maxBatchSize:     _defaultMaxBatchSize,
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
// Supports batch: request.EntityIDs may contain one or more entity IDs.
// Uses IN (...) queries for efficient batch processing. Returns per-entity results.
func (engine *CheckEngine) Check(ctx context.Context, request *invoke.BatchCheckRequest) (response *invoke.BatchCheckResponse, err error) {
	deniedResp := invoke.NewBatchCheckResponse(base.CheckResult_CHECK_RESULT_DENIED, request.EntityIDs...)

	// Read entity definition once (all entities share the same type)
	var en *base.EntityDefinition
	en, _, err = engine.schemaReader.ReadEntityDefinition(ctx, request.TenantID, request.EntityType, request.Metadata.GetSchemaVersion())
	if err != nil {
		return deniedResp, err
	}

	return engine.check(ctx, request, en)(ctx)
}

// CheckFunction is a type that represents a function that takes a context
// and returns a BatchCheckResponse with per-entity results along with an error.
// It is used to perform permission checks within the CheckEngine.
type CheckFunction func(ctx context.Context) (*invoke.BatchCheckResponse, error)

// CheckCombiner is a type that represents a function which takes a context
// and a slice of CheckFunctions. It combines the per-entity results of
// multiple CheckFunctions according to a specific strategy and returns
// a BatchCheckResponse along with an error.
// expectedEntityCount is the total number of unique entity IDs expected across
// all functions. It is used by checkUnion to gate early exit correctly when
// functions operate on disjoint entity sets.
// Concurrency is controlled by a request-scoped semaphore stored in the context.
type CheckCombiner func(ctx context.Context, functions []CheckFunction, limit int, expectedEntityCount int) (*invoke.BatchCheckResponse, error)

// invoke creates a CheckFunction that invokes a batch check through the full invoke chain
// (DirectInvoker -> Cache -> CheckEngine), ensuring depth tracking, caching, and tracing.
func (engine *CheckEngine) invoke(request *invoke.BatchCheckRequest) CheckFunction {
	return func(ctx context.Context) (*invoke.BatchCheckResponse, error) {
		return engine.invoker.Check(ctx, request)
	}
}

// check constructs a CheckFunction that performs permission checks based on the type of reference in the entity definition.
// All entity IDs in the batch request share the same type and permission, enabling batch SQL queries.
func (engine *CheckEngine) check(
	ctx context.Context,
	request *invoke.BatchCheckRequest,
	en *base.EntityDefinition,
) CheckFunction {
	// Identity check: for each entity, if it matches the subject, mark it as ALLOWED.
	// Collect remaining entities that need further checking.
	identityResults := make(map[string]base.CheckResult)
	var remainingIDs []string
	for _, id := range request.EntityIDs {
		if tuple.AreQueryAndSubjectEqual(&base.Entity{Type: request.EntityType, Id: id}, request.Permission, request.Subject) {
			identityResults[id] = base.CheckResult_CHECK_RESULT_ALLOWED
		} else {
			remainingIDs = append(remainingIDs, id)
		}
	}

	// If all entities matched via identity, return immediately.
	if len(remainingIDs) == 0 {
		return func(ctx context.Context) (*invoke.BatchCheckResponse, error) {
			return &invoke.BatchCheckResponse{
				Results:  identityResults,
				Metadata: emptyResponseMetadata(),
			}, nil
		}
	}

	// Build a sub-request for the remaining entities.
	subRequest := request
	if len(identityResults) > 0 {
		subRequest = request.Clone()
		subRequest.EntityIDs = remainingIDs
	}

	var fn CheckFunction

	// Determine the type of the reference by name in the given entity definition.
	tor, _ := schema.GetTypeOfReferenceByNameInEntityDefinition(en, request.Permission)

	// Based on the type of the reference, define the CheckFunction in different ways.
	switch tor {
	case base.EntityDefinition_REFERENCE_PERMISSION:
		// Get the permission from the entity definition.
		permission, err := schema.GetPermissionByNameInEntityDefinition(en, request.Permission)
		if err != nil {
			// If an error is encountered while getting the permission, a CheckFunction is returned that always fails with this error.
			return checkFail(err)
		}
		// Get the child of the permission.
		child := permission.GetChild()

		// If the child has a rewrite, check the rewrite.
		// If not, check the leaf.
		if child.GetRewrite() != nil {
			fn = engine.checkRewrite(ctx, subRequest, child.GetRewrite())
		} else {
			fn = engine.checkLeaf(subRequest, child.GetLeaf())
		}
	case base.EntityDefinition_REFERENCE_ATTRIBUTE:
		// If the reference is an attribute, check the direct attribute.
		fn = engine.checkDirectAttribute(subRequest)
	case base.EntityDefinition_REFERENCE_RELATION:
		// If the reference is a relation, check the direct relation.
		fn = engine.checkDirectRelation(subRequest)
	default:
		fn = engine.checkDirectCall(subRequest)
	}

	// If the CheckFunction is still undefined after the switch, return a CheckFunction that always fails with an error indicating an undefined child kind.
	if fn == nil {
		return checkFail(errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_CHILD_KIND.String()))
	}

	// Otherwise, return a CheckFunction that checks a union of CheckFunctions.
	return func(ctx context.Context) (*invoke.BatchCheckResponse, error) {
		result, err := checkUnion(ctx, []CheckFunction{fn}, engine.concurrencyLimit, len(request.EntityIDs))
		if err != nil {
			return result, err
		}
		// Merge identity results into the result.
		if len(identityResults) > 0 {
			for id, res := range identityResults {
				result.Results[id] = res
			}
		}
		return result, nil
	}
}

// checkRewrite prepares a CheckFunction according to the provided Rewrite operation.
// It uses a Rewrite object that describes how to combine the results of multiple CheckFunctions.
func (engine *CheckEngine) checkRewrite(ctx context.Context, request *invoke.BatchCheckRequest, rewrite *base.Rewrite) CheckFunction {
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
func (engine *CheckEngine) checkLeaf(request *invoke.BatchCheckRequest, leaf *base.Leaf) CheckFunction {
	// Switch statement depending on the Leaf type
	switch op := leaf.GetType().(type) {
	// In case of TupleToUserSet operation, prepare a CheckFunction that checks
	// if the request's user is in the UserSet referenced by the tuple.
	case *base.Leaf_TupleToUserSet:
		return engine.checkTupleToUserSet(request, op.TupleToUserSet)
	// In case of ComputedUserSet operation, prepare a CheckFunction that checks
	// if the request's user is in the computed UserSet.
	case *base.Leaf_ComputedUserSet:
		return engine.checkComputedUserSet(request, op.ComputedUserSet)
	// In case of ComputedAttribute operation, prepare a CheckFunction that checks
	// the computed attribute's permission.
	case *base.Leaf_ComputedAttribute:
		return engine.checkComputedAttribute(request, op.ComputedAttribute)
	// In case of Call operation, prepare a CheckFunction that checks
	// the Call's permission.
	case *base.Leaf_Call:
		return engine.checkCall(request, op.Call)
	// In case of an undefined type, return a CheckFunction that always fails.
	default:
		return checkFail(errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_CHILD_TYPE.String()))
	}
}

// setChild prepares a CheckFunction according to the provided combiner function and children.
func (engine *CheckEngine) setChild(
	ctx context.Context,
	request *invoke.BatchCheckRequest,
	children []*base.Child,
	combiner CheckCombiner,
) CheckFunction {
	// Create a slice to store the CheckFunctions
	functions := make([]CheckFunction, 0, len(children))
	// Loop over each child node
	for _, child := range children {
		// Switch on the type of the child node
		switch child.GetType().(type) {
		// In case of a Rewrite node, create a CheckFunction for the Rewrite and append it
		case *base.Child_Rewrite:
			functions = append(functions, engine.checkRewrite(ctx, request, child.GetRewrite()))
		// In case of a Leaf node, create a CheckFunction for the Leaf and append it
		case *base.Child_Leaf:
			functions = append(functions, engine.checkLeaf(request, child.GetLeaf()))
		// In case of an undefined type, return a CheckFunction that always fails
		default:
			return checkFail(errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_CHILD_TYPE.String()))
		}
	}

	// Return a function that when called, runs the appropriate combiner function
	// (union, intersection, exclusion) on the prepared CheckFunctions.
	// Concurrency is controlled by the request-scoped semaphore in the context.
	return func(ctx context.Context) (*invoke.BatchCheckResponse, error) {
		return combiner(ctx, functions, engine.concurrencyLimit, len(request.EntityIDs))
	}
}

// checkDirectRelation is a method of CheckEngine struct that returns a CheckFunction.
// It's responsible for directly checking the permissions on an entity
func (engine *CheckEngine) checkDirectRelation(request *invoke.BatchCheckRequest) CheckFunction {
	// The returned CheckFunction is a closure over the provided context and request
	return func(ctx context.Context) (result *invoke.BatchCheckResponse, err error) {
		// Define a TupleFilter. This specifies which tuples we're interested in.
		// We want tuples that match the entity type and ID from the request, and have a specific relation.
		filter := &base.TupleFilter{
			Entity: &base.EntityFilter{
				Type: request.EntityType,
				Ids:  request.EntityIDs, // IN (...)
			},
			Relation: request.Permission,
		}

		// Query contextual tuples (supports multiple Ids in EntityFilter)
		var cti *database.TupleIterator
		cti, err = storageContext.NewContextualTuples(request.Context.GetTuples()...).QueryRelationships(filter, database.NewCursorPagination())
		if err != nil {
			// If an error occurred while querying, return a "denied" response and the error.
			return denied(request.EntityIDs, emptyResponseMetadata()), err
		}

		// Batch query with subject push-down
		var rit *database.TupleIterator
		rit, err = engine.dataReader.QueryRelationshipsWithSubjectFilter(ctx, request.TenantID, filter, request.Subject, request.Metadata.GetSnapToken(), database.NewCursorPagination())
		// If there's an error in querying, return a denied permission response along with the error.
		if err != nil {
			return denied(request.EntityIDs, emptyResponseMetadata()), err
		}

		// Create a new UniqueTupleIterator from the two TupleIterators.
		// NewUniqueTupleIterator() ensures that the iterator only returns unique tuples.
		it := database.NewUniqueTupleIterator(rit, cti)

		// Per-entity results: track which entities got a direct match.
		directResults := make(map[string]base.CheckResult)
		// Track which entities still need resolution via userset checks.
		needsResolution := make(map[string]bool, len(request.EntityIDs))
		for _, id := range request.EntityIDs {
			needsResolution[id] = true
		}

		// Collect userset subjects, grouping by (entityType, relation) for batch queries.
		// Also track which parent entity each userset subject came from.
		usersetGroups := map[usersetGroupKey][]string{}
		// usersetToParents maps userset entity -> list of parent entity IDs
		usersetToParents := map[entityRef][]string{}

		for it.HasNext() {
			next, ok := it.GetNext()
			if !ok {
				break
			}
			subject := next.GetSubject()
			parentEntityID := next.GetEntity().GetId()

			if tuple.AreSubjectsEqual(subject, request.Subject) {
				// Direct match: mark this specific entity as ALLOWED.
				directResults[parentEntityID] = base.CheckResult_CHECK_RESULT_ALLOWED
				delete(needsResolution, parentEntityID)
				continue
			}
			if !tuple.IsDirectSubject(subject) && subject.GetRelation() != tuple.ELLIPSIS {
				key := usersetGroupKey{entityType: subject.GetType(), relation: subject.GetRelation()}
				usersetGroups[key] = append(usersetGroups[key], subject.GetId())
				ref := entityRef{entityType: subject.GetType(), entityID: subject.GetId()}
				usersetToParents[ref] = append(usersetToParents[ref], parentEntityID)
			}
		}

		// Early exit: if all entities have direct matches, no need for userset checks.
		if len(needsResolution) == 0 {
			resp := invoke.NewBatchCheckResponse(base.CheckResult_CHECK_RESULT_DENIED, request.EntityIDs...)
			for id, res := range directResults {
				resp.Results[id] = res
			}
			return resp, nil
		}

		// Build check functions for userset groups, chunking large groups.
		var checkFunctions []CheckFunction
		totalExpectedEntities := 0
		for key, ids := range usersetGroups {
			totalExpectedEntities += len(ids)
			for i := 0; i < len(ids); i += engine.maxBatchSize {
				end := min(i+engine.maxBatchSize, len(ids))
				checkFunctions = append(checkFunctions, engine.invoke(&invoke.BatchCheckRequest{
					TenantID:   request.TenantID,
					EntityType: key.entityType,
					EntityIDs:  ids[i:end],
					Permission: key.relation,
					Subject:    request.Subject,
					Metadata:   request.Metadata,
					Context:    request.Context,
				}))
			}
		}

		// Start with the response for all requested entities defaulting to DENIED.
		resp := invoke.NewBatchCheckResponse(base.CheckResult_CHECK_RESULT_DENIED, request.EntityIDs...)

		// Apply direct match results.
		for id, res := range directResults {
			resp.Results[id] = res
		}

		// If there are userset check functions, run them and map results back to parent entities.
		if len(checkFunctions) > 0 {
			usersetResp, err := checkUnion(ctx, checkFunctions, engine.concurrencyLimit, totalExpectedEntities)
			if err != nil {
				resp.Metadata = joinResponseMetas(resp.Metadata, usersetResp.Metadata)
				return resp, err
			}
			resp.Metadata = joinResponseMetas(resp.Metadata, usersetResp.Metadata)

			// Map userset results back to parent entities.
			for ref, parentIDs := range usersetToParents {
				if usersetResult, ok := usersetResp.Results[ref.entityID]; ok && usersetResult == base.CheckResult_CHECK_RESULT_ALLOWED {
					for _, parentID := range parentIDs {
						resp.Results[parentID] = base.CheckResult_CHECK_RESULT_ALLOWED
					}
				}
			}
		}

		return resp, nil
	}
}

// checkTupleToUserSet is a method of CheckEngine that checks permissions using the
// TupleToUserSet data structure. It returns a CheckFunction closure that does the check.
func (engine *CheckEngine) checkTupleToUserSet(
	request *invoke.BatchCheckRequest,
	ttu *base.TupleToUserSet,
) CheckFunction {
	// The returned CheckFunction is a closure over the provided context, request, and ttu.
	return func(ctx context.Context) (*invoke.BatchCheckResponse, error) {
		// Define a TupleFilter. This specifies which tuples we're interested in.
		// We want tuples that match the entity type and ID from the request, and have a specific relation.
		filter := &base.TupleFilter{
			Entity: &base.EntityFilter{
				Type: request.EntityType,
				Ids:  request.EntityIDs, // IN (...)
			},
			Relation: ttu.GetTupleSet().GetRelation(),
		}

		// Use the filter to query for relationships in the given context.
		// NewContextualRelationships() creates a ContextualRelationships instance from tuples in the request.
		// QueryRelationships() then uses the filter to find and return matching relationships.
		cti, err := storageContext.NewContextualTuples(request.Context.GetTuples()...).QueryRelationships(filter, database.NewCursorPagination())
		if err != nil {
			// If an error occurred while querying, return a "denied" response and the error.
			return denied(request.EntityIDs, emptyResponseMetadata()), err
		}

		// Use the filter to query for relationships in the database.
		// relationshipReader.QueryRelationships() uses the filter to find and return matching relationships.
		rit, err := engine.dataReader.QueryRelationships(ctx, request.TenantID, filter, request.Metadata.GetSnapToken(), database.NewCursorPagination())
		if err != nil {
			// If an error occurred while querying, return a "denied" response and the error.
			return denied(request.EntityIDs, emptyResponseMetadata()), err
		}

		// Create a new UniqueTupleIterator from the two TupleIterators.
		// NewUniqueTupleIterator() ensures that the iterator only returns unique tuples.
		it := database.NewUniqueTupleIterator(rit, cti)

		// Group subjects by type for batch processing.
		// Also track which parent entity each subject came from.
		subjectsByType := map[string][]string{}
		// subjectToParents maps subject entity -> list of parent entity IDs
		subjectToParents := map[entityRef][]string{}

		for it.HasNext() {
			next, ok := it.GetNext()
			if !ok {
				break
			}
			s := next.GetSubject()
			parentEntityID := next.GetEntity().GetId()
			subjectsByType[s.GetType()] = append(subjectsByType[s.GetType()], s.GetId())
			ref := entityRef{entityType: s.GetType(), entityID: s.GetId()}
			subjectToParents[ref] = append(subjectToParents[ref], parentEntityID)
		}

		var checkFunctions []CheckFunction
		totalExpectedEntities := 0
		for entityType, ids := range subjectsByType {
			totalExpectedEntities += len(ids)
			for i := 0; i < len(ids); i += engine.maxBatchSize {
				end := min(i+engine.maxBatchSize, len(ids))
				checkFunctions = append(checkFunctions, engine.invoke(&invoke.BatchCheckRequest{
					TenantID:   request.TenantID,
					EntityType: entityType,
					EntityIDs:  ids[i:end],
					Permission: ttu.GetComputed().GetRelation(),
					Subject:    request.Subject,
					Metadata:   request.Metadata,
					Context:    request.Context,
					Arguments:  request.Arguments,
				}))
			}
		}

		// Start with all requested entities defaulting to DENIED.
		resp := invoke.NewBatchCheckResponse(base.CheckResult_CHECK_RESULT_DENIED, request.EntityIDs...)

		if len(checkFunctions) == 0 {
			return resp, nil
		}

		subjectResp, err := checkUnion(ctx, checkFunctions, engine.concurrencyLimit, totalExpectedEntities)
		if err != nil {
			resp.Metadata = joinResponseMetas(resp.Metadata, subjectResp.Metadata)
			return resp, err
		}
		resp.Metadata = joinResponseMetas(resp.Metadata, subjectResp.Metadata)

		// Map subject results back to parent entities.
		for ref, parentIDs := range subjectToParents {
			if subjectResult, ok := subjectResp.Results[ref.entityID]; ok && subjectResult == base.CheckResult_CHECK_RESULT_ALLOWED {
				for _, parentID := range parentIDs {
					resp.Results[parentID] = base.CheckResult_CHECK_RESULT_ALLOWED
				}
			}
		}

		return resp, nil
	}
}

// metadata to determine if the computed user set should be excluded from the result.
// checkComputedUserSet is a method of CheckEngine that checks permissions using the
// ComputedUserSet data structure. It returns a CheckFunction closure that performs the check.
func (engine *CheckEngine) checkComputedUserSet(
	request *invoke.BatchCheckRequest,
	cu *base.ComputedUserSet,
) CheckFunction {
	// The returned CheckFunction invokes a permission check with a new request that is almost the same
	// as the incoming request, but changes the Permission to be the relation defined in the computed user set.
	// This is how the check "descends" into the computed user set to check permissions there.
	return engine.invoke(&invoke.BatchCheckRequest{
		TenantID:   request.TenantID,
		EntityType: request.EntityType,
		EntityIDs:  request.EntityIDs,
		Permission: cu.GetRelation(),
		Subject:    request.Subject,
		Metadata:   request.Metadata,
		Context:    request.Context,
		Arguments:  request.Arguments,
	})
}

// checkComputedAttribute constructs a CheckFunction that checks if a computed attribute
// permission check request is allowed or denied.
func (engine *CheckEngine) checkComputedAttribute(
	request *invoke.BatchCheckRequest,
	ca *base.ComputedAttribute,
) CheckFunction {
	// We're returning a function here - this is the CheckFunction.
	// Instead of performing the check directly here, we're using the 'invoke' method.
	// We pass a new PermissionCheckRequest to 'invoke', copying most of the fields
	// from the original request, but replacing the 'Permission' with the computed
	// attribute's name.
	return engine.invoke(&invoke.BatchCheckRequest{
		TenantID:   request.TenantID,
		EntityType: request.EntityType,
		EntityIDs:  request.EntityIDs,
		Permission: ca.GetName(),
		Subject:    request.Subject,
		Metadata:   request.Metadata,
		Context:    request.Context,
	})
}

// checkDirectAttribute constructs a CheckFunction that checks if a direct attribute
// permission check request is allowed or denied.
// Uses batch QueryAttributes with multiple entity IDs in a single query.
func (engine *CheckEngine) checkDirectAttribute(request *invoke.BatchCheckRequest) CheckFunction {
	return func(ctx context.Context) (*invoke.BatchCheckResponse, error) {
		resp := invoke.NewBatchCheckResponse(base.CheckResult_CHECK_RESULT_DENIED, request.EntityIDs...)

		filter := &base.AttributeFilter{
			Entity: &base.EntityFilter{
				Type: request.EntityType,
				Ids:  request.EntityIDs, // IN (...)
			},
			Attributes: []string{request.Permission},
		}

		// Query contextual attributes (in-memory, supports multiple IDs).
		cta, err := storageContext.NewContextualAttributes(request.Context.GetAttributes()...).QueryAttributes(filter, database.NewCursorPagination())
		if err != nil {
			return resp, err
		}

		// Batch query from database.
		ait, err := engine.dataReader.QueryAttributes(ctx, request.TenantID, filter, request.Metadata.GetSnapToken(), database.NewCursorPagination())
		if err != nil {
			return resp, err
		}

		// Combine attributes from different sources ensuring uniqueness.
		it := database.NewUniqueAttributeIterator(ait, cta)
		for it.HasNext() {
			next, ok := it.GetNext()
			if !ok {
				break
			}

			// Unmarshal the attribute value into a BoolValue message.
			var msg base.BooleanValue
			if err := next.GetValue().UnmarshalTo(&msg); err != nil {
				return resp, err
			}

			if msg.Data {
				resp.Results[next.GetEntity().GetId()] = base.CheckResult_CHECK_RESULT_ALLOWED
			}
		}

		return resp, nil
	}
}

// checkCall creates and returns a CheckFunction based on the provided request and call details.
// It essentially constructs a new BatchCheckRequest based on the call details and then invokes
// the permission check using the engine's invoke method.
func (engine *CheckEngine) checkCall(
	request *invoke.BatchCheckRequest,
	call *base.Call,
) CheckFunction {
	// Construct a new permission check request based on the input request and call details.
	return engine.invoke(&invoke.BatchCheckRequest{
		TenantID:   request.TenantID,
		EntityType: request.EntityType,
		EntityIDs:  request.EntityIDs,
		Permission: call.GetRuleName(),
		Subject:    request.Subject,
		Metadata:   request.Metadata,
		Context:    request.Context,
		Arguments:  call.GetArguments(),
	})
}

// checkDirectCall creates and returns a CheckFunction that performs direct permission checking.
// The function evaluates permissions based on rule definitions, arguments, and attributes.
// Processes each entity individually since each entity may have different attribute values.
func (engine *CheckEngine) checkDirectCall(request *invoke.BatchCheckRequest) CheckFunction {
	return func(ctx context.Context) (*invoke.BatchCheckResponse, error) {
		resp := invoke.NewBatchCheckResponse(base.CheckResult_CHECK_RESULT_DENIED, request.EntityIDs...)

		// Read the rule definition from the schema once (shared across all entities).
		var ru *base.RuleDefinition
		ru, _, err := engine.schemaReader.ReadRuleDefinition(ctx, request.TenantID, request.Permission, request.Metadata.GetSchemaVersion())
		if err != nil {
			return resp, err
		}

		// Prepare the CEL environment once (shared across all entities).
		env, err := utils.ArgumentsAsCelEnv(ru.Arguments)
		if err != nil {
			return resp, err
		}

		// Compile the rule expression once.
		exp := cel.CheckedExprToAst(ru.Expression)
		prg, err := env.Program(exp)
		if err != nil {
			return resp, err
		}

		// Classify arguments once.
		var attributes []string
		baseArguments := map[string]any{
			"context": map[string]any{
				"data": request.Context.GetData().AsMap(),
			},
		}
		for _, arg := range request.Arguments {
			switch actualArg := arg.Type.(type) {
			case *base.Argument_ComputedAttribute:
				attrName := actualArg.ComputedAttribute.GetName()
				emptyValue := getEmptyValueForType(ru.GetArguments()[attrName])
				baseArguments[attrName] = emptyValue
				attributes = append(attributes, attrName)
			default:
				return resp, errors.New(base.ErrorCode_ERROR_CODE_INTERNAL.String())
			}
		}

		// Batch query ALL attributes for ALL entities at once.
		attrsByEntity := map[string]map[string]any{}
		if len(attributes) > 0 {
			filter := &base.AttributeFilter{
				Entity: &base.EntityFilter{
					Type: request.EntityType,
					Ids:  request.EntityIDs,
				},
				Attributes: attributes,
			}

			ait, err := engine.dataReader.QueryAttributes(ctx, request.TenantID, filter, request.Metadata.GetSnapToken(), database.NewCursorPagination())
			if err != nil {
				return resp, err
			}

			cta, err := storageContext.NewContextualAttributes(request.Context.GetAttributes()...).QueryAttributes(filter, database.NewCursorPagination())
			if err != nil {
				return resp, err
			}

			it := database.NewUniqueAttributeIterator(ait, cta)
			for it.HasNext() {
				next, ok := it.GetNext()
				if !ok {
					break
				}
				entityID := next.GetEntity().GetId()
				if attrsByEntity[entityID] == nil {
					attrsByEntity[entityID] = make(map[string]any)
				}
				attrsByEntity[entityID][next.GetAttribute()] = utils.ConvertProtoAnyToInterface(next.GetValue())
			}
		}

		// Evaluate CEL per entity with its specific attributes.
		for _, entityID := range request.EntityIDs {
			arguments := make(map[string]any, len(baseArguments))
			for k, v := range baseArguments {
				arguments[k] = v
			}
			for k, v := range attrsByEntity[entityID] {
				arguments[k] = v
			}

			// Evaluate the rule expression with the arguments for this entity.
			out, _, err := prg.Eval(arguments)
			if err != nil {
				return resp, fmt.Errorf("failed to evaluate expression: %w", err)
			}

			result, ok := out.Value().(bool)
			if !ok {
				return resp, fmt.Errorf("expected boolean result, but got %T", out.Value())
			}

			if result {
				resp.Results[entityID] = base.CheckResult_CHECK_RESULT_ALLOWED
			}
		}

		return resp, nil
	}
}

// checkUnion checks if the subject has permission by running multiple CheckFunctions concurrently.
// Per-entity merge: for each entityID, if ANY function returned ALLOWED -> ALLOWED, else -> DENIED.
func checkUnion(ctx context.Context, functions []CheckFunction, limit int, expectedEntityCount int) (*invoke.BatchCheckResponse, error) {
	// Initialize the response metadata
	responseMetadata := emptyResponseMetadata()

	// If there are no functions, deny the permission and return
	if len(functions) == 0 {
		return &invoke.BatchCheckResponse{
			Results:  map[string]base.CheckResult{},
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

	// Merged per-entity results: for each entityID, if ANY function returned ALLOWED -> ALLOWED.
	mergedResults := map[string]base.CheckResult{}
	// Track how many entities are still DENIED. When this reaches 0, all are ALLOWED and we can exit early.
	deniedCount := 0
	entityIDsSeen := false

	// Iterate over the results of the CheckFunctions
	for range len(functions) {
		select {
		// If a result is received
		case d := <-decisionChan:
			// Merge the response metadata with the received metadata
			responseMetadata = joinResponseMetas(responseMetadata, d.resp.Metadata)
			// If there was an error, return what we have so far with the error
			if d.err != nil {
				return &invoke.BatchCheckResponse{
					Results:  mergedResults,
					Metadata: responseMetadata,
				}, d.err
			}
			// Per-entity union: if this function allowed an entity, mark it as allowed
			for entityID, result := range d.resp.Results {
				if result == base.CheckResult_CHECK_RESULT_ALLOWED {
					if prev, exists := mergedResults[entityID]; !exists {
						mergedResults[entityID] = base.CheckResult_CHECK_RESULT_ALLOWED
						// New entity seen, already allowed, no change to deniedCount
					} else if prev == base.CheckResult_CHECK_RESULT_DENIED {
						mergedResults[entityID] = base.CheckResult_CHECK_RESULT_ALLOWED
						deniedCount--
					}
				} else if _, exists := mergedResults[entityID]; !exists {
					mergedResults[entityID] = base.CheckResult_CHECK_RESULT_DENIED
					deniedCount++
				}
			}
			entityIDsSeen = true

			// Early exit: if all expected entities are now ALLOWED, no further functions can change the result.
			if entityIDsSeen && deniedCount == 0 && len(mergedResults) >= expectedEntityCount {
				return &invoke.BatchCheckResponse{
					Results:  mergedResults,
					Metadata: responseMetadata,
				}, nil
			}
		// If the context is done, return a cancellation error
		case <-ctx.Done():
			return &invoke.BatchCheckResponse{
				Results:  mergedResults,
				Metadata: responseMetadata,
			}, errors.New(base.ErrorCode_ERROR_CODE_CANCELLED.String())
		}
	}

	return &invoke.BatchCheckResponse{
		Results:  mergedResults,
		Metadata: responseMetadata,
	}, nil
}

// checkIntersection checks if the subject has permission by running multiple CheckFunctions concurrently.
// Per-entity merge: for each entityID, ALL functions must return ALLOWED -> ALLOWED, else -> DENIED.
func checkIntersection(ctx context.Context, functions []CheckFunction, limit int, _ int) (*invoke.BatchCheckResponse, error) {
	// Initialize the response metadata
	responseMetadata := emptyResponseMetadata()

	// If there are no functions, deny the permission and return
	if len(functions) == 0 {
		return &invoke.BatchCheckResponse{
			Results:  map[string]base.CheckResult{},
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

	// Track per-entity: how many functions returned ALLOWED and total functions seen.
	allowedCounts := map[string]int{}
	deniedEntities := map[string]struct{}{} // entities that received at least one DENIED
	// Track all entity IDs seen.
	allEntityIDs := map[string]struct{}{}
	entityIDsSeen := false

	// Iterate over the results of the CheckFunctions
	for range len(functions) {
		select {
		// If a result is received
		case d := <-decisionChan:
			// Merge the response metadata with the received metadata
			responseMetadata = joinResponseMetas(responseMetadata, d.resp.Metadata)
			// If there was an error, return denied with the error
			if d.err != nil {
				// Build denied results for all seen entities
				results := make(map[string]base.CheckResult, len(allEntityIDs))
				for id := range allEntityIDs {
					results[id] = base.CheckResult_CHECK_RESULT_DENIED
				}
				return &invoke.BatchCheckResponse{
					Results:  results,
					Metadata: responseMetadata,
				}, d.err
			}
			// Track per-entity allowed counts
			for entityID, result := range d.resp.Results {
				allEntityIDs[entityID] = struct{}{}
				if result == base.CheckResult_CHECK_RESULT_ALLOWED {
					allowedCounts[entityID]++
				} else {
					deniedEntities[entityID] = struct{}{}
				}
			}
			entityIDsSeen = true

			// Early exit: if ALL known entities have been DENIED by at least one function,
			// no further functions can make them ALLOWED (intersection requires all).
			if entityIDsSeen && len(deniedEntities) == len(allEntityIDs) {
				results := make(map[string]base.CheckResult, len(allEntityIDs))
				for id := range allEntityIDs {
					results[id] = base.CheckResult_CHECK_RESULT_DENIED
				}
				return &invoke.BatchCheckResponse{
					Results:  results,
					Metadata: responseMetadata,
				}, nil
			}
		// If the context is done, return a cancellation error
		case <-ctx.Done():
			results := make(map[string]base.CheckResult, len(allEntityIDs))
			for id := range allEntityIDs {
				results[id] = base.CheckResult_CHECK_RESULT_DENIED
			}
			return &invoke.BatchCheckResponse{
				Results:  results,
				Metadata: responseMetadata,
			}, errors.New(base.ErrorCode_ERROR_CODE_CANCELLED.String())
		}
	}

	// Build final results: entity is ALLOWED only if ALL functions returned ALLOWED for it.
	numFunctions := len(functions)
	results := make(map[string]base.CheckResult, len(allEntityIDs))
	for entityID := range allEntityIDs {
		if allowedCounts[entityID] == numFunctions {
			results[entityID] = base.CheckResult_CHECK_RESULT_ALLOWED
		} else {
			results[entityID] = base.CheckResult_CHECK_RESULT_DENIED
		}
	}

	return &invoke.BatchCheckResponse{
		Results:  results,
		Metadata: responseMetadata,
	}, nil
}

// checkExclusion is a function that checks if there are any exclusions for given CheckFunctions.
// Per-entity merge: for each entityID, first function ALLOWED AND all remaining DENIED -> ALLOWED, else -> DENIED.
func checkExclusion(ctx context.Context, functions []CheckFunction, limit int, _ int) (*invoke.BatchCheckResponse, error) {
	// Initialize the response metadata
	responseMetadata := emptyResponseMetadata()

	// Check if there are at least 2 functions, otherwise return an error indicating that exclusion requires more than one function
	if len(functions) <= 1 {
		return &invoke.BatchCheckResponse{
			Results:  map[string]base.CheckResult{},
			Metadata: responseMetadata,
		}, errors.New(base.ErrorCode_ERROR_CODE_EXCLUSION_REQUIRES_MORE_THAN_ONE_FUNCTION.String())
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
		defer wg.Done()
		result, err := functions[0](cancelCtx)
		leftDecisionChan <- CheckResponse{
			resp: result,
			err:  err,
		}
	}()

	// Run the remaining functions concurrently
	clean := checkRun(cancelCtx, functions[1:], decisionChan, limit)

	// Ensure that all resources are properly cleaned up when the function exits
	defer func() {
		cancel()
		clean()
		close(decisionChan)
		wg.Wait()
		close(leftDecisionChan)
	}()

	// Per-entity results from the first (left) function.
	var leftResults map[string]base.CheckResult

	// Process the result from the first function
	select {
	case left := <-leftDecisionChan:
		responseMetadata = joinResponseMetas(responseMetadata, left.resp.Metadata)

		if left.err != nil {
			return &invoke.BatchCheckResponse{
				Results:  map[string]base.CheckResult{},
				Metadata: responseMetadata,
			}, left.err
		}

		leftResults = left.resp.Results

	case <-ctx.Done():
		return &invoke.BatchCheckResponse{
			Results:  map[string]base.CheckResult{},
			Metadata: responseMetadata,
		}, errors.New(base.ErrorCode_ERROR_CODE_CANCELLED.String())
	}

	// Early exit: if ALL entities in the left result are DENIED, exclusion cannot produce ALLOWED.
	allLeftDenied := true
	leftAllowedCount := 0
	for _, result := range leftResults {
		if result == base.CheckResult_CHECK_RESULT_ALLOWED {
			allLeftDenied = false
			leftAllowedCount++
		}
	}
	if allLeftDenied {
		results := make(map[string]base.CheckResult, len(leftResults))
		for id := range leftResults {
			results[id] = base.CheckResult_CHECK_RESULT_DENIED
		}
		return &invoke.BatchCheckResponse{
			Results:  results,
			Metadata: responseMetadata,
		}, nil
	}

	// Track per-entity: whether any remaining function returned ALLOWED (which would exclude).
	excludedEntities := map[string]bool{}

	// Process the results from the remaining functions
	for range len(functions) - 1 {
		select {
		case d := <-decisionChan:
			responseMetadata = joinResponseMetas(responseMetadata, d.resp.Metadata)

			if d.err != nil {
				// On error, return denied for all entities
				results := make(map[string]base.CheckResult, len(leftResults))
				for id := range leftResults {
					results[id] = base.CheckResult_CHECK_RESULT_DENIED
				}
				return &invoke.BatchCheckResponse{
					Results:  results,
					Metadata: responseMetadata,
				}, d.err
			}

			// If any remaining function allowed an entity, that entity is excluded (denied)
			for entityID, result := range d.resp.Results {
				if result == base.CheckResult_CHECK_RESULT_ALLOWED {
					if !excludedEntities[entityID] && leftResults[entityID] == base.CheckResult_CHECK_RESULT_ALLOWED {
						leftAllowedCount--
					}
					excludedEntities[entityID] = true
				}
			}

			// Early exit: if all initially-ALLOWED entities are now excluded, result is all DENIED
			if leftAllowedCount <= 0 {
				results := make(map[string]base.CheckResult, len(leftResults))
				for id := range leftResults {
					results[id] = base.CheckResult_CHECK_RESULT_DENIED
				}
				return &invoke.BatchCheckResponse{
					Results:  results,
					Metadata: responseMetadata,
				}, nil
			}

		case <-ctx.Done():
			results := make(map[string]base.CheckResult, len(leftResults))
			for id := range leftResults {
				results[id] = base.CheckResult_CHECK_RESULT_DENIED
			}
			return &invoke.BatchCheckResponse{
				Results:  results,
				Metadata: responseMetadata,
			}, errors.New(base.ErrorCode_ERROR_CODE_CANCELLED.String())
		}
	}

	// Build final results: entity is ALLOWED if first function ALLOWED AND no remaining function ALLOWED
	results := make(map[string]base.CheckResult, len(leftResults))
	for entityID, leftResult := range leftResults {
		if leftResult == base.CheckResult_CHECK_RESULT_ALLOWED && !excludedEntities[entityID] {
			results[entityID] = base.CheckResult_CHECK_RESULT_ALLOWED
		} else {
			results[entityID] = base.CheckResult_CHECK_RESULT_DENIED
		}
	}

	return &invoke.BatchCheckResponse{
		Results:  results,
		Metadata: responseMetadata,
	}, nil
}

// checkRun executes a list of CheckFunctions concurrently.
// DB-level concurrency is controlled by the semaphore DataReader proxy, not here.
// checkRun executes a list of CheckFunctions concurrently with a local concurrency limit.
// DB-level concurrency is controlled by the semaphore DataReader proxy.
// The local limit here prevents excessive goroutine fan-out and depth exhaustion.
func checkRun(ctx context.Context, functions []CheckFunction, decisionChan chan<- CheckResponse, limit int) func() {
	var wg sync.WaitGroup
	cl := make(chan struct{}, limit)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, fun := range functions {
			child := fun
			select {
			case cl <- struct{}{}:
			case <-ctx.Done():
				return
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				result, err := child(ctx)
				decisionChan <- CheckResponse{
					resp: result,
					err:  err,
				}
				<-cl
			}()
		}
	}()

	// Return a cleanup function that waits for all goroutines to finish and then closes the concurrency limit channel
	return func() {
		wg.Wait()
		close(cl)
	}
}

// checkFail is a helper function that returns a CheckFunction that always returns a denied BatchCheckResponse
// with the provided error and an empty PermissionCheckResponseMetadata.
//
// The function works as follows:
//  1. The function takes an error as input parameter.
//  2. The function returns a CheckFunction that takes a context as input parameter and always returns a denied
//     BatchCheckResponse with the provided error and an empty PermissionCheckResponseMetadata.
func checkFail(err error) CheckFunction {
	return func(ctx context.Context) (*invoke.BatchCheckResponse, error) {
		return &invoke.BatchCheckResponse{
			Results:  map[string]base.CheckResult{},
			Metadata: &base.PermissionCheckResponseMetadata{},
		}, err
	}
}

// usersetGroupKey identifies a group of userset subjects that share the same entity type and relation,
// allowing their next-level queries to be batched into a single SQL call with IN (...).
type usersetGroupKey struct {
	entityType string
	relation   string
}

// entityRef identifies a specific entity by type and ID (used as map key).
type entityRef struct {
	entityType string
	entityID   string
}

// denied is a helper function that returns a denied BatchCheckResponse for the given entity IDs
// with the provided PermissionCheckResponseMetadata.
//
// The function works as follows:
// 1. The function takes entity IDs and a PermissionCheckResponseMetadata as input parameters.
// 2. The function returns a denied BatchCheckResponse with RESULT_DENIED for each entity and the provided metadata.
func denied(entityIDs []string, meta *base.PermissionCheckResponseMetadata) *invoke.BatchCheckResponse {
	return &invoke.BatchCheckResponse{
		Results: func() map[string]base.CheckResult {
			results := make(map[string]base.CheckResult, len(entityIDs))
			for _, id := range entityIDs {
				results[id] = base.CheckResult_CHECK_RESULT_DENIED
			}
			return results
		}(),
		Metadata: meta,
	}
}

// emptyResponseMetadata creates and returns an empty PermissionCheckResponseMetadata.
//
// Returns:
// - A pointer to PermissionCheckResponseMetadata with the CheckCount initialized to 0.
func emptyResponseMetadata() *base.PermissionCheckResponseMetadata {
	return &base.PermissionCheckResponseMetadata{
		CheckCount: 0,
	}
}
