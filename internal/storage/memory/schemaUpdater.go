package memory

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Permify/permify/internal/servers"
	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/internal/storage/memory/constants"
	db "github.com/Permify/permify/pkg/database/memory"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/hashicorp/go-memdb"
)

// SchemaUpdater - Structure for Schema Updater
type SchemaUpdater struct {
	database *db.Memory
}

// NewSchemaUpdater - Creates a new SchemaUpdater
func NewSchemaUpdater(database *db.Memory) *SchemaUpdater {
	return &SchemaUpdater{
		database: database,
	}
}

// UpdateSchema - Update entity config in the database
func (u *SchemaUpdater) UpdateSchema(ctx context.Context, tenantID, version string, definitions map[string]map[string][]string) (schema []string, err error) {
	txn := u.database.DB.Txn(false)
	defer txn.Abort()
	var it memdb.ResultIterator
	it, err = txn.Get(constants.SchemaDefinitionsTable, "version", tenantID, version)
	if err != nil {
		return schema, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}

	var storedDefinitions []string
	for obj := it.Next(); obj != nil; obj = it.Next() {
		storedDefinitions = append(storedDefinitions, obj.(storage.SchemaDefinition).Serialized())
	}
		
	var allErrors []error

	for entity, updates := range definitions {
		entityIdx, err := getEntityIndex(storedDefinitions, entity)
		if err != nil {
			allErrors = append(allErrors, err)
		} else {
			for action, stmts := range updates {
				switch action {
				case "write":
					updatedEntity := writeStatementsToSchema(storedDefinitions[entityIdx], stmts)
					storedDefinitions[entityIdx] = updatedEntity
				case "delete":
					updatedEntity, errs := deleteStatementsFromSchema(storedDefinitions[entityIdx], stmts)
					allErrors = append(allErrors, errs...)
					storedDefinitions[entityIdx] = updatedEntity
				case "update":
					updatedEntity, errs := updateStatementsInSchema(storedDefinitions[entityIdx], stmts)
					allErrors = append(allErrors, errs...)
					storedDefinitions[entityIdx] = updatedEntity
				default:
					allErrors = append(allErrors, fmt.Errorf("action not allowed in partial write for entity %s", entity))
				}
			}
		}
	}

	if len(allErrors) > 0 {
		return nil, &servers.MultiError{Errors: allErrors}
	}
	return storedDefinitions, nil
}

func getEntityIndex(definitions []string, entity string) (int, error) {
	entityString := fmt.Sprintf("entity %s", entity)
	for idx, definition := range definitions {
		if strings.HasPrefix(definition, entityString) {
			return idx, nil
		}
	}

	return 1, fmt.Errorf("%s entity not found in schema", entity)
}

func writeStatementsToSchema(entity string, statements []string) string {
	for _, stmt := range statements {
    	// Find the index of the last closing curly brace
    	lastClosingBraceIndex := strings.LastIndex(entity, "}")

		// Construct the modified DSL string
		entity = entity[:lastClosingBraceIndex] + stmt + "\n" + entity[lastClosingBraceIndex:]
	}
	return entity
}

func deleteStatementsFromSchema(entity string, statements []string) (string, []error) {
	var allErrors []error
	for _, stmt := range statements {
		// Check if the stmt exists
		if !strings.Contains(entity, stmt) {
			allErrors = append(allErrors, fmt.Errorf("statement %s does not exist", stmt))
		}
		entity = strings.Replace(entity, stmt, "", -1)
	}

	return entity, allErrors
}

func updateStatementsInSchema(entity string, statements []string) (string, []error) {
	var allErrors []error
	lines := strings.Split(entity, "\n")
	for _, stmt := range statements {
		found := false
		parts := strings.Fields(stmt)
		if len(parts) < 3 {
			allErrors = append(allErrors, fmt.Errorf("Invalid update statement: %s", stmt)) 
		}	
		identifier, target, _ := parts[0], parts[1], parts[2:]
		for i, line := range lines {
			if strings.HasPrefix(strings.TrimSpace(line), identifier+" "+target) {
				// Replace the line with the updated value
				lines[i] = stmt
				found = true
				break
			}
		}
		if !found {
			allErrors = append(allErrors, fmt.Errorf("Invalid update statement, relation does not exist"))
		}
	}
	entity = strings.Join(lines, "\n")
	return entity, allErrors
}