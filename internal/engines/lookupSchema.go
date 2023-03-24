package engines

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/exp/slices"

	otelCodes "go.opentelemetry.io/otel/codes"

	"github.com/Permify/permify/internal/repositories"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// LookupSchemaEngine represents an engine used for looking up schemas. It is used to find a schema from a given
// entity type.
type LookupSchemaEngine struct {
	// schemaReader is the repository used for reading schemas.
	schemaReader repositories.SchemaReader
}

// NewLookupSchemaEngine returns a new instance of LookupSchemaEngine with the provided SchemaReader.
func NewLookupSchemaEngine(schemaReader repositories.SchemaReader) *LookupSchemaEngine {
	return &LookupSchemaEngine{
		schemaReader: schemaReader,
	}
}

// Run method executes the lookup schema engine by taking the context and a lookup schema request and returning
// the lookup schema response or an error. It starts an OpenTelemetry span, initializes the response and determines
// the schema version for the request. It then retrieves the entity definition from the schema reader and iterates
// over each action within it, checking if the request has the necessary permissions to execute the action. The
// allowed action names are added to the response and returned, or an error is returned if encountered.
func (command *LookupSchemaEngine) Run(ctx context.Context, request *base.PermissionLookupSchemaRequest) (*base.PermissionLookupSchemaResponse, error) {
	ctx, span := tracer.Start(ctx, "permissions.lookup-schema.execute")
	defer span.End()

	var err error

	response := &base.PermissionLookupSchemaResponse{
		ActionNames: []string{},
	}

	if request.GetMetadata().GetSchemaVersion() == "" {
		request.Metadata.SchemaVersion, err = command.schemaReader.HeadVersion(ctx, request.GetTenantId())
		if err != nil {
			return response, err
		}
	}

	var en *base.EntityDefinition
	en, _, err = command.schemaReader.ReadSchemaDefinition(ctx, request.GetTenantId(), request.GetEntityType(), request.GetMetadata().GetSchemaVersion())
	if err != nil {
		return nil, err
	}

	for _, action := range en.GetActions() {
		var can bool
		can, err = command.l(ctx, request, action.Child)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())
			return nil, err
		}
		if can {
			response.ActionNames = append(response.ActionNames, action.Name)
		}
	}
	return response, nil
}

// SchemaLookupDecision - This is a type definition for a decision made during a schema lookup operation.
// It contains a boolean flag indicating whether the operation is permitted or not, and an error object
// in case of any errors encountered during the operation.
type SchemaLookupDecision struct {
	Can bool
	Err error
}

// sendSchemaLookupDecision - This function returns a SchemaLookupDecision struct with a boolean can and an
// error err indicating whether a permission check can be performed and any error occurred during the process,
// respectively. The function is named sendSchemaLookupDecision.
func sendSchemaLookupDecision(can bool, err error) SchemaLookupDecision {
	return SchemaLookupDecision{
		Can: can,
		Err: err,
	}
}

// SchemaLookupFunction represents a function that performs a schema lookup operation and sends the result via a channel
// The result is a SchemaLookupDecision that contains a boolean value indicating if the lookup operation was successful or not
// and an error in case there was an issue with the lookup operation.
type SchemaLookupFunction func(ctx context.Context, lookupChan chan<- SchemaLookupDecision)

// SchemaLookupCombiner represents a function that takes a slice of SchemaLookupFunctions and combines their results to produce
// a single SchemaLookupDecision result.
// The result is a SchemaLookupDecision that contains a boolean value indicating if the lookup operation was successful or not
// and an error in case there was an issue with the lookup operation.
type SchemaLookupCombiner func(ctx context.Context, functions []SchemaLookupFunction) SchemaLookupDecision

