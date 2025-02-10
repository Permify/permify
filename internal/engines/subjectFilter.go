package engines

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/google/cel-go/cel"

	"github.com/Permify/permify/internal/schema"
	"github.com/Permify/permify/internal/storage"
	storageContext "github.com/Permify/permify/internal/storage/context"
	"github.com/Permify/permify/pkg/database"
	"github.com/Permify/permify/pkg/dsl/utils"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
)

const ALL = "<>"

type SubjectFilter struct {
	// schemaReader is responsible for reading schema information
	schemaReader storage.SchemaReader
	// dataReader is responsible for reading relationship information
	dataReader storage.DataReader
	// concurrencyLimit is the maximum number of concurrent permission checks allowed
	concurrencyLimit int
}

func NewSubjectFilter(schemaReader storage.SchemaReader, dataReader storage.DataReader, opts ...SubjectFilterOption) *SubjectFilter {
	filter := &SubjectFilter{
		dataReader:       dataReader,
		schemaReader:     schemaReader,
		concurrencyLimit: _defaultConcurrencyLimit,
	}

	for _, opt := range opts {
		opt(filter)
	}

	return filter
}

// SubjectFilterFunction defines the type for a function that takes a context and
// returns a pointer to a PermissionSubjectFilterResponse and an error.
// This type is often used when you want to pass around functions with this specific signature.
type SubjectFilterFunction func(ctx context.Context) ([]string, error)

// SubjectFilterCombiner defines the type for a function that takes a context, a slice of SubjectFilterFunctions,
// an integer as a limit and returns a pointer to a PermissionSubjectFilterResponse and an error.
// This type is useful when you want to define a function that can execute multiple SubjectFilterFunctions in a specific way
// (like concurrently with a limit or sequentially) and combine their results into a single PermissionSubjectFilterResponse.
type SubjectFilterCombiner func(ctx context.Context, functions []SubjectFilterFunction, limit int) ([]string, error)

// SubjectFilter is a method for the SubjectFilterEngine struct.
// It takes a context and a pointer to a PermissionSubjectFilterRequest
// and returns a pointer to a PermissionSubjectFilterResponse and an error.
func (engine *SubjectFilter) SubjectFilter(ctx context.Context, request *base.PermissionLookupSubjectRequest) (response []string, err error) {
	// ReadEntityDefinition method of the SchemaReader interface is used to retrieve the entity's schema definition.
	// GetTenantId, GetType and GetSchemaVersion methods are used to provide necessary arguments to ReadEntityDefinition.
	var en *base.EntityDefinition
	en, _, err = engine.schemaReader.ReadEntityDefinition(ctx, request.GetTenantId(), request.GetEntity().GetType(), request.GetMetadata().GetSchemaVersion())
	if err != nil {
		// If an error is encountered while reading the schema definition, return an empty response and the error.
		return subjectFilterEmpty(), err
	}

	var res []string

	// Call the subjectFilter method of the engine, which returns a function.
	// That function is then immediately called with the context to perform the actual subject lookup.
	res, err = engine.subjectFilter(ctx, request, en)(ctx)
	if err != nil {
		// If an error is encountered during the lookup, return an empty response and the error.
		return subjectFilterEmpty(), err
	}

	// If everything went smoothly, return the lookup result and nil error.
	return res, nil
}

// subjectFilter is a method for the SubjectFilterEngine struct.
// It determines the type of relational reference for the permission field in the request,
// prepares and returns a SubjectFilterFunction accordingly.
func (engine *SubjectFilter) subjectFilter(
	ctx context.Context,
	request *base.PermissionLookupSubjectRequest,
	en *base.EntityDefinition,
) SubjectFilterFunction {
	// The GetTypeOfRelationalReferenceByNameInEntityDefinition method retrieves the type of relational reference for the permission field.
	tor, _ := schema.GetTypeOfReferenceByNameInEntityDefinition(en, request.GetPermission())

	var fn SubjectFilterFunction

	switch tor {
	case base.EntityDefinition_REFERENCE_PERMISSION:
		// Get the permission by its name in the entity definition.
		permission, err := schema.GetPermissionByNameInEntityDefinition(en, request.GetPermission())
		if err != nil {
			// If an error is encountered while getting the permission, a SubjectFilterFunction is returned that always fails with this error.
			return subjectFilterFail(err)
		}
		child := permission.GetChild()

		// Depending on whether the child permission has a rewrite rule,
		// we prepare a SubjectFilterFunction to handle it accordingly.
		if child.GetRewrite() != nil {
			fn = engine.subjectFilterRewrite(ctx, request, child.GetRewrite())
		} else {
			fn = engine.subjectFilterLeaf(request, child.GetLeaf())
		}
	case base.EntityDefinition_REFERENCE_ATTRIBUTE:
		fn = engine.subjectFilterDirectAttribute(request)
	case base.EntityDefinition_REFERENCE_RELATION:
		fn = engine.subjectFilterDirectRelation(request)
	default:
		fn = engine.subjectFilterDirectCall(request)
	}

	// If we could not prepare a SubjectFilterFunction, we return a function that always fails with an error indicating undefined child kind.
	if fn == nil {
		return subjectFilterFail(errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_CHILD_KIND.String()))
	}

	// Finally, we return a function that combines results from the prepared function.
	return func(ctx context.Context) ([]string, error) {
		return subjectFilterUnion(ctx, []SubjectFilterFunction{fn}, engine.concurrencyLimit)
	}
}

