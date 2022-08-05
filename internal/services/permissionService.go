package services

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/Permify/permify/internal/entities"
	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/pkg/dsl/schema"
	"github.com/Permify/permify/pkg/tuple"
)

var (
	DepthError              = errors.New("depth error")
	CanceledError           = errors.New("canceled error")
	ActionCannotFoundError  = errors.New("action cannot found")
	EntityCannotFoundError  = errors.New("entity cannot found")
	UndefinedChildTypeError = errors.New("undefined child type")
	UndefinedChildKindError = errors.New("undefined child kind")
)

// VisitMap -
type VisitMap map[string]Decision

// isVisited -
func (v VisitMap) isVisited(key string) bool {
	if _, ok := v[key]; ok {
		return true
	}
	return false
}

// set -
func (v VisitMap) set(key string, decision Decision) {
	v[key] = decision
}

// get -
func (v VisitMap) get(key string) (decision Decision) {
	return v[key]
}

// Decision -
type Decision struct {
	Prefix string `json:"prefix"`
	Can    bool   `json:"can"`
	Err    error  `json:"err"`
}

// sendDecision -
func sendDecision(can bool, prefix string, err error) Decision {
	return Decision{
		Prefix: prefix,
		Can:    can,
		Err:    err,
	}
}

// CheckFunction -
type CheckFunction func(ctx context.Context, decisionChan chan<- Decision)

// Combiner .
type Combiner func(ctx context.Context, requests []CheckFunction) Decision

// IPermissionService -
type IPermissionService interface {
	Check(ctx context.Context, subject tuple.Subject, action string, entity tuple.Entity, d int) (bool, *VisitMap, int, error)
}

// PermissionService -
type PermissionService struct {
	relationTupleRepository repositories.IRelationTupleRepository
	schemaService           ISchemaService
}

// NewPermissionService -
func NewPermissionService(rr repositories.IRelationTupleRepository, ss ISchemaService) *PermissionService {
	return &PermissionService{
		relationTupleRepository: rr,
		schemaService:           ss,
	}
}

// Request -
type Request struct {
	Entity  tuple.Entity
	Subject tuple.Subject
	depth   int
	mux     sync.Mutex
}

// SetDepth -
func (r *Request) SetDepth(i int) {
	r.depth = i
}

// decrease -
func (r *Request) decrease() *Request {
	r.mux.Lock()
	defer r.mux.Unlock()
	r.depth--
	return r
}

// isDepthFinish -
func (r *Request) isDepthFinish() bool {
	r.mux.Lock()
	defer r.mux.Unlock()
	return r.depth <= 0
}

// Check -
func (service *PermissionService) Check(ctx context.Context, subject tuple.Subject, action string, entity tuple.Entity, d int) (can bool, visit *VisitMap, remainingDepth int, err error) {
	can = false

	var sch schema.Schema
	sch, err = service.schemaService.Schema(ctx)
	if err != nil {
		return false, nil, d, EntityCannotFoundError
	}

	var child schema.Child
	child = sch.GetEntityByName(entity.Type).GetAction(action).Child
	if child == nil {
		return false, nil, d, ActionCannotFoundError
	}

	vm := &VisitMap{}

	re := Request{
		Entity:  entity,
		Subject: subject,
	}

	re.SetDepth(d)

	can, visit, err = service.c(ctx, &re, child, vm)
	remainingDepth = re.depth

	return
}

// c -
func (service *PermissionService) c(ctx context.Context, request *Request, child schema.Child, vm *VisitMap) (bool, *VisitMap, error) {
	var fn CheckFunction

	switch child.GetKind() {
	case schema.RewriteKind.String():
		fn = service.checkRewrite(ctx, request, child.(schema.Rewrite), vm)
	case schema.LeafKind.String():
		fn = service.checkLeaf(ctx, request, child.(schema.Leaf), vm)
	}

	if fn == nil {
		return false, nil, UndefinedChildKindError
	}

	result := union(ctx, []CheckFunction{fn})

	return result.Can, vm, result.Err
}

// checkRewrite -
func (service *PermissionService) checkRewrite(ctx context.Context, request *Request, child schema.Rewrite, vm *VisitMap) CheckFunction {
	switch child.GetType() {
	case schema.Union.String():
		return service.set(ctx, request, child.Children, union, vm)
	case schema.Intersection.String():
		return service.set(ctx, request, child.Children, intersection, vm)
	default:
		return fail(UndefinedChildTypeError)
	}
}

// checkLeaf -
func (service *PermissionService) checkLeaf(ctx context.Context, request *Request, child schema.Leaf, vm *VisitMap) CheckFunction {
	switch child.GetType() {
	case schema.TupleToUserSetType.String():
		return service.check(ctx, request.Entity, tuple.Relation(child.Value), request, child.Exclusion, vm)
	case schema.ComputedUserSetType.String():
		return service.check(ctx, request.Entity, tuple.Relation(child.Value), request, child.Exclusion, vm)
	default:
		return fail(UndefinedChildTypeError)
	}
}

