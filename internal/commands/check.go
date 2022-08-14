package commands

import (
	"context"

	"github.com/Permify/permify/internal/entities"
	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/pkg/dsl/schema"
	"github.com/Permify/permify/pkg/tuple"
)

// CheckVisitMap -
type CheckVisitMap map[string]CheckDecision

// isVisited -
func (v CheckVisitMap) isVisited(key string) (CheckDecision, bool) {
	if _, ok := v[key]; ok {
		return v[key], true
	}
	return CheckDecision{}, false
}

// set -
func (v CheckVisitMap) set(key string, decision CheckDecision) {
	v[key] = decision
}

// CheckDecision -
type CheckDecision struct {
	Prefix string `json:"prefix"`
	Can    bool   `json:"can"`
	Err    error  `json:"err"`
}

// sendCheckDecision -
func sendCheckDecision(can bool, prefix string, err error) CheckDecision {
	return CheckDecision{
		Prefix: prefix,
		Can:    can,
		Err:    err,
	}
}

// CheckFunction -
type CheckFunction func(ctx context.Context, decisionChan chan<- CheckDecision)

// CheckCombiner .
type CheckCombiner func(ctx context.Context, functions []CheckFunction) CheckDecision

// CheckCommand -
type CheckCommand struct {
	relationTupleRepository repositories.IRelationTupleRepository
}

// NewCheckCommand -
func NewCheckCommand(rr repositories.IRelationTupleRepository) *CheckCommand {
	return &CheckCommand{
		relationTupleRepository: rr,
	}
}

// CheckQuery -
type CheckQuery struct {
	Entity  tuple.Entity
	Subject tuple.Subject
	depth   int
}

// SetDepth -
func (r *CheckQuery) SetDepth(i int) {
	r.depth = i
}

// decrease -
func (r *CheckQuery) decrease() *CheckQuery {
	r.depth--
	return r
}

// isDepthFinish -
func (r *CheckQuery) isDepthFinish() bool {
	return r.depth <= 0
}

// CheckResponse -
type CheckResponse struct {
	Can            bool
	Visit          *CheckVisitMap
	RemainingDepth int
	Error          error
}

// Execute -
func (command *CheckCommand) Execute(ctx context.Context, q *CheckQuery, child schema.Child) (response CheckResponse) {
	response.Can = false
	vm := &CheckVisitMap{}
	response.Can, response.Visit, response.Error = command.c(ctx, q, child, vm)
	response.RemainingDepth = q.depth
	return
}

// c -
func (command *CheckCommand) c(ctx context.Context, q *CheckQuery, child schema.Child, vm *CheckVisitMap) (bool, *CheckVisitMap, error) {
	var fn CheckFunction
	switch child.GetKind() {
	case schema.RewriteKind.String():
		fn = command.checkRewrite(ctx, q, child.(schema.Rewrite), vm)
	case schema.LeafKind.String():
		fn = command.checkLeaf(ctx, q, child.(schema.Leaf), vm)
	}

	if fn == nil {
		return false, nil, UndefinedChildKindError
	}

	result := checkUnion(ctx, []CheckFunction{fn})

	return result.Can, vm, result.Err
}

// checkRewrite -
func (command *CheckCommand) checkRewrite(ctx context.Context, q *CheckQuery, child schema.Rewrite, vm *CheckVisitMap) CheckFunction {
	switch child.GetType() {
	case schema.Union.String():
		return command.set(ctx, q, child.Children, checkUnion, vm)
	case schema.Intersection.String():
		return command.set(ctx, q, child.Children, checkIntersection, vm)
	default:
		return checkFail(UndefinedChildTypeError)
	}
}

// checkLeaf -
func (command *CheckCommand) checkLeaf(ctx context.Context, q *CheckQuery, child schema.Leaf, vm *CheckVisitMap) CheckFunction {
	switch child.GetType() {
	case schema.TupleToUserSetType.String():
		return command.check(ctx, q.Entity, tuple.Relation(child.Value), q, child.Exclusion, vm)
	case schema.ComputedUserSetType.String():
		return command.check(ctx, q.Entity, tuple.Relation(child.Value), q, child.Exclusion, vm)
	default:
		return checkFail(UndefinedChildTypeError)
	}
}

