//go:build integration

package postgres

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Permify/permify/internal/config"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	postgres    testcontainers.Container
	postgresDSN string
	cfg         config.Database
)

func TestMain(m *testing.M) {
	postgresDSN = GetPostgres()

	// Give the PostgreSQL container some time to start up
	time.Sleep(5 * time.Second)

	exitCode := m.Run()

	TeardownPostgreSQL()

	os.Exit(exitCode)
}

func GetPostgres() string {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "postgres:14-alpine",
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor:   wait.ForLog("database system is ready to accept connections"),
		Env:          map[string]string{"POSTGRES_USER": "postgres", "POSTGRES_PASSWORD": "postgres", "POSTGRES_DB": "permify"},
	}

	postgres, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		fmt.Println("Error starting PostgreSQL container:", err)
		os.Exit(1)
	}

	cmd := []string{"sh", "-c", "export PGPASSWORD=postgres" + "; psql -U postgres -d permify -c 'DROP SCHEMA public CASCADE; CREATE SCHEMA public;'"}

	_, _, err = postgres.Exec(context.Background(), cmd)
	if err != nil {
		fmt.Println("Error resetting PostgreSQL schema:", err)
		os.Exit(1)
	}

	host, err := postgres.Host(ctx)
	if err != nil {
		fmt.Println("Error getting PostgreSQL container host:", err)
		os.Exit(1)
	}

	port, err := postgres.MappedPort(ctx, "5432")
	if err != nil {
		fmt.Println("Error getting PostgreSQL container port:", err)
		os.Exit(1)
	}
	dbAddr := fmt.Sprintf("%s:%s", host, port.Port())
	postgresDSN = fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", "postgres", "postgres", dbAddr, "permify")
	cfg = config.Database{
		Engine:                "postgres",
		URI:                   postgresDSN,
		AutoMigrate:           true,
		MaxOpenConnections:    20,
		MaxIdleConnections:    1,
		MaxConnectionLifetime: 300,
		MaxConnectionIdleTime: 60,
	}

	return postgresDSN
}

func TeardownPostgreSQL() {
	if postgres != nil {
		postgres.Terminate(context.Background())
	}
}
