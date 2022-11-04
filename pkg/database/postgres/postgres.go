package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v4"

	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/Masterminds/squirrel"

	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// Postgres -.
type Postgres struct {
	maxPoolSize  int
	connAttempts int
	connTimeout  time.Duration

	Builder squirrel.StatementBuilderType
	Pool    *pgxpool.Pool
}

// New -.
func New(uri string, database string, opts ...Option) (*Postgres, error) {
	pg := &Postgres{
		maxPoolSize:  _defaultMaxPoolSize,
		connAttempts: _defaultConnAttempts,
		connTimeout:  _defaultConnTimeout,
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

	poolConfig.MaxConns = int32(pg.maxPoolSize)

	pg.Pool, err = pgxpool.ConnectConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, err
	}

	_, err = pg.IsReady(context.Background())
	if err != nil {
		return nil, err
	}

	return pg, nil
}

// IsReady -
func (p *Postgres) IsReady(ctx context.Context) (bool, error) {
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, p.connTimeout)
	defer cancel()

	if err := p.Pool.Ping(ctx); err != nil {
		return false, err
	}

	return true, nil
}

// Migrate -
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

// GetEngineType -
func (p *Postgres) GetEngineType() string {
	return "postgres"
}

// Close -.
func (p *Postgres) Close() error {
	if p.Pool != nil {
		p.Pool.Close()
	}
	return nil
}