// subjectFilterRewrite is a method for the SubjectFilterEngine struct.
// It generates a SubjectFilterFunction based on the rewrite operation type present in the provided request.
func (engine *SubjectFilter) subjectFilterRewrite(
	ctx context.Context,
	request *base.PermissionLookupSubjectRequest,
	rewrite *base.Rewrite,
) SubjectFilterFunction {
	// Check the type of the rewrite operation in the request
	switch rewrite.GetRewriteOperation() {
	// If the operation type is UNION, we prepare a function using subjectFilterUnion
	// If the request has an exclusion, we use subjectFilterIntersection instead
	case *base.Rewrite_OPERATION_UNION.Enum():
		// Use the chosen function to set the child entities for lookup
		return engine.setChild(ctx, request, rewrite.GetChildren(), subjectFilterUnion)

	// If the operation type is INTERSECTION, we prepare a function using subjectFilterIntersection
	// If the request has an exclusion, we use subjectFilterUnion instead
	case *base.Rewrite_OPERATION_INTERSECTION.Enum():

		// Use the chosen function to set the child entities for lookup
		return engine.setChild(ctx, request, rewrite.GetChildren(), subjectFilterIntersection)

	// If the operation type is not recognized, we return a function that always fails with an error indicating undefined child type.
	case *base.Rewrite_OPERATION_EXCLUSION.Enum():

		// implement exclusion
		return engine.setChild(ctx, request, rewrite.GetChildren(), subjectFilterExclusion)
	default:
		return subjectFilterFail(errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_CHILD_TYPE.String()))
	}
}

// subjectFilterLeaf is a method for the SubjectFilterEngine struct.
// It generates a SubjectFilterFunction based on the type of the leaf node in the provided request.
func (engine *SubjectFilter) subjectFilterLeaf(
	request *base.PermissionLookupSubjectRequest,
	leaf *base.Leaf,
) SubjectFilterFunction {
	// Check the type of the leaf node in the request
	switch op := leaf.GetType().(type) {
	// If the type is TupleToUserSet, we prepare a function using subjectFilterTupleToUserSet
	case *base.Leaf_TupleToUserSet:
		return engine.subjectFilterTupleToUserSet(request, op.TupleToUserSet)

	// If the type is ComputedUserSet, we prepare a function using subjectFilterComputedUserSet
	case *base.Leaf_ComputedUserSet:
		return engine.subjectFilterComputedUserSet(request, op.ComputedUserSet)

	case *base.Leaf_ComputedAttribute:

		return engine.subjectfFlterComputedAttribute(request, op.ComputedAttribute)
	case *base.Leaf_Call:

		return engine.subjectFilterCall(request, op.Call)
	// If the leaf type is not recognized, we return a function that always fails with an error indicating undefined child type.
	default:
		return subjectFilterFail(errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_CHILD_TYPE.String()))
	}
}

// checkComputedAttribute constructs a CheckFunction that checks if a computed attribute
// permission check request is allowed or denied.
func (engine *SubjectFilter) subjectfFlterComputedAttribute(
	request *base.PermissionLookupSubjectRequest,
	ca *base.ComputedAttribute,
) SubjectFilterFunction {
	// This function creates a new SubjectFilterFunction. This function is defined to call SubjectFilter on the engine,
	// with a new PermissionSubjectFilterRequest based on the current request and the ComputedUserSet.
	return func(ctx context.Context) ([]string, error) {
		return engine.SubjectFilter(ctx, &base.PermissionLookupSubjectRequest{
			// The tenant ID is preserved from the original request.
			TenantId: request.GetTenantId(),

			// The entity type and ID are preserved from the original request.
			Entity: &base.Entity{
				Type: request.GetEntity().GetType(),
				Id:   request.GetEntity().GetId(),
			},

			// The permission field is replaced with the relation from the ComputedUserSet.
			Permission: ca.GetName(),

			// The subject reference is preserved from the original request.
			SubjectReference: request.GetSubjectReference(),

			// The metadata is preserved from the original request.
			Metadata: request.GetMetadata(),

			// The contextual tuples are preserved from the original request.
			Context: request.GetContext(),

			ContinuousToken: request.GetContinuousToken(),
		})
	}
}

