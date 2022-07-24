package schema

import (
	"github.com/Permify/permify/pkg/helper"
)

// OPType -
type OPType string

const (
	Union        OPType = "union"
	Intersection OPType = "intersection"
	// Exclusion    OPType = "exclusion"
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

// RelationType -
type RelationType string

const (
	BelongsTo  RelationType = "belongs-to"
	ManyToMany RelationType = "many-to-many"
	Custom     RelationType = "custom"
)

// TableType -
type TableType string

const (
	Main  TableType = "main"
	Pivot TableType = "pivot"
)

type Schema struct {
	// all entities
	Entities map[string]Entity

	// all tables
	Tables map[string]TableType

	// pivot name to entities
	PivotToEntity map[string]Entity

	// pivot name to relation
	PivotToRelation map[string]Relation
}

// GetEntity -
func (s Schema) GetEntity(name string) Entity {
	return s.Entities[name]
}

// NewSchema -
func NewSchema(entities ...Entity) (schema Schema) {
	schema.Tables = map[string]TableType{}
	schema.Entities = map[string]Entity{}
	schema.PivotToEntity = map[string]Entity{}
	schema.PivotToRelation = map[string]Relation{}

	for _, entity := range entities {
		schema.Tables[entity.EntityOption.Table] = Main
		for _, relation := range entity.Relations {
			if relation.RelationOption.Rel == ManyToMany {
				schema.Tables[relation.RelationOption.Table] = Pivot
				schema.PivotToEntity[relation.RelationOption.Table] = entity
				schema.PivotToRelation[relation.RelationOption.Table] = relation
			}
		}
		schema.Entities[entity.Name] = entity
	}

	return
}

// Entity -
type Entity struct {
	Name      string
	Relations []Relation
	Actions   []Action

	// option
	EntityOption EntityOption
}

// EntityOption -
type EntityOption struct {
	Table      string
	Identifier string
}

// GetAction -
func (e Entity) GetAction(name string) Action {
	for _, en := range e.Actions {
		if en.Name == name {
			return en
		}
	}
	return Action{}
}

// Relation -
type Relation struct {
	Name string
	Type string

	// option
	RelationOption RelationOption
}

// RelationOption -
type RelationOption struct {
	Table string
	Rel   RelationType
	Cols  []string
}

// Action -
type Action struct {
	Name  string
	Child Exp
}

// Child -
type Child Exp

type Exp interface {
	GetType() string
	GetKind() string
}

// Rewrite -
type Rewrite struct {
	Type     OPType // union or intersection
	Children []Child
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
	Type  LeafType // tupleToUserSet or computedUserSet
	Value string
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

// Filter -
func (r Relations) Filter(relationTypes ...RelationType) (relations Relations) {
	for _, relation := range r {
		if len(relationTypes) > 0 {
			if helper.InArray(relation.RelationOption.Rel, relationTypes) {
				relations = append(relations, relation)
			}
		} else {
			relations = append(relations, relation)
		}
	}
	return
}

// GetRelationByName -
func (r Relations) GetRelationByName(name string) (relation Relation) {
	for _, rel := range r {
		if rel.Name == name {
			return rel
		}
	}
	return
}
