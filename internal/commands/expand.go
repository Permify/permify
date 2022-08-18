package commands

import (
	"context"

	"github.com/Permify/permify/internal/entities"
	internalErrors "github.com/Permify/permify/internal/internal-errors"
	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/pkg/dsl/schema"
	"github.com/Permify/permify/pkg/logger"
	"github.com/Permify/permify/pkg/tuple"
)

type NodeKind string

const (
	EXPAND NodeKind = "expand"
	LEAF   NodeKind = "leaf"
	BRANCH NodeKind = "branch"
)

// Node -
type Node interface {
	GetKind() NodeKind
	Error() error
}

// ExpandNode -
type ExpandNode struct {
	Kind      NodeKind      `json:"kind"`
	Operation schema.OPType `json:"operation"`
	Children  []Node        `json:"children"`
	Err       error         `json:"-"`
}

// GetKind -
func (ExpandNode) GetKind() NodeKind {
	return EXPAND
}

// Error -
func (e ExpandNode) Error() error {
	return e.Err
}

// LeafNode -
type LeafNode struct {
	Kind    NodeKind      `json:"kind"`
	Subject tuple.Subject `json:"subject"`
	Err     error         `json:"-"`
}

// GetKind -
func (LeafNode) GetKind() NodeKind {
	return LEAF
}

// Error -
func (e LeafNode) Error() error {
	return e.Err
}

// BranchNode -
type BranchNode struct {
	Kind     NodeKind                `json:"kind"`
	Target   tuple.EntityAndRelation `json:"target"`
	Children []Node                  `json:"children"`
	Err      error                   `json:"-"`
}

// GetKind -
func (BranchNode) GetKind() NodeKind {
	return BRANCH
}

// Error -
func (e BranchNode) Error() error {
	return e.Err
}

// ExpandFunction -
type ExpandFunction func(ctx context.Context, expandChan chan<- Node)

// ExpandCombiner .
type ExpandCombiner func(ctx context.Context, requests []ExpandFunction) Node

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

// ExpandQuery -
type ExpandQuery struct {
	Entity tuple.Entity
}

// ExpandResponse -
type ExpandResponse struct {
	Tree Node
}

// Execute -
func (command *ExpandCommand) Execute(ctx context.Context, q *ExpandQuery, child schema.Child) (response ExpandResponse, err error) {
	response.Tree, err = command.e(ctx, q, child)
	return
}

// e -
func (command *ExpandCommand) e(ctx context.Context, q *ExpandQuery, child schema.Child) (Node, error) {
	var fn ExpandFunction
	switch child.GetKind() {
	case schema.RewriteKind.String():
		fn = command.expandRewrite(ctx, q, child.(schema.Rewrite)) // ExpandNode
	case schema.LeafKind.String():
		fn = command.expandLeaf(ctx, q, child.(schema.Leaf)) // Branch or Leaf
	}
	result := expandRoot(ctx, []ExpandFunction{fn})
	return result, nil
}

// expandRewrite -
func (command *ExpandCommand) expandRewrite(ctx context.Context, q *ExpandQuery, child schema.Rewrite) ExpandFunction {
	switch child.GetType() {
	case schema.Union.String():
		return command.set(ctx, q, child.Children, expandUnion)
	case schema.Intersection.String():
		return command.set(ctx, q, child.Children, expandIntersection)
	default:
		return expandFail(internalErrors.UndefinedChildTypeError)
	}
}

// expandLeaf -
func (command *ExpandCommand) expandLeaf(ctx context.Context, q *ExpandQuery, child schema.Leaf) ExpandFunction {
	switch child.GetType() {
	case schema.TupleToUserSetType.String():
		return command.expand(ctx, q.Entity, tuple.Relation(child.Value), q)
	case schema.ComputedUserSetType.String():
		return command.expand(ctx, q.Entity, tuple.Relation(child.Value), q)
	default:
		return expandFail(internalErrors.UndefinedChildTypeError)
	}
}

// set -
func (command *ExpandCommand) set(ctx context.Context, q *ExpandQuery, children []schema.Child, combiner ExpandCombiner) ExpandFunction {
	var functions []ExpandFunction
	for _, child := range children {
		switch child.GetKind() {
		case schema.RewriteKind.String():
			functions = append(functions, command.expandRewrite(ctx, q, child.(schema.Rewrite)))
		case schema.LeafKind.String():
			functions = append(functions, command.expandLeaf(ctx, q, child.(schema.Leaf)))
		default:
			return expandFail(internalErrors.UndefinedChildKindError)
		}
	}

	return func(ctx context.Context, resultChan chan<- Node) {
		resultChan <- combiner(ctx, functions)
	}
}

