package postgres

import (
	"time"
)

// Option - Option type
type Option func(*Postgres)

// MaxOpenConnections - Deprecated: use MaxConnections instead for consistency with pgxpool.
// Kept for backward compatibility and internally forwards to MaxConnections.
func MaxOpenConnections(size int) Option {
	return MaxConnections(size)
}

// MaxConnections - Defines maximum number of connections in the pool (maps to pgxpool MaxConns)
func MaxConnections(size int) Option {
	return func(c *Postgres) {
		c.maxConnections = size
	}
}

// MaxIdleConnections - Deprecated: use MinConnections instead.
// Kept for backward compatibility and only used as a fallback for MinConnections when
// MinConnections is not set (0). MinIdleConns is only honored when explicitly configured.
func MaxIdleConnections(c int) Option {
	return func(p *Postgres) {
		p.maxIdleConnections = c
	}
}

// MinConnections - Defines minimum number of connections in the pool
// If not set (0) and MaxIdleConnections is set, MaxIdleConnections will be used for backward compatibility (old behavior).
func MinConnections(c int) Option {
	return func(p *Postgres) {
		p.minConnections = c
	}
}

// MinIdleConnections - Defines minimum number of idle connections in the pool.
// This is superior to MinConnections for ensuring idle connections are always available.
// Note: MaxIdleConnections only affects MinConnections when MinConnections is not set; it does
// not implicitly set MinIdleConnections.
func MinIdleConnections(c int) Option {
	return func(p *Postgres) {
		p.minIdleConnections = c
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

// HealthCheckPeriod - Defines the period between health checks on idle connections
func HealthCheckPeriod(d time.Duration) Option {
	return func(p *Postgres) {
		p.healthCheckPeriod = d
	}
}

// MaxConnectionLifetimeJitter - Defines the jitter added to MaxConnLifetime to prevent all connections from expiring at once
func MaxConnectionLifetimeJitter(d time.Duration) Option {
	return func(p *Postgres) {
		p.maxConnectionLifetimeJitter = d
	}
}

// ConnectTimeout - Defines the maximum time to wait when establishing a new connection
func ConnectTimeout(d time.Duration) Option {
	return func(p *Postgres) {
		p.connectTimeout = d
	}
}

func MaxDataPerWrite(v int) Option {
	return func(c *Postgres) {
		c.maxDataPerWrite = v
	}
}

func WatchBufferSize(v int) Option {
	return func(c *Postgres) {
		c.watchBufferSize = v
	}
}

func MaxRetries(v int) Option {
	return func(c *Postgres) {
		c.maxRetries = v
	}
}
