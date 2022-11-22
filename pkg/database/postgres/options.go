package postgres

import "time"

// Option - Option type
type Option func(*Postgres)

// MaxPoolSize - Defines maximum pool size for postgresql db
func MaxPoolSize(size int) Option {
	return func(c *Postgres) {
		c.maxPoolSize = size
	}
}

// ConnAttempts - Returns connection attempts
func ConnAttempts(attempts int) Option {
	return func(c *Postgres) {
		c.connAttempts = attempts
	}
}

// ConnTimeout - Returns connection timeout
func ConnTimeout(timeout time.Duration) Option {
	return func(c *Postgres) {
		c.connTimeout = timeout
	}
}
