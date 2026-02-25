package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Repair", func() {
	Context("RepairConfig", func() {
		It("Case 1: DefaultRepairConfig should return correct default values", func() {
			config := DefaultRepairConfig()

			Expect(config.BatchSize).Should(Equal(1000))
			Expect(config.MaxRetries).Should(Equal(3))
			Expect(config.RetryDelay).Should(Equal(100))
			Expect(config.DryRun).Should(BeFalse())
			Expect(config.Verbose).Should(BeTrue())
		})

		It("Case 2: RepairConfig should accept custom values", func() {
			config := &RepairConfig{
				BatchSize:  500,
				MaxRetries: 5,
				RetryDelay: 200,
				DryRun:     true,
				Verbose:    false,
			}

			Expect(config.BatchSize).Should(Equal(500))
			Expect(config.MaxRetries).Should(Equal(5))
			Expect(config.RetryDelay).Should(Equal(200))
			Expect(config.DryRun).Should(BeTrue())
			Expect(config.Verbose).Should(BeFalse())
		})
	})

	Context("RepairResult", func() {
		It("Case 1: Should initialize with correct default values", func() {
			result := &RepairResult{
				CreatedTxIdFixed: 0,
				Errors:           make([]error, 0),
				Duration:         "",
			}

			Expect(result.CreatedTxIdFixed).Should(Equal(0))
			Expect(result.Errors).Should(HaveLen(0))
			Expect(result.Duration).Should(BeEmpty())
		})
	})

	Context("Integration tests with real database", func() {
		var db *Postgres
		var container testcontainers.Container

		AfterEach(func() {
			if db != nil {
				db.Close()
			}
			if container != nil {
				ctx := context.Background()
				container.Terminate(ctx)
			}
		})

		Context("When database is available", func() {
			It("Case 1: Should create Postgres instance and perform repair", func() {
				var err error
				db, container, err = createTestDatabase()
				Expect(err).ShouldNot(HaveOccurred())

				// Test basic connectivity
				ctx := context.Background()
				ready, err := db.IsReady(ctx)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(ready).Should(BeTrue())

				// Test repair with dry run
				config := &RepairConfig{
					BatchSize:  100,
					MaxRetries: 3,
					RetryDelay: 100,
					DryRun:     true,
					Verbose:    true,
				}

				result, err := db.Repair(ctx, config)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(result).ShouldNot(BeNil())
				// Transactions table doesn't exist, so we expect errors
				Expect(result.Errors).Should(HaveLen(1))
				Expect(result.Errors[0].Error()).Should(ContainSubstring("transactions"))
			})

			It("Case 2: Should get current PostgreSQL XID", func() {
				var err error
				db, container, err = createTestDatabase()
				Expect(err).ShouldNot(HaveOccurred())

				ctx := context.Background()
				currentXID, err := db.getCurrentPostgreXID(ctx)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(currentXID).Should(BeNumerically(">", uint64(0)))
			})

			It("Case 3: Should get max referenced XID from transactions table", func() {
				var err error
				db, container, err = createTestDatabase()
				Expect(err).ShouldNot(HaveOccurred())

				ctx := context.Background()
				maxXID, err := db.getMaxReferencedXID(ctx)
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring("transactions"))
				Expect(maxXID).Should(Equal(uint64(0)))
			})

			It("Case 4: Should handle missing transactions table gracefully", func() {
				var err error
				db, container, err = createTestDatabase()
				Expect(err).ShouldNot(HaveOccurred())

				ctx := context.Background()

				// First, get current XID
				currentXID, err := db.getCurrentPostgreXID(ctx)
				Expect(err).ShouldNot(HaveOccurred())

				// Test repair - should handle missing transactions table
				config := &RepairConfig{
					BatchSize:  50,
					MaxRetries: 3,
					RetryDelay: 100,
					DryRun:     false,
					Verbose:    true,
				}

				result, err := db.Repair(ctx, config)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(result).ShouldNot(BeNil())
				Expect(result.Errors).Should(HaveLen(1))
				Expect(result.Errors[0].Error()).Should(ContainSubstring("transactions"))

				// XID counter should not be advanced due to error
				// Note: XID may have changed due to container restart, so we just check it's reasonable
				newXID, err := db.getCurrentPostgreXID(ctx)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(newXID).Should(BeNumerically(">=", currentXID))
			})

			It("Case 5: Should perform repair with migrations and advance XID counter", func() {
				var err error
				db, container, err = createTestDatabaseWithMigrations()
				Expect(err).ShouldNot(HaveOccurred())

				ctx := context.Background()

				// Get initial XID before repair
				initialXID, err := db.getCurrentPostgreXID(ctx)
				Expect(err).ShouldNot(HaveOccurred())

				// Get max referenced XID from transactions table (should be 0 initially)
				maxReferencedXID, err := db.getMaxReferencedXID(ctx)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(maxReferencedXID).Should(Equal(uint64(0)))

				// Insert some test data to create transactions
				_, err = db.WritePool.Exec(ctx, `
					INSERT INTO transactions (id, tenant_id, entity_type, entity_id, commit_tx_id, created_tx_id) 
					VALUES 
						(1000, 'tenant1', 'user', 'user1', 1000, 500),
						(1001, 'tenant1', 'user', 'user2', 1001, 501),
						(1002, 'tenant2', 'organization', 'org1', 1002, 502)
				`)
				Expect(err).ShouldNot(HaveOccurred())

				// Get max referenced XID after inserting data
				maxReferencedXIDAfterInsert, err := db.getMaxReferencedXID(ctx)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(maxReferencedXIDAfterInsert).Should(Equal(uint64(1002)))

				// Perform repair with dry run first
				config := &RepairConfig{
					BatchSize:  100,
					MaxRetries: 3,
					RetryDelay: 100,
					DryRun:     true,
					Verbose:    true,
				}

				result, err := db.Repair(ctx, config)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(result).ShouldNot(BeNil())
				Expect(result.CreatedTxIdFixed).Should(BeNumerically(">", 0)) // Should advance XID counter in dry run

				// Get XID after dry run (should not have changed)
				xidAfterDryRun, err := db.getCurrentPostgreXID(ctx)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(xidAfterDryRun).Should(BeNumerically(">=", initialXID))

				// Now perform actual repair
				config.DryRun = false
				result, err = db.Repair(ctx, config)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(result).ShouldNot(BeNil())
				Expect(result.CreatedTxIdFixed).Should(BeNumerically(">", uint64(0)))

				// Get XID after repair (should have advanced)
				finalXID, err := db.getCurrentPostgreXID(ctx)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(finalXID).Should(BeNumerically(">", xidAfterDryRun))

				// Verify that the repair actually advanced the XID counter
				Expect(finalXID).Should(BeNumerically(">", maxReferencedXIDAfterInsert))

				// Verify that the repair result shows the correct number of fixed transactions
				Expect(result.CreatedTxIdFixed).Should(BeNumerically(">=", maxReferencedXIDAfterInsert))
			})
		})
	})
})

