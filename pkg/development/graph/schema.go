package graph

import (
	"errors"
	"fmt"
	"strings"

	"github.com/rs/xid"

	"github.com/Permify/permify/internal/schema"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
)

// SchemaToGraph takes a schema definition and converts it into a graph
// representation, returning the created graph and an error if any occurs.
func SchemaToGraph(schema *base.SchemaDefinition) (g Graph, err error) {
	// Iterate through entity definitions in the schema
	for _, en := range schema.GetEntityDefinitions() {
		// Convert each entity into a graph
		eg, err := EntityToGraph(en)
		if err != nil {
			return Graph{}, err
		}
		// Add the nodes and edges from the entity graph to the schema graph
		g.AddNodes(eg.Nodes())
		g.AddEdges(eg.Edges())
	}
	return
}

// EntityToGraph takes an entity definition and converts it into a graph
// representation, returning the created graph and an error if any occurs.
func EntityToGraph(entity *base.EntityDefinition) (g Graph, err error) {
	// Create a node for the entity
	enNode := &Node{
		Type:  "entity",
		ID:    entity.GetName(),
		Label: entity.GetName(),
	}
	g.AddNode(enNode)

	// Iterate through the relations in the entity
	for _, re := range entity.GetRelations() {
		// Create a node for each relation
		reNode := &Node{
			Type:  "relation",
			ID:    fmt.Sprintf("%s#%s", entity.GetName(), re.GetName()),
			Label: re.Name,
		}

		// Iterate through the relation references
		for _, ref := range re.GetRelationReferences() {
			if ref.GetRelation() != "" {
				g.AddEdge(reNode, &Node{
					Type:  "relation",
					ID:    fmt.Sprintf("%s#%s", ref.GetType(), ref.GetRelation()),
					Label: re.Name,
				}, nil)
			} else {
				g.AddEdge(reNode, &Node{
					Type:  "entity",
					ID:    fmt.Sprintf("%s", ref.GetType()),
					Label: re.Name,
				}, nil)
			}
		}

		// Add relation node and edge to the graph
		g.AddNode(reNode)
		g.AddEdge(enNode, reNode, nil)
	}

	// Iterate through the permissions in the entity
	for _, permission := range entity.GetPermissions() {
		// Create a node for each permission
		acNode := &Node{
			Type:  "permission",
			ID:    fmt.Sprintf("%s#%s", entity.GetName(), permission.GetName()),
			Label: permission.GetName(),
		}
		g.AddNode(acNode)
		g.AddEdge(enNode, acNode, nil)
		// Build permission graph for each permission
		ag, err := buildPermissionGraph(entity, acNode, []*base.Child{permission.GetChild()})
		if err != nil {
			return Graph{}, err
		}
		// Add nodes and edges from permission graph to entity graph
		g.AddNodes(ag.Nodes())
		g.AddEdges(ag.Edges())
	}
	return
}

// buildActionGraph creates a permission graph for the given entity and node,
// and recursively processes the children of the node. Returns the created
// graph and an error if any occurs.
func buildPermissionGraph(entity *base.EntityDefinition, from *Node, children []*base.Child) (g Graph, err error) {
	// Iterate through the children
	for _, child := range children {
		switch child.GetType().(type) {
		case *base.Child_Rewrite:
			// Create a node for the rewrite operation
			rw := &Node{
				Type:  "operation",
				ID:    xid.New().String(),
				Label: child.GetRewrite().GetRewriteOperation().String(),
			}

			// Add the rewrite node to the graph and connect it to the parent node
			g.AddNode(rw)
			g.AddEdge(from, rw, child.GetExclusion())
			// Recursively process the children of the rewrite node
			ag, err := buildPermissionGraph(entity, rw, child.GetRewrite().GetChildren())
			if err != nil {
				return Graph{}, err
			}
			// Add the nodes and edges from the child graph to the current graph
			g.AddNodes(ag.Nodes())
			g.AddEdges(ag.Edges())
		case *base.Child_Leaf:
			// Process the leaf node
			leaf := child.GetLeaf()

			switch leaf.GetType().(type) {
			case *base.Leaf_TupleToUserSet:
				// Find the relation in the entity definition
				re, err := schema.GetRelationByNameInEntityDefinition(entity, leaf.GetTupleToUserSet().GetTupleSet().GetRelation())
				if err != nil {
					return Graph{}, errors.New(base.ErrorCode_ERROR_CODE_RELATION_DEFINITION_NOT_FOUND.String())
				}

				// Add an edge between the parent node and the tuple set relation node
				g.AddEdge(from, &Node{
					Type:  "relation",
					ID:    fmt.Sprintf("%s#%s", GetTupleSetReferenceReference(re), leaf.GetTupleToUserSet().GetComputed().GetRelation()),
					Label: leaf.GetTupleToUserSet().GetComputed().GetRelation(),
				}, child.GetExclusion())

			case *base.Leaf_ComputedUserSet:
				// Add an edge between the parent node and the computed user set relation node
				g.AddEdge(from, &Node{
					Type:  "relation",
					ID:    fmt.Sprintf("%s#%s", entity.GetName(), leaf.GetComputedUserSet().GetRelation()),
					Label: leaf.GetComputedUserSet().GetRelation(),
				}, child.GetExclusion())
			default:
				break
			}
		}
	}
	return
}

// GetTupleSetReferenceReference iterates through the relation references
// and returns the first reference that doesn't contain a "#" symbol.
// If no such reference is found, it returns the tuple.USER constant.
func GetTupleSetReferenceReference(definition *base.RelationDefinition) string {
	for _, ref := range definition.GetRelationReferences() {
		if !strings.Contains(ref.String(), "#") {
			return ref.GetType()
		}
	}
	return tuple.USER
}
