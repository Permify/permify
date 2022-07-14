package services

import (
	`github.com/dgraph-io/ristretto`
	`golang.org/x/net/context`

	e `github.com/Permify/permify/internal/entities`
	`github.com/Permify/permify/internal/migrations`
	`github.com/Permify/permify/internal/repositories`
	PQDatabase `github.com/Permify/permify/pkg/database/postgres`
	`github.com/Permify/permify/pkg/dsl/schema`
	`github.com/Permify/permify/pkg/migration`
)

// ISchemaService -
type ISchemaService interface {
	AllSchema(ctx context.Context) (sch schema.Schema, err error)
	Schema(ctx context.Context, entity string) (sch schema.Entity, err error)
	ReplaceSchema(ctx context.Context, configs []e.EntityConfig) error
}

// SchemaService -
type SchemaService struct {
	repository repositories.IEntityConfigRepository
	cache      *ristretto.Cache
	listen     *PQDatabase.Postgres
}

// NewSchemaService -
func NewSchemaService(repo repositories.IEntityConfigRepository, cache *ristretto.Cache, write *PQDatabase.Postgres) *SchemaService {
	return &SchemaService{
		repository: repo,
		cache:      cache,
		listen:     write,
	}
}

// AllSchema -
func (service *SchemaService) AllSchema(ctx context.Context) (sch schema.Schema, err error) {
	//service.repository.All(ctx)
	return
}

// Schema -
func (service *SchemaService) Schema(ctx context.Context, entity string) (sch schema.Entity, err error) {
	//service.repository.First(ctx, entity)
	return
}

// ReplaceSchema -
func (service *SchemaService) ReplaceSchema(ctx context.Context, configs []e.EntityConfig) (err error) {

	var sch schema.Schema
	sch, err = service.AllSchema(context.TODO())
	if err != nil {
		return err
	}

	var notifierMigrations *migration.Migration
	notifierMigrations, err = migrations.RegisterNotifierMigrations(sch.GetTableNames())
	if err != nil {
		return err
	}

	err = service.listen.Migrate(*notifierMigrations)
	if err != nil {
		return err
	}

	return service.repository.Replace(ctx, configs)
}