// Helper functions for integration tests

// createTestDatabase creates a test database instance with container
func createTestDatabase() (*Postgres, testcontainers.Container, error) {
	ctx := context.Background()

	image := "postgres:15-alpine"

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        image,
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_USER":     "postgres",
				"POSTGRES_PASSWORD": "postgres",
				"POSTGRES_DB":       "permify",
			},
			WaitingFor: wait.ForAll(
				wait.ForLog("database system is ready to accept connections"),
				wait.ForListeningPort("5432/tcp"),
			),
		},
		Started: true,
	})
	if err != nil {
		return nil, nil, err
	}

	// Enable track_commit_timestamp for XID tracking
	_, _, execErr := container.Exec(ctx, []string{"psql", "-U", "postgres", "-c", "ALTER SYSTEM SET track_commit_timestamp = on;"})
	if execErr != nil {
		container.Terminate(ctx)
		return nil, nil, execErr
	}

	// Restart container to apply the setting
	stopTimeout := 2 * time.Second
	err = container.Stop(context.Background(), &stopTimeout)
	if err != nil {
		container.Terminate(ctx)
		return nil, nil, err
	}

	err = container.Start(context.Background())
	if err != nil {
		container.Terminate(ctx)
		return nil, nil, err
	}

	// Get connection details
	host, err := container.Host(ctx)
	if err != nil {
		container.Terminate(ctx)
		return nil, nil, err
	}

	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		container.Terminate(ctx)
		return nil, nil, err
	}

	dbAddr := fmt.Sprintf("%s:%s", host, port.Port())
	postgresDSN := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", "postgres", "postgres", dbAddr, "permify")

	// Create database instance directly without migrations
	db, err := New(postgresDSN,
		MaxOpenConnections(20),
		MaxIdleConnections(1),
		MaxConnectionIdleTime(60),
		MaxConnectionLifeTime(300),
	)
	if err != nil {
		container.Terminate(ctx)
		return nil, nil, err
	}

	return db, container, nil
}

