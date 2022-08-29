package schema

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
	return entity, EntityCanNotFoundErr
}

// NewSchema -
func NewSchema(entities ...Entity) (schema Schema) {
	schema = Schema{
		Entities: map[string]Entity{},
	}

	for _, entity := range entities {
		schema.Entities[entity.Name] = entity
	}

	return
}

// Entity -
type Entity struct {
	Name      string     `json:"name"`
	Relations []Relation `json:"relations"`
	Actions   []Action   `json:"actions"`
}

// GetAction -
func (e Entity) GetAction(name string) (action Action, err error) {
	for _, en := range e.Actions {
		if en.Name == name {
			return en, nil
		}
	}
	return action, ActionCanNotFoundErr
}

// GetRelation -
func (e Entity) GetRelation(name string) (relation Relation, err error) {
	for _, re := range e.Relations {
		if re.Name == name {
			return re, nil
		}
	}
	return relation, RelationCanNotFoundErr
}

// Relation -
type Relation struct {
	Name  string   `json:"name"`
	Types []string `json:"type"`
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
	return relation, RelationCanNotFoundErr
}
