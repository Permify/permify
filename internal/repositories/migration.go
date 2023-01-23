package repositories

import (
	"database/sql"
	"embed"
	"fmt"
	"log"

	"github.com/pressly/goose/v3"

	"github.com/Permify/permify/internal/config"
	"github.com/Permify/permify/pkg/database"
)

const (
	postgresMigrationDir = "postgres/migrations"
)

//go:embed postgres/migrations/*.sql
var postgresMigrations embed.FS

// Migrate - migrate the database
func Migrate(conf config.Database) (err error) {
	switch conf.Engine {
	case database.POSTGRES.String():

		var db *sql.DB
		db, err = sql.Open("pgx", conf.URI)
		if err != nil {
			return err
		}

		defer func() {
			if err = db.Close(); err != nil {
				log.Fatal("failed to close the db", err)
			}
		}()

		goose.SetTableName("migrations")

		if err = goose.SetDialect("postgres"); err != nil {
			log.Fatal("failed to initialize the migrate command", err)
		}

		goose.SetBaseFS(postgresMigrations)

		if err = goose.Up(db, postgresMigrationDir); err != nil {
			return err
		}

		return nil
	case database.MEMORY.String():
		return nil
	default:
		return fmt.Errorf("%s connection is unsupported", conf.Engine)
	}

	return err
}
