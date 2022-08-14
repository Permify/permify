package commands

import (
	"context"

	"github.com/Permify/permify/internal/entities"
	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/pkg/dsl/schema"
	"github.com/Permify/permify/pkg/tuple"
)

type Node interface {
	GetNodeType() string
}

type ExpandNode struct {
	Type     string `json:"type"`
	Children []Node `json:"nodes"`
}

func (ExpandNode) GetNodeType() string {
	return "expand"
}

type LeafNode struct {
	Subject tuple.Subject `json:"subject"`
}

func (LeafNode) GetNodeType() string {
	return "leaf"
}

type BranchNode struct {
	Expand   tuple.EntityAndRelation `json:"expand"`
	Children []Node                  `json:"children"`
}

func (BranchNode) GetNodeType() string {
	return "branch"
}

// ExpandFunction -
type ExpandFunction func(ctx context.Context, expandChan chan<- Node)

// ExpandCombiner .
type ExpandCombiner func(ctx context.Context, requests []ExpandFunction) Node

// ExpandCommand -
type ExpandCommand struct {
	relationTupleRepository repositories.IRelationTupleRepository
}

// NewExpandCommand -
func NewExpandCommand(rr repositories.IRelationTupleRepository) *ExpandCommand {
	return &ExpandCommand{
		relationTupleRepository: rr,
	}
}

// ExpandQuery -
type ExpandQuery struct {
	Entity tuple.Entity
	depth  int
}

// SetDepth -
func (r *ExpandQuery) SetDepth(i int) {
	r.depth = i
}

// decrease -
func (r *ExpandQuery) decrease() *ExpandQuery {
	r.depth--
	return r
}

// isDepthFinish -
func (r *ExpandQuery) isDepthFinish() bool {
	return r.depth <= 0
}

// ExpandResponse -
type ExpandResponse struct {
	Tree           Node
	RemainingDepth int
	Error          error
}

// Execute -
func (command *ExpandCommand) Execute(ctx context.Context, q *ExpandQuery, child schema.Child) (response ExpandResponse) {
	expandNode, _ := command.e(ctx, q, child)
	response.Tree = expandNode
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
		return expandFail(UndefinedChildTypeError)
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
		return expandFail(UndefinedChildTypeError)
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
			return expandFail(UndefinedChildKindError)
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
		branch.Expand = tuple.EntityAndRelation{
			Entity:   entity,
			Relation: relation,
		}
		branch.Children = []Node{}

		var expandFunctions []ExpandFunction

		for iterator.HasNext() {
			subject := iterator.GetNext()
			if subject.IsUser() {
				branch.Children = append(branch.Children, &LeafNode{
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

func expandSetOperation(
	ctx context.Context,
	functions []ExpandFunction,
	op string,
) Node {
	children := make([]Node, 0, len(functions))

	if len(functions) == 0 {
		return &ExpandNode{
			Type:     op,
			Children: children,
		}
	}

	c, cancelFn := context.WithCancel(ctx)
	defer cancelFn()

	result := make([]chan Node, 0, len(functions))
	for _, fn := range functions {
		en := make(chan Node)
		result = append(result, en)
		go fn(c, en)
	}

	for _, resultChan := range result {
		select {
		case res := <-resultChan:
			children = append(children, res)
		case <-ctx.Done():
			return nil
		}
	}

	return &ExpandNode{
		Type:     op,
		Children: children,
	}
}

// expandUnion -
func expandRoot(ctx context.Context, functions []ExpandFunction) Node {
	return expandSetOperation(ctx, functions, "root")
}

// expandUnion -
func expandUnion(ctx context.Context, functions []ExpandFunction) Node {
	return expandSetOperation(ctx, functions, "union")
}

// expandIntersection -
func expandIntersection(ctx context.Context, functions []ExpandFunction) Node {
	return expandSetOperation(ctx, functions, "intersection")
}

// expandFail -
func expandFail(err error) ExpandFunction {
	return func(ctx context.Context, expandChan chan<- Node) {
		// expandChan <- sendExpandNode("", err)
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
