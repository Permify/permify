package ast

import (
	"errors"
	"fmt"
	"strings"

	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// Schema represents the parsed schema, which contains all the statements
// and extracted entity and relational references used by the schema. It
// is used as an intermediate representation before generating the
// corresponding metadata.
type Schema struct {
	// The list of statements in the schema
	Statements []Statement

	// references - Map of all relational references extracted from the schema
	references *References
}

func NewSchema() *Schema {
	return &Schema{
		Statements: []Statement{},
		references: NewReferences(),
	}
}

// String -
func (sch *Schema) String() string {
	stmts := make([]string, 0, len(sch.Statements))
	for _, stmt := range sch.Statements {
		stmts = append(stmts, stmt.String())
	}
	return strings.Join(stmts, "\n")
}

// GetReferences -
func (sch *Schema) GetReferences() *References {
	return sch.references
}

// SetReferences -
func (sch *Schema) SetReferences(refs *References) {
	sch.references = refs
}

// AddStatement adds a new statement to an entity within the schema. It also updates the schema's references.
func (sch *Schema) AddStatement(entityName string, stmt Statement) error {
	// Loop through all statements in the schema to find the one matching the entity name.
	var entityStmt *EntityStatement
	for _, statement := range sch.Statements {
		if statement.GetName() == entityName {
			// Attempt to cast the found statement to an EntityStatement.
			es, ok := statement.(*EntityStatement)
			if ok {
				entityStmt = es // Successfully found and cast the EntityStatement.
				break
			} else {
				return errors.New(base.ErrorCode_ERROR_CODE_CANNOT_CONVERT_TO_ENTITY_STATEMENT.String()) // Casting failed.
			}
		}
	}

	// If no matching entity was found, return an error.
	if entityStmt == nil {
		return errors.New(base.ErrorCode_ERROR_CODE_ENTITY_STATEMENT_NOT_FOUND.String())
	}

	// Construct a unique key for the new statement.
	refKey := fmt.Sprintf("%s#%s", entityName, stmt.GetName())

	// Check if a reference for this statement already exists in the schema.
	if sch.GetReferences().IsReferenceExist(refKey) {
		return errors.New(base.ErrorCode_ERROR_CODE_ALREADY_EXIST.String()) // Avoid duplicating references.
	}

	// Add the new statement to the appropriate list in the EntityStatement, based on its type.
	switch stmt.StatementType() {
	case PERMISSION_STATEMENT:
		// Append the new permission statement to the list of permission statements for the entity.
		entityStmt.PermissionStatements = append(entityStmt.PermissionStatements, stmt)
		// Add a new permission reference to the schema's references.
		return sch.GetReferences().AddPermissionReference(refKey)

	case RELATION_STATEMENT:
		// Append the new relation statement to the list of relation statements for the entity.
		entityStmt.RelationStatements = append(entityStmt.RelationStatements, stmt)
		// Check if the statement can be cast to a RelationStatement and add its references.
		if rs, ok := stmt.(*RelationStatement); ok {
			return sch.GetReferences().AddRelationReferences(refKey, rs.RelationTypes)
		} else {
			return errors.New(base.ErrorCode_ERROR_CODE_CANNOT_CONVERT_TO_RELATION_STATEMENT.String())
		}

	case ATTRIBUTE_STATEMENT:
		// Append the new attribute statement to the list of attribute statements for the entity.
		entityStmt.AttributeStatements = append(entityStmt.AttributeStatements, stmt)
		// Check if the statement can be cast to an AttributeStatement and add its references.
		if as, ok := stmt.(*AttributeStatement); ok {
			return sch.GetReferences().AddAttributeReferences(refKey, as.AttributeType)
		} else {
			return errors.New(base.ErrorCode_ERROR_CODE_CANNOT_CONVERT_TO_ATTRIBUTE_STATEMENT.String())
		}

	default:
		// The statement type is unknown, return an error.
		return errors.New(base.ErrorCode_ERROR_CODE_UNKNOWN_STATEMENT_TYPE.String())
	}
}

// UpdateStatement updates a statement of a specific type (permission, relation, or attribute)
// for a given entity within the schema. It either updates an existing statement or appends a new one if not found.
func (sch *Schema) UpdateStatement(entityName string, newStmt Statement) error {
	// Iterate through all statements in the schema to find the one matching the entity name.
	for _, statement := range sch.Statements {
		if statement.GetName() == entityName {
			// Convert the generic statement interface to a specific EntityStatement type.
			entityStmt, ok := statement.(*EntityStatement)
			if !ok {
				return errors.New(base.ErrorCode_ERROR_CODE_CANNOT_CONVERT_TO_ENTITY_STATEMENT.String())
			}

			// Construct a unique reference key for the new statement.
			referenceKey := fmt.Sprintf("%s#%s", entityName, newStmt.GetName())

			// Check if a reference for the statement already exists within the schema.
			if !sch.GetReferences().IsReferenceExist(referenceKey) {
				return errors.New(base.ErrorCode_ERROR_CODE_REFERENCE_NOT_FOUND.String())
			}

			// Based on the statement type, update the corresponding list in the EntityStatement.
			switch newStmt.StatementType() {
			case PERMISSION_STATEMENT, RELATION_STATEMENT, ATTRIBUTE_STATEMENT:
				var stmts *[]Statement // Pointer to the slice of statements to update.

				// Assign the correct slice based on the type of the new statement.
				switch newStmt.StatementType() {
				case PERMISSION_STATEMENT:
					stmts = &entityStmt.PermissionStatements
				case RELATION_STATEMENT:
					stmts = &entityStmt.RelationStatements
				case ATTRIBUTE_STATEMENT:
					stmts = &entityStmt.AttributeStatements
				}

				// Flag to check if the statement has been updated.
				updated := false

				// Iterate over the statements to find and update the one with the matching name.
				for i, stmt := range *stmts {
					if stmt.GetName() == newStmt.GetName() {
						(*stmts)[i] = newStmt
						updated = true
						break // Stop iterating once the statement is updated.
					}
				}

				// If the statement was not found and updated, append it to the slice.
				if !updated {
					*stmts = append(*stmts, newStmt)
				}

				// Update the reference in the schema based on the statement type.
				switch newStmt.StatementType() {
				case PERMISSION_STATEMENT:
					return sch.GetReferences().UpdatePermissionReference(referenceKey)
				case RELATION_STATEMENT:
					if rs, ok := newStmt.(*RelationStatement); ok {
						return sch.GetReferences().UpdateRelationReferences(referenceKey, rs.RelationTypes)
					}
					return errors.New(base.ErrorCode_ERROR_CODE_CANNOT_CONVERT_TO_RELATION_STATEMENT.String())
				case ATTRIBUTE_STATEMENT:
					if as, ok := newStmt.(*AttributeStatement); ok {
						return sch.GetReferences().UpdateAttributeReferences(referenceKey, as.AttributeType)
					}
					return errors.New(base.ErrorCode_ERROR_CODE_CANNOT_CONVERT_TO_ATTRIBUTE_STATEMENT.String())
				}
			default:
				return errors.New(base.ErrorCode_ERROR_CODE_UNKNOWN_STATEMENT_TYPE.String())
			}
			// Return nil to indicate successful update.
			return nil
		}
	}

	// If no matching entity statement is found, return an error.
	return errors.New(base.ErrorCode_ERROR_CODE_ENTITY_STATEMENT_NOT_FOUND.String())
}

// DeleteStatement removes a specific statement from an entity within the schema.
// It identifies the statement by its name and the name of the entity it belongs to.
// If successful, it also removes the corresponding reference from the schema.
func (sch *Schema) DeleteStatement(entityName, name string) error {
	// Iterate over all statements in the schema to find the entity by name.
	for _, statement := range sch.Statements {
		if statement.GetName() == entityName {
			// Try to convert the found statement into an EntityStatement.
			entityStmt, ok := statement.(*EntityStatement)
			if !ok {
				// If conversion fails, return an error indicating wrong type.
				return errors.New(base.ErrorCode_ERROR_CODE_CANNOT_CONVERT_TO_ENTITY_STATEMENT.String())
			}

			// Construct the reference key from the entity and statement names.
			referenceKey := fmt.Sprintf("%s#%s", entityName, name)
			// Retrieve the type of reference associated with the key.
			refType, ok := sch.GetReferences().GetReferenceType(referenceKey)
			if !ok {
				// If no reference is found, return an error.
				return errors.New(base.ErrorCode_ERROR_CODE_REFERENCE_NOT_FOUND.String())
			}

			// Declare a variable to hold the target list of statements for modification.
			var targetStatements *[]Statement
			// Assign the targetStatements pointer based on the reference type.
			switch refType {
			case PERMISSION:
				targetStatements = &entityStmt.PermissionStatements
			case RELATION:
				targetStatements = &entityStmt.RelationStatements
			case ATTRIBUTE:
				targetStatements = &entityStmt.AttributeStatements
			default:
				// If the reference type is unknown, return an error.
				return errors.New(base.ErrorCode_ERROR_CODE_UNKNOWN_REFERENCE_TYPE.String())
			}

			// Create a new slice to hold all statements except the one to be deleted.
			var newStatements []Statement
			for _, stmt := range *targetStatements {
				if stmt.GetName() != name {
					// If the statement is not the one to delete, add it to the new slice.
					newStatements = append(newStatements, stmt)
				}
			}
			// Replace the old slice with the new slice, effectively deleting the statement.
			*targetStatements = newStatements

			// Remove the corresponding reference based on the reference type.
			switch refType {
			case PERMISSION:
				return sch.GetReferences().RemovePermissionReference(referenceKey)
			case RELATION:
				return sch.GetReferences().RemoveRelationReferences(referenceKey)
			case ATTRIBUTE:
				return sch.GetReferences().RemoveAttributeReferences(referenceKey)
			}

			// If the statement was successfully removed, return nil to indicate success.
			return nil
		}
	}

	// If no matching entity is found in the schema, return an error.
	return errors.New(base.ErrorCode_ERROR_CODE_ENTITY_STATEMENT_NOT_FOUND.String())
}
