package factories

import (
	"fmt"

	"github.com/Permify/permify/internal/config"
	"github.com/Permify/permify/internal/storage/memory/migrations"
	"github.com/Permify/permify/internal/storage/postgres/utils"
	"github.com/Permify/permify/pkg/database"
	IMDatabase "github.com/Permify/permify/pkg/database/memory"
	PQDatabase "github.com/Permify/permify/pkg/database/postgres"
)

// DatabaseFactory is a factory function that creates a database instance according to the given configuration.
// It supports different types of databases, such as PostgreSQL and in-memory databases.
//
// conf: the configuration object containing the necessary information to create a database connection.
//
//	It should have the following properties:
//	- Engine: the type of the database, e.g., POSTGRES or MEMORY
//	- URI: the connection string for the database (only required for some database engines, e.g., POSTGRES)
//	- MaxOpenConnections: the maximum number of open connections to the database
//	- MaxIdleConnections: the maximum number of idle connections in the connection pool
//	- MaxConnectionIdleTime: the maximum amount of time a connection can be idle before being closed
//	- MaxConnectionLifetime: the maximum amount of time a connection can be reused before being closed
//	- WatchBufferSize: specifies the buffer size for database watch operations, impacting how many changes can be queued
//	- MaxDataPerWrite: sets the maximum amount of data per write operation to the database
//	- MaxRetries: defines the maximum number of retries for database operations in case of failure
//
// Returns a database.Database instance if the database connection is successfully created, or an error if the
// creation fails or the specified database engine is unsupported.
func DatabaseFactory(conf config.Database) (db database.Database, err error) {
	switch conf.Engine {
	case database.POSTGRES.String():

		if conf.URI == "" {
			db, err = PQDatabase.NewWithSeparateURIs(conf.Writer.URI, conf.Reader.URI,
				PQDatabase.MaxOpenConnections(conf.MaxOpenConnections),
				PQDatabase.MaxIdleConnections(conf.MaxIdleConnections),
				PQDatabase.MaxConnectionIdleTime(conf.MaxConnectionIdleTime),
				PQDatabase.MaxConnectionLifeTime(conf.MaxConnectionLifetime),
				PQDatabase.WatchBufferSize(conf.WatchBufferSize),
				PQDatabase.MaxDataPerWrite(conf.MaxDataPerWrite),
				PQDatabase.MaxRetries(conf.MaxRetries),
			)
			if err != nil {
				return nil, err
			}
		} else {
			db, err = PQDatabase.New(conf.URI,
				PQDatabase.MaxOpenConnections(conf.MaxOpenConnections),
				PQDatabase.MaxIdleConnections(conf.MaxIdleConnections),
				PQDatabase.MaxConnectionIdleTime(conf.MaxConnectionIdleTime),
				PQDatabase.MaxConnectionLifeTime(conf.MaxConnectionLifetime),
				PQDatabase.WatchBufferSize(conf.WatchBufferSize),
				PQDatabase.MaxDataPerWrite(conf.MaxDataPerWrite),
				PQDatabase.MaxRetries(conf.MaxRetries),
			)
			if err != nil {
				return nil, err
			}
		}

		// check postgres version
		_, err = utils.EnsureDBVersion(db.(*PQDatabase.Postgres).ReadPool)
		if err != nil {
			return nil, err
		}

		return
	case database.MEMORY.String():
		db, err = IMDatabase.New(migrations.Schema)
		if err != nil {
			return nil, err
		}
		return
	default:
		return nil, fmt.Errorf("%s connection is unsupported", conf.Engine)
	}
}
