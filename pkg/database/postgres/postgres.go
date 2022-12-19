package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/cenkalti/backoff/v4"

	"github.com/jackc/pgx/v4"

	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/Masterminds/squirrel"

	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// Postgres - Structure for Postresql instance
type Postgres struct {
	maxOpenConnections int
	connectionTimeout  time.Duration

	Builder squirrel.StatementBuilderType
	Pool    *pgxpool.Pool
}

// New - Creates new postgresql db instance
func New(uri, database string, opts ...Option) (*Postgres, error) {
	pg := &Postgres{
		maxOpenConnections: _defaultMaxOpenConnections,
		connectionTimeout:  _defaultConnectionTimeout,
	}

	// Custom options
	for _, opt := range opts {
		opt(pg)
	}

	pg.Builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	poolConfig, err := pgxpool.ParseConfig(uri + "/" + database)
	if err != nil {
		return nil, err
	}

	poolConfig.MaxConns = int32(pg.maxOpenConnections)

	pg.Pool, err = pgxpool.ConnectConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, err
	}

	policy := backoff.NewExponentialBackOff()
	policy.MaxElapsedTime = 1 * time.Minute
	err = backoff.Retry(func() error {
		err = pg.Pool.Ping(context.Background())
		if err != nil {
			return err
		}
		return nil
	}, policy)
	if err != nil {
		return nil, err
	}

	return pg, nil
}

// Migrate - Migration operations for postgresql db
func (p *Postgres) Migrate(statements []string) (err error) {
	ctx := context.Background()

	var tx pgx.Tx
	tx, err = p.Pool.Begin(ctx)
	if err != nil {
		return errors.New(base.ErrorCode_ERROR_CODE_MIGRATION.String())
	}

	for _, statement := range statements {
		_, err = tx.Exec(ctx, statement)
		if err != nil {
			return errors.New(base.ErrorCode_ERROR_CODE_MIGRATION.String())
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return errors.New(base.ErrorCode_ERROR_CODE_MIGRATION.String())
	}

	return nil
}

// GetEngineType - Get the engine type which is postgresql in string
func (p *Postgres) GetEngineType() string {
	return "postgres"
}

// Close - Close postgresql instance
func (p *Postgres) Close() error {
	if p.Pool != nil {
		p.Pool.Close()
	}
	return nil
}