// l is a helper function that takes a context, a lookup schema request, and a child.
// It returns a boolean and an error indicating whether the given child is allowed for the given request.
// It selects an appropriate lookup function based on the type of child (either a leaf or a rewrite),
// and applies this function to the given request and child. The resulting schema lookup decisions are combined
// using the union function.
func (command *LookupSchemaEngine) l(ctx context.Context, request *base.PermissionLookupSchemaRequest, child *base.Child) (bool, error) {
	var fn SchemaLookupFunction
	switch child.Type.(type) {
	case *base.Child_Rewrite:
		fn = command.lookupRewrite(ctx, request, child.GetRewrite())
	case *base.Child_Leaf:
		fn = command.lookupLeaf(ctx, request, child.GetLeaf())
	}

	if fn == nil {
		return false, errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_CHILD_KIND.String())
	}

	result := schemaLookupUnion(ctx, []SchemaLookupFunction{fn})
	return result.Can, result.Err
}

// lookupRewrite takes a PermissionLookupSchemaRequest, a Rewrite object and returns a SchemaLookupFunction that is used to determine
// whether the specified schema is accessible or not by looking up its children. Depending on the rewrite operation (UNION or INTERSECTION)
// it uses either the schemaLookupUnion or schemaLookupIntersection function. It returns a function that is used by schemaLookupCombiner.
// If the operation is not defined, it returns a schemaLookupFail function with an error.
func (command *LookupSchemaEngine) lookupRewrite(ctx context.Context, request *base.PermissionLookupSchemaRequest, rewrite *base.Rewrite) SchemaLookupFunction {
	switch rewrite.GetRewriteOperation() {
	case *base.Rewrite_OPERATION_UNION.Enum():
		return command.setChild(ctx, request, rewrite.GetChildren(), schemaLookupUnion)
	case *base.Rewrite_OPERATION_INTERSECTION.Enum():
		return command.setChild(ctx, request, rewrite.GetChildren(), schemaLookupIntersection)
	default:
		return schemaLookupFail(errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_CHILD_TYPE.String()))
	}
}

// checkLeaf - This function takes a Leaf object which contains either a TupleToUserSet or ComputedUserSet type and returns a SchemaLookupFunction.
// It calls the lookup function with the computed relation name and the request and exclusion objects.
func (command *LookupSchemaEngine) lookupLeaf(ctx context.Context, request *base.PermissionLookupSchemaRequest, leaf *base.Leaf) SchemaLookupFunction {
	switch leaf.GetType().(type) {
	case *base.Leaf_TupleToUserSet:
		return command.lookup(ctx, fmt.Sprintf("%s.%s", leaf.GetTupleToUserSet().GetTupleSet().GetRelation(), leaf.GetTupleToUserSet().GetComputed().GetRelation()), request, leaf.GetExclusion())
	case *base.Leaf_ComputedUserSet:
		return command.lookup(ctx, leaf.GetComputedUserSet().GetRelation(), request, leaf.GetExclusion())
	default:
		return schemaLookupFail(errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_CHILD_TYPE.String()))
	}
}

// setChild creates a SchemaLookupFunction that sets multiple children of a parent node and uses a SchemaLookupCombiner
// to compute the final decision. It loops over the children and creates corresponding lookup functions for each
// child based on its type (Rewrite or Leaf). It then creates a SchemaLookupFunction that invokes each child lookup
// function and uses the given SchemaLookupCombiner to compute the final decision based on the results of the child
// lookup functions. The resulting SchemaLookupFunction is a function that takes a context.Context and a channel
// (resultChan) as arguments. The function sends the result of the lookup operation (whether it can or cannot access
// the schema) to the resultChan.
func (command *LookupSchemaEngine) setChild(ctx context.Context, request *base.PermissionLookupSchemaRequest, children []*base.Child, combiner SchemaLookupCombiner) SchemaLookupFunction {
	var functions []SchemaLookupFunction
	for _, child := range children {
		switch child.GetType().(type) {
		case *base.Child_Rewrite:
			functions = append(functions, command.lookupRewrite(ctx, request, child.GetRewrite()))
		case *base.Child_Leaf:
			functions = append(functions, command.lookupLeaf(ctx, request, child.GetLeaf()))
		default:
			return schemaLookupFail(errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_CHILD_KIND.String()))
		}
	}

	return func(ctx context.Context, resultChan chan<- SchemaLookupDecision) {
		resultChan <- combiner(ctx, functions)
	}
}