// set -
func (service *PermissionService) set(ctx context.Context, request *Request, children []schema.Child, combiner Combiner, vm *VisitMap) CheckFunction {
	var functions []CheckFunction
	for _, child := range children {
		switch child.GetKind() {
		case schema.RewriteKind.String():
			functions = append(functions, service.checkRewrite(ctx, request, child.(schema.Rewrite), vm))
		case schema.LeafKind.String():
			functions = append(functions, service.checkLeaf(ctx, request, child.(schema.Leaf), vm))
		default:
			return fail(UndefinedChildKindError)
		}
	}

	return func(ctx context.Context, resultChan chan<- Decision) {
		resultChan <- combiner(ctx, functions)
	}
}

// union -
func union(ctx context.Context, functions []CheckFunction) Decision {
	if len(functions) == 0 {
		return sendDecision(false, "", nil)
	}

	decisionChan := make(chan Decision, len(functions))
	childCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	for _, fn := range functions {
		go fn(childCtx, decisionChan)
	}

	for i := 0; i < len(functions); i++ {
		select {
		case result := <-decisionChan:
			if result.Err == nil && result.Can {
				return sendDecision(true, result.Prefix, nil)
			}
			if result.Err != nil {
				return sendDecision(false, result.Prefix, result.Err)
			}
		case <-ctx.Done():
			return sendDecision(false, "", CanceledError)
		}
	}

	return sendDecision(false, "", nil)
}

// intersection -
func intersection(ctx context.Context, functions []CheckFunction) Decision {
	if len(functions) == 0 {
		return sendDecision(false, "", nil)
	}

	decisionChan := make(chan Decision, len(functions))
	childCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	for _, fn := range functions {
		go fn(childCtx, decisionChan)
	}

	for i := 0; i < len(functions); i++ {
		select {
		case result := <-decisionChan:
			if result.Err == nil && !result.Can {
				return sendDecision(false, result.Prefix, nil)
			}
			if result.Err != nil {
				return sendDecision(false, result.Prefix, result.Err)
			}
		case <-ctx.Done():
			return sendDecision(false, "", CanceledError)
		}
	}

	return sendDecision(true, "", nil)
}

// getUsers -
func (service *PermissionService) getUsers(ctx context.Context, entity tuple.Entity, relation tuple.Relation) (subjects []tuple.Subject, err error) {
	r := relation.Split()

	var tuples []entities.RelationTuple
	tuples, err = service.relationTupleRepository.QueryTuples(ctx, entity.Type, entity.ID, r[0].String())
	if err != nil {
		return nil, err
	}

	for _, tup := range tuples {
		ct := tup.ToTuple()
		if !ct.Subject.IsUser() {
			subject := ct.Subject

			if tup.UsersetRelation == tuple.ELLIPSIS {
				subject.Relation = r[1]
			} else {
				subject.Relation = ct.Subject.Relation
			}

			subjects = append(subjects, subject)
		} else {
			subjects = append(subjects, tuple.Subject{
				Type: tuple.USER,
				ID:   tup.UsersetObjectID,
			})
		}
	}

	return
}

// check -
func (service *PermissionService) check(ctx context.Context, entity tuple.Entity, relation tuple.Relation, re *Request, exclusion bool, vm *VisitMap) CheckFunction {
	return func(ctx context.Context, decisionChan chan<- Decision) {
		var err error

		key := fmt.Sprintf(tuple.ENTITY+tuple.RELATION, entity.Type, entity.ID, relation.String())

		if vm.isVisited(key) {
			decisionChan <- vm.get(key)
			return
		}

		re.decrease()

		if re.isDepthFinish() {
			decisionChan <- sendDecision(false, "", DepthError)
			return
		}

		var subjects []tuple.Subject
		subjects, err = service.getUsers(ctx, entity, relation)
		if err != nil {
			fail(err)
			return
		}

		for _, sub := range subjects {
			if sub.Equals(re.Subject) {
				var dec Decision
				if exclusion {
					dec = sendDecision(false, "not", err)
				} else {
					dec = sendDecision(true, "", err)
				}
				vm.set(key, dec)
				decisionChan <- dec
				return
			} else {
				if !sub.IsUser() {
					decisionChan <- union(ctx, []CheckFunction{service.check(ctx, tuple.Entity{ID: sub.ID, Type: sub.Type}, sub.Relation, re, exclusion, vm)})
					return
				}
			}
		}

		var dec Decision
		if exclusion {
			dec = sendDecision(true, "not", err)
		} else {
			dec = sendDecision(false, "", err)
		}
		vm.set(key, dec)
		decisionChan <- dec
		return
	}
}

// fail -
func fail(err error) CheckFunction {
	return func(ctx context.Context, decisionChan chan<- Decision) {
		decisionChan <- sendDecision(false, "", err)
	}
}
