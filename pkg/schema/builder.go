package schema

import (
	"strings"

	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// Schema is a function that returns a pointer to a SchemaDefinition structure. It takes a variable number of pointers to
// EntityDefinition structures as input arguments, which are added to the schema being constructed.
// The function creates a new SchemaDefinition structure, initializes its EntityDefinitions field to an empty map, and
// then adds each input EntityDefinition structure to this map with the entity name as the key.
// Finally, it returns the pointer to the constructed SchemaDefinition structure.
func Schema(entities []*base.EntityDefinition, rules []*base.RuleDefinition) *base.SchemaDefinition {
	// create a new SchemaDefinition structure
	def := &base.SchemaDefinition{
		EntityDefinitions: map[string]*base.EntityDefinition{},
		RuleDefinitions:   map[string]*base.RuleDefinition{},
		References:        map[string]base.SchemaDefinition_Reference{},
	}

	for _, entity := range entities {
		def.EntityDefinitions[entity.Name] = entity
		def.References[entity.Name] = base.SchemaDefinition_REFERENCE_ENTITY
	}

	for _, rule := range rules {
		def.RuleDefinitions[rule.Name] = rule
		def.References[rule.Name] = base.SchemaDefinition_REFERENCE_RULE
	}

	// return the pointer to the constructed SchemaDefinition structure
	return def
}

// Entity - Entity builder
// This function creates and returns a new instance of EntityDefinition.
// It takes in the name of the entity, an array of relations, and an array of actions.
// It then initializes the EntityDefinition with the provided values and returns it.
// The EntityDefinition contains information about the entity's name, its relations, actions, and references.
func Entity(name string, relations []*base.RelationDefinition, attributes []*base.AttributeDefinition, permissions []*base.PermissionDefinition) *base.EntityDefinition {
	def := &base.EntityDefinition{
		Name:        name,
		Relations:   map[string]*base.RelationDefinition{},
		Attributes:  map[string]*base.AttributeDefinition{},
		Permissions: map[string]*base.PermissionDefinition{},
		References:  map[string]base.EntityDefinition_Reference{},
	}

	for _, relation := range relations {
		def.Relations[relation.Name] = relation
		def.References[relation.Name] = base.EntityDefinition_REFERENCE_RELATION
	}

	for _, attribute := range attributes {
		def.Attributes[attribute.Name] = attribute
		def.References[attribute.Name] = base.EntityDefinition_REFERENCE_ATTRIBUTE
	}

	for _, permission := range permissions {
		def.Permissions[permission.Name] = permission
		def.References[permission.Name] = base.EntityDefinition_REFERENCE_PERMISSION
	}

	return def
}

//func Rule(name string, relations []*base.RelationDefinition, attributes []*base.AttributeDefinition, permissions []*base.PermissionDefinition) *base.RuleDefinition {
//	def := &base.RuleDefinition{
//		Name:        name,
//		Relations:   map[string]*base.RelationDefinition{},
//		Attributes:  map[string]*base.AttributeDefinition{},
//		Permissions: map[string]*base.PermissionDefinition{},
//		References:  map[string]base.EntityDefinition_Reference{},
//	}
//
//	for _, relation := range relations {
//		def.Relations[relation.Name] = relation
//		def.References[relation.Name] = base.EntityDefinition_REFERENCE_RELATION
//	}
//
//	for _, attribute := range attributes {
//		def.Attributes[attribute.Name] = attribute
//		def.References[attribute.Name] = base.EntityDefinition_REFERENCE_ATTRIBUTE
//	}
//
//	for _, permission := range permissions {
//		def.Permissions[permission.Name] = permission
//		def.References[permission.Name] = base.EntityDefinition_REFERENCE_PERMISSION
//	}
//
//	return def
//}

// Relation - Relation builder function that creates a new RelationDefinition instance
// with the given name and references.
//
// Parameters:
// - name: a string representing the name of the relation.
// - references: a variadic parameter representing the relation references associated
// with this relation.
//
// Returns:
// - a pointer to a new RelationDefinition instance with the given name and references.
func Relation(name string, references ...*base.RelationReference) *base.RelationDefinition {
	return &base.RelationDefinition{
		Name:               name,
		RelationReferences: references,
	}
}

func Attribute(name string, typ base.AttributeType) *base.AttributeDefinition {
	return &base.AttributeDefinition{
		Name: name,
		Type: typ,
	}
}

// Attributes - Attributes builder
func Attributes(defs ...*base.AttributeDefinition) []*base.AttributeDefinition {
	return defs
}

// Relations - Relations builder
func Relations(defs ...*base.RelationDefinition) []*base.RelationDefinition {
	return defs
}

// Reference - Reference builder
func Reference(name string) *base.RelationReference {
	// Split the name parameter into a type and relation if it contains a "#"
	s := strings.Split(name, "#")
	if len(s) == 1 {
		// If no relation is specified in the name, return a RelationReference with an empty relation field
		return &base.RelationReference{
			Type:     s[0],
			Relation: "",
		}
	}
	// If a relation is specified in the name, return a RelationReference with both type and relation fields populated
	return &base.RelationReference{
		Type:     s[0],
		Relation: s[1],
	}
}

// Permission - Permission builder creates a new action definition with the given name and child entity
func Permission(name string, child *base.Child) *base.PermissionDefinition {
	return &base.PermissionDefinition{
		Name:  name,
		Child: child,
	}
}

// Permissions - Permissions builder creates a slice of action definitions from the given variadic list of action definitions
func Permissions(defs ...*base.PermissionDefinition) []*base.PermissionDefinition {
	return defs
}

// ComputedUserSet - returns a Child definition that represents a computed set of users based on a relation
// relation: the name of the relation on which the computed set is based
// exclusion: a boolean indicating if the computed set should exclude or include the users in the set
func ComputedUserSet(relation string) *base.Child {
	return &base.Child{
		Type: &base.Child_Leaf{
			Leaf: &base.Leaf{
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
// Returns a pointer to a base.Child struct, containing a Leaf struct with a TupleToUserSet struct,
// that represents a child computation where the tuple set is passed to the computed user set.
// Takes a reference string, relation string, and exclusion boolean as arguments.
// reference: the name of the reference to the tuple set
// relation: the name of the relation for the computed user set
// exclusion: a boolean indicating whether to exclude the computed user set
// Returns a pointer to a base.Child struct.
func TupleToUserSet(reference, relation string) *base.Child {
	return &base.Child{
		Type: &base.Child_Leaf{
			Leaf: &base.Leaf{
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

// Union takes a variable number of Child arguments and returns a new Child representing the union of all the sets obtained by evaluating each child.
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

// Intersection - Returns a child element that represents the intersection of the given children. This child element can be used in defining entity relations and actions.
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

// Exclusion - Returns a child element that represents the exclusion of the given children. This child element can be used in defining entity relations and actions.
func Exclusion(children ...*base.Child) *base.Child {
	return &base.Child{
		Type: &base.Child_Rewrite{
			Rewrite: &base.Rewrite{
				RewriteOperation: base.Rewrite_OPERATION_EXCLUSION,
				Children:         children,
			},
		},
	}
}
