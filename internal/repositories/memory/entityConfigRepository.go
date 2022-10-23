package memory

import (
	"context"

	"github.com/pkg/errors"

	"github.com/hashicorp/go-memdb"

	"github.com/Permify/permify/internal/repositories"
	db "github.com/Permify/permify/pkg/database/memory"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// EntityConfigRepository -.
type EntityConfigRepository struct {
	Database *db.Memory
}

// NewEntityConfigRepository -.
func NewEntityConfigRepository(mm *db.Memory) *EntityConfigRepository {
	return &EntityConfigRepository{mm}
}

// Migrate -
func (r *EntityConfigRepository) Migrate() (err error) {
	return nil
}

// All -
func (r *EntityConfigRepository) All(ctx context.Context, version string) ([]repositories.EntityConfig, error) {
	var configs []repositories.EntityConfig
	var err error
	if version == "" {
		version, err = r.findLastVersion(ctx)
		if err != nil {
			return configs, err
		}
	}

	txn := r.Database.DB.Txn(false)
	defer txn.Abort()
	var it memdb.ResultIterator
	it, err = txn.Get("entity_config", "version", version)
	if err != nil {
		return configs, errors.New(base.ErrorCode_execution.String())
	}
	for obj := it.Next(); obj != nil; obj = it.Next() {
		configs = append(configs, obj.(repositories.EntityConfig))
	}
	return configs, nil
}

// Read -
func (r *EntityConfigRepository) Read(ctx context.Context, name string, version string) (repositories.EntityConfig, error) {
	var config repositories.EntityConfig
	var err error
	if version == "" {
		version, err = r.findLastVersion(ctx)
		if err != nil {
			return config, err
		}
	}

	txn := r.Database.DB.Txn(false)
	defer txn.Abort()
	var raw interface{}
	raw, err = txn.First("entity_config", "id", name, version)
	if err != nil {
		return config, errors.New(base.ErrorCode_execution.String())
	}
	if _, ok := raw.(repositories.EntityConfig); ok {
		return raw.(repositories.EntityConfig), nil
	}
	return repositories.EntityConfig{}, errors.New(base.ErrorCode_schema_not_found.String())
}

// Write -
func (r *EntityConfigRepository) Write(ctx context.Context, configs []repositories.EntityConfig, version string) error {
	var err error
	txn := r.Database.DB.Txn(true)
	defer txn.Abort()
	for _, config := range configs {
		config.Version = version
		if err = txn.Insert("entity_config", config); err != nil {
			return errors.New(base.ErrorCode_execution.String())
		}
	}
	txn.Commit()
	return nil
}

// findLastVersion -
func (r *EntityConfigRepository) findLastVersion(ctx context.Context) (string, error) {
	var err error
	txn := r.Database.DB.Txn(false)
	defer txn.Abort()
	var raw interface{}
	raw, err = txn.Last("entity_config", "version")
	if err != nil {
		return "", errors.New(base.ErrorCode_execution.String())
	}
	if _, ok := raw.(repositories.EntityConfig); ok {
		return raw.(repositories.EntityConfig).Version, nil
	}
	return "", errors.New(base.ErrorCode_schema_not_found.String())
}
