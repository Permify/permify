package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/Permify/permify/internal/servers"
	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/internal/storage/postgres/utils"
	db "github.com/Permify/permify/pkg/database/postgres"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// SchemaUpdater - Structure for SchemaUpdater
type SchemaUpdater struct {
	database *db.Postgres
	// options
	txOptions sql.TxOptions
}

// NewSchemaUpdater - Creates a new SchemaUpdater
func NewSchemaUpdater(database *db.Postgres) *SchemaUpdater {
	return &SchemaUpdater{
		database:  database,
		txOptions: sql.TxOptions{Isolation: sql.LevelReadCommitted, ReadOnly: true},
	}
}

// UpdateSchema - Update entity config in the database
func (u *SchemaUpdater) UpdateSchema(ctx context.Context, tenantID, version string, definitions map[string]map[string][]string) (schema []string, err error) {
	ctx, span := tracer.Start(ctx, "schema-reader.read-schema")
	defer span.End()

	slog.Debug("reading schema", slog.Any("tenant_id", tenantID), slog.Any("version", version))

	builder := u.database.Builder.Select("name, serialized_definition, version").From(SchemaDefinitionTable).Where(squirrel.Eq{"version": version, "tenant_id": tenantID})

	var query string
	var args []interface{}

	query, args, err = builder.ToSql()
	if err != nil {
		return nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_SQL_BUILDER)
	}

	slog.Debug("executing sql query", slog.Any("query", query), slog.Any("arguments", args))

	var rows *sql.Rows
	rows, err = u.database.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_EXECUTION)
	}
	defer rows.Close()

	var storedDefinitions []string
	for rows.Next() {
		sd := storage.SchemaDefinition{}
		err = rows.Scan(&sd.Name, &sd.SerializedDefinition, &sd.Version)
		if err != nil {
			return nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_SCAN)
		}
		storedDefinitions = append(storedDefinitions, sd.Serialized())
	}
	if err = rows.Err(); err != nil {
		return nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_SCAN)
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
			allErrors = append(allErrors, fmt.Errorf("Invalid update statement: %s", stmt))
		}
	}
	entity = strings.Join(lines, "\n")
	return entity, allErrors
}