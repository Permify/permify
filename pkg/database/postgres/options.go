package postgres

import (
	"time"
)

// Option - Option type
type Option func(*Postgres)

// MaxConns - Defines maximum number of connections in the pool (maps to pgxpool MaxConns)
// Deprecated: Use MaxConns instead of MaxOpenConnections for consistency with pgxpool.
// MaxOpenConnections is kept for backward compatibility and internally uses MaxConns.
func MaxOpenConnections(size int) Option {
	return MaxConns(size)
}

// MaxConns - Defines maximum number of connections in the pool (maps to pgxpool MaxConns)
func MaxConns(size int) Option {
	return func(c *Postgres) {
		c.maxConns = size
	}
}

// MaxIdleConnections - Defines maximum idle connections for postgresql db
// Deprecated: Use MinConns and/or MinIdleConns instead. This is kept for backward compatibility.
// If MinConns is not set, MaxIdleConnections will be used as MinConns (old behavior).
// If MinIdleConns is not set, MaxIdleConnections will also be used as MinIdleConns.
func MaxIdleConnections(c int) Option {
	return func(p *Postgres) {
		p.maxIdleConnections = c
	}
}

// MinConns - Defines minimum number of connections in the pool
// If not set (0) and MaxIdleConnections is set, MaxIdleConnections will be used for backward compatibility (old behavior).
func MinConns(c int) Option {
	return func(p *Postgres) {
		p.minConns = c
	}
}

// MinIdleConns - Defines minimum number of idle connections in the pool
// This is superior to MinConns for ensuring idle connections are always available.
// If not set and MaxIdleConnections is set, MaxIdleConnections will be used for backward compatibility.
// Note: MaxIdleConnections also affects MinConns if MinConns is not set.
func MinIdleConns(c int) Option {
	return func(p *Postgres) {
		p.minIdleConns = c
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

// MaxConnLifetimeJitter - Defines the jitter added to MaxConnLifetime to prevent all connections from expiring at once
func MaxConnLifetimeJitter(d time.Duration) Option {
	return func(p *Postgres) {
		p.maxConnLifetimeJitter = d
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