// lookup is a function that returns a SchemaLookupFunction for the given relation name and exclusion flag. The returned function
// will check whether the relation name is present in the list of relation names in the given PermissionLookupSchemaRequest and
// return a SchemaLookupDecision indicating whether the check was successful or not, based on the exclusion flag. If the exclusion flag is set,
// the function will return true if the relation name is not present in the list, and false if it is. If the exclusion flag is not set, the function
// will return true if the relation name is present in the list, and false if it is not. The SchemaLookupFunction is returned as a closure that
// takes a context and a channel, and sends the SchemaLookupDecision through the channel.
func (command *LookupSchemaEngine) lookup(ctx context.Context, relation string, request *base.PermissionLookupSchemaRequest, exclusion bool) SchemaLookupFunction {
	return func(ctx context.Context, lookupChan chan<- SchemaLookupDecision) {
		var err error
		if exclusion {
			lookupChan <- sendSchemaLookupDecision(!slices.Contains(request.GetRelationNames(), relation), err)
			return
		}
		lookupChan <- sendSchemaLookupDecision(slices.Contains(request.GetRelationNames(), relation), err)
	}
}

// This function takes a slice of SchemaLookupFunction and performs a union operation on them. It creates a separate goroutine for each function
// and passes the result to a channel. It checks the result of each function in the channel and returns true if any of the functions return true
// without any error. If all the functions return false or any of them return an error, it returns false along with the error. It takes a context
// object that can be used to cancel the operation.
func schemaLookupUnion(ctx context.Context, functions []SchemaLookupFunction) SchemaLookupDecision {
	if len(functions) == 0 {
		return sendSchemaLookupDecision(true, nil)
	}

	lookupChan := make(chan SchemaLookupDecision, len(functions))
	childCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	for _, fn := range functions {
		go fn(childCtx, lookupChan)
	}

	for i := 0; i < len(functions); i++ {
		select {
		case result := <-lookupChan:
			if result.Err == nil && result.Can {
				return sendSchemaLookupDecision(true, nil)
			}
			if result.Err != nil {
				return sendSchemaLookupDecision(false, result.Err)
			}
		case <-ctx.Done():
			return sendSchemaLookupDecision(false, errors.New(base.ErrorCode_ERROR_CODE_CANCELLED.String()))
		}
	}

	return sendSchemaLookupDecision(false, nil)
}

// This is a function that takes in a context and a slice of SchemaLookupFunctions, and returns a SchemaLookupDecision
// indicating whether the intersection of the decisions from all the input functions is true or false. If any of the
// input functions return an error, the output SchemaLookupDecision will also contain that error. This function uses
// goroutines to call each input function concurrently, and sends the results to a channel to wait for all the results
// before returning.
func schemaLookupIntersection(ctx context.Context, functions []SchemaLookupFunction) SchemaLookupDecision {
	if len(functions) == 0 {
		return sendSchemaLookupDecision(true, nil)
	}

	lookupChan := make(chan SchemaLookupDecision, len(functions))
	childCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	for _, fn := range functions {
		go fn(childCtx, lookupChan)
	}

	for i := 0; i < len(functions); i++ {
		select {
		case result := <-lookupChan:
			if result.Err == nil && !result.Can {
				return sendSchemaLookupDecision(false, nil)
			}
			if result.Err != nil {
				return sendSchemaLookupDecision(false, result.Err)
			}
		case <-ctx.Done():
			return sendSchemaLookupDecision(false, errors.New(base.ErrorCode_ERROR_CODE_CANCELLED.String()))
		}
	}

	return sendSchemaLookupDecision(true, nil)
}

// This is a function that creates a new SchemaLookupFunction that always returns a failure decision with the given error message.
// It takes an error parameter and returns a SchemaLookupFunction. The SchemaLookupFunction takes a context.Context and
// a chan<- SchemaLookupDecision parameter, and sends a failure SchemaLookupDecision to the provided channel with the given
// error message.
func schemaLookupFail(err error) SchemaLookupFunction {
	return func(ctx context.Context, decisionChan chan<- SchemaLookupDecision) {
		decisionChan <- sendSchemaLookupDecision(false, err)
	}
}
