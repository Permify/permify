package graph

import (
	"errors"
	"fmt"

	"github.com/rs/xid"

	"github.com/Permify/permify/internal/schema"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

type Builder struct {
	schema *base.SchemaDefinition
}

// NewBuilder creates a new Builder object.
func NewBuilder(schema *base.SchemaDefinition) Builder {
	return Builder{schema: schema}
}

// SchemaToGraph converts a schema definition into a graph representation.
func (b Builder) SchemaToGraph() (g Graph, err error) {
	for _, entity := range b.schema.GetEntityDefinitions() {
		eg, err := b.EntityToGraph(entity)
		if err != nil {
			return Graph{}, fmt.Errorf("failed to convert entity to graph: %w", err)
		}
		g.AddNodes(eg.Nodes())
		g.AddEdges(eg.Edges())
	}
	return
}

// EntityToGraph takes an entity definition and converts it into a graph
// representation, returning the created graph and an error if any occurs.
func (b Builder) EntityToGraph(entity *base.EntityDefinition) (g Graph, err error) {
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
			Label: re.GetName(),
		}

		// Iterate through the relation references
		for _, ref := range re.GetRelationReferences() {
			if ref.GetRelation() != "" {
				g.AddEdge(reNode, &Node{
					Type:  "relation",
					ID:    fmt.Sprintf("%s#%s", ref.GetType(), ref.GetRelation()),
					Label: re.GetName(),
				})
			} else {
				g.AddEdge(reNode, &Node{
					Type:  "entity",
					ID:    ref.GetType(),
					Label: re.GetName(),
				})
			}
		}

		// Add relation node and edge to the graph
		g.AddNode(reNode)
		g.AddEdge(enNode, reNode)
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
		g.AddEdge(enNode, acNode)
		// Build permission graph for each permission
		ag, err := b.buildPermissionGraph(entity, acNode, []*base.Child{permission.GetChild()})
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
func (b Builder) buildPermissionGraph(entity *base.EntityDefinition, from *Node, children []*base.Child) (g Graph, err error) {
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
			g.AddEdge(from, rw)
			// Recursively process the children of the rewrite node
			ag, err := b.buildPermissionGraph(entity, rw, child.GetRewrite().GetChildren())
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

				for _, r := range re.GetRelationReferences() {
					ag, err := b.addEdgeFromRelation(from, r, leaf)
					if err != nil {
						return Graph{}, err
					}
					g.AddNodes(ag.Nodes())
					g.AddEdges(ag.Edges())
				}
			case *base.Leaf_ComputedUserSet:
				// Add an edge between the parent node and the computed user set relation node
				g.AddEdge(from, &Node{
					Type:  "relation",
					ID:    fmt.Sprintf("%s#%s", entity.GetName(), leaf.GetComputedUserSet().GetRelation()),
					Label: leaf.GetComputedUserSet().GetRelation(),
				})
			default:
				break
			}
		}
	}
	return
}

// AddEdgeFromRelation adds an edge to the graph from the relation information
func (b Builder) addEdgeFromRelation(from *Node, reference *base.RelationReference, leaf *base.Leaf) (g Graph, err error) {
	if reference.GetRelation() != "" {
		upperen, err := schema.GetEntityByName(b.schema, reference.GetType())
		if err != nil {
			return Graph{}, err
		}
		re, err := schema.GetRelationByNameInEntityDefinition(upperen, leaf.GetTupleToUserSet().GetTupleSet().GetRelation())
		if err != nil {
			return Graph{}, errors.New(base.ErrorCode_ERROR_CODE_RELATION_DEFINITION_NOT_FOUND.String())
		}
		for _, r := range re.GetRelationReferences() {
			ag, err := b.addEdgeFromRelation(from, r, leaf)
			if err != nil {
				return Graph{}, err
			}
			g.AddNodes(ag.Nodes())
			g.AddEdges(ag.Edges())
		}
	} else {
		// Add an edge between the parent node and the tuple set relation node
		g.AddEdge(from, &Node{
			Type:  "relation",
			ID:    fmt.Sprintf("%s#%s", reference.GetType(), leaf.GetTupleToUserSet().GetComputed().GetRelation()),
			Label: leaf.GetTupleToUserSet().GetComputed().GetRelation(),
		})
	}
	return
}
