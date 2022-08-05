package schema

import (
	"github.com/Permify/permify/pkg/helper"
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
	Entities map[string]Entity `json:"entities"`

	// all tables
	Tables map[string]TableType

	// table name to entities
	TableToEntity map[string]Entity

	// pivot name to entities
	PivotToEntity map[string]Entity

	// pivot name to relation
	PivotToRelation map[string]Relation
}

// GetEntityByName -
func (s Schema) GetEntityByName(name string) Entity {
	return s.Entities[name]
}

// GetEntityByTableName -
func (s Schema) GetEntityByTableName(table string) Entity {
	return s.TableToEntity[table]
}

// GetTableType -
func (s Schema) GetTableType(table string) TableType {
	return s.Tables[table]
}

// NewSchema -
func NewSchema(entities ...Entity) (schema Schema) {
	schema = Schema{
		Tables:          map[string]TableType{},
		Entities:        map[string]Entity{},
		TableToEntity:   map[string]Entity{},
		PivotToEntity:   map[string]Entity{},
		PivotToRelation: map[string]Relation{},
	}

	for _, entity := range entities {
		schema.Tables[entity.EntityOption.Table] = Main
		schema.TableToEntity[entity.EntityOption.Table] = entity
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
	Name      string     `json:"name"`
	Relations []Relation `json:"relations"`
	Actions   []Action   `json:"actions"`

	// option
	EntityOption EntityOption `json:"entity_option"`
}

// EntityOption -
type EntityOption struct {
	Table      string `json:"table"`
	Identifier string `json:"identifier"`
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
	Name string `json:"name"`
	Type string `json:"type"`

	// option
	RelationOption RelationOption `json:"relation_option"`
}

// RelationOption -
type RelationOption struct {
	Table string       `json:"table"`
	Rel   RelationType `json:"rel"`
	Cols  []string     `json:"cols"`
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
