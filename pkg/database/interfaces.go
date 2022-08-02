package database

import (
	"context"
)

// Database -
type Database interface {
	IsReady(ctx context.Context) (bool, error)
	GetConnectionType() string
	Close()
}