func (engine *SubjectFilter) subjectFilterDirectAttribute(
	request *base.PermissionLookupSubjectRequest,
) SubjectFilterFunction {
	// We're returning a function here - this is the actual CheckFunction.
	return func(ctx context.Context) (result []string, err error) {
		// Create a new AttributeFilter with the entity type and ID from the request
		// and the requested permission.
		filter := &base.AttributeFilter{
			Entity: &base.EntityFilter{
				Type: request.GetEntity().GetType(),
				Ids:  []string{request.GetEntity().GetId()},
			},
			Attributes: []string{request.GetPermission()},
		}

		var val *base.Attribute

		// storageContext.NewContextualAttributes creates a new instance of ContextualAttributes based on the attributes
		// retrieved from the request context.
		val, err = storageContext.NewContextualAttributes(request.GetContext().GetAttributes()...).QuerySingleAttribute(filter)
		// An error occurred while querying the single attribute, so we return a denied response with empty metadata
		// and the error.
		if err != nil {
			return subjectFilterEmpty(), err
		}

		if val == nil {
			// Use the data reader's QuerySingleAttribute method to find the relevant attribute
			val, err = engine.dataReader.QuerySingleAttribute(ctx, request.GetTenantId(), filter, request.GetMetadata().GetSnapToken())
			// If there was an error, return a denied response and the error.
			if err != nil {
				return subjectFilterEmpty(), err
			}
		}

		// No attribute was found matching the provided filter. In this case, we return a denied response with empty metadata
		// and no error.
		if val == nil {
			return subjectFilterEmpty(), nil
		}

		// Unmarshal the attribute value into a BoolValue message.
		var msg base.BooleanValue
		if err := val.GetValue().UnmarshalTo(&msg); err != nil {
			// If there was an error unmarshaling, return a denied response and the error.
			return subjectFilterEmpty(), err
		}

		// If the attribute's value is true, return an allowed response.
		if msg.Data {
			return []string{ALL}, nil
		}

		// If the attribute's value is not true, return a denied response.
		return subjectFilterEmpty(), nil
	}
}

func (engine *SubjectFilter) subjectFilterCall(
	request *base.PermissionLookupSubjectRequest,
	call *base.Call,
) SubjectFilterFunction {
	// Construct a new permission check request based on the input request and call details.
	return func(ctx context.Context) ([]string, error) {
		return engine.SubjectFilter(ctx, &base.PermissionLookupSubjectRequest{
			TenantId:         request.GetTenantId(),
			Entity:           request.GetEntity(),
			Permission:       call.GetRuleName(),
			SubjectReference: request.GetSubjectReference(),
			Metadata:         request.GetMetadata(),
			Context:          request.GetContext(),
			Arguments:        call.GetArguments(),
		})
	}
}

