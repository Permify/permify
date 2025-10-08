package utils // Postgres utility functions
import (      // Package imports
	"context" // Context
	"fmt"     // Formatting
	"strconv" // String conversion

	"github.com/jackc/pgx/v5/pgxpool" // Postgres connection pool
) // End imports
// Version constants
const ( // Version constants
	earliestPostgresVersion = 130008 // The earliest supported version of PostgreSQL is 13.8
) // End constants
// EnsureDBVersion checks the version of the given database connection
// and returns an error if the version is not supported.
func EnsureDBVersion(db *pgxpool.Pool) (version string, err error) { // Check database version
	err = db.QueryRow(context.Background(), "SHOW server_version_num;").Scan(&version) // Query version
	if err != nil {                                                                    // Query failed
		return // Return error
	} // Query succeeded
	v, err := strconv.Atoi(version)  // Convert to int
	if v < earliestPostgresVersion { // Check minimum version
		err = fmt.Errorf("unsupported postgres version: %s, expected >= %d", version, earliestPostgresVersion) // Version too old
	} // Version check done
	return // Return version
} // End EnsureDBVersion
