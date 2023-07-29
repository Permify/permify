package schema

import (
	"errors"

	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// Walker is a struct used for traversing a schema
type Walker struct {
	schema *base.SchemaDefinition
}

// NewWalker is a constructor for the Walker struct
func NewWalker(schema *base.SchemaDefinition) *Walker {
	return &Walker{
		schema: schema,
	}
}

// Walk traverses the schema based on entity type and permission
func (w *Walker) Walk(
	entityType string,
	permission string,
) error {
	// Lookup the entity definition in the schema
	def, ok := w.schema.EntityDefinitions[entityType]
	if !ok {
		// Error is returned if entity definition is not found
		return errors.New("entity definition not found")
	}

	// Switch on the type of reference specified by the permission
	switch def.References[permission] {
	case base.EntityDefinition_REFERENCE_PERMISSION:
		// If the reference type is a permission, look up the permission
		permission, ok := def.Permissions[permission]
		if !ok {
			// Error is returned if permission is not found
			return errors.New(base.ErrorCode_ERROR_CODE_PERMISSION_NOT_FOUND.String())
		}
		// Check if the permission has a child element
		child := permission.GetChild()
		// If the child has a rewrite rule, walk the rewrite rule
		if child.GetRewrite() != nil {
			return w.WalkRewrite(entityType, child.GetRewrite())
		}
		// Otherwise, walk the leaf node
		return w.WalkLeaf(entityType, child.GetLeaf())
	case base.EntityDefinition_REFERENCE_RELATION:
		// If the reference type is a relation, nothing to do, return nil
		return nil
	case base.EntityDefinition_REFERENCE_ATTRIBUTE:
		// If the reference type is an attribute, not implemented, return error
		return ErrUnimplemented
	default:
		// For any other reference type, not implemented, return error
		return ErrUnimplemented
	}
}

// WalkRewrite is a method that walks through the rewrite part of the schema
func (w *Walker) WalkRewrite(
	entityType string,
	rewrite *base.Rewrite,
) error {
	// Loop through each child in the rewrite
	for _, child := range rewrite.GetChildren() {
		// Switch on the type of the child
		switch child.GetType().(type) {
		case *base.Child_Rewrite:
			// If the child is a rewrite, recursively walk the rewrite
			return w.WalkRewrite(entityType, child.GetRewrite())
		case *base.Child_Leaf:
			// If the child is a leaf, walk the leaf
			return w.WalkLeaf(entityType, child.GetLeaf())
		default:
			// For any other child type, return an error indicating an undefined child type
			return errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_CHILD_KIND.String())
		}
	}
	// If no errors occurred during the loop, return nil
	return nil
}

// WalkLeaf is a method that walks through the leaf part of the schema
func (w *Walker) WalkLeaf(
	entityType string,
	leaf *base.Leaf,
) error {
	// Switch on the type of the leaf
	switch t := leaf.GetType().(type) {
	case *base.Leaf_TupleToUserSet:
		// Handle case where the leaf is a tuple to user set
		tupleSet := t.TupleToUserSet.GetTupleSet().GetRelation()
		computedUserSet := t.TupleToUserSet.GetComputed().GetRelation()

		// Look up the entity definition
		entityDefinitions, exists := w.schema.EntityDefinitions[entityType]
		if !exists {
			// Return error if entity definition is not found
			return errors.New(base.ErrorCode_ERROR_CODE_ENTITY_DEFINITION_NOT_FOUND.String())
		}

		// Look up the relations in the entity definition
		relations, exists := entityDefinitions.Relations[tupleSet]
		if !exists {
			// Return error if relations is not found
			return errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_RELATION_REFERENCE.String())
		}

		// Walk each relation reference
		for _, rel := range relations.GetRelationReferences() {
			return w.Walk(
				rel.GetType(),
				computedUserSet,
			)
		}

		// If no errors occur, return nil
		return nil
	case *base.Leaf_ComputedUserSet:
		// Handle case where the leaf is a computed user set
		// Walk the entity type and relation
		return w.Walk(
			entityType,
			t.ComputedUserSet.GetRelation(),
		)
	case *base.Leaf_ComputedAttribute:
		// Handle case where the leaf is a computed attribute
		// This is currently unimplemented, so return an error
		return ErrUnimplemented
	case *base.Leaf_Call:
		// Handle case where the leaf is a call
		// This is currently unimplemented, so return an error
		return ErrUnimplemented
	default:
		// Handle any other type of leaf
		// Return an error indicating the leaf type is undefined
		return ErrUndefinedLeafType
	}
}