// checkDirectCall creates and returns a CheckFunction that performs direct permission checking.
// The function evaluates permissions based on rule definitions, arguments, and attributes.
func (engine *SubjectFilter) subjectFilterDirectCall(
	request *base.PermissionLookupSubjectRequest,
) SubjectFilterFunction {
	return func(ctx context.Context) (ids []string, err error) {
		// Read the rule definition from the schema. If an error occurs, return the default denied response.
		var ru *base.RuleDefinition
		ru, _, err = engine.schemaReader.ReadRuleDefinition(ctx, request.GetTenantId(), request.GetPermission(), request.GetMetadata().GetSchemaVersion())
		if err != nil {
			return subjectFilterEmpty(), err
		}

		// Initialize an arguments map to hold argument values.
		arguments := map[string]interface{}{
			"context": map[string]interface{}{
				"data": request.GetContext().GetData().AsMap(),
			},
		}

		// List to store computed attributes.
		attributes := make([]string, 0)

		// Iterate over request arguments to classify and process them.
		for _, arg := range request.GetArguments() {
			switch actualArg := arg.Type.(type) {
			case *base.Argument_ComputedAttribute:
				// Handle computed attributes: Set them to a default empty value.
				attrName := actualArg.ComputedAttribute.GetName()
				emptyValue := getEmptyValueForType(ru.GetArguments()[attrName])
				arguments[attrName] = emptyValue
				attributes = append(attributes, attrName)
			default:
				// Return an error for any unsupported argument types.
				return subjectFilterEmpty(), errors.New(base.ErrorCode_ERROR_CODE_INTERNAL.String())
			}
		}

		// If there are computed attributes, fetch them from the data source.
		if len(attributes) > 0 {
			filter := &base.AttributeFilter{
				Entity: &base.EntityFilter{
					Type: request.GetEntity().GetType(),
					Ids:  []string{request.GetEntity().GetId()},
				},
				Attributes: attributes,
			}

			ait, err := engine.dataReader.QueryAttributes(ctx, request.GetTenantId(), filter, request.GetMetadata().GetSnapToken(), database.NewCursorPagination())
			if err != nil {
				return subjectFilterEmpty(), err
			}

			cta, err := storageContext.NewContextualAttributes(request.GetContext().GetAttributes()...).QueryAttributes(filter, database.NewCursorPagination())
			if err != nil {
				return subjectFilterEmpty(), err
			}

			// Combine attributes from different sources ensuring uniqueness.
			it := database.NewUniqueAttributeIterator(ait, cta)
			for it.HasNext() {
				next, ok := it.GetNext()
				if !ok {
					break
				}
				arguments[next.GetAttribute()] = utils.ConvertProtoAnyToInterface(next.GetValue())
			}
		}

		// Prepare the CEL environment with the argument values.
		env, err := utils.ArgumentsAsCelEnv(ru.Arguments)
		if err != nil {
			return subjectFilterEmpty(), err
		}

		// Compile the rule expression into an executable form.
		exp := cel.CheckedExprToAst(ru.Expression)
		prg, err := env.Program(exp)
		if err != nil {
			return subjectFilterEmpty(), err
		}

		// Evaluate the rule expression with the provided arguments.
		out, _, err := prg.Eval(arguments)
		if err != nil {
			return subjectFilterEmpty(), fmt.Errorf("failed to evaluate expression: %w", err)
		}

		// Ensure the result of evaluation is boolean and decide on permission.
		result, ok := out.Value().(bool)
		if !ok {
			return subjectFilterEmpty(), fmt.Errorf("expected boolean result, but got %T", out.Value())
		}

		// If the result of the CEL evaluation is true, return an "allowed" response, otherwise return a "denied" response
		if result {
			return []string{ALL}, nil
		}

		return subjectFilterEmpty(), err
	}
}

