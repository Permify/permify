package schema

import (
	"errors"

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
	AttributeLinkedEntrance       LinkedEntranceKind = "attribute"
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
	TargetEntrance   *base.Entrance
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
func (g *LinkedSchemaGraph) LinkedEntrances(target, source *base.Entrance) ([]*LinkedEntrance, error) {
	entries, err := g.findEntrance(target, source, map[string]struct{}{})
	if err != nil {
		return nil, err
	}

	return entries, nil
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
func (g *LinkedSchemaGraph) findEntrance(target, source *base.Entrance, visited map[string]struct{}) ([]*LinkedEntrance, error) {
	key := utils.Key(target.GetType(), target.GetValue())
	if _, ok := visited[key]; ok {
		return nil, nil
	}
	visited[key] = struct{}{}

	def, ok := g.schema.EntityDefinitions[target.GetType()]
	if !ok {
		return nil, errors.New("entity definition not found")
	}

	switch def.References[target.GetValue()] {
	case base.EntityDefinition_REFERENCE_PERMISSION:
		permission, ok := def.Permissions[target.GetValue()]
		if !ok {
			return nil, errors.New("permission not found")
		}
		child := permission.GetChild()
		if child.GetRewrite() != nil {
			return g.findEntranceRewrite(target, source, child.GetRewrite(), visited)
		}
		return g.findEntranceLeaf(target, source, child.GetLeaf(), visited)
	case base.EntityDefinition_REFERENCE_ATTRIBUTE:
		attribute, ok := def.Attributes[target.GetValue()]
		if !ok {
			return nil, errors.New("attribute not found")
		}
		return []*LinkedEntrance{
			{
				Kind: AttributeLinkedEntrance,
				TargetEntrance: &base.Entrance{
					Type:  target.GetType(),
					Value: attribute.GetName(),
				},
			},
		}, nil
	case base.EntityDefinition_REFERENCE_RELATION:
		return g.findRelationEntrance(target, source, visited)
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
func (g *LinkedSchemaGraph) findRelationEntrance(target, source *base.Entrance, visited map[string]struct{}) ([]*LinkedEntrance, error) {
	var res []*LinkedEntrance

	entity, ok := g.schema.EntityDefinitions[target.GetType()]
	if !ok {
		return nil, errors.New("entity definition not found")
	}

	relation, ok := entity.Relations[target.GetValue()]
	if !ok {
		return nil, errors.New("relation definition not found")
	}

	if IsDirectlyRelated(relation, source) {
		res = append(res, &LinkedEntrance{
			Kind: RelationLinkedEntrance,
			TargetEntrance: &base.Entrance{
				Type:  target.GetType(),
				Value: target.GetValue(),
			},
		})
	}

	for _, rel := range relation.GetRelationReferences() {
		if rel.GetRelation() != "" {
			entrances, err := g.findEntrance(&base.Entrance{
				Type:  rel.GetType(),
				Value: rel.GetRelation(),
			}, source, visited)
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
func (g *LinkedSchemaGraph) findEntranceLeaf(target, source *base.Entrance, leaf *base.Leaf, visited map[string]struct{}) ([]*LinkedEntrance, error) {
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
			if rel.GetType() == source.GetType() && source.GetValue() == computedUserSet {
				res = append(res, &LinkedEntrance{
					Kind:             TupleToUserSetLinkedEntrance,
					TargetEntrance:   target,
					TupleSetRelation: tupleSet,
				})
			}

			results, err := g.findEntrance(
				&base.Entrance{
					Type:  rel.GetType(),
					Value: computedUserSet,
				},
				source,
				visited,
			)
			if err != nil {
				return nil, err
			}
			res = append(res, results...)
		}
		return res, nil
	case *base.Leaf_ComputedUserSet:
		var entrances []*LinkedEntrance

		if target.GetType() == source.GetType() && t.ComputedUserSet.GetRelation() == source.GetValue() {
			entrances = append(entrances, &LinkedEntrance{
				Kind:           ComputedUserSetLinkedEntrance,
				TargetEntrance: target,
			})
		}

		results, err := g.findEntrance(
			&base.Entrance{
				Type:  target.GetType(),
				Value: t.ComputedUserSet.GetRelation(),
			},
			source,
			visited,
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
		var entrances []*LinkedEntrance
		entrances = append(entrances, &LinkedEntrance{
			Kind: AttributeLinkedEntrance,
			TargetEntrance: &base.Entrance{
				Type:  target.GetType(),
				Value: t.ComputedAttribute.GetName(),
			},
		})
		return entrances, nil
	case *base.Leaf_Call:
		var entrances []*LinkedEntrance
		for _, arg := range t.Call.GetArguments() {
			computedAttr := arg.GetComputedAttribute()
			if computedAttr != nil {
				entrances = append(entrances, &LinkedEntrance{
					Kind: AttributeLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  target.GetType(),
						Value: computedAttr.GetName(),
					},
				})
			}
		}
		return entrances, nil
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
func (g *LinkedSchemaGraph) findEntranceRewrite(target, source *base.Entrance, rewrite *base.Rewrite, visited map[string]struct{}) (results []*LinkedEntrance, err error) {
	var res []*LinkedEntrance
	for _, child := range rewrite.GetChildren() {
		switch child.GetType().(type) {
		case *base.Child_Rewrite:
			results, err = g.findEntranceRewrite(target, source, child.GetRewrite(), visited)
			if err != nil {
				return nil, err
			}
		case *base.Child_Leaf:
			results, err = g.findEntranceLeaf(target, source, child.GetLeaf(), visited)
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
