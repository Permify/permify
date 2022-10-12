package commands

import (
	"context"

	internalErrors `github.com/Permify/permify/internal/errors`
	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/pkg/errors"
	"github.com/Permify/permify/pkg/logger"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	`github.com/Permify/permify/pkg/tuple`
)

// ExpandFunction -
type ExpandFunction func(ctx context.Context, expandChan chan<- *base.Expand)

// ExpandCombiner .
type ExpandCombiner func(ctx context.Context, functions []ExpandFunction, expand ...*base.Expand) *base.Expand

// ExpandCommand -
type ExpandCommand struct {
	relationTupleRepository repositories.IRelationTupleRepository
	logger                  logger.Interface
}

// NewExpandCommand -
func NewExpandCommand(rr repositories.IRelationTupleRepository, l logger.Interface) *ExpandCommand {
	return &ExpandCommand{
		relationTupleRepository: rr,
		logger:                  l,
	}
}

// GetRelationTupleRepository -
func (command *ExpandCommand) GetRelationTupleRepository() repositories.IRelationTupleRepository {
	return command.relationTupleRepository
}

// ExpandQuery -
type ExpandQuery struct {
	Entity *base.Entity
}

// ExpandResponse -
type ExpandResponse struct {
	Tree *base.Expand `json:"tree"`
}

// Execute -
func (command *ExpandCommand) Execute(ctx context.Context, q *ExpandQuery, child *base.Child) (response ExpandResponse, err errors.Error) {
	response.Tree, err = command.e(ctx, q, child)
	return
}

// e -
func (command *ExpandCommand) e(ctx context.Context, q *ExpandQuery, child *base.Child) (*base.Expand, errors.Error) {
	var fn ExpandFunction
	switch op := child.GetType().(type) {
	case *base.Child_Rewrite:
		fn = command.expandRewrite(ctx, q, op.Rewrite)
	case *base.Child_Leaf:
		fn = command.expandLeaf(ctx, q, op.Leaf)
	}
	result := expandRoot(ctx, fn)
	return result, nil
}

// expandRewrite -
func (command *ExpandCommand) expandRewrite(ctx context.Context, q *ExpandQuery, rewrite *base.Rewrite) ExpandFunction {
	switch rewrite.GetRewriteOperation() {
	case *base.Rewrite_UNION.Enum():
		return command.set(ctx, q, rewrite.GetChildren(), expandUnion)
	case *base.Rewrite_INTERSECTION.Enum():
		return command.set(ctx, q, rewrite.GetChildren(), expandIntersection)
	default:
		return expandFail(internalErrors.UndefinedChildTypeError)
	}
}

// expandLeaf -
func (command *ExpandCommand) expandLeaf(ctx context.Context, q *ExpandQuery, leaf *base.Leaf) ExpandFunction {
	switch op := leaf.GetType().(type) {
	case *base.Leaf_TupleToUserSet:
		return command.expand(ctx, &base.EntityAndRelation{
			Entity:   q.Entity,
			Relation: op.TupleToUserSet.GetRelation(),
		}, q, leaf.GetExclusion())
	case *base.Leaf_ComputedUserSet:
		return command.expand(ctx, &base.EntityAndRelation{
			Entity:   q.Entity,
			Relation: op.ComputedUserSet.GetRelation(),
		}, q, leaf.GetExclusion())
	default:
		return expandFail(internalErrors.UndefinedChildTypeError)
	}
}

// set -
func (command *ExpandCommand) set(ctx context.Context, q *ExpandQuery, children []*base.Child, combiner ExpandCombiner) ExpandFunction {
	var functions []ExpandFunction
	for _, child := range children {
		switch child.GetType().(type) {
		case *base.Child_Rewrite:
			functions = append(functions, command.expandRewrite(ctx, q, child.GetRewrite()))
		case *base.Child_Leaf:
			functions = append(functions, command.expandLeaf(ctx, q, child.GetLeaf()))
		default:
			return expandFail(internalErrors.UndefinedChildKindError)
		}
	}

	return func(ctx context.Context, resultChan chan<- *base.Expand) {
		resultChan <- combiner(ctx, functions)
	}
}

// expand -
func (command *ExpandCommand) expand(ctx context.Context, entityAndRelation *base.EntityAndRelation, q *ExpandQuery, exclusion bool) ExpandFunction {
	return func(ctx context.Context, expandChan chan<- *base.Expand) {
		var err errors.Error

		var iterator tuple.ISubjectIterator
		iterator, err = getSubjects(ctx, command, entityAndRelation.GetEntity(), entityAndRelation.GetRelation())
		if err != nil {
			expandFail(err)
			return
		}

		var subjects = &base.Subjects{
			Exclusion: exclusion,
		}

		var expandFunctions []ExpandFunction

		for iterator.HasNext() {
			subject := iterator.GetNext()
			subjects.Subjects = append(subjects.Subjects, subject)
			if !tuple.IsSubjectUser(subject) {
				expandFunctions = append(expandFunctions, command.expand(ctx, &base.EntityAndRelation{
					Entity: &base.Entity{
						Type: subject.GetType(),
						Id:   subject.GetId(),
					},
					Relation: subject.GetRelation(),
				}, q, exclusion))
			}
		}

		var node = &base.Expand{
			Target: entityAndRelation,
			Node:   &base.Expand_Leaf{Leaf: subjects},
		}

		if len(expandFunctions) > 0 {
			expandChan <- expandUnion(ctx, expandFunctions, node)
			return
		}

		expandChan <- node
		return
	}
}

// expandOperation -
func expandOperation(
	ctx context.Context,
	functions []ExpandFunction,
	op base.ExpandTreeNode_Operation,
	node ...*base.Expand,
) *base.Expand {
	c, cancel := context.WithCancel(ctx)
	defer cancel()

	result := make([]chan *base.Expand, 0, len(functions))

	for _, fn := range functions {
		en := make(chan *base.Expand)
		result = append(result, en)
		go fn(c, en)
	}

	children := make([]*base.Expand, 0, len(functions)+len(node))

	if len(node) > 0 {
		children = append(children, node...)
	}

	for _, resultChan := range result {
		select {
		case child := <-resultChan:
			children = append(children, child)
		case <-ctx.Done():
			return nil
		}
	}

	return &base.Expand{
		Node: &base.Expand_Expand{Expand: &base.ExpandTreeNode{
			Operation: op,
			Children:  children,
		}},
	}
}

// expandRoot -
func expandRoot(ctx context.Context, fn ExpandFunction) *base.Expand {
	resultChan := make(chan *base.Expand, 1)
	go fn(ctx, resultChan)
	select {
	case result := <-resultChan:
		return result
	case <-ctx.Done():
		return nil
	}
}

// expandUnion -
func expandUnion(ctx context.Context, functions []ExpandFunction, expand ...*base.Expand) *base.Expand {
	return expandOperation(ctx, functions, base.ExpandTreeNode_UNION, expand...)
}

// expandIntersection -
func expandIntersection(ctx context.Context, functions []ExpandFunction, expand ...*base.Expand) *base.Expand {
	return expandOperation(ctx, functions, base.ExpandTreeNode_INTERSECTION, expand...)
}

// expandFail -
func expandFail(err error) ExpandFunction {
	return func(ctx context.Context, expandChan chan<- *base.Expand) {
		expandChan <- &base.Expand{}
	}
}