// subjectFilterDirect is a method of the SubjectFilterEngine struct.
// It creates a SubjectFilterFunction for a direct lookup of a subject.
func (engine *SubjectFilter) subjectFilterDirectRelation(
	request *base.PermissionLookupSubjectRequest,
) SubjectFilterFunction {
	// The returned SubjectFilterFunction first queries relationships of the entity in the request using the request's permission.
	// Then it separates the subjects into foundedUsers and foundedUserSets depending on the relation and exclusion flag.
	// Finally, it adds the ids of all foundedUsers and foundedUserSets to the response.
	return func(ctx context.Context) (result []string, err error) {
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
		cti, err = storageContext.NewContextualTuples(request.GetContext().GetTuples()...).QueryRelationships(filter, database.NewCursorPagination(database.Cursor(request.GetContinuousToken()), database.Sort("subject_id")))
		if err != nil {
			return subjectFilterEmpty(), err
		}

		// Query the relationships for the entity in the request.
		// TupleFilter helps in filtering out the relationships for a specific entity and a permission.
		var rit *database.TupleIterator
		rit, err = engine.dataReader.QueryRelationships(ctx, request.GetTenantId(), filter, request.GetMetadata().GetSnapToken(), database.NewCursorPagination(database.Cursor(request.GetContinuousToken()), database.Sort("subject_id")))
		if err != nil {
			return subjectFilterEmpty(), err
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
		var re []string

		// Add the ids of all foundedUsers to the response.
		for _, s := range foundedUsers.GetSubjects() {
			re = append(re, s.GetId())
		}

		// Iterate over the foundedUserSets.
		fi := foundedUserSets.CreateSubjectIterator()
		for fi.HasNext() {
			// For each user set, make a SubjectFilter request.
			subject := fi.GetNext()
			resp, err := engine.SubjectFilter(ctx, &base.PermissionLookupSubjectRequest{
				TenantId: request.GetTenantId(),
				Entity: &base.Entity{
					Type: subject.GetType(),
					Id:   subject.GetId(),
				},
				Permission:       subject.GetRelation(),
				SubjectReference: request.GetSubjectReference(),
				Metadata:         request.GetMetadata(),
				Context:          request.GetContext(),
				ContinuousToken:  request.GetContinuousToken(),
			})
			if err != nil {
				return subjectFilterEmpty(), err
			}

			// Add the subject ids from the response to the final response.
			re = append(re, resp...)
		}

		// Return the final response with all subject ids and nil error.
		return re, nil
	}
}

// setChild generates a SubjectFilterFunction by applying a SubjectFilterCombiner
// to a set of child permission lookups, given a request and a list of Child objects.
func (engine *SubjectFilter) setChild(
	ctx context.Context, // The context for carrying out the operation
	request *base.PermissionLookupSubjectRequest, // The request containing parameters for lookup
	children []*base.Child, // The children of a particular node in the permission schema
	combiner SubjectFilterCombiner, // A function to combine the results from multiple lookup functions
) SubjectFilterFunction {
	var functions []SubjectFilterFunction // Array of functions to store lookup functions for each child

	// Iterating over the list of child nodes
	for _, child := range children {
		// Check the type of the child node and generate the appropriate lookup function
		switch child.GetType().(type) {
		case *base.Child_Rewrite:
			// If child is of type Rewrite, generate a lookup function using subjectFilterRewrite
			functions = append(functions, engine.subjectFilterRewrite(ctx, request, child.GetRewrite()))
		case *base.Child_Leaf:
			// If child is of type Leaf, generate a lookup function using subjectFilterLeaf
			functions = append(functions, engine.subjectFilterLeaf(request, child.GetLeaf()))
		default:
			// If child type is not recognised, return a failed lookup function with error
			return subjectFilterFail(errors.New("set child error"))
		}
	}

	// Return a function that when executed, applies the SubjectFilterCombiner on the generated lookup functions
	return func(ctx context.Context) ([]string, error) {
		return combiner(ctx, functions, engine.concurrencyLimit)
	}
}

// subjectFilterComputedUserSet is a method for the SubjectFilterEngine struct.
// It creates a SubjectFilterFunction for a specific ComputedUserSet.
func (engine *SubjectFilter) subjectFilterComputedUserSet(
	request *base.PermissionLookupSubjectRequest,
	cu *base.ComputedUserSet,
) SubjectFilterFunction {
	// This function creates a new SubjectFilterFunction. This function is defined to call SubjectFilter on the engine,
	// with a new PermissionSubjectFilterRequest based on the current request and the ComputedUserSet.
	return func(ctx context.Context) ([]string, error) {
		return engine.SubjectFilter(ctx, &base.PermissionLookupSubjectRequest{
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
			Context: request.GetContext(),

			ContinuousToken: request.GetContinuousToken(),
		})
	}
}

// subjectFilterTupleToUserSet is a method of the SubjectFilterEngine struct.
// It creates a SubjectFilterFunction for a specific TupleToUserSet.
func (engine *SubjectFilter) subjectFilterTupleToUserSet(
	request *base.PermissionLookupSubjectRequest,
	ttu *base.TupleToUserSet,
) SubjectFilterFunction {
	// The returned SubjectFilterFunction first queries relationships of the entity in the request using the relation from the TupleToUserSet.
	// For each subject in the relationships, it generates a SubjectFilterFunction by treating it as a ComputedUserSet.
	// Finally, it combines all these functions into a single response.
	return func(ctx context.Context) ([]string, error) {
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
		cti, err := storageContext.NewContextualTuples(request.GetContext().GetTuples()...).QueryRelationships(filter, database.NewCursorPagination())
		if err != nil {
			// If an error occurs during querying, an empty response with the error is returned.
			return subjectFilterEmpty(), err
		}

		// Query the relationships for the entity in the request.
		// TupleFilter helps in filtering out the relationships for a specific entity and a permission.
		var rit *database.TupleIterator
		rit, err = engine.dataReader.QueryRelationships(ctx, request.GetTenantId(), filter, request.GetMetadata().GetSnapToken(), database.NewCursorPagination())
		if err != nil {
			// If an error occurs during querying, an empty response with the error is returned.
			return subjectFilterEmpty(), err
		}

		// Create a new UniqueTupleIterator from the two TupleIterators.
		// NewUniqueTupleIterator() ensures that the iterator only returns unique tuples.
		it := database.NewUniqueTupleIterator(rit, cti)

		// Initialize the slice for storing SubjectFilterFunction instances.
		var subjectFilterFunctions []SubjectFilterFunction
		for it.HasNext() {
			// For each subject in the relationships, create a SubjectFilterFunction by treating it as a ComputedUserSet.
			// Get the next tuple's subject.
			next, ok := it.GetNext()
			if !ok {
				break
			}
			subject := next.GetSubject()

			subjectFilterFunctions = append(subjectFilterFunctions, engine.subjectFilterComputedUserSet(&base.PermissionLookupSubjectRequest{
				TenantId: request.GetTenantId(),
				Entity: &base.Entity{
					Type: subject.GetType(),
					Id:   subject.GetId(),
				},
				Permission:       subject.GetRelation(),
				SubjectReference: request.GetSubjectReference(),
				Metadata:         request.GetMetadata(),
				Context:          request.GetContext(),
				ContinuousToken:  request.GetContinuousToken(),
			}, ttu.GetComputed()))
		}

		// Combine all the subjectFilterFunctions into a single response using the subjectFilterUnion method.
		return subjectFilterUnion(ctx, subjectFilterFunctions, engine.concurrencyLimit)
	}
}

// subjectFilterUnion function is used to find the union of subjects
// returned by executing multiple lookup subject functions concurrently.
func subjectFilterUnion(ctx context.Context, functions []SubjectFilterFunction, limit int) ([]string, error) {
	// If there are no functions to be executed, return an empty response
	if len(functions) == 0 {
		return subjectFilterEmpty(), nil
	}

	// Create channels to handle asynchronous responses from lookup functions.
	decisionChan := make(chan SubjectFilterResponse, len(functions))

	// Create a cancellable context
	cancelCtx, cancel := context.WithCancel(ctx)

	// Run the functions concurrently
	clean := subjectFiltersRun(cancelCtx, functions, decisionChan, limit)

	// Ensure resources are cleaned up correctly after functions are done executing
	defer func() {
		cancel()
		clean()
		close(decisionChan)
	}()

	var res []string
	var excludedIds []string
	encounteredWildcard := false

	// For each function, collect results
	for i := 0; i < len(functions); i++ {
		select {
		case d := <-decisionChan:
			// If an error occurred in the function, return the error
			if d.err != nil {
				return subjectFilterEmpty(), d.err
			}
			// Check if the response contains "<>" or "<> plus additional IDs"
			if containsWildcard(d.resp) {
				encounteredWildcard = true
				// Collect any additional IDs alongside "<>", treat them as exclusions
				for _, id := range d.resp {
					if id != ALL && !slices.Contains(excludedIds, id) {
						excludedIds = append(excludedIds, id)
					}
				}
				continue
			}
			// Otherwise, append the unique subject IDs to the result
			for _, id := range d.resp {
				if !encounteredWildcard && !slices.Contains(res, id) {
					res = append(res, id)
				} else if encounteredWildcard {
					// If wildcard was encountered, treat these as exclusions
					if !slices.Contains(excludedIds, id) {
						excludedIds = append(excludedIds, id)
					}
				}
			}
		case <-ctx.Done():
			return subjectFilterEmpty(), errors.New(base.ErrorCode_ERROR_CODE_CANCELLED.String())
		}
	}

	// If a wildcard ("<>") was encountered, include everything except the exclusions.
	if encounteredWildcard {
		// Return the wildcard result with exclusions as "<>-1,2,3"
		finalRes := []string{ALL}
		if len(excludedIds) > 0 {
			exclusions := "-" + strings.Join(excludedIds, ",")
			finalRes = []string{finalRes[0] + exclusions}
		}
		return finalRes, nil
	}

	// Return the union of all IDs collected
	return res, nil
}

// subjectFilterIntersection function is used to find the intersection of subjects
// returned by executing multiple lookup subject functions concurrently.
func subjectFilterIntersection(ctx context.Context, functions []SubjectFilterFunction, limit int) ([]string, error) {
	// If there are no functions to be executed, return an empty response
	if len(functions) == 0 {
		return subjectFilterEmpty(), nil
	}

	// Create channels to handle asynchronous responses from lookup functions.
	decisionChan := make(chan SubjectFilterResponse, len(functions))

	// Create a cancellable context
	cancelCtx, cancel := context.WithCancel(ctx)

	// Run the functions concurrently
	clean := subjectFiltersRun(cancelCtx, functions, decisionChan, limit)

	// Ensure resources are cleaned up correctly after functions are done executing
	defer func() {
		cancel()
		clean()
		close(decisionChan)
	}()

	var commonIds []string
	initialized := false
	encounteredWildcard := false
	var excludedIds []string

	// For each function, collect results
	for i := 0; i < len(functions); i++ {
		select {
		case d := <-decisionChan:
			// If an error occurred in the function, return the error
			if d.err != nil {
				return subjectFilterEmpty(), d.err
			}
			// If "<>" is encountered, handle any exclusions that come with it
			if containsWildcard(d.resp) {
				encounteredWildcard = true
				for _, id := range d.resp {
					if id != ALL && !slices.Contains(excludedIds, id) {
						excludedIds = append(excludedIds, id)
					}
				}
				continue
			}
			// If it's the first function, initialize the common IDs list
			if !initialized {
				commonIds = append(commonIds, d.resp...)
				initialized = true
			} else {
				// Intersect the current common IDs with the new set of IDs
				commonIds = intersect(commonIds, d.resp)
			}
		case <-ctx.Done():
			return subjectFilterEmpty(), errors.New(base.ErrorCode_ERROR_CODE_CANCELLED.String())
		}
	}

	// If wildcard was encountered, we exclude the IDs in `excludedIds`
	if encounteredWildcard {
		if len(commonIds) == 0 {
			return []string{}, nil
		}

		// Exclude IDs from commonIds that are in excludedIds
		filteredCommonIds := []string{}
		for _, id := range commonIds {
			if !slices.Contains(excludedIds, id) {
				filteredCommonIds = append(filteredCommonIds, id)
			}
		}
		return filteredCommonIds, nil
	}

	// Return the final intersection of IDs
	return commonIds, nil
}

// subjectFilterExclusion is a function that checks for a subject's exclusion
// among different lookup subject functions and returns the subjects not found
// in the excluded list.
func subjectFilterExclusion(ctx context.Context, functions []SubjectFilterFunction, limit int) ([]string, error) {
	// Ensure that we have more than one function for exclusion comparison
	if len(functions) <= 1 {
		return subjectFilterEmpty(), errors.New(base.ErrorCode_ERROR_CODE_EXCLUSION_REQUIRES_MORE_THAN_ONE_FUNCTION.String())
	}

	// Create channels to handle asynchronous responses from lookup functions
	leftDecisionChan := make(chan SubjectFilterResponse, 1)
	decisionChan := make(chan SubjectFilterResponse, len(functions)-1)

	// Create a cancellable context
	cancelCtx, cancel := context.WithCancel(ctx)

	// Use a WaitGroup to ensure all goroutines have completed
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		// Run the first lookup function in a goroutine
		result, err := functions[0](cancelCtx)
		// Send the result to the leftDecisionChan
		leftDecisionChan <- SubjectFilterResponse{
			resp: result,
			err:  err,
		}
		wg.Done()
	}()

	// Run the rest of the lookup functions in parallel
	clean := subjectFiltersRun(cancelCtx, functions[1:], decisionChan, limit-1)

	// Ensure resources are cleaned up properly
	defer func() {
		cancel()
		clean()
		close(decisionChan)
		wg.Wait()
		close(leftDecisionChan)
	}()

	var leftIds []string
	var res []string

	// Retrieve the result of the first lookup function (base set for exclusion)
	select {
	case left := <-leftDecisionChan:
		// If there's an error, return it
		if left.err != nil {
			return subjectFilterEmpty(), left.err
		}
		// Get the list of subject IDs from the first lookup function
		leftIds = left.resp

	// If the context is cancelled, return a cancellation error
	case <-ctx.Done():
		return subjectFilterEmpty(), errors.New(base.ErrorCode_ERROR_CODE_CANCELLED.String())
	}

	var exIds []string
	encounteredRightWildcard := false

	// Retrieve the results of the remaining lookup functions (to use for exclusion)
	for i := 0; i < len(functions)-1; i++ {
		select {
		case d := <-decisionChan:
			// If there's an error, return it
			if d.err != nil {
				return subjectFilterEmpty(), d.err
			}
			// If "<>" is encountered in the right functions, mark the wildcard for exclusion
			if containsWildcard(d.resp) {
				encounteredRightWildcard = true
				// If a wildcard is encountered, no need to process further, as all should be excluded
				break
			}
			// Otherwise, append the IDs to the exclusion list
			exIds = append(exIds, d.resp...)
		case <-ctx.Done():
			return subjectFilterEmpty(), errors.New(base.ErrorCode_ERROR_CODE_CANCELLED.String())
		}
	}

	// Handle wildcard logic in the rightIds (i.e., if "<>" is in rightIds)
	if encounteredRightWildcard {
		// If any of the right-side functions contains "<>", return an empty result (exclude all)
		return []string{}, nil
	}

	// Handle wildcard logic in the leftIds (i.e., if "<>" is in leftIds)
	if containsWildcard(leftIds) {
		// If left side contains "<>", return it with exclusions
		if len(exIds) > 0 {
			// Format the result as "<>-id1,id2,id3" where id1, id2, id3 are the exclusions
			exclusions := "-" + strings.Join(exIds, ",")
			return []string{ALL + exclusions}, nil
		}
		// If no exclusions are provided, return just "<>"
		return []string{ALL}, nil
	}

	// Normal case (no wildcard): Exclude the IDs from `leftIds` that are in `exIds`
	for _, id := range leftIds {
		// Only include IDs from `leftIds` that are NOT in `exIds`
		if !slices.Contains(exIds, id) {
			res = append(res, id)
		}
	}

	// Return the filtered result after exclusion
	return res, nil
}

