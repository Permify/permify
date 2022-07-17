package services

import (
	"context"
	"errors"
	`fmt`
	`sync`

	"github.com/Permify/permify/internal/entities"
	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/pkg/dsl/schema"
	"github.com/Permify/permify/pkg/tuple"
)

var (
	DepthError              = errors.New("depth error")
	CanceledError           = errors.New("canceled error")
	ActionCannotFoundError  = errors.New("action cannot found")
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

// add -
func (v VisitMap) set(key string, decision Decision) {
	v[key] = decision
}

// add -
func (v VisitMap) get(key string) (decision Decision) {
	return v[key]
}

// Decision -
type Decision struct {
	Can bool  `json:"can"`
	Err error `json:"err"`
}

// sendDecision -
func sendDecision(can bool, err error) Decision {
	return Decision{
		Can: can,
		Err: err,
	}
}

// CheckFunction -
type CheckFunction func(ctx context.Context, decisionChan chan<- Decision)

// Combiner .
type Combiner func(ctx context.Context, requests []CheckFunction) Decision

// IPermissionService -
type IPermissionService interface {
	Check(ctx context.Context, s string, a string, o string, d int) (bool, *VisitMap, int, error)
}

// PermissionService -
type PermissionService struct {
	repository repositories.IRelationTupleRepository
	schema     schema.Schema
}

// NewPermissionService -
func NewPermissionService(repo repositories.IRelationTupleRepository, schema schema.Schema) *PermissionService {
	return &PermissionService{
		repository: repo,
		schema:     schema,
	}
}

// Request -
type Request struct {
	Object  tuple.Object
	Subject tuple.User
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
func (service *PermissionService) Check(ctx context.Context, s string, a string, o string, d int) (can bool, visit *VisitMap, remainingDepth int, err error) {
	can = false

	var object tuple.Object
	object, err = tuple.ConvertObject(o)
	if err != nil {
		return false, nil, d, err
	}

	entity := service.schema.Entities[object.Namespace]

	var child schema.Child

	for _, act := range entity.Actions {
		if act.Name == a {
			child = act.Child
			goto check
		}
	}

	if child == nil {
		return false, nil, d, ActionCannotFoundError
	}

check:

	var vm = &VisitMap{}

	re := Request{
		Object:  object,
		Subject: tuple.ConvertUser(s),
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
		return service.check(ctx, request.Object, tuple.Relation(child.Value), request, vm)
	case schema.ComputedUserSetType.String():
		return service.check(ctx, request.Object, tuple.Relation(child.Value), request, vm)
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
		return sendDecision(false, nil)
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
				return sendDecision(true, nil)
			}
			if result.Err != nil {
				return sendDecision(false, result.Err)
			}
		case <-ctx.Done():
			return sendDecision(false, CanceledError)
		}
	}

	return sendDecision(false, nil)
}

// intersection -
func intersection(ctx context.Context, functions []CheckFunction) Decision {
	if len(functions) == 0 {
		return sendDecision(false, nil)
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
				return sendDecision(false, nil)
			}
			if result.Err != nil {
				return sendDecision(false, result.Err)
			}
		case <-ctx.Done():
			return sendDecision(false, CanceledError)
		}
	}

	return sendDecision(true, nil)
}

// getUsers -
func (service *PermissionService) getUsers(ctx context.Context, object tuple.Object, relation tuple.Relation) (users []tuple.User, err error) {

	r := relation.Split()

	var en []entities.RelationTuple
	en, err = service.repository.QueryTuples(ctx, object.Namespace, object.ID, r[0].String())
	if err != nil {
		return nil, err
	}

	for _, entity := range en {
		if entity.UsersetEntity != "" {
			user := tuple.User{
				UserSet: tuple.UserSet{
					Object: tuple.Object{
						Namespace: entity.UsersetEntity,
						ID:        entity.UsersetObjectID,
					},
				},
			}

			if entity.UsersetRelation == tuple.ELLIPSIS {
				user.UserSet.Relation = r[1]
			} else {
				user.UserSet.Relation = tuple.Relation(entity.UsersetRelation)
			}

			users = append(users, user)
		} else {
			users = append(users, tuple.User{
				ID: entity.UsersetObjectID,
			})
		}
	}

	return
}

// check -
func (service *PermissionService) check(ctx context.Context, object tuple.Object, relation tuple.Relation, re *Request, vm *VisitMap) CheckFunction {
	return func(ctx context.Context, decisionChan chan<- Decision) {
		var err error

		var key = fmt.Sprintf(tuple.OBJECT+tuple.RELATION, object.Namespace, object.ID, relation.String())

		if vm.isVisited(key) {
			decisionChan <- vm.get(key)
			return
		}

		re.decrease()

		if re.isDepthFinish() {
			decisionChan <- sendDecision(false, DepthError)
			return
		}

		var users []tuple.User
		users, err = service.getUsers(ctx, object, relation)
		if err != nil {
			fail(err)
			return
		}

		for _, t := range users {
			if t.Equals(re.Subject) {
				dec := sendDecision(true, err)
				vm.set(key, dec)
				decisionChan <- dec
				return
			} else {
				if !t.IsUser() {
					decisionChan <- union(ctx, []CheckFunction{service.check(ctx, t.UserSet.Object, t.UserSet.Relation, re, vm)})
					return
				}
			}
		}

		dec := sendDecision(false, err)
		vm.set(key, dec)
		decisionChan <- dec
		return
	}
}

// fail -
func fail(err error) CheckFunction {
	return func(ctx context.Context, decisionChan chan<- Decision) {
		decisionChan <- sendDecision(false, err)
	}
}
