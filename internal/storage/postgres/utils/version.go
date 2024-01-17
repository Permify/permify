package utils

import (
	"database/sql"
	"fmt"
)

const (
	// The earliest supported version of PostgreSQL is 13.8
	earliestPostgresVersion = 130008
)

// EnsureDBVersion checks the version of the given database connection and returns an error if the version is not
// supported.
func EnsureDBVersion(db *sql.DB) (version int, err error) {
	err = db.QueryRow("SHOW server_version_num;").Scan(&version)
	if err != nil {
		return
	}
	if version < earliestPostgresVersion {
		err = fmt.Errorf("unsupported postgres version: %d, expected >= %d", version, earliestPostgresVersion)
	}
	return
}
