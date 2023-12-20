package storage

import (
	"embed"
	"fmt"
	"log"

	"github.com/pressly/goose/v3"

	"github.com/Permify/permify/internal/config"
	"github.com/Permify/permify/pkg/database"
	PQDatabase "github.com/Permify/permify/pkg/database/postgres"
)

const (
	postgresMigrationDir = "postgres/migrations"
	postgresDialect      = "postgres"
	migrationsTable      = "migrations"
)

//go:embed postgres/migrations/*.sql
var postgresMigrations embed.FS

// Migrate performs database migrations depending on the given configuration.
func Migrate(conf config.Database) (err error) {
	switch conf.Engine {
	case database.POSTGRES.String():
		// Create a new Postgres database connection
		var db *PQDatabase.Postgres
		db, err = PQDatabase.New(conf.URI)
		if err != nil {
			return err
		}
		// Ensure database connection is closed when function returns
		defer closeDB(db)

		// Set table name for migrations
		goose.SetTableName(migrationsTable)

		// Set dialect to be used for migration
		if err = goose.SetDialect(postgresDialect); err != nil {
			return err
		}

		// Set file system for migration scripts
		goose.SetBaseFS(postgresMigrations)

		// Perform migration
		if err = goose.Up(db.DB, postgresMigrationDir); err != nil {
			return err
		}

		return nil
	case database.MEMORY.String():
		// No migrations needed for in-memory database
		return nil
	default:
		// Unsupported database engine
		return fmt.Errorf("%s connection is unsupported", conf.Engine)
	}
}

// MigrateUp performs all available database migrations to update the schema to the latest version.
func MigrateUp(engine, uri string) (err error) {
	switch engine {
	case database.POSTGRES.String():
		var db *PQDatabase.Postgres
		db, err = PQDatabase.New(uri)
		if err != nil {
			return err
		}
		defer closeDB(db)

		goose.SetTableName(migrationsTable)

		if err = goose.SetDialect(postgresDialect); err != nil {
			return err
		}

		goose.SetBaseFS(postgresMigrations)

		if err = goose.Up(db.DB, postgresMigrationDir); err != nil {
			return err
		}

		return nil
	case database.MEMORY.String():
		return nil
	default:
		return fmt.Errorf("%s connection is unsupported", engine)
	}
}

// MigrateUpTo performs database migrations up to a specific version.
func MigrateUpTo(engine, uri string, p int64) (err error) {
	switch engine {
	case database.POSTGRES.String():
		var db *PQDatabase.Postgres
		db, err = PQDatabase.New(uri)
		if err != nil {
			return err
		}
		defer closeDB(db)

		goose.SetTableName(migrationsTable)

		if err = goose.SetDialect(postgresDialect); err != nil {
			return err
		}

		goose.SetBaseFS(postgresMigrations)

		if err = goose.UpTo(db.DB, postgresMigrationDir, p); err != nil {
			return err
		}

		return nil
	case database.MEMORY.String():
		return nil
	default:
		return fmt.Errorf("%s connection is unsupported", engine)
	}
}

// MigrateDown undoes all database migrations, reverting the schema to the initial state.
func MigrateDown(engine, uri string) (err error) {
	switch engine {
	case database.POSTGRES.String():
		var db *PQDatabase.Postgres
		db, err = PQDatabase.New(uri)
		if err != nil {
			return err
		}
		defer closeDB(db)

		goose.SetTableName(migrationsTable)

		if err = goose.SetDialect(postgresDialect); err != nil {
			return err
		}

		goose.SetBaseFS(postgresMigrations)

		if err = goose.Down(db.DB, postgresMigrationDir); err != nil {
			return err
		}

		return nil
	case database.MEMORY.String():
		return nil
	default:
		return fmt.Errorf("%s connection is unsupported", engine)
	}
}

// MigrateDownTo undoes database migrations down to a specific version.
func MigrateDownTo(engine, uri string, p int64) (err error) {
	switch engine {
	case database.POSTGRES.String():
		var db *PQDatabase.Postgres
		db, err = PQDatabase.New(uri)
		if err != nil {
			return err
		}
		defer closeDB(db)

		goose.SetTableName(migrationsTable)

		if err = goose.SetDialect(postgresDialect); err != nil {
			return err
		}

		goose.SetBaseFS(postgresMigrations)

		if err = goose.DownTo(db.DB, postgresMigrationDir, p); err != nil {
			return err
		}

		return nil
	case database.MEMORY.String():
		return nil
	default:
		return fmt.Errorf("%s connection is unsupported", engine)
	}
}

// MigrateReset roll back all migrations.
func MigrateReset(engine, uri string) (err error) {
	switch engine {
	case database.POSTGRES.String():
		var db *PQDatabase.Postgres
		db, err = PQDatabase.New(uri)
		if err != nil {
			return err
		}
		defer closeDB(db)

		goose.SetTableName(migrationsTable)

		if err = goose.SetDialect(postgresDialect); err != nil {
			return err
		}

		goose.SetBaseFS(postgresMigrations)

		if err = goose.Reset(db.DB, postgresMigrationDir); err != nil {
			return err
		}

		return nil
	case database.MEMORY.String():
		return nil
	default:
		return fmt.Errorf("%s connection is unsupported", engine)
	}
}

// MigrateStatus displays the status of all migrations.
func MigrateStatus(engine, uri string) (err error) {
	switch engine {
	case database.POSTGRES.String():
		var db *PQDatabase.Postgres
		db, err = PQDatabase.New(uri)
		if err != nil {
			return err
		}
		defer closeDB(db)

		goose.SetTableName(migrationsTable)

		if err = goose.SetDialect(postgresDialect); err != nil {
			return err
		}

		goose.SetBaseFS(postgresMigrations)

		if err = goose.Status(db.DB, postgresMigrationDir); err != nil {
			return err
		}

		return nil
	case database.MEMORY.String():
		return nil
	default:
		return fmt.Errorf("%s connection is unsupported", engine)
	}
}

// closeDB cleanly closes the database connection and logs if an error occurs.
func closeDB(db *PQDatabase.Postgres) {
	if err := db.Close(); err != nil {
		log.Printf("failed to close the database: %v", err)
	}
}
