package schema

import (
	"errors"
	"sync"

	"github.com/Permify/permify/pkg/dsl/utils"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// LinkedSchemaGraph represents a graph of linked schema objects. The schema object contains definitions for entities,
// relationships, and permissions, and the graph is constructed by linking objects together based on their dependencies. The
// graph is used by the PermissionEngine to resolve permissions and expand user sets for a given request.
//
// Fields:
//   - schema: pointer to the base.SchemaDefinition that defines the schema objects in the graph
type LinkedSchemaGraph struct {
	schema *base.SchemaDefinition
}

// NewLinkedGraph returns a new instance of LinkedSchemaGraph with the specified base.SchemaDefinition as its schema.
// The schema object contains definitions for entities, relationships, and permissions, and is used to construct a graph of
// linked schema objects. The graph is used by the PermissionEngine to resolve permissions and expand user sets for a
// given request.
//
// Parameters:
//   - schema: pointer to the base.SchemaDefinition that defines the schema objects in the graph
//
// Returns:
//   - pointer to a new instance of LinkedSchemaGraph with the specified schema object
func NewLinkedGraph(schema *base.SchemaDefinition) *LinkedSchemaGraph {
	return &LinkedSchemaGraph{
		schema: schema,
	}
}

// LinkedEntranceKind is a string type that represents the kind of LinkedEntrance object. An LinkedEntrance object defines an entry point
// into the LinkedSchemaGraph, which is used to resolve permissions and expand user sets for a given request.
//
// Values:
//   - RelationLinkedEntrance: represents an entry point into a relationship object in the schema graph
//   - TupleToUserSetLinkedEntrance: represents an entry point into a tuple-to-user-set object in the schema graph
//   - ComputedUserSetLinkedEntrance: represents an entry point into a computed user set object in the schema graph
type LinkedEntranceKind string

const (
	RelationLinkedEntrance        LinkedEntranceKind = "relation"
	TupleToUserSetLinkedEntrance  LinkedEntranceKind = "tuple_to_user_set"
	ComputedUserSetLinkedEntrance LinkedEntranceKind = "computed_user_set"
)

// LinkedEntrance represents an entry point into the LinkedSchemaGraph, which is used to resolve permissions and expand user
// sets for a given request. The object contains a kind that specifies the type of entry point (e.g. relation, tuple-to-user-set),
// an entry point reference that identifies the specific entry point in the graph, and a tuple set relation reference that
// specifies the relation to use when expanding user sets for the entry point.
//
// Fields:
//   - Kind: LinkedEntranceKind representing the type of entry point
//   - LinkedEntrance: pointer to a base.RelationReference that identifies the entry point in the schema graph
//   - TupleSetRelation: pointer to a base.RelationReference that specifies the relation to use when expanding user sets
//     for the entry point
type LinkedEntrance struct {
	Kind             LinkedEntranceKind
	TargetEntrance   *base.RelationReference
	TupleSetRelation string
}

// LinkedEntranceKind returns the kind of the LinkedEntrance object. The kind specifies the type of entry point (e.g. relation,
// tuple-to-user-set, computed user set).
//
// Returns:
//   - LinkedEntranceKind representing the type of entry point
func (re LinkedEntrance) LinkedEntranceKind() LinkedEntranceKind {
	return re.Kind
}

// RelationshipLinkedEntrances returns a slice of LinkedEntrance objects that represent entry points into the LinkedSchemaGraph
// for the specified target and source relations. The function recursively searches the graph for all entry points that can
// be reached from the target relation through the specified source relation. The resulting entry points contain a reference
// to the relation object in the schema graph and the relation used to expand user sets for the entry point. If the target or
// source relation does not exist in the schema graph, the function returns an error.
//
// Parameters:
//   - target: pointer to a base.RelationReference that identifies the target relation
//   - source: pointer to a base.RelationReference that identifies the source relation used to reach the target relation
//
// Returns:
//   - slice of LinkedEntrance objects that represent entry points into the LinkedSchemaGraph, or an error if the target or
//     source relation does not exist in the schema graph
func (g *LinkedSchemaGraph) RelationshipLinkedEntrances(targets []*base.RelationReference, source *base.RelationReference) ([]*LinkedEntrance, error) {
	visited := map[string]struct{}{}
	var mu sync.Mutex
	var wg sync.WaitGroup
	errChan := make(chan error, len(targets))
	done := make(chan struct{})
	totalEntries := make([]*LinkedEntrance, 0)
	for _, target := range targets {
		wg.Add(1)
		go func(target *base.RelationReference) {
			defer wg.Done()
			entries, err := g.findEntrance(target, source, visited, &mu)
			if err != nil {
				errChan <- err
			}
			mu.Lock()
			defer mu.Unlock()
			totalEntries = append(totalEntries, entries...)
		}(target)
	}

	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// All goroutines finished
		return totalEntries, nil
	case err := <-errChan:
		// Return on the first error encountered
		return nil, err
	}
}

