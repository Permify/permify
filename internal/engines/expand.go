package engines

import (
	"context"
	"errors"

	"google.golang.org/protobuf/types/known/anypb"

	"github.com/Permify/permify/internal/schema"
	"github.com/Permify/permify/internal/storage"
	storageContext "github.com/Permify/permify/internal/storage/context"
	"github.com/Permify/permify/pkg/database"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
)

// ExpandEngine - This comment is describing a type called ExpandEngine. The ExpandEngine type contains two fields: schemaReader,
// which is a storage.SchemaReader object, and relationshipReader, which is a storage.RelationshipReader object.
// The ExpandEngine type is used to expand permission scopes based on a given user ID and a set of permission requirements.
type ExpandEngine struct {
	// schemaReader is responsible for reading schema information
	schemaReader storage.SchemaReader
	// relationshipReader is responsible for reading relationship information
	dataReader storage.DataReader
}

// NewExpandEngine - This function creates a new instance of ExpandEngine by taking a SchemaReader and a RelationshipReader as
// parameters and returning a pointer to the created instance. The SchemaReader is used to read schema definitions, while the
// RelationshipReader is used to read relationship definitions.
func NewExpandEngine(sr storage.SchemaReader, rr storage.DataReader) *ExpandEngine {
	return &ExpandEngine{
		schemaReader: sr,
		dataReader:   rr,
	}
}

// Expand - This is the Run function of the ExpandEngine type, which takes a context, a PermissionExpandRequest,
// and returns a PermissionExpandResponse and an error.
// The function begins by starting a new OpenTelemetry span, with the name "permissions.expand.execute".
// It then checks if a snap token and schema version are included in the request. If not, it retrieves the head
// snapshot and head schema version, respectively, from the appropriate repository.
//
// Finally, the function calls the expand function of the ExpandEngine type with the context, PermissionExpandRequest,
// and false value, and returns the resulting PermissionExpandResponse and error. If there is an error, the span records
// the error and sets the status to indicate an error.
func (engine *ExpandEngine) Expand(ctx context.Context, request *base.PermissionExpandRequest) (response *base.PermissionExpandResponse, err error) {
	resp := engine.expand(ctx, request)
	if resp.Err != nil {
		return nil, resp.Err
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
type ExpandCombiner func(ctx context.Context, entity *base.Entity, permission string, arguments []*base.Argument, functions []ExpandFunction) ExpandResponse

// 'expand' is a method of ExpandEngine which takes a context and a PermissionExpandRequest,
// and returns an ExpandResponse. This function is the main entry point for expanding permissions.
func (engine *ExpandEngine) expand(ctx context.Context, request *base.PermissionExpandRequest) ExpandResponse {
	var fn ExpandFunction // Declare an ExpandFunction variable.

	// Read entity definition based on the entity type in the request.
	en, _, err := engine.schemaReader.ReadEntityDefinition(ctx, request.GetTenantId(), request.GetEntity().GetType(), request.GetMetadata().GetSchemaVersion())
	if err != nil {
		// If an error occurred while reading entity definition, return an ExpandResponse with the error.
		return ExpandResponse{Err: err}
	}

	var tor base.EntityDefinition_Reference
	// Get the type of reference by name in the entity definition.
	tor, _ = schema.GetTypeOfReferenceByNameInEntityDefinition(en, request.GetPermission())

	// Depending on the type of reference, execute different branches of code.
	switch tor {
	case base.EntityDefinition_REFERENCE_PERMISSION:
		// If the reference is a permission, get the permission by name from the entity definition.
		permission, err := schema.GetPermissionByNameInEntityDefinition(en, request.GetPermission())
		if err != nil {
			// If an error occurred while getting the permission, return an ExpandResponse with the error.
			return ExpandResponse{Err: err}
		}

		// Get the child of the permission.
		child := permission.GetChild()
		// If the child has a rewrite rule, use the 'expandRewrite' method.
		if child.GetRewrite() != nil {
			fn = engine.expandRewrite(ctx, request, child.GetRewrite())
		} else {
			// If the child doesn't have a rewrite rule, use the 'expandLeaf' method.
			fn = engine.expandLeaf(request, child.GetLeaf())
		}
	case base.EntityDefinition_REFERENCE_ATTRIBUTE:
		// If the reference is an attribute, use the 'expandDirectAttribute' method.
		fn = engine.expandDirectAttribute(request)
	case base.EntityDefinition_REFERENCE_RELATION:
		// If the reference is a relation, use the 'expandDirectRelation' method.
		fn = engine.expandDirectRelation(request)
	default:
		// If the reference is neither permission, attribute, nor relation, use the 'expandCall' method.
		fn = engine.expandDirectCall(request)
	}

	if fn == nil {
		// If no expand function was set, return an ExpandResponse with an error.
		return ExpandResponse{Err: errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_CHILD_KIND.String())}
	}

	// Execute the expand function with the root context.
	return expandRoot(ctx, fn)
}

// 'expandRewrite' is a method of ExpandEngine which takes a context, a PermissionExpandRequest,
// and a rewrite rule. It returns an ExpandFunction which represents the expansion of the rewrite rule.
func (engine *ExpandEngine) expandRewrite(ctx context.Context, request *base.PermissionExpandRequest, rewrite *base.Rewrite) ExpandFunction {
	// The rewrite rule can have different operations: UNION, INTERSECTION, EXCLUSION.
	// Depending on the operation type, it calls different methods.
	switch rewrite.GetRewriteOperation() {
	// If the operation is UNION, call the 'setChild' method with 'expandUnion' as the expand function.
	case *base.Rewrite_OPERATION_UNION.Enum():
		return engine.setChild(ctx, request, rewrite.GetChildren(), expandUnion)

	// If the operation is INTERSECTION, call the 'setChild' method with 'expandIntersection' as the expand function.
	case *base.Rewrite_OPERATION_INTERSECTION.Enum():
		return engine.setChild(ctx, request, rewrite.GetChildren(), expandIntersection)

	// If the operation is EXCLUSION, call the 'setChild' method with 'expandExclusion' as the expand function.
	case *base.Rewrite_OPERATION_EXCLUSION.Enum():
		return engine.setChild(ctx, request, rewrite.GetChildren(), expandExclusion)

	// If the operation is not any of the defined types, return an error.
	default:
		return expandFail(errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_CHILD_TYPE.String()))
	}
}

