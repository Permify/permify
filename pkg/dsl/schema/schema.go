package schema

import (
	"errors"
	"fmt"
	"strings"

	"github.com/rs/xid"

	"github.com/Permify/permify/pkg/graph"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// NewSchema -
func NewSchema(entities ...*base.EntityDefinition) (schema *base.Schema) {
	schema = &base.Schema{
		EntityDefinitions:   map[string]*base.EntityDefinition{},
		RelationDefinitions: map[string]*base.RelationDefinition{},
		ActionDefinitions:   map[string]*base.ActionDefinition{},
	}
	for _, entity := range entities {
		if entity.Relations == nil {
			entity.Relations = map[string]*base.RelationDefinition{}
		}
		if entity.Actions == nil {
			entity.Actions = map[string]*base.ActionDefinition{}
		}
		schema.EntityDefinitions[entity.Name] = entity
		for _, relation := range entity.GetRelations() {
			schema.RelationDefinitions[fmt.Sprintf("%v#%v", entity.GetName(), relation.GetName())] = relation
		}
		for _, action := range entity.GetActions() {
			schema.ActionDefinitions[fmt.Sprintf("%v#%v", entity.GetName(), action.GetName())] = action
		}
	}
	return
}

// GetEntityByName -
func GetEntityByName(schema *base.Schema, name string) (entityDefinition *base.EntityDefinition, err error) {
	if en, ok := schema.GetEntityDefinitions()[name]; ok {
		return en, nil
	}
	return nil, errors.New(base.ErrorCode_entity_definition_not_found.String())
}

// GetRelationWithKey Key -> entity_name#relation_name
func GetRelationWithKey(schema *base.Schema, key string) (relationDefinition *base.RelationDefinition, err error) {
	if en, ok := schema.GetRelationDefinitions()[key]; ok {
		return en, nil
	}
	return nil, errors.New(base.ErrorCode_relation_definition_not_found.String())
}

// GetActionWithKey Key -> entity_name#action_name
func GetActionWithKey(schema *base.Schema, key string) (actionDefinition *base.ActionDefinition, err error) {
	if en, ok := schema.GetActionDefinitions()[key]; ok {
		return en, nil
	}
	return nil, errors.New(base.ErrorCode_action_definition_not_found.String())
}

// GetActionByNameInEntityDefinition -
func GetActionByNameInEntityDefinition(entityDefinition *base.EntityDefinition, name string) (actionDefinition *base.ActionDefinition, err error) {
	if re, ok := entityDefinition.GetActions()[name]; ok {
		return re, nil
	}
	return nil, errors.New(base.ErrorCode_action_can_not_found.String())
}

// GetRelationByNameInEntityDefinition -
func GetRelationByNameInEntityDefinition(entityDefinition *base.EntityDefinition, name string) (relationDefinition *base.RelationDefinition, err error) {
	if re, ok := entityDefinition.GetRelations()[name]; ok {
		return re, nil
	}
	return nil, errors.New(base.ErrorCode_action_definition_not_found.String())
}

// GetEntityReference -
func GetEntityReference(definition *base.RelationDefinition) string {
	return definition.GetEntityReference().GetName()
}

// GetTable -
func GetTable(definition *base.EntityDefinition) string {
	if en, ok := definition.GetOption()["table"]; ok {
		return string(en.Value)
	}
	return definition.GetName()
}

// GetIdentifier -
func GetIdentifier(definition *base.EntityDefinition) string {
	if en, ok := definition.GetOption()["identifier"]; ok {
		return string(en.Value)
	}
	return "id"
}

// GetColumn -
func GetColumn(definition *base.RelationDefinition) (string, bool) {
	if col, ok := definition.GetOption()["column"]; ok {
		return string(col.Value), true
	}
	return "", false
}

// GraphSchema -
func GraphSchema(schema *base.Schema) (g graph.Graph, error error) {
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
func GraphEntity(entity *base.EntityDefinition) (g graph.Graph, error error) {
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
func buildActionGraph(entity *base.EntityDefinition, from *graph.Node, children []*base.Child) (g graph.Graph, error error) {
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
				re, err := GetRelationByNameInEntityDefinition(entity, v[0])
				if err != nil {
					return graph.Graph{}, errors.New(base.ErrorCode_relation_definition_not_found.String())
				}
				g.AddEdge(from, &graph.Node{
					Type:  "relation",
					ID:    fmt.Sprintf("entity:%s:relation:%s", GetEntityReference(re), v[1]),
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
