package testinstance

import (
	"context"
	"fmt"
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
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

	// Execute the command in the container
	_, _, execErr := postgres.Exec(ctx, []string{"psql", "-U", "postgres", "-c", "ALTER SYSTEM SET track_commit_timestamp = on;"})
	gomega.Expect(execErr).ShouldNot(gomega.HaveOccurred())

	stopTimeout := 2 * time.Second
	err = postgres.Stop(context.Background(), &stopTimeout)
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

	err = postgres.Start(context.Background())
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

	cmd := []string{"sh", "-c", "export PGPASSWORD=postgres" + "; psql -U postgres -d permify -c 'DROP SCHEMA public CASCADE; CREATE SCHEMA public;'"}

	_, _, err = postgres.Exec(ctx, cmd)
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

	host, err := postgres.Host(ctx)
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

	port, err := postgres.MappedPort(ctx, "5432")
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

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
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

	var db database.Database
	db, err = PQDatabase.New(cfg.URI,
		PQDatabase.MaxOpenConnections(cfg.MaxOpenConnections),
		PQDatabase.MaxIdleConnections(cfg.MaxIdleConnections),
		PQDatabase.MaxConnectionIdleTime(cfg.MaxConnectionIdleTime),
		PQDatabase.MaxConnectionLifeTime(cfg.MaxConnectionLifetime),
	)
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

	_, err = utils.EnsureDBVersion(db.(*PQDatabase.Postgres).WritePool)
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

	return db
}
