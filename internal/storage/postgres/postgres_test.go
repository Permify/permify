package postgres

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/Permify/permify/internal/config"
	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/pkg/database"
	PQDatabase "github.com/Permify/permify/pkg/database/postgres"
)

func TestPostgres(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "postgres-suite")
}

func postgresDB(postgresVersion string) database.Database {
	ctx := context.Background()

	image := fmt.Sprintf("postgres:%s-alpine", postgresVersion)

	postgres, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        image,
			ExposedPorts: []string{"5432/tcp"},
			WaitingFor:   wait.ForLog("database system is ready to accept connections"),
			Env:          map[string]string{"POSTGRES_USER": "postgres", "POSTGRES_PASSWORD": "postgres", "POSTGRES_DB": "permify"},
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

	return db
}

// isSameArray - check if two arrays are the same
func isSameArray(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	sortedA := make([]string, len(a))
	copy(sortedA, a)
	sort.Strings(sortedA)

	sortedB := make([]string, len(b))
	copy(sortedB, b)
	sort.Strings(sortedB)

	for i := range sortedA {
		if sortedA[i] != sortedB[i] {
			return false
		}
	}

	return true
}
