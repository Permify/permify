package factories

import (
	"fmt"

	"github.com/Permify/permify/internal/config"
	"github.com/Permify/permify/internal/storage/memory/migrations"
	"github.com/Permify/permify/internal/storage/postgres/utils" // Postgres utilities
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
//	- MaxConnections: the maximum number of connections in the pool (maps to pgxpool MaxConns)
//	- MaxOpenConnections: deprecated, use MaxConnections instead
//	- MinConnections: the minimum number of connections in the pool (maps to pgxpool MinConns)
//	- MinIdleConns: the minimum number of idle connections in the pool (maps to pgxpool MinIdleConns)
//	- MaxIdleConnections: deprecated, use MinConnections instead (maps to MinConnections if MinConnections is not set)
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

		opts := []PQDatabase.Option{
			PQDatabase.MaxConnectionIdleTime(conf.MaxConnectionIdleTime),
			PQDatabase.MaxConnectionLifeTime(conf.MaxConnectionLifetime),
			PQDatabase.WatchBufferSize(conf.WatchBufferSize),
			PQDatabase.MaxDataPerWrite(conf.MaxDataPerWrite),
			PQDatabase.MaxRetries(conf.MaxRetries),
		}

		// Add MinConnections if set (takes precedence over MaxIdleConnections for backward compatibility)
		if conf.MinConnections > 0 {
			opts = append(opts, PQDatabase.MinConnections(conf.MinConnections))
		}

		// Use MaxConnections if set, otherwise fall back to MaxOpenConnections for backward compatibility
		// Note: MaxConnections defaults to 0 in config, so we check MaxOpenConnections if MaxConnections is 0
		if conf.MaxConnections > 0 {
			opts = append(opts, PQDatabase.MaxConnections(conf.MaxConnections))
		} else {
			// Backward compatibility: if MaxConnections is not set, use MaxOpenConnections
			opts = append(opts, PQDatabase.MaxOpenConnections(conf.MaxOpenConnections))
		}

		// Kept for backward compatibility
		if conf.MaxIdleConnections > 0 {
			opts = append(opts, PQDatabase.MaxIdleConnections(conf.MaxIdleConnections))
		}

		// Add MinIdleConnections if set (takes precedence over MaxIdleConnections)
		if conf.MinIdleConns > 0 {
			opts = append(opts, PQDatabase.MinIdleConnections(conf.MinIdleConns))
		}

		// Add optional pool configuration options if set
		if conf.HealthCheckPeriod > 0 {
			opts = append(opts, PQDatabase.HealthCheckPeriod(conf.HealthCheckPeriod))
		}
		if conf.MaxConnectionLifetimeJitter > 0 {
			opts = append(opts, PQDatabase.MaxConnectionLifetimeJitter(conf.MaxConnectionLifetimeJitter))
		}
		if conf.ConnectTimeout > 0 {
			opts = append(opts, PQDatabase.ConnectTimeout(conf.ConnectTimeout))
		}

		if conf.URI == "" {
			db, err = PQDatabase.NewWithSeparateURIs(conf.Writer.URI, conf.Reader.URI, opts...)
			if err != nil {
				return nil, err
			}
		} else {
			db, err = PQDatabase.New(conf.URI, opts...)
			if err != nil {
				return nil, err
			}
		}

		// Verify postgres version compatibility
		// Check postgres version compatibility with the database
		_, err = utils.EnsureDBVersion(db.(*PQDatabase.Postgres).ReadPool)
		if err != nil { // Version check failed
			return nil, err // Return version error
		}
		// Return database instance

		return db, err
	case database.MEMORY.String():
		db, err = IMDatabase.New(migrations.Schema)
		if err != nil {
			return nil, err
		}
		return db, err
	default:
		return nil, fmt.Errorf("%s connection is unsupported", conf.Engine)
	}
}