// subjectFiltersRun executes a list of SubjectFilterFunction concurrently, with a limit on the maximum number of concurrent executions.
// It sends the responses from these functions to the provided decisionChan.
// It returns a function that blocks until all SubjectFilterFunctions have completed.
func subjectFiltersRun(ctx context.Context, functions []SubjectFilterFunction, decisionChan chan<- SubjectFilterResponse, limit int) func() {
	// cl is a channel used to control the concurrency level.
	// Its capacity is set to the limit argument to restrict the maximum number of concurrent goroutines.
	cl := make(chan struct{}, limit)

	// wg is a WaitGroup that is used to wait until all SubjectFilterFunctions have completed.
	var wg sync.WaitGroup

	// subjectFilter is a function that executes a SubjectFilterFunction and sends its result to the decisionChan.
	subjectFilter := func(child SubjectFilterFunction) {
		result, err := child(ctx)
		decisionChan <- SubjectFilterResponse{
			resp: result,
			err:  err,
		}
		// Release a spot in the concurrency limit channel and notify the WaitGroup that a function has completed.
		<-cl
		wg.Done()
	}

	// Start a goroutine that loops through all the SubjectFilterFunctions and starts a new goroutine for each one, up to the concurrency limit.
	wg.Add(1)
	go func() {
	run:
		for _, fun := range functions {
			child := fun
			select {
			// If there is room in the concurrency limit channel, start a new goroutine.
			case cl <- struct{}{}:
				wg.Add(1)
				go subjectFilter(child)
			// If the context is done, break out of the loop.
			case <-ctx.Done():
				break run
			}
		}
		// Notify the WaitGroup that this function is done.
		wg.Done()
	}()

	// Return a function that waits until all SubjectFilterFunctions have completed, then closes the concurrency limit channel.
	return func() {
		wg.Wait()
		close(cl)
	}
}

