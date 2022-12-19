package postgres

import "time"

// Option - Option type
type Option func(*Postgres)

// MaxOpenConnections - Defines maximum open connections for postgresql db
func MaxOpenConnections(size int) Option {
	return func(c *Postgres) {
		c.maxOpenConnections = size
	}
}

// ConnectionTimeout - Returns connection timeout
func ConnectionTimeout(timeout time.Duration) Option {
	return func(c *Postgres) {
		c.connectionTimeout = timeout
	}
}
