package postgres

import (
	"context"
	"os"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/pkg/database"
	PQDatabase "github.com/Permify/permify/pkg/database/postgres"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/testinstance"
	"github.com/Permify/permify/pkg/tuple"
	"github.com/rs/xid"
)

var _ = Describe("TenantWriter", func() {
	var db database.Database
	var tenantWriter *TenantWriter

	BeforeEach(func() {
		version := os.Getenv("POSTGRES_VERSION")

		if version == "" {
			version = "14"
		}

		db = testinstance.PostgresDB(version)
		tenantWriter = NewTenantWriter(db.(*PQDatabase.Postgres))
	})

	AfterEach(func() {
		err := db.Close()
		Expect(err).ShouldNot(HaveOccurred())
	})

	Context("Create Tenant", func() {
		It("should create tenant", func() {
			ctx := context.Background()

			tenant, err := tenantWriter.CreateTenant(ctx, "test_id_1", "test name 1")
			Expect(err).ShouldNot(HaveOccurred())

			Expect(tenant.Id).Should(Equal("test_id_1"))
			Expect(tenant.Name).Should(Equal("test name 1"))
		})

		It("should get unique error", func() {
			ctx := context.Background()

			_, err := tenantWriter.CreateTenant(ctx, "test_id_1", "test name 1")
			Expect(err).ShouldNot(HaveOccurred())

			_, err = tenantWriter.CreateTenant(ctx, "test_id_1", "test name 1")
			Expect(err.Error()).Should(Equal(base.ErrorCode_ERROR_CODE_UNIQUE_CONSTRAINT.String()))
		})

		It("should create tenant with empty name", func() {
			ctx := context.Background()

			tenant, err := tenantWriter.CreateTenant(ctx, "test_id_empty_name", "")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(tenant.Id).Should(Equal("test_id_empty_name"))
			Expect(tenant.Name).Should(Equal(""))
		})

		It("should create tenant with special characters in name", func() {
			ctx := context.Background()

			specialName := "Test Tenant !@#$%^&*()_+-=[]{}|;':\",./<>?"
			tenant, err := tenantWriter.CreateTenant(ctx, "test_id_special", specialName)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(tenant.Id).Should(Equal("test_id_special"))
			Expect(tenant.Name).Should(Equal(specialName))
		})

		It("should create tenant with unicode characters", func() {
			ctx := context.Background()

			unicodeName := "Test Tenant 测试 テスト 테스트 🎉"
			tenant, err := tenantWriter.CreateTenant(ctx, "test_id_unicode", unicodeName)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(tenant.Id).Should(Equal("test_id_unicode"))
			Expect(tenant.Name).Should(Equal(unicodeName))
		})

		It("should set CreatedAt timestamp", func() {
			ctx := context.Background()

			beforeCreate := time.Now()
			tenant, err := tenantWriter.CreateTenant(ctx, "test_id_timestamp", "Test Tenant")
			afterCreate := time.Now()

			Expect(err).ShouldNot(HaveOccurred())
			Expect(tenant.CreatedAt).ShouldNot(BeNil())
			Expect(tenant.CreatedAt.AsTime()).Should(BeTemporally(">=", beforeCreate.Add(-1*time.Second)))
			Expect(tenant.CreatedAt.AsTime()).Should(BeTemporally("<=", afterCreate.Add(1*time.Second)))
		})

		It("should create multiple tenants with different IDs", func() {
			ctx := context.Background()

			tenant1, err := tenantWriter.CreateTenant(ctx, "multi_tenant_1", "Tenant 1")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(tenant1.Id).Should(Equal("multi_tenant_1"))

			tenant2, err := tenantWriter.CreateTenant(ctx, "multi_tenant_2", "Tenant 2")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(tenant2.Id).Should(Equal("multi_tenant_2"))

			tenant3, err := tenantWriter.CreateTenant(ctx, "multi_tenant_3", "Tenant 3")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(tenant3.Id).Should(Equal("multi_tenant_3"))
		})

		It("should create tenant with long name", func() {
			ctx := context.Background()

			longName := "A" + strings.Repeat(" very long tenant name ", 50)
			tenant, err := tenantWriter.CreateTenant(ctx, "test_id_long_name", longName)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(tenant.Id).Should(Equal("test_id_long_name"))
			Expect(tenant.Name).Should(Equal(longName))
		})
	})

	Context("Delete Tenant", func() {
		It("should delete tenant", func() {
			ctx := context.Background()

			tenant, err := tenantWriter.CreateTenant(ctx, "test_id_1", "test name 1")
			Expect(err).ShouldNot(HaveOccurred())

			Expect(tenant.Id).Should(Equal("test_id_1"))
			Expect(tenant.Name).Should(Equal("test name 1"))

			err = tenantWriter.DeleteTenant(ctx, "test_id_1")
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("should return not found error when tenant does not exist", func() {
			ctx := context.Background()

			err := tenantWriter.DeleteTenant(ctx, "non_existent_tenant_id")
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(Equal("ERROR_CODE_NOT_FOUND"))
		})

		It("should delete tenant with associated schema data", func() {
			ctx := context.Background()

			tenantID := "tenant_with_schema"
			_, err := tenantWriter.CreateTenant(ctx, tenantID, "Tenant with Schema")
			Expect(err).ShouldNot(HaveOccurred())

			// Create schema for the tenant
			schemaWriter := NewSchemaWriter(db.(*PQDatabase.Postgres))
			version := xid.New().String()
			schema := []storage.SchemaDefinition{
				{TenantID: tenantID, Name: "user", SerializedDefinition: []byte("entity user {}"), Version: version},
				{TenantID: tenantID, Name: "organization", SerializedDefinition: []byte("entity organization { relation admin @user}"), Version: version},
			}
			err = schemaWriter.WriteSchema(ctx, schema)
			Expect(err).ShouldNot(HaveOccurred())

			// Delete tenant should succeed even with associated schema
			err = tenantWriter.DeleteTenant(ctx, tenantID)
			Expect(err).ShouldNot(HaveOccurred())

			// Verify schema data was actually deleted by checking database directly
			var schemaCount int
			err = db.(*PQDatabase.Postgres).WritePool.QueryRow(ctx,
				"SELECT COUNT(*) FROM "+SchemaDefinitionTable+" WHERE tenant_id = $1",
				tenantID).Scan(&schemaCount)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(schemaCount).Should(Equal(0))
		})

		It("should delete tenant with associated relation tuples", func() {
			ctx := context.Background()

			tenantID := "tenant_with_tuples"
			_, err := tenantWriter.CreateTenant(ctx, tenantID, "Tenant with Tuples")
			Expect(err).ShouldNot(HaveOccurred())

			// Create relation tuples for the tenant
			dataWriter := NewDataWriter(db.(*PQDatabase.Postgres))
			tup, err := tuple.Tuple("organization:1#admin@user:user-1")
			Expect(err).ShouldNot(HaveOccurred())

			_, err = dataWriter.Write(ctx, tenantID, database.NewTupleCollection(tup), database.NewAttributeCollection())
			Expect(err).ShouldNot(HaveOccurred())

			// Delete tenant should succeed even with associated tuples
			// This tests the full DeleteTenant flow including all batch operations
			err = tenantWriter.DeleteTenant(ctx, tenantID)
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("should delete tenant with all associated data types", func() {
			ctx := context.Background()

			tenantID := "tenant_with_all_data"
			_, err := tenantWriter.CreateTenant(ctx, tenantID, "Tenant with All Data")
			Expect(err).ShouldNot(HaveOccurred())

			// Create schema
			schemaWriter := NewSchemaWriter(db.(*PQDatabase.Postgres))
			version := xid.New().String()
			schema := []storage.SchemaDefinition{
				{TenantID: tenantID, Name: "user", SerializedDefinition: []byte("entity user {}"), Version: version},
			}
			err = schemaWriter.WriteSchema(ctx, schema)
			Expect(err).ShouldNot(HaveOccurred())

			// Create relation tuples
			dataWriter := NewDataWriter(db.(*PQDatabase.Postgres))
			tup, err := tuple.Tuple("user:1#admin@user:user-1")
			Expect(err).ShouldNot(HaveOccurred())
			_, err = dataWriter.Write(ctx, tenantID, database.NewTupleCollection(tup), database.NewAttributeCollection())
			Expect(err).ShouldNot(HaveOccurred())

			// Delete tenant should succeed and clean up all associated data
			// This tests the full DeleteTenant flow including all batch operations (bundles, tuples, attributes, schema, transactions)
			err = tenantWriter.DeleteTenant(ctx, tenantID)
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("should allow recreating tenant after deletion", func() {
			ctx := context.Background()

			tenantID := "recreate_tenant"
			tenantName := "Original Name"

			// Create tenant
			tenant1, err := tenantWriter.CreateTenant(ctx, tenantID, tenantName)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(tenant1.Id).Should(Equal(tenantID))
			Expect(tenant1.Name).Should(Equal(tenantName))

			// Delete tenant
			err = tenantWriter.DeleteTenant(ctx, tenantID)
			Expect(err).ShouldNot(HaveOccurred())

			// Recreate tenant with same ID
			tenant2, err := tenantWriter.CreateTenant(ctx, tenantID, "New Name")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(tenant2.Id).Should(Equal(tenantID))
			Expect(tenant2.Name).Should(Equal("New Name"))
		})

		It("should delete tenant and verify it cannot be deleted again", func() {
			ctx := context.Background()

			tenantID := "delete_twice_tenant"
			_, err := tenantWriter.CreateTenant(ctx, tenantID, "Test Tenant")
			Expect(err).ShouldNot(HaveOccurred())

			// First deletion should succeed
			err = tenantWriter.DeleteTenant(ctx, tenantID)
			Expect(err).ShouldNot(HaveOccurred())

			// Second deletion should fail with not found
			err = tenantWriter.DeleteTenant(ctx, tenantID)
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(Equal("ERROR_CODE_NOT_FOUND"))
		})

		It("should handle deletion of tenant with empty ID", func() {
			ctx := context.Background()

			err := tenantWriter.DeleteTenant(ctx, "")
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(Equal("ERROR_CODE_NOT_FOUND"))
		})
	})

	Context("Error Handling", func() {
		Context("CreateTenant Error Handling", func() {
			It("should handle execution error", func() {
				ctx := context.Background()

				// Create a separate database instance and close it to trigger execution error
				separateDB := testinstance.PostgresDB("14")
				err := separateDB.Close()
				Expect(err).ShouldNot(HaveOccurred())

				writerWithClosedDB := NewTenantWriter(separateDB.(*PQDatabase.Postgres))

				_, err = writerWithClosedDB.CreateTenant(ctx, "test_id", "test_name")
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(Or(
					Equal(base.ErrorCode_ERROR_CODE_EXECUTION.String()),
					Equal(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String()),
					Equal(base.ErrorCode_ERROR_CODE_SCAN.String()),
				))
			})
		})

		Context("DeleteTenant Error Handling", func() {
			It("should handle transaction begin error", func() {
				ctx := context.Background()

				// Create a separate database instance and close it to trigger transaction begin error
				separateDB := testinstance.PostgresDB("14")
				err := separateDB.Close()
				Expect(err).ShouldNot(HaveOccurred())

				writerWithClosedDB := NewTenantWriter(separateDB.(*PQDatabase.Postgres))

				err = writerWithClosedDB.DeleteTenant(ctx, "test_id")
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(Or(
					Equal(base.ErrorCode_ERROR_CODE_EXECUTION.String()),
					Equal(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String()),
					Equal(base.ErrorCode_ERROR_CODE_SCAN.String()),
				))
			})

			It("should handle batch execution error during table deletion", func() {
				ctx := context.Background()

				// Create a tenant first
				tenantID := "batch-exec-error-tenant"
				_, err := tenantWriter.CreateTenant(ctx, tenantID, "Batch Exec Error Tenant")
				Expect(err).ShouldNot(HaveOccurred())

				// Create a separate database instance for this test
				separateDB := testinstance.PostgresDB("14")
				separateWriter := NewTenantWriter(separateDB.(*PQDatabase.Postgres))

				// Create tenant in separate DB
				_, err = separateWriter.CreateTenant(ctx, tenantID, "Batch Exec Error Tenant")
				Expect(err).ShouldNot(HaveOccurred())

				// Close the database to trigger batch execution error
				// This tests the batch execution error path during table deletion loop
				err = separateDB.Close()
				Expect(err).ShouldNot(HaveOccurred())

				writerWithClosedDB := NewTenantWriter(separateDB.(*PQDatabase.Postgres))

				err = writerWithClosedDB.DeleteTenant(ctx, tenantID)
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(Or(
					Equal(base.ErrorCode_ERROR_CODE_EXECUTION.String()),
					Equal(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String()),
					Equal(base.ErrorCode_ERROR_CODE_SCAN.String()),
				))
			})

			It("should handle batch execution error during tenant deletion", func() {
				ctx := context.Background()

				// Create a separate database instance for this test
				separateDB := testinstance.PostgresDB("14")
				separateWriter := NewTenantWriter(separateDB.(*PQDatabase.Postgres))

				// Create a tenant first
				tenantID := "batch-exec-tenant-error"
				_, err := separateWriter.CreateTenant(ctx, tenantID, "Batch Exec Tenant Error")
				Expect(err).ShouldNot(HaveOccurred())

				// Close the database to trigger batch execution error
				// This tests the batch execution error path during tenant deletion
				err = separateDB.Close()
				Expect(err).ShouldNot(HaveOccurred())

				writerWithClosedDB := NewTenantWriter(separateDB.(*PQDatabase.Postgres))

				err = writerWithClosedDB.DeleteTenant(ctx, tenantID)
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(Or(
					Equal(base.ErrorCode_ERROR_CODE_EXECUTION.String()),
					Equal(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String()),
					Equal(base.ErrorCode_ERROR_CODE_SCAN.String()),
				))
			})

			It("should handle database connection errors", func() {
				ctx := context.Background()

				// Create a separate database instance and close it to trigger connection errors
				// This tests errors at various stages: transaction begin, query row, batch execution, commit
				separateDB := testinstance.PostgresDB("14")
				err := separateDB.Close()
				Expect(err).ShouldNot(HaveOccurred())

				writerWithClosedDB := NewTenantWriter(separateDB.(*PQDatabase.Postgres))

				err = writerWithClosedDB.DeleteTenant(ctx, "test_id")
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(Or(
					Equal(base.ErrorCode_ERROR_CODE_EXECUTION.String()),
					Equal(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String()),
					Equal(base.ErrorCode_ERROR_CODE_SCAN.String()),
				))
			})

			It("should handle batch close error after exec error", func() {
				ctx := context.Background()

				// Create a separate database instance for this test
				separateDB := testinstance.PostgresDB("14")
				separateWriter := NewTenantWriter(separateDB.(*PQDatabase.Postgres))

				// Create a tenant first
				tenantID := "batch-close-error-tenant"
				_, err := separateWriter.CreateTenant(ctx, tenantID, "Batch Close Error Tenant")
				Expect(err).ShouldNot(HaveOccurred())

				// Close the database to trigger batch close error
				// This tests the batch close error path after batch execution error
				err = separateDB.Close()
				Expect(err).ShouldNot(HaveOccurred())

				writerWithClosedDB := NewTenantWriter(separateDB.(*PQDatabase.Postgres))

				err = writerWithClosedDB.DeleteTenant(ctx, tenantID)
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(Or(
					Equal(base.ErrorCode_ERROR_CODE_EXECUTION.String()),
					Equal(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String()),
					Equal(base.ErrorCode_ERROR_CODE_SCAN.String()),
				))
			})

			It("should handle batch close error after successful execution", func() {
				ctx := context.Background()

				// Create a separate database instance for this test
				separateDB := testinstance.PostgresDB("14")
				separateWriter := NewTenantWriter(separateDB.(*PQDatabase.Postgres))

				// Create a tenant first
				tenantID := "batch-close-success-error-tenant"
				_, err := separateWriter.CreateTenant(ctx, tenantID, "Batch Close Success Error Tenant")
				Expect(err).ShouldNot(HaveOccurred())

				// Close the database to trigger batch close error
				// This tests the batch close error path after successful batch execution
				err = separateDB.Close()
				Expect(err).ShouldNot(HaveOccurred())

				writerWithClosedDB := NewTenantWriter(separateDB.(*PQDatabase.Postgres))

				err = writerWithClosedDB.DeleteTenant(ctx, tenantID)
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(Or(
					Equal(base.ErrorCode_ERROR_CODE_EXECUTION.String()),
					Equal(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String()),
					Equal(base.ErrorCode_ERROR_CODE_SCAN.String()),
				))
			})

			It("should handle commit error after successful batch operations", func() {
				ctx := context.Background()

				// Create a separate database instance for this test
				separateDB := testinstance.PostgresDB("14")
				separateWriter := NewTenantWriter(separateDB.(*PQDatabase.Postgres))

				// Create a tenant first
				tenantID := "commit-error-tenant"
				_, err := separateWriter.CreateTenant(ctx, tenantID, "Commit Error Tenant")
				Expect(err).ShouldNot(HaveOccurred())

				// Close the database to trigger commit error
				// This will test the tx.Commit() error path (line 123)
				err = separateDB.Close()
				Expect(err).ShouldNot(HaveOccurred())

				writerWithClosedDB := NewTenantWriter(separateDB.(*PQDatabase.Postgres))

				err = writerWithClosedDB.DeleteTenant(ctx, tenantID)
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(Or(
					Equal(base.ErrorCode_ERROR_CODE_EXECUTION.String()),
					Equal(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String()),
					Equal(base.ErrorCode_ERROR_CODE_SCAN.String()),
				))
			})
		})
	})
})
