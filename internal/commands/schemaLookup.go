package commands

import (
	"context"

	internalErrors "github.com/Permify/permify/internal/errors"
	"github.com/Permify/permify/pkg/errors"
	"github.com/Permify/permify/pkg/helper"
	"github.com/Permify/permify/pkg/logger"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// SchemaLookupDecision -
type SchemaLookupDecision struct {
	Prefix string       `json:"prefix"`
	Can    bool         `json:"can"`
	Err    errors.Error `json:"-"`
}

// sendSchemaLookupDecision -
func sendSchemaLookupDecision(can bool, prefix string, err errors.Error) SchemaLookupDecision {
	return SchemaLookupDecision{
		Prefix: prefix,
		Can:    can,
		Err:    err,
	}
}

// SchemaLookupFunction -
type SchemaLookupFunction func(ctx context.Context, lookupChan chan<- SchemaLookupDecision)

// SchemaLookupCombiner .
type SchemaLookupCombiner func(ctx context.Context, functions []SchemaLookupFunction) SchemaLookupDecision

// SchemaLookupCommand -
type SchemaLookupCommand struct {
	logger logger.Interface
}

// NewSchemaLookupCommand -
func NewSchemaLookupCommand(l logger.Interface) *SchemaLookupCommand {
	return &SchemaLookupCommand{
		logger: l,
	}
}

// SchemaLookupQuery -
type SchemaLookupQuery struct {
	Relations []string
}

// SchemaLookupResponse -
type SchemaLookupResponse struct {
	ActionNames []string
}

// Execute -
func (command *SchemaLookupCommand) Execute(ctx context.Context, q *SchemaLookupQuery, actions []*base.ActionDefinition) (response SchemaLookupResponse, err errors.Error) {
	response.ActionNames = []string{}
	for _, action := range actions {
		var can bool
		can, err = command.l(ctx, q, action.Child)
		if err != nil {
			return
		}
		if can {
			response.ActionNames = append(response.ActionNames, action.Name)
		}
	}
	return
}

// c -
func (command *SchemaLookupCommand) l(ctx context.Context, q *SchemaLookupQuery, child *base.Child) (bool, errors.Error) {
	var fn SchemaLookupFunction
	switch child.Type.(type) {
	case *base.Child_Rewrite:
		fn = command.lookupRewrite(ctx, q, child.GetRewrite())
	case *base.Child_Leaf:
		fn = command.lookupLeaf(ctx, q, child.GetLeaf())
	}

	if fn == nil {
		return false, internalErrors.UndefinedChildKindError
	}

	result := schemaLookupUnion(ctx, []SchemaLookupFunction{fn})
	return result.Can, result.Err
}

// lookupRewrite -
func (command *SchemaLookupCommand) lookupRewrite(ctx context.Context, q *SchemaLookupQuery, rewrite *base.Rewrite) SchemaLookupFunction {
	switch rewrite.GetRewriteOperation() {
	case *base.Rewrite_UNION.Enum():
		return command.set(ctx, q, rewrite.GetChildren(), schemaLookupUnion)
	case *base.Rewrite_INTERSECTION.Enum():
		return command.set(ctx, q, rewrite.GetChildren(), schemaLookupIntersection)
	default:
		return schemaLookupFail(internalErrors.UndefinedChildTypeError)
	}
}

// checkLeaf -
func (command *SchemaLookupCommand) lookupLeaf(ctx context.Context, q *SchemaLookupQuery, leaf *base.Leaf) SchemaLookupFunction {
	switch leaf.GetType().(type) {
	case *base.Leaf_TupleToUserSet:
		return command.lookup(ctx, leaf.GetTupleToUserSet().GetRelation(), q, leaf.GetExclusion())
	case *base.Leaf_ComputedUserSet:
		return command.lookup(ctx, leaf.GetComputedUserSet().GetRelation(), q, leaf.GetExclusion())
	default:
		return schemaLookupFail(internalErrors.UndefinedChildTypeError)
	}
}

// set -
func (command *SchemaLookupCommand) set(ctx context.Context, q *SchemaLookupQuery, children []*base.Child, combiner SchemaLookupCombiner) SchemaLookupFunction {
	var functions []SchemaLookupFunction
	for _, child := range children {
		switch child.GetType().(type) {
		case *base.Child_Rewrite:
			functions = append(functions, command.lookupRewrite(ctx, q, child.GetRewrite()))
		case *base.Child_Leaf:
			functions = append(functions, command.lookupLeaf(ctx, q, child.GetLeaf()))
		default:
			return schemaLookupFail(internalErrors.UndefinedChildKindError)
		}
	}

	return func(ctx context.Context, resultChan chan<- SchemaLookupDecision) {
		resultChan <- combiner(ctx, functions)
	}
}

// check -
func (command *SchemaLookupCommand) lookup(ctx context.Context, relation string, q *SchemaLookupQuery, exclusion bool) SchemaLookupFunction {
	return func(ctx context.Context, lookupChan chan<- SchemaLookupDecision) {
		var err errors.Error
		if exclusion {
			lookupChan <- sendSchemaLookupDecision(!helper.InArray(relation, q.Relations), "not", err)
			return
		}
		lookupChan <- sendSchemaLookupDecision(helper.InArray(relation, q.Relations), "", err)
		return
	}
}

// union -
func schemaLookupUnion(ctx context.Context, functions []SchemaLookupFunction) SchemaLookupDecision {
	if len(functions) == 0 {
		return sendSchemaLookupDecision(true, "", nil)
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
				return sendSchemaLookupDecision(true, result.Prefix, nil)
			}
			if result.Err != nil {
				return sendSchemaLookupDecision(false, result.Prefix, result.Err)
			}
		case <-ctx.Done():
			return sendSchemaLookupDecision(false, "", internalErrors.CanceledError)
		}
	}

	return sendSchemaLookupDecision(false, "", nil)
}

// intersection -
func schemaLookupIntersection(ctx context.Context, functions []SchemaLookupFunction) SchemaLookupDecision {
	if len(functions) == 0 {
		return sendSchemaLookupDecision(true, "", nil)
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
				return sendSchemaLookupDecision(false, result.Prefix, nil)
			}
			if result.Err != nil {
				return sendSchemaLookupDecision(false, result.Prefix, result.Err)
			}
		case <-ctx.Done():
			return sendSchemaLookupDecision(false, "", internalErrors.CanceledError)
		}
	}

	return sendSchemaLookupDecision(true, "", nil)
}

// schemaLookupFail -
func schemaLookupFail(err errors.Error) SchemaLookupFunction {
	return func(ctx context.Context, decisionChan chan<- SchemaLookupDecision) {
		decisionChan <- sendSchemaLookupDecision(false, "", err)
	}
}
