package schema

import (
	"fmt"
	"strings"

	"github.com/rs/xid"

	"github.com/Permify/permify/pkg/errors"
	"github.com/Permify/permify/pkg/graph"
	base `github.com/Permify/permify/pkg/pb/base/v1`
)

// GetEntityByName -
func GetEntityByName(schema *base.Schema, name string) (entityDefinition *base.EntityDefinition, err errors.Error) {
	if en, ok := schema.GetEntityDefinitions()[name]; ok {
		return en, nil
	}
	return nil, errors.NewError(errors.Service).SetMessage("entity not found")
}

// NewSchema -
func NewSchema(entities ...*base.EntityDefinition) (schema *base.Schema) {
	schema = &base.Schema{
		EntityDefinitions: map[string]*base.EntityDefinition{},
	}
	for _, entity := range entities {
		if entity.Relations == nil {
			entity.Relations = []*base.RelationDefinition{}
		}
		if entity.Actions == nil {
			entity.Actions = []*base.ActionDefinition{}
		}
		schema.EntityDefinitions[entity.Name] = entity
	}
	return
}

// GetAction -
func GetAction(entityDefinition *base.EntityDefinition, name string) (actionDefinition *base.ActionDefinition, err errors.Error) {
	for _, en := range entityDefinition.Actions {
		if en.GetName() == name {
			return en, nil
		}
	}
	return nil, errors.NewError(errors.Validation).AddParam("action", "action con not found")
}

// GetRelation -
func GetRelation(entityDefinition *base.EntityDefinition, name string) (relationDefinition *base.RelationDefinition, err errors.Error) {
	for _, re := range entityDefinition.Relations {
		if re.GetName() == name {
			return re, nil
		}
	}
	return nil, errors.NewError(errors.Validation).AddParam("relation", "relation con not found")
}

// GetTable -
func GetTable(definition *base.EntityDefinition) string {
	if en, ok := definition.GetOption()["table"]; ok {
		return en.String()
	}
	return definition.GetName()
}

// GetIdentifier -
func GetIdentifier(definition *base.EntityDefinition) string {
	if en, ok := definition.GetOption()["identifier"]; ok {
		return en.String()
	}
	return "id"
}

// Type -
func Type(definition *base.RelationDefinition) string {
	for _, typ := range definition.Types {
		if !strings.Contains(typ.GetName(), "#") {
			return typ.GetName()
		}
	}
	return ""
}

// GetColumn -
func GetColumn(definition *base.RelationDefinition) (string, bool) {
	if col, ok := definition.GetOption()["column"]; ok {
		return col.String(), true
	}
	return "", false
}

// COLLECTIONS

type Relations []*base.RelationDefinition

// GetRelationByName -
func (r Relations) GetRelationByName(name string) (definition *base.RelationDefinition, err error) {
	for _, rel := range r {
		if rel.Name == name {
			return rel, nil
		}
	}
	return nil, errors.NewError(errors.Service).SetMessage("relation not found")
}

// GraphSchema -
func GraphSchema(schema *base.Schema) (g graph.Graph, error errors.Error) {
	for _, en := range schema.GetEntityDefinitions() {
		eg, err := GraphEntity(en)
		if err != nil {
			return graph.Graph{}, err
		}
		g.AddNodes(eg.Nodes())
		g.AddEdges(eg.Edges())
	}
	return
}

// GraphEntity -
func GraphEntity(entity *base.EntityDefinition) (g graph.Graph, error errors.Error) {
	enNode := &graph.Node{
		Type:  "entity",
		ID:    fmt.Sprintf("entity:%s", entity.GetName()),
		Label: entity.GetName(),
	}
	g.AddNode(enNode)

	for _, re := range entity.GetRelations() {
		reNode := &graph.Node{
			Type:  "relation",
			ID:    fmt.Sprintf("entity:%s:relation:%s", entity.GetName(), re.GetName()),
			Label: re.Name,
		}
		g.AddNode(reNode)
		g.AddEdge(enNode, reNode, nil)
	}

	for _, action := range entity.GetActions() {
		acNode := &graph.Node{
			Type:  "action",
			ID:    fmt.Sprintf("entity:%s:action:%s", entity.GetName(), action.GetName()),
			Label: action.GetName(),
		}
		g.AddNode(acNode)
		g.AddEdge(enNode, acNode, nil)
		ag, err := buildActionGraph(entity, acNode, []*base.Child{action.GetChild()})
		if err != nil {
			return graph.Graph{}, err
		}
		g.AddNodes(ag.Nodes())
		g.AddEdges(ag.Edges())
	}
	return
}

// buildActionGraph -
func buildActionGraph(entity *base.EntityDefinition, from *graph.Node, children []*base.Child) (g graph.Graph, error errors.Error) {
	for _, child := range children {
		switch child.GetType().(type) {
		case *base.Child_Rewrite:
			rw := &graph.Node{
				Type:  "logic",
				ID:    xid.New().String(),
				Label: child.String(),
			}
			g.AddNode(rw)
			g.AddEdge(from, rw, nil)
			ag, err := buildActionGraph(entity, rw, child.GetRewrite().GetChildren())
			if err != nil {
				return graph.Graph{}, err
			}
			g.AddNodes(ag.Nodes())
			g.AddEdges(ag.Edges())
		case *base.Child_Leaf:
			leaf := child.GetLeaf()
			switch leaf.GetType().(type) {
			case *base.Leaf_TupleToUserSet:
				v := strings.Split(leaf.GetTupleToUserSet().GetRelation(), ".")
				re, err := GetRelation(entity, v[0])
				if err != nil {
					return graph.Graph{}, errors.NewError(errors.Service).SetMessage("relation not found")
				}
				g.AddEdge(from, &graph.Node{
					Type:  "relation",
					ID:    fmt.Sprintf("entity:%s:relation:%s", Type(re), v[1]),
					Label: v[1],
				}, leaf.GetExclusion())
				break
			case *base.Leaf_ComputedUserSet:
				g.AddEdge(from, &graph.Node{
					Type:  "relation",
					ID:    fmt.Sprintf("entity:%s:relation:%s", entity.GetName(), leaf.GetComputedUserSet().GetRelation()),
					Label: leaf.GetComputedUserSet().GetRelation(),
				}, leaf.GetExclusion())
				break
			default:
				break
			}
		}
	}
	return
}
