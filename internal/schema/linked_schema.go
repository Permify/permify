package schema

import (
	"errors"

	"github.com/Permify/permify/pkg/dsl/utils"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// LinkedSchemaGraph represents a graph of linked schema objects. The schema object contains definitions for entities,
// relationships, and permissions, and the graph is constructed by linking objects together based on their dependencies. The
// graph is used by the PermissionEngine to resolve permissions and expand user sets for a given request.
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
	PathChainLinkedEntrance       LinkedEntranceKind = "path_chain"
)

// LinkedEntrance represents an entry point into the LinkedSchemaGraph, which is used to resolve permissions and expand user
// sets for a given request. The object contains a kind that specifies the type of entry point (e.g. relation, tuple-to-user-set),
// an entry point reference that identifies the specific entry point in the graph, and a tuple set relation reference that
// specifies the relation to use when expanding user sets for the entry point.
//
// Fields:
//   - Kind: LinkedEntranceKind representing the type of entry point
//   - TargetEntrance: pointer to a base.Entrance that identifies the entry point in the schema graph
//   - TupleSetRelation: string that specifies the relation to use when expanding user sets for the entry point
//   - PathChain: complete chain of relation references for multi-hop nested attributes
type LinkedEntrance struct {
	Kind             LinkedEntranceKind
	TargetEntrance   *base.Entrance
	TupleSetRelation string
	PathChain        []*base.RelationReference // Complete chain for multi-hop nested attributes
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

		// Cache computed relation path chains to avoid duplicate BuildRelationPathChain calls
		relationPathChainCache := make(map[string][]*base.RelationReference)

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

			// Handle nested attributes: create PathChainLinkedEntrance for cases with PathChain
			var filteredResults []*LinkedEntrance

			for _, result := range results {
				if target.GetType() != result.TargetEntrance.GetType() &&
					(result.Kind == AttributeLinkedEntrance || result.Kind == PathChainLinkedEntrance) {
					if result.Kind == PathChainLinkedEntrance && len(result.PathChain) > 0 {
						// Compose the existing path chain with the tuple-set relation to preserve the exact path.
						pathChain := make([]*base.RelationReference, 0, len(result.PathChain)+1)
						pathChain = append(pathChain, &base.RelationReference{
							Type:     target.GetType(),
							Relation: tupleSet,
						})
						pathChain = append(pathChain, result.PathChain...)
						res = append(res, &LinkedEntrance{
							Kind:             PathChainLinkedEntrance,
							TargetEntrance:   result.TargetEntrance,
							TupleSetRelation: "",
							PathChain:        pathChain,
						})
						continue
					}
					cacheKey := target.GetType() + "->" + result.TargetEntrance.GetType()

					var pathChain []*base.RelationReference
					var exists bool

					if pathChain, exists = relationPathChainCache[cacheKey]; !exists {
						var err error
						pathChain, err = g.BuildRelationPathChain(target.GetType(), result.TargetEntrance.GetType())
						if err == nil {
							relationPathChainCache[cacheKey] = pathChain
						}
					}

					if len(pathChain) > 0 {
						// Create PathChainLinkedEntrance for cases with PathChain
						res = append(res, &LinkedEntrance{
							Kind:             PathChainLinkedEntrance,
							TargetEntrance:   result.TargetEntrance,
							TupleSetRelation: "",
							PathChain:        pathChain,
						})
						// Skip adding AttributeLinkedEntrance for cases with PathChain
						continue
					}
				}
				// Non-nested or other types: keep as-is
				filteredResults = append(filteredResults, result)
			}

			res = append(res, filteredResults...)
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

// GetInverseRelation finds which relation connects the given entity types.
// Returns the relation name that connects sourceEntityType to targetEntityType.
func (g *LinkedSchemaGraph) GetInverseRelation(sourceEntityType, targetEntityType string) (string, error) {
	entityDef, exists := g.schema.EntityDefinitions[sourceEntityType]
	if !exists {
		return "", errors.New("source entity definition not found")
	}

	for relationName, relationDef := range entityDef.Relations {
		for _, relRef := range relationDef.RelationReferences {
			if relRef.GetType() == targetEntityType {
				return relationName, nil
			}
		}
	}

	return "", errors.New("no relation found connecting source to target entity type")
}

// BuildRelationPathChain builds the complete relation path chain for multi-hop nested attributes
func (g *LinkedSchemaGraph) BuildRelationPathChain(sourceEntityType, targetEntityType string) ([]*base.RelationReference, error) {
	// Try direct relation first
	relationName, err := g.GetInverseRelation(sourceEntityType, targetEntityType)
	if err == nil {
		// Direct relation exists, return single hop
		return []*base.RelationReference{
			{
				Type:     sourceEntityType,
				Relation: relationName,
			},
		}, nil
	}

	// Use BFS to find multi-hop path
	visited := make(map[string]bool)
	queue := []struct {
		entityType string
		path       []*base.RelationReference
	}{{sourceEntityType, []*base.RelationReference{}}}

	visited[sourceEntityType] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current.entityType == targetEntityType {
			return current.path, nil
		}

		entityDef, exists := g.schema.EntityDefinitions[current.entityType]
		if !exists {
			continue
		}

		// Explore all relations from current entity
		for relationName, relationDef := range entityDef.Relations {
			for _, relRef := range relationDef.RelationReferences {
				nextEntityType := relRef.GetType()
				if !visited[nextEntityType] {
					visited[nextEntityType] = true

					// Build new path
					newPath := make([]*base.RelationReference, len(current.path)+1)
					copy(newPath, current.path)
					newPath[len(current.path)] = &base.RelationReference{
						Type:     current.entityType,
						Relation: relationName,
					}

					queue = append(queue, struct {
						entityType string
						path       []*base.RelationReference
					}{nextEntityType, newPath})
				}
			}
		}
	}

	return nil, errors.New("no path found between entity types")
}

