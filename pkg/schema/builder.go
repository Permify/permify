package schema

import (
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// Schema - Schema builder
func Schema(entities ...*base.EntityDefinition) *base.SchemaDefinition {
	def := &base.SchemaDefinition{
		EntityDefinitions: map[string]*base.EntityDefinition{},
	}
	for _, entity := range entities {
		def.EntityDefinitions[entity.Name] = entity
	}
	return def
}

// Entity - Entity builder
func Entity(name string, relations []*base.RelationDefinition, actions []*base.ActionDefinition) *base.EntityDefinition {
	def := &base.EntityDefinition{
		Name:       name,
		Relations:  map[string]*base.RelationDefinition{},
		Actions:    map[string]*base.ActionDefinition{},
		References: map[string]base.EntityDefinition_RelationalReference{},
	}
	for _, relation := range relations {
		def.Relations[relation.Name] = relation
		def.References[relation.Name] = base.EntityDefinition_RELATIONAL_REFERENCE_RELATION
	}
	for _, action := range actions {
		def.Actions[action.Name] = action
		def.References[action.Name] = base.EntityDefinition_RELATIONAL_REFERENCE_ACTION
	}
	return def
}

// Relation - Relation builder
func Relation(name string, references ...*base.RelationReference) *base.RelationDefinition {
	return &base.RelationDefinition{
		Name:               name,
		RelationReferences: references,
	}
}

// Relations  - Relations builder
func Relations(defs ...*base.RelationDefinition) []*base.RelationDefinition {
	return defs
}

// Reference - Reference builder
func Reference(name string) *base.RelationReference {
	return &base.RelationReference{
		Name: name,
	}
}

// Action - Action builder
func Action(name string, child *base.Child) *base.ActionDefinition {
	return &base.ActionDefinition{
		Name:  name,
		Child: child,
	}
}

// Actions - Actions builder
func Actions(defs ...*base.ActionDefinition) []*base.ActionDefinition {
	return defs
}

// ComputedUserSet -
func ComputedUserSet(relation string, exclusion bool) *base.Child {
	return &base.Child{
		Type: &base.Child_Leaf{
			Leaf: &base.Leaf{
				Exclusion: exclusion,
				Type: &base.Leaf_ComputedUserSet{
					ComputedUserSet: &base.ComputedUserSet{
						Relation: relation,
					},
				},
			},
		},
	}
}

// TupleToUserSet -
func TupleToUserSet(reference, relation string, exclusion bool) *base.Child {
	return &base.Child{
		Type: &base.Child_Leaf{
			Leaf: &base.Leaf{
				Exclusion: exclusion,
				Type: &base.Leaf_TupleToUserSet{
					TupleToUserSet: &base.TupleToUserSet{
						TupleSet: &base.TupleSet{
							Relation: reference,
						},
						Computed: &base.ComputedUserSet{
							Relation: relation,
						},
					},
				},
			},
		},
	}
}

// Union -
func Union(children ...*base.Child) *base.Child {
	return &base.Child{
		Type: &base.Child_Rewrite{
			Rewrite: &base.Rewrite{
				RewriteOperation: base.Rewrite_OPERATION_UNION,
				Children:         children,
			},
		},
	}
}

// Intersection -
func Intersection(children ...*base.Child) *base.Child {
	return &base.Child{
		Type: &base.Child_Rewrite{
			Rewrite: &base.Rewrite{
				RewriteOperation: base.Rewrite_OPERATION_INTERSECTION,
				Children:         children,
			},
		},
	}
}