// 'expandLeaf' is a method of ExpandEngine which takes a PermissionExpandRequest and a leaf object.
// It returns an ExpandFunction, a function which performs the action associated with a given leaf type.
func (engine *ExpandEngine) expandLeaf(
	request *base.PermissionExpandRequest,
	leaf *base.Leaf,
) ExpandFunction {
	// The leaf object can have different types, each associated with a different expansion function.
	// Depending on the type of the leaf, different expansion functions are returned.
	switch op := leaf.GetType().(type) {

	// If the type of the leaf is TupleToUserSet, the method 'expandTupleToUserSet' is called.
	case *base.Leaf_TupleToUserSet:
		return engine.expandTupleToUserSet(request, op.TupleToUserSet)

	// If the type of the leaf is ComputedUserSet, the method 'expandComputedUserSet' is called.
	case *base.Leaf_ComputedUserSet:
		return engine.expandComputedUserSet(request, op.ComputedUserSet)

	// If the type of the leaf is ComputedAttribute, the method 'expandComputedAttribute' is called.
	case *base.Leaf_ComputedAttribute:
		return engine.expandComputedAttribute(request, op.ComputedAttribute)

	// If the type of the leaf is Call, the method 'expandCall' is called.
	case *base.Leaf_Call:
		return engine.expandCall(request, op.Call)

	// If the leaf type is none of the above, an error is returned.
	default:
		return expandFail(errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_CHILD_TYPE.String()))
	}
}