// GetSubjectRelationForPathWalk determines the correct subject relation for a given path walk
// This is needed to fix the Subject.Relation field in path chain traversal for complex relations like @group#member
func (g *LinkedSchemaGraph) GetSubjectRelationForPathWalk(leftEntityType, relationName, rightEntityType string) string {
	if entityDef, exists := g.schema.EntityDefinitions[leftEntityType]; exists {
		if relationDef, exists := entityDef.Relations[relationName]; exists {
			// Look for RelationReference that matches rightEntityType
			for _, relRef := range relationDef.GetRelationReferences() {
				if relRef.GetType() == rightEntityType {
					return relRef.GetRelation()
				}
			}
		}
	}
	return ""
}

// SelfCycleRelationsForPermission returns tuple-set relations that cause a permission
// to reference itself (e.g., view = parent.view).
func (g *LinkedSchemaGraph) SelfCycleRelationsForPermission(entityType, permission string) []string {
	entityDef, exists := g.schema.EntityDefinitions[entityType]
	if !exists {
		return nil
	}

	permDef, exists := entityDef.Permissions[permission]
	if !exists {
		return nil
	}

	seen := make(map[string]struct{})
	res := make([]string, 0)

	child := permDef.GetChild()
	if child == nil {
		return nil
	}

	g.collectSelfCycleRelations(entityType, permission, child, seen, &res)
	return res
}

func (g *LinkedSchemaGraph) collectSelfCycleRelations(entityType, permission string, child *base.Child, seen map[string]struct{}, res *[]string) {
	if child == nil {
		return
	}

	if child.GetRewrite() != nil {
		for _, c := range child.GetRewrite().GetChildren() {
			g.collectSelfCycleRelations(entityType, permission, c, seen, res)
		}
		return
	}

	leaf := child.GetLeaf()
	if leaf == nil {
		return
	}

	switch t := leaf.GetType().(type) {
	case *base.Leaf_TupleToUserSet:
		tupleSet := t.TupleToUserSet.GetTupleSet().GetRelation()
		computed := t.TupleToUserSet.GetComputed().GetRelation()
		if tupleSet == "" || computed == "" {
			return
		}
		if computed != permission {
			return
		}
		if _, ok := seen[tupleSet]; ok {
			return
		}
		entityDef, exists := g.schema.EntityDefinitions[entityType]
		if !exists {
			return
		}
		relDef, exists := entityDef.Relations[tupleSet]
		if !exists {
			return
		}
		// Only include relations that point back to the same entity type.
		for _, ref := range relDef.GetRelationReferences() {
			if ref.GetType() == entityType {
				seen[tupleSet] = struct{}{}
				*res = append(*res, tupleSet)
				return
			}
		}
	case *base.Leaf_ComputedUserSet, *base.Leaf_ComputedAttribute, *base.Leaf_Call:
		return
	}
}
