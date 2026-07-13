package memory

import (
	"context"
	"errors"
	"sync"

	"github.com/hashicorp/go-memdb"
	"github.com/rs/xid"

	"github.com/Permify/permify/internal/schema"
	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/internal/storage/memory/constants"
	"github.com/Permify/permify/internal/storage/memory/utils"
	"github.com/Permify/permify/pkg/database"
	db "github.com/Permify/permify/pkg/database/memory"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

var (
	sharedHeadVersion   map[string]string
	sharedHeadVersionMu sync.Mutex
)

func init() {
	sharedHeadVersion = make(map[string]string)
}

// SharedSchemaReader - Structure for Shared Schema Reader
type SharedSchemaReader struct {
	database *db.Memory
}

// NewSharedSchemaReader - Creates a new SharedSchemaReader
func NewSharedSchemaReader(database *db.Memory) *SharedSchemaReader {
	return &SharedSchemaReader{
		database: database,
	}
}

// ReadSharedSchema reads a shared schema from repository
func (r *SharedSchemaReader) ReadSharedSchema(_ context.Context, sharedSchemaID, version string) (sch *base.SchemaDefinition, err error) {
	txn := r.database.DB.Txn(false)
	defer txn.Abort()

	var it memdb.ResultIterator
	it, err = txn.Get(constants.SharedSchemaDefinitionsTable, "version", sharedSchemaID, version)
	if err != nil {
		return sch, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}

	var definitions []string
	for obj := it.Next(); obj != nil; obj = it.Next() {
		definitions = append(definitions, obj.(storage.SharedSchemaDefinition).Serialized())
	}

	sch, err = schema.NewSchemaFromStringDefinitions(false, definitions...)
	if err != nil {
		return nil, err
	}

	return sch, nil
}

// ReadSharedSchemaString returns the shared schema as string definitions.
func (r *SharedSchemaReader) ReadSharedSchemaString(_ context.Context, sharedSchemaID, version string) (definitions []string, err error) {
	txn := r.database.DB.Txn(false)
	defer txn.Abort()

	var it memdb.ResultIterator
	it, err = txn.Get(constants.SharedSchemaDefinitionsTable, "version", sharedSchemaID, version)
	if err != nil {
		return []string{}, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}

	for obj := it.Next(); obj != nil; obj = it.Next() {
		definitions = append(definitions, obj.(storage.SharedSchemaDefinition).Serialized())
	}

	return definitions, nil
}

// ReadSharedEntityDefinition reads a single entity definition from a shared schema.
func (r *SharedSchemaReader) ReadSharedEntityDefinition(_ context.Context, sharedSchemaID, entityName, version string) (definition *base.EntityDefinition, v string, err error) {
	txn := r.database.DB.Txn(false)
	defer txn.Abort()

	var raw interface{}
	raw, err = txn.First(constants.SharedSchemaDefinitionsTable, "id", sharedSchemaID, entityName, version)
	if err != nil {
		return nil, "", errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}

	def, ok := raw.(storage.SharedSchemaDefinition)
	if ok {
		var sch *base.SchemaDefinition
		sch, err = schema.NewSchemaFromStringDefinitions(false, def.Serialized())
		if err != nil {
			return nil, "", err
		}
		definition, err = schema.GetEntityByName(sch, entityName)
		if err != nil {
			return nil, "", err
		}
		return definition, def.Version, err
	}

	return nil, "", errors.New(base.ErrorCode_ERROR_CODE_SCHEMA_NOT_FOUND.String())
}

// ReadSharedRuleDefinition reads a single rule definition from a shared schema.
func (r *SharedSchemaReader) ReadSharedRuleDefinition(_ context.Context, sharedSchemaID, ruleName, version string) (definition *base.RuleDefinition, v string, err error) {
	txn := r.database.DB.Txn(false)
	defer txn.Abort()

	var raw interface{}
	raw, err = txn.First(constants.SharedSchemaDefinitionsTable, "id", sharedSchemaID, ruleName, version)
	if err != nil {
		return nil, "", errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}

	def, ok := raw.(storage.SharedSchemaDefinition)
	if ok {
		var sch *base.SchemaDefinition
		sch, err = schema.NewSchemaFromStringDefinitions(false, def.Serialized())
		if err != nil {
			return nil, "", err
		}
		definition, err = schema.GetRuleByName(sch, ruleName)
		if err != nil {
			return nil, "", err
		}
		return definition, def.Version, err
	}

	return nil, "", errors.New(base.ErrorCode_ERROR_CODE_SCHEMA_NOT_FOUND.String())
}

// SharedHeadVersion returns the latest version for a shared schema.
func (r *SharedSchemaReader) SharedHeadVersion(_ context.Context, sharedSchemaID string) (string, error) {
	sharedHeadVersionMu.Lock()
	defer sharedHeadVersionMu.Unlock()

	version, ok := sharedHeadVersion[sharedSchemaID]
	if !ok {
		return "", errors.New(base.ErrorCode_ERROR_CODE_SCHEMA_NOT_FOUND.String())
	}

	return version, nil
}

// ListSharedSchemas lists all shared schema IDs with pagination.
func (r *SharedSchemaReader) ListSharedSchemas(_ context.Context, pagination database.Pagination) (schemas []*base.SharedSchemaListItem, ct database.EncodedContinuousToken, err error) {
	sharedHeadVersionMu.Lock()
	defer sharedHeadVersionMu.Unlock()

	// Collect all shared schema IDs and their head versions
	type entry struct {
		id      string
		version string
	}
	var entries []entry
	for id, ver := range sharedHeadVersion {
		entries = append(entries, entry{id: id, version: ver})
	}

	// Sort by ID for consistent pagination
	for i := 0; i < len(entries); i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[i].id > entries[j].id {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}

	var lowerBound string
	startPage := false
	if pagination.Token() != "" {
		var t database.ContinuousToken
		t, err = utils.EncodedContinuousToken{Value: pagination.Token()}.Decode()
		if err != nil {
			return nil, nil, err
		}
		lowerBound = t.(utils.ContinuousToken).Value
	}

	schemas = make([]*base.SharedSchemaListItem, 0, pagination.PageSize()+1)
	for _, e := range entries {
		if lowerBound != "" && !startPage {
			if e.id == lowerBound {
				startPage = true
			}
			continue
		}

		id, parseErr := xid.FromString(e.version)
		if parseErr != nil {
			continue
		}
		schemas = append(schemas, &base.SharedSchemaListItem{
			SharedSchemaId: e.id,
			HeadVersion:    e.version,
			CreatedAt:      id.Time().String(),
		})
		if len(schemas) > int(pagination.PageSize()) {
			return schemas[:pagination.PageSize()], utils.NewContinuousToken(e.id).Encode(), nil
		}
	}

	return schemas, database.NewNoopContinuousToken().Encode(), nil
}

// GetTenantSharedSchemaID returns the shared_schema_id for a tenant, or empty string if none.
func (r *SharedSchemaReader) GetTenantSharedSchemaID(_ context.Context, tenantID string) (string, error) {
	txn := r.database.DB.Txn(false)
	defer txn.Abort()

	raw, err := txn.First(constants.TenantsTable, "id", tenantID)
	if err != nil {
		return "", nil
	}
	if raw == nil {
		return "", nil
	}

	tenant, ok := raw.(storage.Tenant)
	if !ok {
		return "", nil
	}

	return tenant.SharedSchemaID, nil
}
