package postgres

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"

	"github.com/exaring/otelpgx"

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

// New -
func New(uri string, opts ...Option) (*Postgres, error) {
	return newDB(uri, uri, opts...)
}

// NewWithSeparateURIs -
func NewWithSeparateURIs(writerUri, readerUri string, opts ...Option) (*Postgres, error) {
	return newDB(writerUri, readerUri, opts...)
}

// new - Creates new postgresql db instance
func newDB(writerUri, readerUri string, opts ...Option) (*Postgres, error) {
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

	writeConfig, err := pgxpool.ParseConfig(writerUri)
	if err != nil {
		return nil, err
	}

	readConfig, err := pgxpool.ParseConfig(readerUri)
	if err != nil {
		return nil, err
	}

	// Set the default execution mode for queries using the write and read configurations.
	setDefaultQueryExecMode(writeConfig.ConnConfig)
	setDefaultQueryExecMode(readConfig.ConnConfig)

	// Set the plan cache mode for both write and read configurations to optimize query planning.
	setPlanCacheMode(writeConfig.ConnConfig)
	setPlanCacheMode(readConfig.ConnConfig)

	// Set the minimum number of idle connections in the pool for both write and read configurations.
	writeConfig.MinConns = int32(pg.maxIdleConnections)
	readConfig.MinConns = int32(pg.maxIdleConnections)

	// Set the maximum number of active connections in the pool for both write and read configurations.
	writeConfig.MaxConns = int32(pg.maxOpenConnections)
	readConfig.MaxConns = int32(pg.maxOpenConnections)

	// Set the maximum amount of time a connection may be idle before being closed for both configurations.
	writeConfig.MaxConnIdleTime = pg.maxConnectionIdleTime
	readConfig.MaxConnIdleTime = pg.maxConnectionIdleTime

	// Set the maximum lifetime of a connection in the pool for both configurations.
	writeConfig.MaxConnLifetime = pg.maxConnectionLifeTime
	readConfig.MaxConnLifetime = pg.maxConnectionLifeTime

	// Set a jitter to the maximum connection lifetime to prevent all connections from expiring at the same time.
	writeConfig.MaxConnLifetimeJitter = time.Duration(0.2 * float64(pg.maxConnectionLifeTime))
	readConfig.MaxConnLifetimeJitter = time.Duration(0.2 * float64(pg.maxConnectionLifeTime))

	writeConfig.ConnConfig.Tracer = otelpgx.NewTracer()
	readConfig.ConnConfig.Tracer = otelpgx.NewTracer()

	// Create connection pools for both writing and reading operations using the configured settings.
	pg.WritePool, pg.ReadPool, err = createPools(
		context.Background(), // Context used to control the lifecycle of the pools.
		writeConfig,          // Configuration settings for the write pool.
		readConfig,           // Configuration settings for the read pool.
	)
	// Handle errors during the creation of the connection pools.
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
	"disable":           "disable",
}

func setPlanCacheMode(config *pgx.ConnConfig) {
	// Default plan cache mode
	const defaultMode = "auto"

	// Extract connection string
	connStr := config.ConnString()
	planCacheMode := defaultMode

	// Check for specific plan cache modes in the connection string
	for key, value := range planCacheModes {
		if strings.Contains(connStr, "plan_cache_mode="+key) {
			if key == "disable" {
				delete(config.Config.RuntimeParams, "plan_cache_mode")
				slog.Info("setPlanCacheMode", slog.String("mode", "disabled"))
				return
			}
			planCacheMode = value
			slog.Info("setPlanCacheMode", slog.String("mode", key))
			break
		}
	}

	// Set the plan cache mode
	config.Config.RuntimeParams["plan_cache_mode"] = planCacheMode
	if planCacheMode == defaultMode {
		slog.Warn("setPlanCacheMode", slog.String("mode", defaultMode))
	}
}

// createPools initializes read and write connection pools with appropriate configurations and error handling.
func createPools(ctx context.Context, wConfig, rConfig *pgxpool.Config) (*pgxpool.Pool, *pgxpool.Pool, error) {
	// Context with timeout for creating the pools
	initCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Create write pool
	writePool, err := pgxpool.NewWithConfig(initCtx, wConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create write pool: %w", err)
	}

	// Create read pool using the same configuration
	readPool, err := pgxpool.NewWithConfig(initCtx, rConfig)
	if err != nil {
		writePool.Close() // Ensure write pool is closed on failure
		return nil, nil, fmt.Errorf("failed to create read pool: %w", err)
	}

	// Set up retry policy for pinging pools
	retryPolicy := backoff.NewExponentialBackOff()
	retryPolicy.MaxElapsedTime = 1 * time.Minute

	// Attempt to ping both pools to confirm connectivity
	err = backoff.Retry(func() error {
		pingCtx, pingCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer pingCancel()

		if err := writePool.Ping(pingCtx); err != nil {
			return fmt.Errorf("write pool ping failed: %w", err)
		}
		if err := readPool.Ping(pingCtx); err != nil {
			return fmt.Errorf("read pool ping failed: %w", err)
		}
		return nil
	}, retryPolicy)
	// Handle errors from pinging
	if err != nil {
		writePool.Close()
		readPool.Close()
		return nil, nil, fmt.Errorf("pinging pools failed: %w", err)
	}

	return writePool, readPool, nil
}
