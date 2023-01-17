package postgres

import (
	"time"
)

// Option - Option type
type Option func(*Postgres)

// MaxOpenConnections - Defines maximum open connections for postgresql db
func MaxOpenConnections(size int) Option {
	return func(c *Postgres) {
		c.maxOpenConnections = size
	}
}

// MaxIdleConnections - Defines maximum idle connections for postgresql db
func MaxIdleConnections(c int) Option {
	return func(p *Postgres) {
		p.maxIdleConnections = c
	}
}

// MaxConnectionIdleTime - Defines maximum connection idle for postgresql db
func MaxConnectionIdleTime(d time.Duration) Option {
	return func(p *Postgres) {
		p.maxConnectionIdleTime = d
	}
}

// MaxConnectionLifeTime - Defines maximum connection lifetime for postgresql db
func MaxConnectionLifeTime(d time.Duration) Option {
	return func(p *Postgres) {
		p.maxConnectionLifeTime = d
	}
}