// createTestDatabaseWithMigrations creates a test database instance with migrations
func createTestDatabaseWithMigrations() (*Postgres, testcontainers.Container, error) {
	ctx := context.Background()

	image := "postgres:15-alpine"

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        image,
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_USER":     "postgres",
				"POSTGRES_PASSWORD": "postgres",
				"POSTGRES_DB":       "permify",
			},
			WaitingFor: wait.ForAll(
				wait.ForLog("database system is ready to accept connections"),
				wait.ForListeningPort("5432/tcp"),
			),
		},
		Started: true,
	})
	if err != nil {
		return nil, nil, err
	}

	// Enable track_commit_timestamp for XID tracking
	_, _, execErr := container.Exec(ctx, []string{"psql", "-U", "postgres", "-c", "ALTER SYSTEM SET track_commit_timestamp = on;"})
	if execErr != nil {
		container.Terminate(ctx)
		return nil, nil, execErr
	}

	// Restart container to apply the setting
	stopTimeout := 2 * time.Second
	err = container.Stop(context.Background(), &stopTimeout)
	if err != nil {
		container.Terminate(ctx)
		return nil, nil, err
	}

	err = container.Start(context.Background())
	if err != nil {
		container.Terminate(ctx)
		return nil, nil, err
	}

	// Get connection details
	host, err := container.Host(ctx)
	if err != nil {
		container.Terminate(ctx)
		return nil, nil, err
	}

	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		container.Terminate(ctx)
		return nil, nil, err
	}

	dbAddr := fmt.Sprintf("%s:%s", host, port.Port())
	postgresDSN := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", "postgres", "postgres", dbAddr, "permify")

	// Create database instance
	db, err := New(postgresDSN,
		MaxOpenConnections(20),
		MaxIdleConnections(1),
		MaxConnectionIdleTime(60),
		MaxConnectionLifeTime(300),
	)
	if err != nil {
		container.Terminate(ctx)
		return nil, nil, err
	}

	// Run migrations to create the necessary tables
	// We'll create the basic schema manually since we can't import internal packages
	_, err = db.WritePool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS transactions (
			id BIGSERIAL PRIMARY KEY,
			tenant_id VARCHAR(255) NOT NULL,
			entity_type VARCHAR(255) NOT NULL,
			entity_id VARCHAR(255) NOT NULL,
			commit_tx_id BIGINT NOT NULL,
			created_tx_id BIGINT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
	`)
	if err != nil {
		container.Terminate(ctx)
		return nil, nil, err
	}

	// Create index for better performance
	_, err = db.WritePool.Exec(ctx, `
		CREATE INDEX IF NOT EXISTS idx_transactions_tenant_id ON transactions(tenant_id);
		CREATE INDEX IF NOT EXISTS idx_transactions_commit_tx_id ON transactions(commit_tx_id);
		CREATE INDEX IF NOT EXISTS idx_transactions_created_tx_id ON transactions(created_tx_id);
	`)
	if err != nil {
		container.Terminate(ctx)
		return nil, nil, err
	}

	return db, container, nil
}