// expand -
func (command *ExpandCommand) expand(ctx context.Context, entity tuple.Entity, relation tuple.Relation, q *ExpandQuery) ExpandFunction {
	return func(ctx context.Context, expandChan chan<- Node) {
		var err error

		var iterator tuple.ISubjectIterator
		iterator, err = command.getSubjects(ctx, entity, relation)
		if err != nil {
			checkFail(err)
			return
		}

		var branch BranchNode
		branch.Kind = BRANCH
		branch.Target = tuple.EntityAndRelation{
			Entity:   entity,
			Relation: relation,
		}
		branch.Children = []Node{}

		var expandFunctions []ExpandFunction

		for iterator.HasNext() {
			subject := iterator.GetNext()
			if subject.IsUser() {
				branch.Children = append(branch.Children, &LeafNode{
					Kind: LEAF,
					Subject: tuple.Subject{
						Type: tuple.USER,
						ID:   subject.ID,
					},
				})
			} else {
				expandFunctions = append(expandFunctions, command.expand(ctx, tuple.Entity{ID: subject.ID, Type: subject.Type}, subject.Relation, q))
			}
		}

		if len(expandFunctions) > 0 {
			branch.Children = append(branch.Children, expandUnion(ctx, expandFunctions))
		}

		expandChan <- &branch
		return
	}
}

// expandSetOperation -
func expandSetOperation(
	ctx context.Context,
	functions []ExpandFunction,
	op schema.OPType,
) Node {
	children := make([]Node, 0, len(functions))

	if len(functions) == 0 {
		return &ExpandNode{
			Kind:      EXPAND,
			Operation: op,
			Children:  children,
		}
	}

	c, cancel := context.WithCancel(ctx)
	defer cancel()

	result := make([]chan Node, 0, len(functions))
	for _, fn := range functions {
		en := make(chan Node)
		result = append(result, en)
		go fn(c, en)
	}

	for _, resultChan := range result {
		select {
		case res := <-resultChan:
			if res.Error() != nil {
				return ExpandNode{
					Err: res.Error(),
				}
			}
			children = append(children, res)
		case <-ctx.Done():
			return nil
		}
	}

	return &ExpandNode{
		Kind:      EXPAND,
		Operation: op,
		Children:  children,
	}
}

// expandUnion -
func expandRoot(ctx context.Context, functions []ExpandFunction) Node {
	return expandSetOperation(ctx, functions, "root")
}

// expandUnion -
func expandUnion(ctx context.Context, functions []ExpandFunction) Node {
	return expandSetOperation(ctx, functions, schema.Union)
}

// expandIntersection -
func expandIntersection(ctx context.Context, functions []ExpandFunction) Node {
	return expandSetOperation(ctx, functions, schema.Intersection)
}

// expandFail -
func expandFail(err error) ExpandFunction {
	return func(ctx context.Context, expandChan chan<- Node) {
		expandChan <- ExpandNode{
			Err: err,
		}
	}
}

// getSubjects -
func (command *ExpandCommand) getSubjects(ctx context.Context, entity tuple.Entity, relation tuple.Relation) (iterator tuple.ISubjectIterator, err error) {
	r := relation.Split()

	var tuples []entities.RelationTuple
	tuples, err = command.relationTupleRepository.QueryTuples(ctx, entity.Type, entity.ID, r[0].String())
	if err != nil {
		return nil, err
	}

	var subjects []*tuple.Subject
	for _, tup := range tuples {
		ct := tup.ToTuple()
		if !ct.Subject.IsUser() {
			subject := ct.Subject
			if tup.UsersetRelation == tuple.ELLIPSIS {
				subject.Relation = r[1]
			} else {
				subject.Relation = ct.Subject.Relation
			}
			subjects = append(subjects, &subject)
		} else {
			subjects = append(subjects, &tuple.Subject{
				Type: tuple.USER,
				ID:   tup.UsersetObjectID,
			})
		}
	}

	return tuple.NewSubjectIterator(subjects), err
}
