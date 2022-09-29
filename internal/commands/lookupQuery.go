package commands

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"

	internalErrors "github.com/Permify/permify/internal/errors"
	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/internal/repositories/entities"
	"github.com/Permify/permify/pkg/dsl/schema"
	"github.com/Permify/permify/pkg/errors"
	"github.com/Permify/permify/pkg/logger"
	"github.com/Permify/permify/pkg/tuple"
)

// BuildStatementResult -
type BuildStatementResult struct {
	Query query
	Err   errors.Error `json:"err"`
}

// query -
type query struct {
	Resolver ResolverFunction
	Map      string
	Join     string
	Vars     []string
}

func (q query) String() string {
	return ""
}

// sendBuildResult -
func sendBuildResult(query query, err errors.Error) BuildStatementResult {
	return BuildStatementResult{
		Query: query,
		Err:   err,
	}
}

// LookupQueryCommand -
type LookupQueryCommand struct {
	relationTupleRepository repositories.IRelationTupleRepository
	logger                  logger.Interface
	builder                 squirrel.StatementBuilderType
}

// BuildFunction -
type BuildFunction func(ctx context.Context, resultChan chan<- BuildStatementResult)

// ResolverFunction -
type ResolverFunction func() []string

// BuildCombiner .
type BuildCombiner func(ctx context.Context, functions []BuildFunction) BuildStatementResult

// NewLookupQueryCommand -
func NewLookupQueryCommand(rr repositories.IRelationTupleRepository, l logger.Interface) *LookupQueryCommand {
	return &LookupQueryCommand{
		logger:                  l,
		builder:                 squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		relationTupleRepository: rr,
	}
}

// LookupQueryQuery -
type LookupQueryQuery struct {
	EntityType string
	Action     string
	Subject    tuple.Subject
	schema     schema.Schema
}

func (l *LookupQueryQuery) SetSchema(sch schema.Schema) {
	l.schema = sch
}

// LookupQueryResponse -
type LookupQueryResponse struct {
	Results []BuildStatementResult
	Query   string
	Table   string
}

// Execute -
func (command *LookupQueryCommand) Execute(ctx context.Context, q *LookupQueryQuery, child schema.Child) (response LookupQueryResponse, err errors.Error) {
	response.Query, err = command.l(ctx, q, child)
	return
}

func (command *LookupQueryCommand) l(ctx context.Context, q *LookupQueryQuery, child schema.Child) (query string, err errors.Error) {
	var fn BuildFunction
	switch child.GetKind() {
	case schema.RewriteKind.String():
		fn = command.buildRewrite(ctx, q, child.(schema.Rewrite))
	case schema.LeafKind.String():
		fn = command.buildLeaf(ctx, q, child.(schema.Leaf))
	}

	if fn == nil {
		return "", internalErrors.UndefinedChildKindError
	}

	// result := buildUnion(ctx, []BuildFunction{fn})
	return "", nil
}

// set -
func (command *LookupQueryCommand) set(ctx context.Context, q *LookupQueryQuery, children []schema.Child, combiner BuildCombiner) BuildFunction {
	var functions []BuildFunction
	for _, child := range children {
		switch child.GetKind() {
		case schema.RewriteKind.String():
			functions = append(functions, command.buildRewrite(ctx, q, child.(schema.Rewrite)))
		case schema.LeafKind.String():
			functions = append(functions, command.buildLeaf(ctx, q, child.(schema.Leaf)))
		default:
			return buildFail(internalErrors.UndefinedChildKindError)
		}
	}

	return func(ctx context.Context, resultChan chan<- BuildStatementResult) {
		resultChan <- combiner(ctx, functions)
	}
}

// buildRewrite -
func (command *LookupQueryCommand) buildRewrite(ctx context.Context, q *LookupQueryQuery, child schema.Rewrite) BuildFunction {
	switch child.GetType() {
	case schema.Union.String():
		return command.set(ctx, q, child.Children, buildCombiner)
	case schema.Intersection.String():
		return command.set(ctx, q, child.Children, buildCombiner)
	default:
		return buildFail(internalErrors.UndefinedChildTypeError)
	}
}

// buildLeaf -
func (command *LookupQueryCommand) buildLeaf(ctx context.Context, q *LookupQueryQuery, child schema.Leaf) BuildFunction {
	switch child.GetType() {
	case schema.TupleToUserSetType.String():
		rel := tuple.Relation(child.Value)
		return command.build(ctx, q.EntityType, rel, child.Exclusion, command.buildTupleToUserSetQuery(q.Subject, q.EntityType, rel, q.schema))
	case schema.ComputedUserSetType.String():
		rel := tuple.Relation(child.Value)
		return command.build(ctx, q.EntityType, tuple.Relation(child.Value), child.Exclusion, command.buildComputedUserSetQuery(q.Subject, q.EntityType, rel, q.schema))
	default:
		return buildFail(internalErrors.UndefinedChildTypeError)
	}
}

// build -
func (command *LookupQueryCommand) build(ctx context.Context, entityType string, relation tuple.Relation, exclusion bool, q query) BuildFunction {
	return func(ctx context.Context, decisionChan chan<- BuildStatementResult) {
		qu := q
		qu.Vars = append(qu.Vars, q.Resolver()...)
		fmt.Println(qu.Map)
		fmt.Println(qu.Join)
		return
	}
}