// findEntrance is a recursive helper function that searches the LinkedSchemaGraph for all entry points that can be reached
// from the specified target relation through the specified source relation. The function uses a depth-first search to traverse
// the schema graph and identify entry points, marking visited nodes in a map to avoid infinite recursion. If the target or
// source relation does not exist in the schema graph, the function returns an error. If the source relation is an action
// reference, the function recursively searches the graph for entry points reachable from the action child. If the source
// relation is a regular relational reference, the function delegates to findRelationEntrance to search for entry points.
//
// Parameters:
//   - target: pointer to a base.RelationReference that identifies the target relation
//   - source: pointer to a base.RelationReference that identifies the source relation used to reach the target relation
//   - visited: map used to track visited nodes and avoid infinite recursion
//
// Returns:
//   - slice of LinkedEntrance objects that represent entry points into the LinkedSchemaGraph, or an error if the target or
//     source relation does not exist in the schema graph
func (g *LinkedSchemaGraph) findEntrance(target, source *base.RelationReference, visited map[string]struct{}, mu *sync.Mutex) ([]*LinkedEntrance, error) {
	key := utils.Key(target.GetType(), target.GetRelation())
	mu.Lock()
	mu.Lock()

	if _, ok := visited[key]; ok {
		mu.Unlock()
		return nil, nil
	}
	visited[key] = struct{}{}
	mu.Unlock()

	def, ok := g.schema.EntityDefinitions[target.GetType()]
	if !ok {
		return nil, errors.New("entity definition not found")
	}

	switch def.References[target.GetRelation()] {
	case base.EntityDefinition_REFERENCE_PERMISSION:
		permission, ok := def.Permissions[target.GetRelation()]
		if !ok {
			return nil, errors.New("permission not found")
		}
		child := permission.GetChild()
		if child.GetRewrite() != nil {
			return g.findEntranceRewrite(target, source, child.GetRewrite(), visited, mu)
		}
		return g.findEntranceLeaf(target, source, child.GetLeaf(), visited, mu)
	case base.EntityDefinition_REFERENCE_ATTRIBUTE:
		return nil, ErrUnimplemented
	case base.EntityDefinition_REFERENCE_RELATION:
		return g.findRelationEntrance(target, source, visited, mu)
	default:
		return nil, ErrUnimplemented
	}
}

// findRelationEntrance is a helper function that searches the LinkedSchemaGraph for entry points that can be reached from
// the specified target relation through the specified source relation. The function only returns entry points that are directly
// related to the target relation (i.e. the relation specified by the source reference is one of the relation's immediate children).
// The function recursively searches the children of the target relation and returns all reachable entry points. If the target
// or source relation does not exist in the schema graph, the function returns an error.
//
// Parameters:
//   - target: pointer to a base.RelationReference that identifies the target relation
//   - source: pointer to a base.RelationReference that identifies the source relation used to reach the target relation
//   - visited: map used to track visited nodes and avoid infinite recursion
//
// Returns:
//   - slice of LinkedEntrance objects that represent entry points into the LinkedSchemaGraph, or an error if the target or
//     source relation does not exist in the schema graph
func (g *LinkedSchemaGraph) findRelationEntrance(target, source *base.RelationReference, visited map[string]struct{}, mu *sync.Mutex) ([]*LinkedEntrance, error) {
	var res []*LinkedEntrance

	entity, ok := g.schema.EntityDefinitions[target.GetType()]
	if !ok {
		return nil, errors.New("entity definition not found")
	}

	relation, ok := entity.Relations[target.GetRelation()]
	if !ok {
		return nil, errors.New("relation definition not found")
	}

	if IsDirectlyRelated(relation, source) {
		res = append(res, &LinkedEntrance{
			Kind: RelationLinkedEntrance,
			TargetEntrance: &base.RelationReference{
				Type:     target.GetType(),
				Relation: target.GetRelation(),
			},
		})
	}

	for _, rel := range relation.GetRelationReferences() {
		if rel.GetRelation() != "" {
			entrances, err := g.findEntrance(rel, source, visited, mu)
			if err != nil {
				return nil, err
			}
			res = append(res, entrances...)
		}
	}

	return res, nil
}