// set -
func (command *CheckCommand) set(ctx context.Context, q *CheckQuery, children []schema.Child, combiner CheckCombiner, vm *CheckVisitMap) CheckFunction {
	var functions []CheckFunction
	for _, child := range children {
		switch child.GetKind() {
		case schema.RewriteKind.String():
			functions = append(functions, command.checkRewrite(ctx, q, child.(schema.Rewrite), vm))
		case schema.LeafKind.String():
			functions = append(functions, command.checkLeaf(ctx, q, child.(schema.Leaf), vm))
		default:
			return checkFail(UndefinedChildKindError)
		}
	}

	return func(ctx context.Context, resultChan chan<- CheckDecision) {
		resultChan <- combiner(ctx, functions)
	}
}

// check -
func (command *CheckCommand) check(ctx context.Context, entity tuple.Entity, relation tuple.Relation, q *CheckQuery, exclusion bool, vm *CheckVisitMap) CheckFunction {
	return func(ctx context.Context, decisionChan chan<- CheckDecision) {
		var err error

		q.decrease()

		if q.isDepthFinish() {
			decisionChan <- sendCheckDecision(false, "", DepthError)
			return
		}

		var iterator tuple.ISubjectIterator
		iterator, err = command.getSubjects(ctx, entity, relation)
		if err != nil {
			checkFail(err)
			return
		}

		var checkFunctions []CheckFunction

		for iterator.HasNext() {
			subject := iterator.GetNext()
			if subject.Equals(q.Subject) {
				var dec CheckDecision
				if exclusion {
					dec = sendCheckDecision(false, "not", err)
				} else {
					dec = sendCheckDecision(true, "", err)
				}
				decisionChan <- dec
				return
			} else {
				if !subject.IsUser() {
					checkFunctions = append(checkFunctions, command.check(ctx, tuple.Entity{ID: subject.ID, Type: subject.Type}, subject.Relation, q, exclusion, vm))
				}
			}
		}

		if len(checkFunctions) > 0 {
			decisionChan <- checkUnion(ctx, checkFunctions)
			return
		}

		var dec CheckDecision
		if exclusion {
			dec = sendCheckDecision(true, "not", err)
		} else {
			dec = sendCheckDecision(false, "", err)
		}
		decisionChan <- dec
		return
	}
}

// union -
func checkUnion(ctx context.Context, functions []CheckFunction) CheckDecision {
	if len(functions) == 0 {
		return sendCheckDecision(false, "", nil)
	}

	decisionChan := make(chan CheckDecision, len(functions))
	childCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	for _, fn := range functions {
		go fn(childCtx, decisionChan)
	}

	for i := 0; i < len(functions); i++ {
		select {
		case result := <-decisionChan:
			if result.Err == nil && result.Can {
				return sendCheckDecision(true, result.Prefix, nil)
			}
			if result.Err != nil {
				return sendCheckDecision(false, result.Prefix, result.Err)
			}
		case <-ctx.Done():
			return sendCheckDecision(false, "", CanceledError)
		}
	}

	return sendCheckDecision(false, "", nil)
}

// intersection -
func checkIntersection(ctx context.Context, functions []CheckFunction) CheckDecision {
	if len(functions) == 0 {
		return sendCheckDecision(false, "", nil)
	}

	decisionChan := make(chan CheckDecision, len(functions))
	childCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	for _, fn := range functions {
		go fn(childCtx, decisionChan)
	}

	for i := 0; i < len(functions); i++ {
		select {
		case result := <-decisionChan:
			if result.Err == nil && !result.Can {
				return sendCheckDecision(false, result.Prefix, nil)
			}
			if result.Err != nil {
				return sendCheckDecision(false, result.Prefix, result.Err)
			}
		case <-ctx.Done():
			return sendCheckDecision(false, "", CanceledError)
		}
	}

	return sendCheckDecision(true, "", nil)
}

// checkFail -
func checkFail(err error) CheckFunction {
	return func(ctx context.Context, decisionChan chan<- CheckDecision) {
		decisionChan <- sendCheckDecision(false, "", err)
	}
}

// getSubjects -
func (command *CheckCommand) getSubjects(ctx context.Context, entity tuple.Entity, relation tuple.Relation) (iterator tuple.ISubjectIterator, err error) {
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
