package memory

import (
	"context"
	e "errors"

	"github.com/hashicorp/go-memdb"
	"github.com/jackc/pgx/v4"

	"github.com/Permify/permify/internal/repositories/entities"
	"github.com/Permify/permify/pkg/database"
	db "github.com/Permify/permify/pkg/database/memory"
	"github.com/Permify/permify/pkg/errors"
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
func (r *EntityConfigRepository) Migrate() (err errors.Error) {
	return nil
}

// All -
func (r *EntityConfigRepository) All(ctx context.Context, version string) (entities.EntityConfigs, errors.Error) {
	var configs entities.EntityConfigs
	var err error
	if version == "" {
		version, err = r.findLastVersion(ctx)
		if err != nil {
			if e.Is(err, pgx.ErrNoRows) {
				return configs, errors.DatabaseError.SetSubKind(database.ErrRecordNotFound)
			}
			return configs, errors.DatabaseError.SetMessage(err.Error())
		}
	}

	txn := r.Database.DB.Txn(false)
	defer txn.Abort()
	var it memdb.ResultIterator
	it, err = txn.Get(entities.EntityConfig{}.Table(), "version", version)
	if err != nil {
		return configs, errors.DatabaseError.SetSubKind(database.ErrBuilder)
	}
	for obj := it.Next(); obj != nil; obj = it.Next() {
		configs = append(configs, obj.(entities.EntityConfig))
	}
	return configs, nil
}

// Read -
func (r *EntityConfigRepository) Read(ctx context.Context, name string, version string) (entities.EntityConfig, errors.Error) {
	var config entities.EntityConfig
	var err error
	if version == "" {
		version, err = r.findLastVersion(ctx)
		if err != nil {
			if e.Is(err, pgx.ErrNoRows) {
				return config, errors.DatabaseError.SetSubKind(database.ErrRecordNotFound)
			}
			return config, errors.DatabaseError.SetMessage(err.Error())
		}
	}

	txn := r.Database.DB.Txn(false)
	defer txn.Abort()
	var raw interface{}
	raw, err = txn.First(entities.EntityConfig{}.Table(), "id", name, version)
	if err != nil {
		return config, errors.DatabaseError.SetSubKind(database.ErrExecution)
	}
	if _, ok := raw.(entities.EntityConfig); ok {
		return raw.(entities.EntityConfig), nil
	}
	return entities.EntityConfig{}, errors.DatabaseError.SetSubKind(database.ErrScan)
}

// Write -
func (r *EntityConfigRepository) Write(ctx context.Context, configs entities.EntityConfigs, version string) errors.Error {
	var err error
	txn := r.Database.DB.Txn(true)
	defer txn.Abort()
	for _, config := range configs {
		config.Version = version
		if err = txn.Insert(entities.EntityConfig{}.Table(), config); err != nil {
			return errors.DatabaseError.SetSubKind(database.ErrExecution)
		}
	}
	txn.Commit()
	return nil
}

// findLastVersion -
func (r *EntityConfigRepository) findLastVersion(ctx context.Context) (string, errors.Error) {
	var err error
	txn := r.Database.DB.Txn(false)
	defer txn.Abort()
	var raw interface{}
	raw, err = txn.Last(entities.EntityConfig{}.Table(), "version")
	if err != nil {
		return "", errors.DatabaseError.SetSubKind(database.ErrExecution)
	}
	if _, ok := raw.(entities.EntityConfig); ok {
		return raw.(entities.EntityConfig).Version, nil
	}
	return "", errors.DatabaseError.SetSubKind(database.ErrScan)
}