// subjectFilterFail returns a function conforming to the SubjectFilterFunction signature.
// The returned function, when invoked, will always return an error that is provided as an argument to subjectFilterFail.
// The SubjectFilterFunction could be used as a mock function in unit tests to simulate failure scenarios.
func subjectFilterFail(err error) SubjectFilterFunction {
	return func(ctx context.Context) ([]string, error) {
		// We return a default PermissionSubjectFilterResponse with an empty slice of subject IDs
		// and the error that was passed into the subjectFilterFail function.
		return subjectFilterEmpty(), err
	}
}

// empty is a helper function that returns a pointer to a PermissionSubjectFilterResponse
// with an empty SubjectIds slice.
// This function could be used in tests where a default or "empty" PermissionSubjectFilterResponse is needed.
func subjectFilterEmpty() []string {
	return []string{}
}

// getIntersection finds the common elements (intersection) across multiple slices of ids
// Helper function to intersect two slices
func intersect(a, b []string) []string {
	set := make(map[string]bool)
	var result []string

	// Add all elements from `a` to the set
	for _, v := range a {
		set[v] = true
	}

	// Check elements in `b` and only keep those that are in `a`
	for _, v := range b {
		if set[v] {
			result = append(result, v)
		}
	}

	return result
}

// containsWildcard checks if the wildcard "<>" is present in the id list
func containsWildcard(ids []string) bool {
	for _, id := range ids {
		if id == ALL {
			return true
		}
	}
	return false
}
