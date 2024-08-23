package instance

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go/wait"

	. "github.com/onsi/gomega"
	"github.com/testcontainers/testcontainers-go"

	"github.com/Permify/permify/internal/config"
	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/internal/storage/postgres/utils"
	"github.com/Permify/permify/pkg/database"
	PQDatabase "github.com/Permify/permify/pkg/database/postgres"
)

func PostgresDB(postgresVersion string) database.Database {
	ctx := context.Background()

	image := fmt.Sprintf("postgres:%s-alpine", postgresVersion)

	postgres, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        image,
			ExposedPorts: []string{"5432/tcp"},
			Env:          map[string]string{"POSTGRES_USER": "postgres", "POSTGRES_PASSWORD": "postgres", "POSTGRES_DB": "permify"},
			WaitingFor: wait.ForAll(
				wait.ForLog("database system is ready to accept connections"),
				wait.ForListeningPort("5432/tcp"),
			),
		},
		Started: true,
	})
	Expect(err).ShouldNot(HaveOccurred())

	// Execute the command in the container
	_, _, execErr := postgres.Exec(ctx, []string{"psql", "-U", "postgres", "-c", "ALTER SYSTEM SET track_commit_timestamp = on;"})
	Expect(execErr).ShouldNot(HaveOccurred())

	stopTimeout := 2 * time.Second
	err = postgres.Stop(context.Background(), &stopTimeout)
	Expect(err).ShouldNot(HaveOccurred())

	err = postgres.Start(context.Background())
	Expect(err).ShouldNot(HaveOccurred())

	cmd := []string{"sh", "-c", "export PGPASSWORD=postgres" + "; psql -U postgres -d permify -c 'DROP SCHEMA public CASCADE; CREATE SCHEMA public;'"}

	_, _, err = postgres.Exec(ctx, cmd)
	Expect(err).ShouldNot(HaveOccurred())

	host, err := postgres.Host(ctx)
	Expect(err).ShouldNot(HaveOccurred())

	port, err := postgres.MappedPort(ctx, "5432")
	Expect(err).ShouldNot(HaveOccurred())

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
	Expect(err).ShouldNot(HaveOccurred())

	var db database.Database
	db, err = PQDatabase.New(cfg.URI,
		PQDatabase.MaxOpenConnections(cfg.MaxOpenConnections),
		PQDatabase.MaxIdleConnections(cfg.MaxIdleConnections),
		PQDatabase.MaxConnectionIdleTime(cfg.MaxConnectionIdleTime),
		PQDatabase.MaxConnectionLifeTime(cfg.MaxConnectionLifetime),
	)

	_, err = utils.EnsureDBVersion(db.(*PQDatabase.Postgres).WritePool)
	Expect(err).ShouldNot(HaveOccurred())

	return db
}
