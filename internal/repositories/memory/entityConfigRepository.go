package memory

import (
	"context"
	e "errors"

	"github.com/hashicorp/go-memdb"
	"github.com/jackc/pgx/v4"

	`github.com/Permify/permify/internal/repositories`
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
func (r *EntityConfigRepository) All(ctx context.Context, version string) ([]repositories.EntityConfig, errors.Error) {
	var configs []repositories.EntityConfig
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
	it, err = txn.Get("entity_config", "version", version)
	if err != nil {
		return configs, errors.DatabaseError.SetSubKind(database.ErrBuilder)
	}
	for obj := it.Next(); obj != nil; obj = it.Next() {
		configs = append(configs, obj.(repositories.EntityConfig))
	}
	return configs, nil
}

// Read -
func (r *EntityConfigRepository) Read(ctx context.Context, name string, version string) (repositories.EntityConfig, errors.Error) {
	var config repositories.EntityConfig
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
	raw, err = txn.First("entity_config", "id", name, version)
	if err != nil {
		return config, errors.DatabaseError.SetSubKind(database.ErrExecution)
	}
	if _, ok := raw.(repositories.EntityConfig); ok {
		return raw.(repositories.EntityConfig), nil
	}
	return repositories.EntityConfig{}, errors.DatabaseError.SetSubKind(database.ErrScan)
}

// Write -
func (r *EntityConfigRepository) Write(ctx context.Context, configs []repositories.EntityConfig, version string) errors.Error {
	var err error
	txn := r.Database.DB.Txn(true)
	defer txn.Abort()
	for _, config := range configs {
		config.Version = version
		if err = txn.Insert("entity_config", config); err != nil {
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
	raw, err = txn.Last("entity_config", "version")
	if err != nil {
		return "", errors.DatabaseError.SetSubKind(database.ErrExecution)
	}
	if _, ok := raw.(repositories.EntityConfig); ok {
		return raw.(repositories.EntityConfig).Version, nil
	}
	return "", errors.DatabaseError.SetSubKind(database.ErrScan)
}
