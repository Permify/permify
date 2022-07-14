package services

import (
	"context"
	"errors"
	"strings"

	"github.com/Permify/permify/internal/entities"
	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/pkg/dsl/schema"
	"github.com/Permify/permify/pkg/tuple"
)

var (
	DepthError             = errors.New("depth error")
	CanceledError          = errors.New("canceled error")
	ActionCannotFoundError = errors.New("action cannot found")
)

// Decision -
type Decision struct {
	Can bool
	Err error
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
	Check(ctx context.Context, s string, a string, o string, d int) (bool, error)
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

type Request struct {
	Object  tuple.Object
	Subject tuple.User
}

// Check -
func (service *PermissionService) Check(ctx context.Context, s string, a string, o string, d int) (can bool, err error) {
	can = false

	var object tuple.Object
	object, err = tuple.ConvertObject(o)
	if err != nil {
		return false, err
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
		return false, ActionCannotFoundError
	}

check:

	re := Request{
		Object:  object,
		Subject: tuple.ConvertUser(s),
	}

	return service.c(ctx, re, child, d)
}

// c -
func (service *PermissionService) c(ctx context.Context, request Request, child schema.Child, d int) (bool, error) {
	var fn CheckFunction

	switch child.GetKind() {
	case schema.RewriteKind.String():
		fn = service.checkRewrite(ctx, request, child.(schema.Rewrite), d)
	case schema.LeafKind.String():
		fn = service.checkLeaf(ctx, request, child.(schema.Leaf), d)
	}

	result := union(ctx, []CheckFunction{fn})

	return result.Can, result.Err
}

// checkRewrite -
func (service *PermissionService) checkRewrite(ctx context.Context, request Request, child schema.Rewrite, d int) CheckFunction {
	switch child.GetType() {
	case schema.Union.String():
		return service.set(ctx, request, child.Children, union, d)
	case schema.Intersection.String():
		return service.set(ctx, request, child.Children, intersection, d)
	default:
		return nil
	}
}

// checkRewrite -
func (service *PermissionService) checkLeaf(ctx context.Context, request Request, child schema.Leaf, d int) CheckFunction {
	switch child.GetType() {
	case schema.TupleToUserSetType.String():
		return service.check(ctx, request.Object, child.Value, request.Subject, d)
	case schema.ComputedUserSetType.String():
		return service.check(ctx, request.Object, child.Value, request.Subject, d)
	default:
		return nil
	}
}

// setOperation -
func (service *PermissionService) set(ctx context.Context, request Request, children []schema.Child, combiner Combiner, d int) CheckFunction {
	var functions []CheckFunction

	for _, child := range children {
		switch child.GetKind() {
		case schema.RewriteKind.String():
			functions = append(functions, service.checkRewrite(ctx, request, child.(schema.Rewrite), d))
		case schema.LeafKind.String():
			functions = append(functions, service.checkLeaf(ctx, request, child.(schema.Leaf), d))
		default:
			return nil
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
func (service *PermissionService) getUsers(ctx context.Context, object tuple.Object, relation string, optRel string) (users []tuple.User, err error) {
	var en []entities.RelationTuple
	en, err = service.repository.QueryTuples(ctx, object.Namespace, object.ID, relation)
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
				user.UserSet.Relation = optRel
			} else {
				user.UserSet.Relation = entity.UsersetRelation
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
func (service *PermissionService) check(ctx context.Context, object tuple.Object, relation string, subject tuple.User, d int) CheckFunction {
	return func(ctx context.Context, decisionChan chan<- Decision) {
		var err error

		if d <= 0 {
			decisionChan <- sendDecision(false, DepthError)
		}

		var users []tuple.User
		r := strings.Split(relation, ".")
		if len(r) > 1 {
			users, err = service.getUsers(ctx, object, r[0], r[1])
		} else {
			users, err = service.getUsers(ctx, object, relation, "")
		}

		if err != nil {
			decisionChan <- sendDecision(false, err)
			return
		}

		for _, t := range users {
			if t.Equals(subject) {
				decisionChan <- sendDecision(true, err)
				return
			} else {
				if !t.IsUser() {
					decisionChan <- union(ctx, []CheckFunction{service.check(ctx, t.UserSet.Object, t.UserSet.Relation, subject, d-1)})
					return
				}
			}
		}

		decisionChan <- sendDecision(false, err)
		return
	}
}
