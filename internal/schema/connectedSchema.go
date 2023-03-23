package schema

import (
	`errors`

	`github.com/Permify/permify/pkg/dsl/utils`
	base `github.com/Permify/permify/pkg/pb/base/v1`
)

// ConnectedSchemaGraph represents a graph of connected schema objects. The schema object contains definitions for entities,
// relationships, and permissions, and the graph is constructed by linking objects together based on their dependencies. The
// graph is used by the PermissionEngine to resolve permissions and expand user sets for a given request.
//
// Fields:
//   - schema: pointer to the base.SchemaDefinition that defines the schema objects in the graph
type ConnectedSchemaGraph struct {
	schema *base.SchemaDefinition
}

// NewConnectedGraph returns a new instance of ConnectedSchemaGraph with the specified base.SchemaDefinition as its schema.
// The schema object contains definitions for entities, relationships, and permissions, and is used to construct a graph of
// connected schema objects. The graph is used by the PermissionEngine to resolve permissions and expand user sets for a
// given request.
//
// Parameters:
//   - schema: pointer to the base.SchemaDefinition that defines the schema objects in the graph
//
// Returns:
//   - pointer to a new instance of ConnectedSchemaGraph with the specified schema object
func NewConnectedGraph(schema *base.SchemaDefinition) *ConnectedSchemaGraph {
	return &ConnectedSchemaGraph{
		schema: schema,
	}
}

// EntrypointKind is a string type that represents the kind of Entrypoint object. An Entrypoint object defines an entry point
// into the ConnectedSchemaGraph, which is used to resolve permissions and expand user sets for a given request.
//
// Values:
//   - RelationEntrypoint: represents an entry point into a relationship object in the schema graph
//   - TupleToUserSetEntrypoint: represents an entry point into a tuple-to-user-set object in the schema graph
//   - ComputedUserSetEntrypoint: represents an entry point into a computed user set object in the schema graph

type EntrypointKind string

const (
	RelationEntrypoint        EntrypointKind = "relation"
	TupleToUserSetEntrypoint  EntrypointKind = "tuple_to_user_set"
	ComputedUserSetEntrypoint EntrypointKind = "computed_user_set"
)

// Entrypoint represents an entry point into the ConnectedSchemaGraph, which is used to resolve permissions and expand user
// sets for a given request. The object contains a kind that specifies the type of entry point (e.g. relation, tuple-to-user-set),
// an entry point reference that identifies the specific entry point in the graph, and a tuple set relation reference that
// specifies the relation to use when expanding user sets for the entry point.
//
// Fields:
//   - Kind: EntrypointKind representing the type of entry point
//   - Entrypoint: pointer to a base.RelationReference that identifies the entry point in the schema graph
//   - TupleSetRelation: pointer to a base.RelationReference that specifies the relation to use when expanding user sets
//     for the entry point
type Entrypoint struct {
	Kind             EntrypointKind
	Entrypoint       *base.RelationReference
	TupleSetRelation *base.RelationReference
}

// EntrypointKind returns the kind of the Entrypoint object. The kind specifies the type of entry point (e.g. relation,
// tuple-to-user-set, computed user set).
//
// Returns:
//   - EntrypointKind representing the type of entry point
func (re Entrypoint) EntrypointKind() EntrypointKind {
	return re.Kind
}

// RelationshipEntryPoints returns a slice of Entrypoint objects that represent entry points into the ConnectedSchemaGraph
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
//   - slice of Entrypoint objects that represent entry points into the ConnectedSchemaGraph, or an error if the target or
//     source relation does not exist in the schema graph
func (g *ConnectedSchemaGraph) RelationshipEntryPoints(target, source *base.RelationReference) ([]Entrypoint, error) {
	entries, err := g.findEntryPoint(target, source, map[string]struct{}{})
	if err != nil {
		return nil, err
	}

	return entries, nil
}

// findEntryPoint is a recursive helper function that searches the ConnectedSchemaGraph for all entry points that can be reached
// from the specified target relation through the specified source relation. The function uses a depth-first search to traverse
// the schema graph and identify entry points, marking visited nodes in a map to avoid infinite recursion. If the target or
// source relation does not exist in the schema graph, the function returns an error. If the source relation is an action
// reference, the function recursively searches the graph for entry points reachable from the action child. If the source
// relation is a regular relational reference, the function delegates to findRelationEntryPoint to search for entry points.
//
// Parameters:
//   - target: pointer to a base.RelationReference that identifies the target relation
//   - source: pointer to a base.RelationReference that identifies the source relation used to reach the target relation
//   - visited: map used to track visited nodes and avoid infinite recursion
//
// Returns:
//   - slice of Entrypoint objects that represent entry points into the ConnectedSchemaGraph, or an error if the target or
//     source relation does not exist in the schema graph
func (g *ConnectedSchemaGraph) findEntryPoint(target, source *base.RelationReference, visited map[string]struct{}) ([]Entrypoint, error) {
	key := utils.Key(target.GetType(), target.GetRelation())
	if _, ok := visited[key]; ok {
		return nil, nil
	}
	visited[key] = struct{}{}

	def, ok := g.schema.EntityDefinitions[target.GetType()]
	if !ok {
		return nil, errors.New("entity definition not found")
	}

	if def.References[target.GetRelation()] == base.EntityDefinition_RELATIONAL_REFERENCE_ACTION {
		action, ok := def.Actions[target.GetRelation()]
		if !ok {
			return nil, nil
		}
		child := action.GetChild()
		if child.GetRewrite() != nil {
			return g.findEntryPointWithRewrite(target, source, action.GetChild().GetRewrite(), visited)
		}
		return g.findEntryPointWithLeaf(target, source, action.GetChild().GetLeaf(), visited)
	}

	return g.findRelationEntryPoint(target, source, visited)
}

