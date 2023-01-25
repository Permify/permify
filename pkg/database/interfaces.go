package database

import (
	"context"
)

// Database - Db interface
type Database interface {
	// GetEngineType get the database type (e.g. postgres, memory, etc.).
	GetEngineType() string

	// Close the database connection.
	Close() error

	// IsReady - Check if database is ready
	IsReady(ctx context.Context) (bool, error)
}
