package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/cenkalti/backoff/v4"

	"github.com/Masterminds/squirrel"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Postgres - Structure for Postresql instance
type Postgres struct {
	DB      *sql.DB
	Builder squirrel.StatementBuilderType
	// options
	maxDataPerWrite       int
	maxRetries            int
	watchBufferSize       int
	maxConnectionLifeTime time.Duration
	maxConnectionIdleTime time.Duration
	maxOpenConnections    int
	maxIdleConnections    int
}

// New - Creates new postgresql db instance
func New(uri string, opts ...Option) (*Postgres, error) {
	pg := &Postgres{
		maxOpenConnections: _defaultMaxOpenConnections,
		maxIdleConnections: _defaultMaxIdleConnections,
		maxDataPerWrite:    _defaultMaxDataPerWrite,
		maxRetries:         _defaultMaxRetries,
		watchBufferSize:    _defaultWatchBufferSize,
	}

	// Custom options
	for _, opt := range opts {
		opt(pg)
	}

	pg.Builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	db, err := sql.Open("pgx", uri)
	if err != nil {
		return nil, err
	}

	if pg.maxOpenConnections != 0 {
		db.SetMaxOpenConns(pg.maxOpenConnections)
	}

	if pg.maxIdleConnections != 0 {
		db.SetMaxIdleConns(pg.maxIdleConnections)
	}

	if pg.maxConnectionLifeTime != 0 {
		db.SetConnMaxLifetime(pg.maxConnectionLifeTime)
	}

	if pg.maxConnectionIdleTime != 0 {
		db.SetConnMaxIdleTime(pg.maxConnectionIdleTime)
	}

	policy := backoff.NewExponentialBackOff()
	policy.MaxElapsedTime = 1 * time.Minute
	err = backoff.Retry(func() error {
		err = db.PingContext(context.Background())
		if err != nil {
			return err
		}
		return nil
	}, policy)
	if err != nil {
		return nil, err
	}

	pg.DB = db
	return pg, nil
}

func (p *Postgres) GetMaxDataPerWrite() int {
	return p.maxDataPerWrite
}

func (p *Postgres) GetMaxRetries() int {
	return p.maxRetries
}

func (p *Postgres) GetWatchBufferSize() int {
	return p.watchBufferSize
}

// GetEngineType - Get the engine type which is postgresql in string
func (p *Postgres) GetEngineType() string {
	return "postgres"
}

// Close - Close postgresql instance
func (p *Postgres) Close() error {
	if p.DB != nil {
		return p.DB.Close()
	}
	return nil
}

// IsReady - Check if database is ready
func (p *Postgres) IsReady(ctx context.Context) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := p.DB.PingContext(ctx); err != nil {
		return false, err
	}
	return true, nil
}