// findRelationEntryPoint is a helper function that searches the ConnectedSchemaGraph for entry points that can be reached from
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
//   - slice of Entrypoint objects that represent entry points into the ConnectedSchemaGraph, or an error if the target or
//     source relation does not exist in the schema graph
func (g *ConnectedSchemaGraph) findRelationEntryPoint(target, source *base.RelationReference, visited map[string]struct{}) ([]Entrypoint, error) {
	var res []Entrypoint

	entity, ok := g.schema.EntityDefinitions[target.GetType()]
	if !ok {
		return nil, errors.New("entity definition not found")
	}

	relation, ok := entity.Relations[target.GetRelation()]
	if !ok {
		return nil, errors.New("relation definition not found")
	}

	if IsDirectlyRelated(relation, source) {
		res = append(res, Entrypoint{
			Kind: RelationEntrypoint,
			Entrypoint: &base.RelationReference{
				Type:     target.GetType(),
				Relation: target.GetRelation(),
			},
		})
	}

	for _, rel := range relation.GetRelationReferences() {
		if rel.GetRelation() != "" {
			entryPoints, err := g.findEntryPoint(rel, source, visited)
			if err != nil {
				return nil, err
			}
			res = append(res, entryPoints...)
		}
	}

	return res, nil
}

// findEntryPointWithLeaf is a helper function that searches the ConnectedSchemaGraph for entry points that can be reached from
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
//   - slice of Entrypoint objects that represent entry points into the ConnectedSchemaGraph, or an error if the target or
//     source relation does not exist in the schema graph
func (g *ConnectedSchemaGraph) findEntryPointWithLeaf(target, source *base.RelationReference, leaf *base.Leaf, visited map[string]struct{}) ([]Entrypoint, error) {
	switch t := leaf.GetType().(type) {
	case *base.Leaf_TupleToUserSet:
		tupleSet := t.TupleToUserSet.GetTupleSet().GetRelation()
		computedUserSet := t.TupleToUserSet.GetComputed().GetRelation()

		var res []Entrypoint

		relations := g.schema.EntityDefinitions[target.GetType()].Relations[tupleSet]

		for _, rel := range relations.GetRelationReferences() {
			if rel.GetType() == source.GetType() && source.GetRelation() == computedUserSet {
				res = append(res, Entrypoint{
					Kind: TupleToUserSetEntrypoint,
					Entrypoint: &base.RelationReference{
						Type:     target.GetType(),
						Relation: target.GetRelation(),
					},
					TupleSetRelation: &base.RelationReference{
						Type:     target.GetType(),
						Relation: tupleSet,
					},
				})
			}
			subResults, err := g.findEntryPoint(
				&base.RelationReference{
					Type:     rel.GetType(),
					Relation: computedUserSet,
				},
				source,
				visited,
			)
			if err != nil {
				return nil, err
			}
			res = append(res, subResults...)
		}
		return res, nil
	case *base.Leaf_ComputedUserSet:
		if target.GetType() == source.GetType() && t.ComputedUserSet.GetRelation() == source.GetRelation() {
			return []Entrypoint{
				{
					Kind: ComputedUserSetEntrypoint,
					Entrypoint: &base.RelationReference{
						Type:     target.GetType(),
						Relation: target.GetRelation(),
					},
				},
			}, nil
		}
		return g.findEntryPoint(
			&base.RelationReference{
				Type:     target.GetType(),
				Relation: t.ComputedUserSet.GetRelation(),
			},
			source,
			visited,
		)
	default:
		return nil, errors.New("undefined leaf type")
	}
}

// findEntryPointWithRewrite is a helper function that searches the ConnectedSchemaGraph for entry points that can be reached from
// the specified target relation through an action reference with a rewrite child. The function recursively searches each child of
// the rewrite and calls either findEntryPointWithRewrite or findEntryPointWithLeaf, depending on the child's type. The function
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
//   - slice of Entrypoint objects that represent entry points into the ConnectedSchemaGraph, or an error if the target or
//     source relation does not exist in the schema graph
func (g *ConnectedSchemaGraph) findEntryPointWithRewrite(
	target *base.RelationReference,
	source *base.RelationReference,
	rewrite *base.Rewrite,
	visited map[string]struct{},
) ([]Entrypoint, error) {
	var err error
	var res []Entrypoint
	for _, child := range rewrite.GetChildren() {
		var results []Entrypoint
		switch child.GetType().(type) {
		case *base.Child_Rewrite:
			results, err = g.findEntryPointWithRewrite(target, source, child.GetRewrite(), visited)
			if err != nil {
				return nil, err
			}
		case *base.Child_Leaf:
			results, err = g.findEntryPointWithLeaf(target, source, child.GetLeaf(), visited)
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