// 'setChild' is a method of the ExpandEngine struct, which aims to create and set an ExpandFunction
// for each child in the provided children array. These functions are derived from the type of child
// (either a rewrite or a leaf) and are passed to a provided 'combiner' function.
func (engine *ExpandEngine) setChild(
	ctx context.Context,
	request *base.PermissionExpandRequest,
	children []*base.Child,
	combiner ExpandCombiner,
) ExpandFunction {
	// Declare an array to hold ExpandFunctions.
	var functions []ExpandFunction

	// Iterate through each child in the provided array.
	for _, child := range children {
		// Based on the type of child, append the appropriate ExpandFunction to the functions array.
		switch child.GetType().(type) {

		// If the child is a 'rewrite' type, use the 'expandRewrite' function.
		case *base.Child_Rewrite:
			functions = append(functions, engine.expandRewrite(ctx, request, child.GetRewrite()))

		// If the child is a 'leaf' type, use the 'expandLeaf' function.
		case *base.Child_Leaf:
			functions = append(functions, engine.expandLeaf(request, child.GetLeaf()))

		// If the child type is not recognized, return an error.
		default:
			return expandFail(errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_CHILD_KIND.String()))
		}
	}

	// Return an ExpandFunction that will pass the prepared functions to the combiner when called.
	return func(ctx context.Context, resultChan chan<- ExpandResponse) {
		resultChan <- combiner(ctx, request.GetEntity(), request.GetPermission(), request.GetArguments(), functions)
	}
}

