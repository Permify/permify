package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Masterminds/squirrel"
)

// Postgres - Structure for Postresql instance
type Postgres struct {
	ReadPool  *pgxpool.Pool
	WritePool *pgxpool.Pool

	Builder squirrel.StatementBuilderType
	// options
	maxDataPerWrite       int
	maxRetries            int
	watchBufferSize       int
	maxConnectionLifeTime time.Duration
	maxConnectionIdleTime time.Duration
	maxOpenConnections    int
	maxIdleConnections    int
	simpleMode            bool
}

// New - Creates new postgresql db instance
func New(uri string, opts ...Option) (*Postgres, error) {
	pg := &Postgres{
		maxOpenConnections: _defaultMaxOpenConnections,
		maxIdleConnections: _defaultMaxIdleConnections,
		maxDataPerWrite:    _defaultMaxDataPerWrite,
		maxRetries:         _defaultMaxRetries,
		watchBufferSize:    _defaultWatchBufferSize,
		simpleMode:         _defaultSimpleMode,
	}

	// Custom options
	for _, opt := range opts {
		opt(pg)
	}

	pg.Builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	writeConfig, err := pgxpool.ParseConfig(uri)
	if err != nil {
		return nil, err
	}

	readConfig, err := pgxpool.ParseConfig(uri)
	if err != nil {
		return nil, err
	}

	if pg.simpleMode {
		writeConfig.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
		readConfig.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	}

	writeConfig.MinConns = int32(pg.maxIdleConnections)
	readConfig.MinConns = int32(pg.maxIdleConnections)

	writeConfig.MaxConns = int32(pg.maxOpenConnections)
	readConfig.MaxConns = int32(pg.maxOpenConnections)

	writeConfig.MaxConnIdleTime = pg.maxConnectionIdleTime
	readConfig.MaxConnIdleTime = pg.maxConnectionIdleTime

	writeConfig.MaxConnLifetime = pg.maxConnectionLifeTime
	readConfig.MaxConnLifetime = pg.maxConnectionLifeTime

	initContext, cancelInit := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelInit()

	pg.WritePool, err = pgxpool.NewWithConfig(initContext, writeConfig)
	if err != nil {
		return nil, err
	}
	pg.ReadPool, err = pgxpool.NewWithConfig(initContext, readConfig)
	if err != nil {
		return nil, err
	}

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
	p.ReadPool.Close()
	p.WritePool.Close()
	return nil
}

// IsReady - Check if database is ready
func (p *Postgres) IsReady(ctx context.Context) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := p.ReadPool.Ping(ctx); err != nil {
		return false, err
	}
	return true, nil
}
