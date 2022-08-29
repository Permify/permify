package commands

import (
	"context"

	internalErrors "github.com/Permify/permify/internal/internal-errors"
	"github.com/Permify/permify/pkg/dsl/schema"
	"github.com/Permify/permify/pkg/helper"
	"github.com/Permify/permify/pkg/logger"
	"github.com/Permify/permify/pkg/tuple"
)

// SchemaLookupDecision -
type SchemaLookupDecision struct {
	Prefix string `json:"prefix"`
	Can    bool   `json:"can"`
	Err    error  `json:"-"`
}

// sendSchemaLookupDecision -
func sendSchemaLookupDecision(can bool, prefix string, err error) SchemaLookupDecision {
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
func (command *SchemaLookupCommand) Execute(ctx context.Context, q *SchemaLookupQuery, actions []schema.Action) (response SchemaLookupResponse, err error) {
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
func (command *SchemaLookupCommand) l(ctx context.Context, q *SchemaLookupQuery, child schema.Child) (bool, error) {
	var fn SchemaLookupFunction
	switch child.GetKind() {
	case schema.RewriteKind.String():
		fn = command.lookupRewrite(ctx, q, child.(schema.Rewrite))
	case schema.LeafKind.String():
		fn = command.lookupLeaf(ctx, q, child.(schema.Leaf))
	}

	if fn == nil {
		return false, internalErrors.UndefinedChildKindError
	}

	result := schemaLookupUnion(ctx, []SchemaLookupFunction{fn})
	return result.Can, result.Err
}

// lookupRewrite -
func (command *SchemaLookupCommand) lookupRewrite(ctx context.Context, q *SchemaLookupQuery, child schema.Rewrite) SchemaLookupFunction {
	switch child.GetType() {
	case schema.Union.String():
		return command.set(ctx, q, child.Children, schemaLookupUnion)
	case schema.Intersection.String():
		return command.set(ctx, q, child.Children, schemaLookupIntersection)
	default:
		return schemaLookupFail(internalErrors.UndefinedChildTypeError)
	}
}

// checkLeaf -
func (command *SchemaLookupCommand) lookupLeaf(ctx context.Context, q *SchemaLookupQuery, child schema.Leaf) SchemaLookupFunction {
	return command.lookup(ctx, tuple.Relation(child.Value), q, child.Exclusion)
}

// set -
func (command *SchemaLookupCommand) set(ctx context.Context, q *SchemaLookupQuery, children []schema.Child, combiner SchemaLookupCombiner) SchemaLookupFunction {
	var functions []SchemaLookupFunction
	for _, child := range children {
		switch child.GetKind() {
		case schema.RewriteKind.String():
			functions = append(functions, command.lookupRewrite(ctx, q, child.(schema.Rewrite)))
		case schema.LeafKind.String():
			functions = append(functions, command.lookupLeaf(ctx, q, child.(schema.Leaf)))
		default:
			return schemaLookupFail(internalErrors.UndefinedChildKindError)
		}
	}

	return func(ctx context.Context, resultChan chan<- SchemaLookupDecision) {
		resultChan <- combiner(ctx, functions)
	}
}

// check -
func (command *SchemaLookupCommand) lookup(ctx context.Context, relation tuple.Relation, q *SchemaLookupQuery, exclusion bool) SchemaLookupFunction {
	return func(ctx context.Context, lookupChan chan<- SchemaLookupDecision) {
		var err error
		if exclusion {
			lookupChan <- sendSchemaLookupDecision(!helper.InArray(relation.String(), q.Relations), "not", err)
			return
		}
		lookupChan <- sendSchemaLookupDecision(helper.InArray(relation.String(), q.Relations), "", err)
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
func schemaLookupFail(err error) SchemaLookupFunction {
	return func(ctx context.Context, decisionChan chan<- SchemaLookupDecision) {
		decisionChan <- sendSchemaLookupDecision(false, "", err)
	}
}
