package utils

import (
	"context"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	// The earliest supported version of PostgreSQL is 13.8
	earliestPostgresVersion = 130008
)

// EnsureDBVersion checks the version of the given database connection and returns an error if the version is not
// supported.
func EnsureDBVersion(db *pgxpool.Pool) (version string, err error) {
	err = db.QueryRow(context.Background(), "SHOW server_version_num;").Scan(&version)
	if err != nil {
		return
	}
	v, err := strconv.Atoi(version)
	if v < earliestPostgresVersion {
		err = fmt.Errorf("unsupported postgres version: %s, expected >= %d", version, earliestPostgresVersion)
	}
	return
}
