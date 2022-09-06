package memory

import (
	"context"
	"errors"

	"github.com/hashicorp/go-memdb"
	"github.com/jackc/pgx/v4"

	internal_errors "github.com/Permify/permify/internal/internal-errors"
	"github.com/Permify/permify/internal/repositories/entities"
	db "github.com/Permify/permify/pkg/database/memory"
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
func (r *EntityConfigRepository) All(ctx context.Context, version string) (configs entities.EntityConfigs, err error) {
	if version == "" {
		version, err = r.findLastVersion(ctx)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return configs, internal_errors.EntityConfigCannotFoundError
			}
			return configs, err
		}
	}

	txn := r.Database.DB.Txn(false)
	defer txn.Abort()
	var it memdb.ResultIterator
	it, err = txn.Get(entities.EntityConfig{}.Table(), "version", version)
	if err != nil {
		return configs, err
	}
	for obj := it.Next(); obj != nil; obj = it.Next() {
		configs = append(configs, obj.(entities.EntityConfig))
	}
	return configs, err
}

// Read -
func (r *EntityConfigRepository) Read(ctx context.Context, name string, version string) (config entities.EntityConfig, err error) {
	if version == "" {
		version, err = r.findLastVersion(ctx)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return config, internal_errors.EntityConfigCannotFoundError
			}
			return config, err
		}
	}

	txn := r.Database.DB.Txn(false)
	defer txn.Abort()
	var raw interface{}
	raw, err = txn.First(entities.EntityConfig{}.Table(), "id", name, version)
	if err != nil {
		return config, err
	}
	if _, ok := raw.(entities.EntityConfig); ok {
		return raw.(entities.EntityConfig), err
	}
	return entities.EntityConfig{}, err
}

// Write -
func (r *EntityConfigRepository) Write(ctx context.Context, configs entities.EntityConfigs, version string) (err error) {
	txn := r.Database.DB.Txn(true)
	for _, config := range configs {
		config.Version = version
		if err = txn.Insert(entities.EntityConfig{}.Table(), config); err != nil {
			return err
		}
	}
	txn.Commit()
	return nil
}

// findLastVersion -
func (r *EntityConfigRepository) findLastVersion(ctx context.Context) (version string, err error) {
	txn := r.Database.DB.Txn(false)
	defer txn.Abort()
	var raw interface{}
	raw, err = txn.Last(entities.EntityConfig{}.Table(), "version")
	if err != nil {
		return "", err
	}
	if _, ok := raw.(entities.EntityConfig); ok {
		return raw.(entities.EntityConfig).Version, err
	}
	return "", err
}
