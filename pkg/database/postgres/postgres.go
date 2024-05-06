package postgres

import (
	"context"
	"strings"
	"time"

	"golang.org/x/exp/slog"

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

	writeConfig, err := pgxpool.ParseConfig(uri)
	if err != nil {
		return nil, err
	}

	readConfig, err := pgxpool.ParseConfig(uri)
	if err != nil {
		return nil, err
	}

	setDefaultQueryExecMode(writeConfig.ConnConfig)
	setDefaultQueryExecMode(readConfig.ConnConfig)
	setPlanCacheMode(writeConfig.ConnConfig)
	setPlanCacheMode(readConfig.ConnConfig)

	writeConfig.MinConns = int32(pg.maxIdleConnections)
	readConfig.MinConns = int32(pg.maxIdleConnections)

	writeConfig.MaxConns = int32(pg.maxOpenConnections)
	readConfig.MaxConns = int32(pg.maxOpenConnections)

	writeConfig.MaxConnIdleTime = pg.maxConnectionIdleTime
	readConfig.MaxConnIdleTime = pg.maxConnectionIdleTime

	writeConfig.MaxConnLifetime = pg.maxConnectionLifeTime
	readConfig.MaxConnLifetime = pg.maxConnectionLifeTime

	writeConfig.MaxConnLifetimeJitter = time.Duration(0.2 * float64(pg.maxConnectionLifeTime))
	readConfig.MaxConnLifetimeJitter = time.Duration(0.2 * float64(pg.maxConnectionLifeTime))

	initialContext, cancelInit := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelInit()

	pg.WritePool, err = pgxpool.NewWithConfig(initialContext, writeConfig)
	if err != nil {
		return nil, err
	}
	pg.ReadPool, err = pgxpool.NewWithConfig(initialContext, readConfig)
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

var queryExecModes = map[string]pgx.QueryExecMode{
	"cache_statement": pgx.QueryExecModeCacheStatement,
	"cache_describe":  pgx.QueryExecModeCacheDescribe,
	"describe_exec":   pgx.QueryExecModeDescribeExec,
	"mode_exec":       pgx.QueryExecModeExec,
	"simple_protocol": pgx.QueryExecModeSimpleProtocol,
}

func setDefaultQueryExecMode(config *pgx.ConnConfig) {
	// Default mode if no specific mode is found in the connection string
	defaultMode := "cache_statement"

	// Iterate through the map keys to check if any are mentioned in the connection string
	for key := range queryExecModes {
		if strings.Contains(config.ConnString(), "default_query_exec_mode="+key) {
			config.DefaultQueryExecMode = queryExecModes[key]
			slog.Info("setDefaultQueryExecMode", slog.String("mode", key))
			return
		}
	}

	// Set to default mode if no matching mode is found
	config.DefaultQueryExecMode = queryExecModes[defaultMode]
	slog.Warn("setDefaultQueryExecMode", slog.String("mode", defaultMode))
}

var planCacheModes = map[string]string{
	"auto":              "auto",
	"force_custom_plan": "force_custom_plan",
}

func setPlanCacheMode(config *pgx.ConnConfig) {
	// Default mode if no specific mode is found in the connection string
	defaultMode := "auto"

	// Check if a plan cache mode is mentioned in the connection string and set it
	for key := range planCacheModes {
		if strings.Contains(config.ConnString(), "plan_cache_mode="+key) {
			config.Config.RuntimeParams["plan_cache_mode"] = planCacheModes[key]
			slog.Info("setPlanCacheMode", slog.String("mode", key))
			return
		}
	}

	// Set to default mode if no matching mode is found
	config.Config.RuntimeParams["plan_cache_mode"] = planCacheModes[defaultMode]
	slog.Warn("setPlanCacheMode", slog.String("mode", defaultMode))
}
