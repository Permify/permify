package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"

	"github.com/Masterminds/squirrel"

	_ "github.com/jackc/pgx/v5/stdlib"

	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// Postgres - Structure for Postresql instance
type Postgres struct {
	maxOpenConnections int

	Builder squirrel.StatementBuilderType
	DB      *sql.DB
}

// New - Creates new postgresql db instance
func New(uri, database string, opts ...Option) (*Postgres, error) {
	pg := &Postgres{
		maxOpenConnections: _defaultMaxOpenConnections,
	}

	// Custom options
	for _, opt := range opts {
		opt(pg)
	}

	pg.Builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	db, err := sql.Open("pgx", fmt.Sprintf("%s/%s", uri, database))
	if err != nil {
		return nil, err
	}

	if pg.maxOpenConnections != 0 {
		db.SetMaxOpenConns(pg.maxOpenConnections)
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

// Migrate - Migration operations for postgresql db
func (p *Postgres) Migrate(statements []string) (err error) {
	ctx := context.Background()

	var tx *sql.Tx
	tx, err = p.DB.Begin()
	if err != nil {
		return errors.New(base.ErrorCode_ERROR_CODE_MIGRATION.String())
	}

	for _, statement := range statements {
		_, err = tx.ExecContext(ctx, statement)
		if err != nil {
			return errors.New(base.ErrorCode_ERROR_CODE_MIGRATION.String())
		}
	}

	err = tx.Commit()
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
