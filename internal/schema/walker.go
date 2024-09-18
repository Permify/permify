package schema

import (
	"errors"

	"github.com/Permify/permify/pkg/dsl/utils"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// Walker is a struct used for traversing a schema
type Walker struct {
	schema *base.SchemaDefinition

	// map used to track visited nodes and avoid infinite recursion
	visited map[string]struct{}
}

// NewWalker is a constructor for the Walker struct
func NewWalker(schema *base.SchemaDefinition) *Walker {
	return &Walker{
		schema:  schema,
		visited: make(map[string]struct{}),
	}
}

// Walk traverses the schema based on entity type and permission
func (w *Walker) Walk(
	entityType string,
	permission string,
) error {
	// Generate a unique key for the entityType and permission combination
	key := utils.Key(entityType, permission)

	// Check if the entity-permission combination has already been visited
	if _, ok := w.visited[key]; ok {
		// If already visited, exit early to avoid redundant processing or infinite recursion
		return nil
	}

	// Mark the entity-permission combination as visited
	w.visited[key] = struct{}{}

	// Lookup the entity definition in the schema
	def, ok := w.schema.GetEntityDefinitions()[entityType]
	if !ok {
		// Error is returned if entity definition is not found
		return errors.New(base.ErrorCode_ERROR_CODE_ENTITY_DEFINITION_NOT_FOUND.String())
	}

	// Switch on the type of reference specified by the permission
	switch def.GetReferences()[permission] {
	case base.EntityDefinition_REFERENCE_PERMISSION:
		// If the reference type is a permission, look up the permission
		permission, ok := def.GetPermissions()[permission]
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
		return errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_CHILD_KIND.String())
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
			if err := w.WalkRewrite(entityType, child.GetRewrite()); err != nil {
				return err
			}
		case *base.Child_Leaf:
			if err := w.WalkLeaf(entityType, child.GetLeaf()); err != nil {
				return err
			}
		default:
			// For any other child type, return an error indicating an undefined child type
			return errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_CHILD_KIND.String())
		}
	}
	// If no errors occurred during the loop, return nil
	return nil
}

// WalkComputedUserSet walk the relation within the ComputedUserSet for the given entityType.
func (w *Walker) WalkComputedUserSet(
	entityType string,
	cu *base.ComputedUserSet,
) error {
	return w.Walk(entityType, cu.GetRelation())
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
		computedUserSet := t.TupleToUserSet.GetComputed()

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
			return w.WalkComputedUserSet(rel.GetType(), computedUserSet)
		}

		// If no errors occur, return nil
		return nil
	case *base.Leaf_ComputedUserSet:
		// Handle case where the leaf is a computed user set
		// Walk the entity type and relation
		return w.WalkComputedUserSet(entityType, t.ComputedUserSet)
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