// expandDirectRelation Expands the target permission for direct subjects, i.e., those whose subject entity has the same type as the target
// entity type and whose subject relation is not an ellipsis. It queries the relationship store for relationships that
// match the target permission, and for each matching relationship, calls Expand on the corresponding subject entity
// and relation. If there are no matching relationships, it returns a leaf node of the expand tree containing the direct
// subjects. If there are matching relationships, it computes the union of the results of calling Expand on each matching
// relationship's subject entity and relation, and attaches the resulting expand nodes as children of a union node in
// the expand tree. Finally, it returns the top-level expand node.
func (engine *ExpandEngine) expandDirectRelation(request *base.PermissionExpandRequest) ExpandFunction {
	return func(ctx context.Context, expandChan chan<- ExpandResponse) {
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
		cti, err := storageContext.NewContextualTuples(request.GetContext().GetTuples()...).QueryRelationships(filter, database.NewCursorPagination())
		if err != nil {
			// If an error occurred while querying, return a "denied" response and the error.
			expandChan <- expandFailResponse(err)
			return
		}

		// Query the relationships for the entity in the request.
		// TupleFilter helps in filtering out the relationships for a specific entity and a permission.
		var rit *database.TupleIterator
		rit, err = engine.dataReader.QueryRelationships(ctx, request.GetTenantId(), filter, request.GetMetadata().GetSnapToken(), database.NewCursorPagination())
		if err != nil {
			expandChan <- expandFailResponse(err)
			return
		}

		// Create a new UniqueTupleIterator from the two TupleIterators.
		// NewUniqueTupleIterator() ensures that the iterator only returns unique tuples.
		it := database.NewUniqueTupleIterator(rit, cti)

		foundedUserSets := database.NewSubjectCollection()
		foundedUsers := database.NewSubjectCollection()

		// it represents an iterator over some collection of subjects.
		for it.HasNext() {
			// Get the next tuple's subject.
			next, ok := it.GetNext()
			if !ok {
				break
			}
			subject := next.GetSubject()

			if tuple.IsDirectSubject(subject) || subject.GetRelation() == tuple.ELLIPSIS {
				foundedUsers.Add(subject)
			} else {
				foundedUserSets.Add(subject)
			}
		}

		// If there are no founded user sets, create and send an expand response.
		if len(foundedUserSets.GetSubjects()) == 0 {
			expandChan <- ExpandResponse{
				Response: &base.PermissionExpandResponse{
					Tree: &base.Expand{
						Entity:     request.GetEntity(),
						Permission: request.GetPermission(),
						Arguments:  request.GetArguments(),
						Node: &base.Expand_Leaf{
							Leaf: &base.ExpandLeaf{
								Type: &base.ExpandLeaf_Subjects{
									Subjects: &base.Subjects{
										Subjects: foundedUsers.GetSubjects(),
									},
								},
							},
						},
					},
				},
			}
			return
		}

		// Define a slice of ExpandFunction.
		var expandFunctions []ExpandFunction

		// Create an iterator for the foundedUserSets.
		si := foundedUserSets.CreateSubjectIterator()

		// Iterate over the foundedUserSets.
		for si.HasNext() {
			sub := si.GetNext()
			// For each subject, append a new function to the expandFunctions slice.
			expandFunctions = append(expandFunctions, func(ctx context.Context, resultChan chan<- ExpandResponse) {
				resultChan <- engine.expand(ctx, &base.PermissionExpandRequest{
					TenantId: request.GetTenantId(),
					Entity: &base.Entity{
						Type: sub.GetType(),
						Id:   sub.GetId(),
					},
					Permission: sub.GetRelation(),
					Metadata:   request.GetMetadata(),
					Context:    request.GetContext(),
				})
			})
		}

		// Use the expandUnion function to process the expandFunctions.
		result := expandUnion(ctx, request.GetEntity(), request.GetPermission(), request.GetArguments(), expandFunctions)

		// If an error occurred, send a failure response and return.
		if result.Err != nil {
			expandChan <- expandFailResponse(result.Err)
			return
		}

		// Get the Expand field from the response tree.
		expand := result.Response.GetTree().GetExpand()

		// Add a new child to the Expand field.
		expand.Children = append(expand.Children, &base.Expand{
			Entity:     request.GetEntity(),
			Permission: request.GetPermission(),
			Arguments:  request.GetArguments(),
			Node: &base.Expand_Leaf{
				Leaf: &base.ExpandLeaf{
					Type: &base.ExpandLeaf_Subjects{
						Subjects: &base.Subjects{
							Subjects: append(foundedUsers.GetSubjects(), foundedUserSets.GetSubjects()...),
						},
					},
				},
			},
		})

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
func (engine *ExpandEngine) expandTupleToUserSet(
	request *base.PermissionExpandRequest,
	ttu *base.TupleToUserSet,
) ExpandFunction {
	return func(ctx context.Context, expandChan chan<- ExpandResponse) {
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
			expandChan <- expandFailResponse(err)
			return
		}

		// Use the filter to query for relationships in the database.
		// relationshipReader.QueryRelationships() uses the filter to find and return matching relationships.
		rit, err := engine.dataReader.QueryRelationships(ctx, request.GetTenantId(), filter, request.GetMetadata().GetSnapToken(), database.NewCursorPagination())
		if err != nil {
			expandChan <- expandFailResponse(err)
			return
		}

		// Create a new UniqueTupleIterator from the two TupleIterators.
		// NewUniqueTupleIterator() ensures that the iterator only returns unique tuples.
		it := database.NewUniqueTupleIterator(rit, cti)

		var expandFunctions []ExpandFunction
		for it.HasNext() {
			// Get the next tuple's subject.
			next, ok := it.GetNext()
			if !ok {
				break
			}
			subject := next.GetSubject()

			expandFunctions = append(expandFunctions, engine.expandComputedUserSet(&base.PermissionExpandRequest{
				TenantId: request.GetTenantId(),
				Entity: &base.Entity{
					Type: subject.GetType(),
					Id:   subject.GetId(),
				},
				Permission: subject.GetRelation(),
				Metadata:   request.GetMetadata(),
				Context:    request.GetContext(),
			}, ttu.GetComputed()))
		}

		expandChan <- expandUnion(
			ctx,
			request.GetEntity(),
			ttu.GetTupleSet().GetRelation(),
			request.GetArguments(),
			expandFunctions,
		)
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
//
// Returns:
//   - ExpandFunction that sends the expanded user set to the provided channel
func (engine *ExpandEngine) expandComputedUserSet(
	request *base.PermissionExpandRequest,
	cu *base.ComputedUserSet,
) ExpandFunction {
	return func(ctx context.Context, resultChan chan<- ExpandResponse) {
		resultChan <- engine.expand(ctx, &base.PermissionExpandRequest{
			TenantId: request.GetTenantId(),
			Entity: &base.Entity{
				Type: request.GetEntity().GetType(),
				Id:   request.GetEntity().GetId(),
			},
			Permission: cu.GetRelation(),
			Metadata:   request.GetMetadata(),
			Context:    request.GetContext(),
		})
	}
}

// The function 'expandDirectAttribute' is a method on the ExpandEngine struct.
// It returns an ExpandFunction which is a type alias for a function that takes a context and an expand response channel as parameters.
func (engine *ExpandEngine) expandDirectAttribute(
	request *base.PermissionExpandRequest, // the request object containing information necessary for the expansion
) ExpandFunction { // returns an ExpandFunction
	return func(ctx context.Context, expandChan chan<- ExpandResponse) { // defining the returned function

		var err error // variable to store any error occurred during the process

		// Defining a filter to get the attribute based on the entity type and ID and the permission requested.
		filter := &base.AttributeFilter{
			Entity: &base.EntityFilter{
				Type: request.GetEntity().GetType(),
				Ids:  []string{request.GetEntity().GetId()},
			},
			Attributes: []string{request.GetPermission()},
		}

		var val *base.Attribute // variable to hold the attribute value

		// Attempt to get the attribute using the defined filter.
		val, err = storageContext.NewContextualAttributes(request.GetContext().GetAttributes()...).QuerySingleAttribute(filter)
		// If there's an error in getting the attribute, send a failure response through the channel and return from the function.
		if err != nil {
			expandChan <- expandFailResponse(err)
			return
		}

		// If no attribute was found, attempt to query it directly from the data reader.
		if val == nil {
			val, err = engine.dataReader.QuerySingleAttribute(ctx, request.GetTenantId(), filter, request.GetMetadata().GetSnapToken())
			if err != nil {
				expandChan <- expandFailResponse(err)
				return
			}
		}

		// If the attribute is still nil, create a new attribute with a false value.
		if val == nil {
			val = &base.Attribute{
				Entity: &base.Entity{
					Type: request.GetEntity().GetType(),
					Id:   request.GetEntity().GetId(),
				},
				Attribute: request.GetPermission(),
			}
			val.Value, err = anypb.New(&base.BooleanValue{Data: false})
			if err != nil {
				expandChan <- expandFailResponse(err)
				return
			}
		}

		// Send an ExpandResponse containing the permission expansion response with the attribute value through the channel.
		expandChan <- ExpandResponse{
			Response: &base.PermissionExpandResponse{
				Tree: &base.Expand{
					Entity:     request.GetEntity(),
					Permission: request.GetPermission(),
					Arguments:  request.GetArguments(),
					Node: &base.Expand_Leaf{
						Leaf: &base.ExpandLeaf{
							Type: &base.ExpandLeaf_Value{
								Value: val.GetValue(),
							},
						},
					},
				},
			},
		}
	}
}

// expandCall returns an ExpandFunction for the given request and call.
// The returned function, when executed, sends the expanded permission result
// to the provided result channel.
func (engine *ExpandEngine) expandCall(
	request *base.PermissionExpandRequest,
	call *base.Call,
) ExpandFunction {
	return func(ctx context.Context, resultChan chan<- ExpandResponse) {
		resultChan <- engine.expand(ctx, &base.PermissionExpandRequest{
			TenantId: request.GetTenantId(),
			Entity: &base.Entity{
				Type: request.GetEntity().GetType(),
				Id:   request.GetEntity().GetId(),
			},
			Permission: call.GetRuleName(),
			Metadata:   request.GetMetadata(),
			Context:    request.GetContext(),
			Arguments:  call.GetArguments(),
		})
	}
}

// The function 'expandCall' is a method on the ExpandEngine struct.
// It takes a PermissionExpandRequest and a Call as parameters and returns an ExpandFunction.
func (engine *ExpandEngine) expandDirectCall(
	request *base.PermissionExpandRequest, // The request object containing information necessary for the expansion.
) ExpandFunction { // The function returns an ExpandFunction.
	return func(ctx context.Context, expandChan chan<- ExpandResponse) { // defining the returned function.

		var err error // variable to store any error occurred during the process.

		var ru *base.RuleDefinition // variable to hold the rule definition.

		// Read the rule definition based on the rule name in the call.
		ru, _, err = engine.schemaReader.ReadRuleDefinition(ctx, request.GetTenantId(), request.GetPermission(), request.GetMetadata().GetSchemaVersion())
		if err != nil {
			// If there's an error in reading the rule definition, send a failure response through the channel and return from the function.
			expandChan <- expandFailResponse(err)
			return
		}

		// Prepare the arguments map to be used in the CEL evaluation
		arguments := make(map[string]*anypb.Any)

		// Prepare a slice for attributes
		attributes := make([]string, 0)

		// For each argument in the call...
		for _, arg := range request.GetArguments() {
			switch actualArg := arg.Type.(type) { // Switch on the type of the argument.
			case *base.Argument_ComputedAttribute: // If the argument is a ComputedAttribute...
				attrName := actualArg.ComputedAttribute.GetName() // get the name of the attribute.

				// Get the empty value for the attribute type.
				emptyValue, err := getEmptyProtoValueForType(ru.GetArguments()[attrName])
				if err != nil {
					expandChan <- expandFailResponse(errors.New(base.ErrorCode_ERROR_CODE_TYPE_CONVERSATION.String()))
					return
				}

				// Set the empty value in the arguments map.
				arguments[attrName] = emptyValue

				// Append the attribute name to the attributes slice.
				attributes = append(attributes, attrName)
			default:
				// If the argument type is unknown, send a failure response and return from the function.
				expandChan <- expandFailResponse(errors.New(base.ErrorCode_ERROR_CODE_INTERNAL.String()))
				return
			}
		}

		// If there are any attributes to query...
		if len(attributes) > 0 {
			// Create an AttributeFilter for the attributes.
			filter := &base.AttributeFilter{
				Entity: &base.EntityFilter{
					Type: request.GetEntity().GetType(),
					Ids:  []string{request.GetEntity().GetId()},
				},
				Attributes: attributes,
			}

			// Query the attributes from the data reader.
			ait, err := engine.dataReader.QueryAttributes(ctx, request.GetTenantId(), filter, request.GetMetadata().GetSnapToken(), database.NewCursorPagination())
			if err != nil {
				expandChan <- expandFailResponse(err)
				return
			}

			// Query the attributes from the context.
			cta, err := storageContext.NewContextualAttributes(request.GetContext().GetAttributes()...).QueryAttributes(filter, database.NewCursorPagination())
			if err != nil {
				expandChan <- expandFailResponse(err)
				return
			}

			// Create an iterator for the unique attributes.
			it := database.NewUniqueAttributeIterator(ait, cta)

			// For each unique attribute...
			for it.HasNext() {
				// Get the next attribute and its value.
				next, ok := it.GetNext()
				if !ok {
					break
				}
				// Set the attribute's value in the arguments map.
				arguments[next.GetAttribute()] = next.GetValue()
			}
		}

		// Send an ExpandResponse containing the permission expansion response with the computed arguments through the channel.
		expandChan <- ExpandResponse{
			Response: &base.PermissionExpandResponse{
				Tree: &base.Expand{
					Entity:     request.GetEntity(),
					Permission: request.GetPermission(),
					Arguments:  request.GetArguments(),
					Node: &base.Expand_Leaf{
						Leaf: &base.ExpandLeaf{
							Type: &base.ExpandLeaf_Values{
								Values: &base.Values{
									Values: arguments,
								},
							},
						},
					},
				},
			},
		}
	}
}

// 'expandComputedAttribute' is a method on the ExpandEngine struct.
// It takes a PermissionExpandRequest and a ComputedAttribute as parameters and returns an ExpandFunction.
func (engine *ExpandEngine) expandComputedAttribute(
	request *base.PermissionExpandRequest, // The request object containing necessary information for the expansion.
	ca *base.ComputedAttribute, // The computed attribute object that has the name of the attribute to be computed.
) ExpandFunction { // The function returns an ExpandFunction.
	return func(ctx context.Context, resultChan chan<- ExpandResponse) { // defining the returned function.

		// The returned function sends the result of an expansion to the result channel.
		// The expansion is performed by calling the 'expand' method of the engine with a new PermissionExpandRequest.
		// The new request is constructed using the tenant ID, entity, computed attribute name, metadata, and context from the original request.
		resultChan <- engine.expand(ctx, &base.PermissionExpandRequest{
			TenantId: request.GetTenantId(), // Tenant ID from the original request.
			Entity: &base.Entity{
				Type: request.GetEntity().GetType(), // Entity type from the original request.
				Id:   request.GetEntity().GetId(),   // Entity ID from the original request.
			},
			Permission: ca.GetName(),          // The name of the computed attribute.
			Metadata:   request.GetMetadata(), // Metadata from the original request.
			Context:    request.GetContext(),  // Context from the original request.
		})
	}
}

// 'expandOperation' is a function that takes a context, an entity, permission string,
// a slice of arguments, slice of ExpandFunctions, and an operation of type base.ExpandTreeNode_Operation.
// It returns an ExpandResponse.
func expandOperation(
	ctx context.Context, // The context of this operation, which may carry deadlines, cancellation signals, etc.
	entity *base.Entity, // The entity on which the operation will be performed.
	permission string, // The permission string required for the operation.
	arguments []*base.Argument, // A slice of arguments required for the operation.
	functions []ExpandFunction, // A slice of functions that will be used to expand the operation.
	op base.ExpandTreeNode_Operation, // The operation to be performed.
) ExpandResponse { // The function returns an ExpandResponse.

	// Initialize an empty slice of base.Expand type.
	children := make([]*base.Expand, 0, len(functions))

	// If there are no functions, return an ExpandResponse with a base.PermissionExpandResponse.
	// This response includes an empty base.Expand with the entity, permission, arguments,
	// and an ExpandTreeNode with the given operation.
	if len(functions) == 0 {
		return ExpandResponse{
			Response: &base.PermissionExpandResponse{
				Tree: &base.Expand{
					Entity:     entity,
					Permission: permission,
					Arguments:  arguments,
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

	// Create a new context that can be cancelled.
	c, cancel := context.WithCancel(ctx)
	// Defer the cancellation, so that it will be called when the function exits.
	defer func() {
		cancel()
	}()

	// Initialize an empty slice of channels which will receive ExpandResponses.
	results := make([]chan ExpandResponse, 0, len(functions))
	// For each function, create a channel and add it to the results slice.
	// Start a goroutine with the function and pass the cancelable context and the channel.
	for _, fn := range functions {
		fc := make(chan ExpandResponse, 1)
		results = append(results, fc)
		go fn(c, fc)
	}

	// For each result channel, wait for a response or for the context to be cancelled.
	for _, result := range results {
		select {
		case resp := <-result:
			// If the response contains an error, return an error response.
			if resp.Err != nil {
				return expandFailResponse(resp.Err)
			}
			// If the response does not contain an error, append the tree of the response to the children slice.
			children = append(children, resp.Response.GetTree())
		case <-ctx.Done():
			// If the context is cancelled, return an error response.
			return expandFailResponse(errors.New(base.ErrorCode_ERROR_CODE_CANCELLED.String()))
		}
	}

	// Return an ExpandResponse with a base.PermissionExpandResponse.
	// This response includes an base.Expand with the entity, permission, arguments,
	// and an ExpandTreeNode with the given operation and the children slice.
	return ExpandResponse{
		Response: &base.PermissionExpandResponse{
			Tree: &base.Expand{
				Entity:     entity,
				Permission: permission,
				Arguments:  arguments,
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
	res := make(chan ExpandResponse, 1)
	go fn(ctx, res)

	select {
	case result := <-res:
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
func expandUnion(
	ctx context.Context,
	entity *base.Entity,
	permission string,
	arguments []*base.Argument,
	functions []ExpandFunction,
) ExpandResponse {
	return expandOperation(ctx, entity, permission, arguments, functions, base.ExpandTreeNode_OPERATION_UNION)
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
func expandIntersection(
	ctx context.Context,
	entity *base.Entity,
	permission string,
	arguments []*base.Argument,
	functions []ExpandFunction,
) ExpandResponse {
	return expandOperation(ctx, entity, permission, arguments, functions, base.ExpandTreeNode_OPERATION_INTERSECTION)
}

// expandExclusion is a helper function that executes multiple ExpandFunctions in parallel and returns an ExpandResponse
// containing the expanded user set that results from the exclusion operation. The function delegates to expandOperation
// with the EXCLUSION operation. If any of the ExpandFunctions return an error, the function returns an ExpandResponse
// with the error. If the context is cancelled before all ExpandFunctions complete, the function returns an ExpandResponse
// with an error indicating that the operation was cancelled.
//
// Parameters:
//   - ctx: context.Context for the request
//   - target: EntityAndRelation containing the entity and its relation for which the exclusion is calculated
//   - functions: slice of ExpandFunctions to execute in parallel
//
// Returns:
//   - ExpandResponse containing the expanded user sets from the exclusion operation, or an error if any of the ExpandFunctions failed
func expandExclusion(
	ctx context.Context,
	entity *base.Entity,
	permission string,
	arguments []*base.Argument,
	functions []ExpandFunction,
) ExpandResponse {
	return expandOperation(ctx, entity, permission, arguments, functions, base.ExpandTreeNode_OPERATION_EXCLUSION)
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