// buildTupleToUserSetQuery -
func (command *LookupQueryCommand) buildTupleToUserSetQuery(subject tuple.Subject, entityType string, relation tuple.Relation, sch schema.Schema) query {
	var qu query
	var err error
	r := relation.Split()
	var re schema.Relation
	var entity schema.Entity
	entity, err = sch.GetEntityByName(entityType)
	if err != nil {
		return query{}
	}
	var rel schema.Relation
	rel, err = schema.Relations(entity.Relations).GetRelationByName(r[0].String())
	if err != nil {
		return query{}
	}
	column, columnExist := rel.GetColumn()
	var parentEntity schema.Entity
	parentEntity, err = sch.GetEntityByName(re.Type())
	if err != nil {
		return query{}
	}
	var parentRel schema.Relation
	parentRel, err = schema.Relations(parentEntity.Relations).GetRelationByName(r[1].String())
	if err != nil {
		return query{}
	}
	parentColumn, parentColumnExist := parentRel.GetColumn()
	if !columnExist {
		qu.Map = fmt.Sprintf("%s.%s", entity.GetTable(), entity.GetIdentifier())
		qu.Resolver = func() []string {
			parentIDs := command.getEntityIDs(context.TODO(), rel.Type(), tuple.Relation(r[1].String()), subject)
			return command.getEntityIDs(context.TODO(), entity.Name, tuple.Relation(rel.Name), tuple.Subject{
				Type:     rel.Type(),
				ID:       parentIDs[0],
				Relation: tuple.ELLIPSIS,
			})
		}
	} else {
		if !parentColumnExist {
			qu.Map = fmt.Sprintf("%s.%s", entity.GetTable(), column)
			qu.Resolver = func() []string {
				return command.getEntityIDs(context.TODO(), rel.Type(), tuple.Relation(r[1].String()), subject)
			}
		} else {
			qu.Map = fmt.Sprintf("%s.%s", parentEntity.GetTable(), parentColumn)
			qu.Join = fmt.Sprintf("%s ON %s.%s = %s.%s", parentEntity.GetTable(), entity.GetTable(), parentColumn, parentEntity.GetTable(), parentEntity.GetIdentifier())
			qu.Resolver = func() []string {
				return command.getEntityIDs(context.TODO(), rel.Type(), tuple.Relation(r[1].String()), subject)
			}
		}
	}
	return qu
}

// buildComputedUserSetQuery -
func (command *LookupQueryCommand) buildComputedUserSetQuery(subject tuple.Subject, entityType string, relation tuple.Relation, sch schema.Schema) (qu query) {
	var err error
	var entity schema.Entity
	entity, err = sch.GetEntityByName(entityType)
	if err != nil {
		return query{}
	}
	var rel schema.Relation
	rel, err = schema.Relations(entity.Relations).GetRelationByName(relation.String())
	if err != nil {
		return query{}
	}
	column, columnExist := rel.GetColumn()
	if !columnExist {
		qu.Map = fmt.Sprintf("%s.%s", entity.GetTable(), entity.GetIdentifier())
		qu.Resolver = func() []string {
			return command.getUserIDs(context.TODO(), subject)
		}
	} else {
		qu.Map = fmt.Sprintf("%s.%s", entity.GetTable(), column)
		qu.Resolver = func() []string {
			return command.getUserIDs(context.TODO(), subject)
		}
	}
	return qu
}

// buildUnion -
func buildCombiner(ctx context.Context, functions []BuildFunction) BuildStatementResult {
	if len(functions) == 0 {
		return sendBuildResult(query{}, nil)
	}

	builderChan := make(chan BuildStatementResult, len(functions))
	childCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	for _, fn := range functions {
		go fn(childCtx, builderChan)
	}

	for i := 0; i < len(functions); i++ {
		select {
		case result := <-builderChan:
			if result.Err == nil {
				return sendBuildResult(result.Query, nil)
			}
		case <-ctx.Done():
			return sendBuildResult(query{}, internalErrors.CanceledError)
		}
	}

	return sendBuildResult(query{}, nil)
}

// buildFail -
func buildFail(err errors.Error) BuildFunction {
	return func(ctx context.Context, decisionChan chan<- BuildStatementResult) {
		decisionChan <- sendBuildResult(query{}, err)
	}
}

// getUserIDs -
func (command *LookupQueryCommand) getUserIDs(ctx context.Context, s tuple.Subject) (r []string) {
	if s.IsUser() {
		r = append(r, s.ID)
	} else {
		iterator, err := command.getSubjects(ctx, tuple.Entity{
			Type: s.Type,
			ID:   s.ID,
		}, s.Relation)
		if err != nil {
			return
		}
		for iterator.HasNext() {
			r = append(r, command.getUserIDs(ctx, *iterator.GetNext())...)
		}
	}
	return
}

// getEntityIDs -
func (command *LookupQueryCommand) getEntityIDs(ctx context.Context, entityType string, relation tuple.Relation, subject tuple.Subject) (r []string) {
	iterator, err := command.getEntities(ctx, entityType, relation, subject)
	if err != nil {
		return
	}
	for iterator.HasNext() {
		entity := *iterator.GetNext()
		r = append(r, entity.ID)
	}
	return
}

// getSubjects -
func (command *LookupQueryCommand) getSubjects(ctx context.Context, entity tuple.Entity, relation tuple.Relation) (iterator tuple.ISubjectIterator, err errors.Error) {
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

// getEntities -
func (command *LookupQueryCommand) getEntities(ctx context.Context, entityType string, relation tuple.Relation, subject tuple.Subject) (iterator tuple.IEntityIterator, err errors.Error) {
	var tuples []entities.RelationTuple
	tuples, err = command.relationTupleRepository.ReverseQueryTuples(ctx, entityType, relation.String(), subject.Type, subject.ID, subject.Relation.String())
	if err != nil {
		return nil, err
	}

	var ent []*tuple.Entity
	for _, tup := range tuples {
		ct := tup.ToTuple()
		ent = append(ent, &ct.Entity)
	}

	return tuple.NewEntityIterator(ent), err
}
