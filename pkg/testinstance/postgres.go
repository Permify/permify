package testinstance

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/onsi/gomega"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/Permify/permify/internal/config"
	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/internal/storage/postgres/utils"
	"github.com/Permify/permify/pkg/database"
	PQDatabase "github.com/Permify/permify/pkg/database/postgres"
)

type PostgresInstance struct {
	Postgres  *PQDatabase.Postgres
	container testcontainers.Container
	closeOnce sync.Once
	closeErr  error
}

var _ database.Database = (*PostgresInstance)(nil)

func (p *PostgresInstance) Close() error {
	p.closeOnce.Do(func() {
		var errs []error

		if p.Postgres != nil {
			errs = append(errs, p.Postgres.Close())
		}

		if p.container != nil {
			terminateCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			errs = append(errs, p.container.Terminate(terminateCtx))
		}

		p.closeErr = errors.Join(errs...)
	})

	return p.closeErr
}

func (p *PostgresInstance) GetEngineType() string {
	return p.Postgres.GetEngineType()
}

func (p *PostgresInstance) IsReady(ctx context.Context) (bool, error) {
	return p.Postgres.IsReady(ctx)
}

func PostgresDB(postgresVersion string) *PostgresInstance {
	ctx := context.Background()

	image := fmt.Sprintf("postgres:%s-alpine", postgresVersion)

	postgres, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        image,
			ExposedPorts: []string{"5432/tcp"},
			Env:          map[string]string{"POSTGRES_USER": "postgres", "POSTGRES_PASSWORD": "postgres", "POSTGRES_DB": "permify"},
			Cmd:          []string{"postgres", "-c", "track_commit_timestamp=on"},
			WaitingFor: wait.ForAll(
				wait.ForLog("database system is ready to accept connections"),
				wait.ForListeningPort("5432/tcp"),
			),
		},
		Started: true,
	})
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

	var db *PQDatabase.Postgres

	cleanup := func() {
		if db != nil {
			_ = db.Close()
		}
		if postgres != nil {
			terminateCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			_ = postgres.Terminate(terminateCtx)
		}
	}

	expectNoError := func(err error) {
		if err != nil {
			cleanup()
		}
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	}

	host, err := postgres.Host(ctx)
	expectNoError(err)

	port, err := postgres.MappedPort(ctx, "5432")
	expectNoError(err)

	dbAddr := fmt.Sprintf("%s:%s", host, port.Port())
	postgresDSN := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", "postgres", "postgres", dbAddr, "permify")

	cfg := config.Database{
		Engine:                "postgres",
		URI:                   postgresDSN,
		AutoMigrate:           true,
		MaxOpenConnections:    20,
		MaxIdleConnections:    1,
		MaxConnectionLifetime: 300,
		MaxConnectionIdleTime: 60,
	}

	err = storage.Migrate(cfg)
	expectNoError(err)

	db, err = PQDatabase.New(cfg.URI,
		PQDatabase.MaxOpenConnections(cfg.MaxOpenConnections),
		PQDatabase.MaxIdleConnections(cfg.MaxIdleConnections),
		PQDatabase.MaxConnectionIdleTime(cfg.MaxConnectionIdleTime),
		PQDatabase.MaxConnectionLifeTime(cfg.MaxConnectionLifetime),
	)
	expectNoError(err)

	_, err = utils.EnsureDBVersion(db.WritePool)
	expectNoError(err)

	return &PostgresInstance{
		Postgres:  db,
		container: postgres,
	}
}
