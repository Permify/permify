package schema

import (
	"fmt"
	"strings"

	"github.com/rs/xid"

	"github.com/Permify/permify/pkg/errors"
	"github.com/Permify/permify/pkg/graph"
)

// OPType -
type OPType string

const (
	Union        OPType = "union"
	Intersection OPType = "intersection"
)

func (o OPType) String() string {
	return string(o)
}

// LeafType -
type LeafType string

const (
	ComputedUserSetType LeafType = "computed_user_set"
	TupleToUserSetType  LeafType = "tuple_to_user_set"
)

func (o LeafType) String() string {
	return string(o)
}

type ChildKind string

const (
	LeafKind    ChildKind = "leaf"
	RewriteKind ChildKind = "rewrite"
)

func (o ChildKind) String() string {
	return string(o)
}

type Schema struct {
	Entities map[string]Entity `json:"entities"`
}

// GetEntityByName -
func (s Schema) GetEntityByName(name string) (entity Entity, err error) {
	if en, ok := s.Entities[name]; ok {
		return en, nil
	}
	return entity, errors.NewError(errors.Service).SetMessage("schema not found")
}

// NewSchema -
func NewSchema(entities ...Entity) (schema Schema) {
	schema = Schema{
		Entities: map[string]Entity{},
	}

	for _, entity := range entities {

		if entity.Relations == nil {
			entity.Relations = []Relation{}
		}

		if entity.Actions == nil {
			entity.Actions = []Action{}
		}

		schema.Entities[entity.Name] = entity
	}

	return
}

// Entity -
type Entity struct {
	Name      string                 `json:"name"`
	Relations []Relation             `json:"relations"`
	Actions   []Action               `json:"actions"`
	Option    map[string]interface{} `json:"option"`
}

// GetAction -
func (e Entity) GetAction(name string) (action Action, err errors.Error) {
	for _, en := range e.Actions {
		if en.Name == name {
			return en, nil
		}
	}
	return action, errors.NewError(errors.Validation).AddParam("action", "action con not found")
}

// GetRelation -
func (e Entity) GetRelation(name string) (relation Relation, err errors.Error) {
	for _, re := range e.Relations {
		if re.Name == name {
			return re, nil
		}
	}
	return relation, errors.NewError(errors.Validation).AddParam("relation", "relation con not found")
}

// GetTable -
func (e Entity) GetTable() string {
	if en, ok := e.Option["table"]; ok {
		return en.(string)
	}
	return e.Name
}

// GetIdentifier -
func (e Entity) GetIdentifier() string {
	if en, ok := e.Option["identifier"]; ok {
		return en.(string)
	}
	return "id"
}

// Relation -
type Relation struct {
	Name   string                 `json:"name"`
	Types  []string               `json:"type"`
	Option map[string]interface{} `json:"option"`
}

// Type -
func (r Relation) Type() string {
	for _, typ := range r.Types {
		if !strings.Contains(typ, "#") {
			return typ
		}
	}
	return ""
}

// GetColumn -
func (r Relation) GetColumn() (string, bool) {
	if col, ok := r.Option["column"]; ok {
		return col.(string), true
	}
	return "", false
}

// Action -
type Action struct {
	Name  string `json:"name"`
	Child Exp    `json:"child"`
}

// Child -
type Child Exp

type Exp interface {
	GetType() string
	GetKind() string
}

// Rewrite -
type Rewrite struct {
	Type     OPType  `json:"type"` // union or intersection
	Children []Child `json:"children"`
}

// GetType -
func (r Rewrite) GetType() string {
	return r.Type.String()
}

// GetKind -
func (Rewrite) GetKind() string {
	return "rewrite"
}

// Leaf -
type Leaf struct {
	Exclusion bool     `json:"exclusion"`
	Type      LeafType `json:"type"` // tupleToUserSet or computedUserSet
	Value     string   `json:"value"`
}

// GetType -
func (l Leaf) GetType() string {
	return l.Type.String()
}

// GetKind -
func (Leaf) GetKind() string {
	return "leaf"
}

// COLLECTIONS

type Relations []Relation

// GetRelationByName -
func (r Relations) GetRelationByName(name string) (relation Relation, err error) {
	for _, rel := range r {
		if rel.Name == name {
			return rel, nil
		}
	}
	return relation, errors.NewError(errors.Service).SetMessage("relation not found")
}

// ToGraph -
func (s Schema) ToGraph() (g graph.Graph, error errors.Error) {
	for _, en := range s.Entities {
		eg, err := en.ToGraph()
		if err != nil {
			return graph.Graph{}, err
		}
		g.AddNodes(eg.Nodes())
		g.AddEdges(eg.Edges())
	}
	return
}

// ToGraph -
func (e Entity) ToGraph() (g graph.Graph, error errors.Error) {
	enNode := &graph.Node{
		Type:  "entity",
		ID:    fmt.Sprintf("entity:%s", e.Name),
		Label: e.Name,
	}
	g.AddNode(enNode)

	for _, re := range e.Relations {
		reNode := &graph.Node{
			Type:  "relation",
			ID:    fmt.Sprintf("entity:%s:relation:%s", e.Name, re.Name),
			Label: re.Name,
		}
		g.AddNode(reNode)
		g.AddEdge(enNode, reNode, nil)
	}

	for _, ac := range e.Actions {
		acNode := &graph.Node{
			Type:  "action",
			ID:    fmt.Sprintf("entity:%s:action:%s", e.Name, ac.Name),
			Label: ac.Name,
		}
		g.AddNode(acNode)
		g.AddEdge(enNode, acNode, nil)
		ag, err := e.buildActionGraph(acNode, []Child{ac.Child})
		if err != nil {
			return graph.Graph{}, err
		}
		g.AddNodes(ag.Nodes())
		g.AddEdges(ag.Edges())
	}
	return
}

// buildActionGraph -
func (e Entity) buildActionGraph(from *graph.Node, children []Child) (g graph.Graph, error errors.Error) {
	for _, child := range children {
		switch child.GetKind() {
		case RewriteKind.String():
			rw := &graph.Node{
				Type:  "logic",
				ID:    xid.New().String(),
				Label: child.GetType(),
			}
			g.AddNode(rw)
			g.AddEdge(from, rw, nil)
			ag, err := e.buildActionGraph(rw, child.(Rewrite).Children)
			if err != nil {
				return graph.Graph{}, err
			}
			g.AddNodes(ag.Nodes())
			g.AddEdges(ag.Edges())
		case LeafKind.String():
			ch := child.(Leaf)
			if ch.Type == ComputedUserSetType {
				g.AddEdge(from, &graph.Node{
					Type:  "relation",
					ID:    fmt.Sprintf("entity:%s:relation:%s", e.Name, ch.Value),
					Label: ch.Value,
				}, ch.Exclusion)
			} else {
				v := strings.Split(ch.Value, ".")
				re, err := e.GetRelation(v[0])
				if err != nil {
					return graph.Graph{}, errors.NewError(errors.Service).SetMessage("relation not found")
				}
				g.AddEdge(from, &graph.Node{
					Type:  "relation",
					ID:    fmt.Sprintf("entity:%s:relation:%s", re.Types[0], v[1]),
					Label: v[1],
				}, ch.Exclusion)
			}
		}
	}
	return
}