// findEntranceWithLeaf is a helper function that searches the LinkedSchemaGraph for entry points that can be reached from
// the specified target relation through an action reference with a leaf child. The function searches for entry points that are
// reachable through a tuple-to-user-set or computed-user-set action. If the action child is a tuple-to-user-set action, the
// function recursively searches for entry points reachable through the child's tuple set relation and the child's computed user
// set relation. If the action child is a computed-user-set action, the function recursively searches for entry points reachable
// through the computed user set relation. The function only returns entry points that can be reached from the target relation
// using the specified source relation. If the target or source relation does not exist in the schema graph, the function returns
// an error.
//
// Parameters:
//   - target: pointer to a base.RelationReference that identifies the target relation
//   - source: pointer to a base.RelationReference that identifies the source relation used to reach the target relation
//   - leaf: pointer to a base.Leaf object that represents the child of an action reference
//   - visited: map used to track visited nodes and avoid infinite recursion
//
// Returns:
//   - slice of LinkedEntrance objects that represent entry points into the LinkedSchemaGraph, or an error if the target or
//     source relation does not exist in the schema graph
func (g *LinkedSchemaGraph) findEntranceLeaf(target, source *base.RelationReference, leaf *base.Leaf, visited map[string]struct{}, mu *sync.Mutex) ([]*LinkedEntrance, error) {
	switch t := leaf.GetType().(type) {
	case *base.Leaf_TupleToUserSet:
		tupleSet := t.TupleToUserSet.GetTupleSet().GetRelation()
		computedUserSet := t.TupleToUserSet.GetComputed().GetRelation()

		var res []*LinkedEntrance
		entityDefinitions, exists := g.schema.EntityDefinitions[target.GetType()]
		if !exists {
			return nil, errors.New("entity definition not found")
		}

		relations, exists := entityDefinitions.Relations[tupleSet]
		if !exists {
			return nil, errors.New("relation definition not found")
		}

		for _, rel := range relations.GetRelationReferences() {
			if rel.GetType() == source.GetType() && source.GetRelation() == computedUserSet {
				res = append(res, &LinkedEntrance{
					Kind:             TupleToUserSetLinkedEntrance,
					TargetEntrance:   target,
					TupleSetRelation: tupleSet,
				})
			}

			results, err := g.findEntrance(
				&base.RelationReference{
					Type:     rel.GetType(),
					Relation: computedUserSet,
				},
				source,
				visited,
				mu,
			)
			if err != nil {
				return nil, err
			}
			res = append(res, results...)
		}
		return res, nil
	case *base.Leaf_ComputedUserSet:

		var entrances []*LinkedEntrance

		if target.GetType() == source.GetType() && t.ComputedUserSet.GetRelation() == source.GetRelation() {
			entrances = append(entrances, &LinkedEntrance{
				Kind:           ComputedUserSetLinkedEntrance,
				TargetEntrance: target,
			})
		}

		results, err := g.findEntrance(
			&base.RelationReference{
				Type:     target.GetType(),
				Relation: t.ComputedUserSet.GetRelation(),
			},
			source,
			visited,
			mu,
		)
		if err != nil {
			return nil, err
		}

		entrances = append(
			entrances,
			results...,
		)
		return entrances, nil
	case *base.Leaf_ComputedAttribute:
		return nil, ErrUnimplemented
	case *base.Leaf_Call:
		return nil, ErrUnimplemented
	default:
		return nil, ErrUndefinedLeafType
	}
}

// findEntranceWithRewrite is a helper function that searches the LinkedSchemaGraph for entry points that can be reached from
// the specified target relation through an action reference with a rewrite child. The function recursively searches each child of
// the rewrite and calls either findEntranceWithRewrite or findEntranceWithLeaf, depending on the child's type. The function
// only returns entry points that can be reached from the target relation using the specified source relation. If the target or
// source relation does not exist in the schema graph, the function returns an error.
//
// Parameters:
//   - target: pointer to a base.RelationReference that identifies the target relation
//   - source: pointer to a base.RelationReference that identifies the source relation used to reach the target relation
//   - rewrite: pointer to a base.Rewrite object that represents the child of an action reference
//   - visited: map used to track visited nodes and avoid infinite recursion
//
// Returns:
//   - slice of LinkedEntrance objects that represent entry points into the LinkedSchemaGraph, or an error if the target or
//     source relation does not exist in the schema graph
func (g *LinkedSchemaGraph) findEntranceRewrite(target, source *base.RelationReference, rewrite *base.Rewrite, visited map[string]struct{}, mu *sync.Mutex) (results []*LinkedEntrance, err error) {
	var res []*LinkedEntrance
	for _, child := range rewrite.GetChildren() {
		switch child.GetType().(type) {
		case *base.Child_Rewrite:
			results, err = g.findEntranceRewrite(target, source, child.GetRewrite(), visited, mu)
			if err != nil {
				return nil, err
			}
		case *base.Child_Leaf:
			results, err = g.findEntranceLeaf(target, source, child.GetLeaf(), visited, mu)
			if err != nil {
				return nil, err
			}
		default:
			return nil, errors.New("undefined child type")
		}
		res = append(res, results...)
	}
	return res, nil
}
