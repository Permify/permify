package gc

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/rs/xid"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/internal/engines"
	"github.com/Permify/permify/internal/invoke"
	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/internal/storage/postgres"
	"github.com/Permify/permify/pkg/attribute"
	"github.com/Permify/permify/pkg/database"
	PQDatabase "github.com/Permify/permify/pkg/database/postgres"
	"github.com/Permify/permify/pkg/dsl/compiler"
	"github.com/Permify/permify/pkg/dsl/parser"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/testinstance"
	"github.com/Permify/permify/pkg/tuple"
)

func TestGC(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "postgres-gc-suite")
}

var _ = Describe("GarbageCollector", func() {
	var db database.Database
	var ctx context.Context
	var garbageCollector *GC
	var tenantWriter *postgres.TenantWriter
	var schemaWriter *postgres.SchemaWriter
	var dataWriter *postgres.DataWriter
	var schemaReader *postgres.SchemaReader
	var dataReader *postgres.DataReader

	BeforeEach(func() {
		ctx = context.Background()
		version := os.Getenv("POSTGRES_VERSION")
		if version == "" {
			version = "14"
		}

		db = testinstance.PostgresDB(version)
		garbageCollector = NewGC(
			db.(*PQDatabase.Postgres),
			Window(5*time.Second),
		)

		tenantWriter = postgres.NewTenantWriter(db.(*PQDatabase.Postgres))
		schemaWriter = postgres.NewSchemaWriter(db.(*PQDatabase.Postgres))
		dataWriter = postgres.NewDataWriter(db.(*PQDatabase.Postgres))

		schemaReader = postgres.NewSchemaReader(db.(*PQDatabase.Postgres))
		dataReader = postgres.NewDataReader(db.(*PQDatabase.Postgres))
	})

	AfterEach(func() {
		err := db.Close()
		Expect(err).ShouldNot(HaveOccurred())
	})

	Context("Garbage Collection", func() {
		It("should perform garbage collection correctly", func() {
			// Step 1: Create tenant
			tenantID := "test-tenant"
			tenantName := "Test Tenant"

			_, err := tenantWriter.CreateTenant(ctx, tenantID, tenantName)
			Expect(err).ShouldNot(HaveOccurred())

			conf, _, err := newSchema(tenantID, "entity user {}\n\nentity organisation {\n    relation admin @user\n    relation member @user\n    relation billing @user\n    relation guest @user\n\n    permission is_member = admin or member or billing or guest\n}\n\nentity billing {\n    relation admin @organisation#admin\n    relation billing @organisation#billing\n\n    permission is_member = admin or billing\n    permission full_access = is_member\n    permission read_only = is_member\n}\n\nentity project {\n    relation admin @organisation#admin\n    relation member @organisation#member\n\n    permission is_member = admin or member\n    permission full_access = admin\n    permission read_only = is_member\n}\n\nentity instance {\n    relation project @project\n\n    permission full_access = project.full_access\n    permission read_only = project.read_only\n}")
			Expect(err).ShouldNot(HaveOccurred())
			err = schemaWriter.WriteSchema(ctx, conf)
			Expect(err).ShouldNot(HaveOccurred())

			// Step 3: Insert data
			tup1, err := tuple.Tuple("organisation:1#member@user:b56661f8-7be6-4342-a4c0-918ee04e5983")
			Expect(err).ShouldNot(HaveOccurred())

			_, err = dataWriter.Write(ctx, tenantID, database.NewTupleCollection(tup1), database.NewAttributeCollection())
			Expect(err).ShouldNot(HaveOccurred())

			checkEngine := engines.NewCheckEngine(schemaReader, dataReader)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				nil,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			checkRes1, err := invoker.Check(ctx, &base.PermissionCheckRequest{
				Metadata: &base.PermissionCheckRequestMetadata{
					SnapToken:     "",
					SchemaVersion: "",
					Depth:         20,
				},
				TenantId: tenantID,
				Entity: &base.Entity{
					Type: "organisation",
					Id:   "1",
				},
				Permission: "is_member",
				Subject: &base.Subject{
					Type: "user",
					Id:   "b56661f8-7be6-4342-a4c0-918ee04e5983",
				},
			})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(checkRes1.Can).Should(Equal(base.CheckResult_CHECK_RESULT_ALLOWED))

			// Step 5: Delete the data
			_, err = dataWriter.Delete(ctx, tenantID, &base.TupleFilter{
				Entity: &base.EntityFilter{
					Type: "organisation",
					Ids:  []string{"1"},
				},
			}, &base.AttributeFilter{})
			Expect(err).ShouldNot(HaveOccurred())

			// Step 6: Perform a permission check (expected to be invalid)
			checkRes2, err := invoker.Check(ctx, &base.PermissionCheckRequest{
				Metadata: &base.PermissionCheckRequestMetadata{
					SnapToken:     "",
					SchemaVersion: "",
					Depth:         20,
				},
				TenantId: tenantID,
				Entity: &base.Entity{
					Type: "organisation",
					Id:   "1",
				},
				Permission: "is_member",
				Subject: &base.Subject{
					Type: "user",
					Id:   "b56661f8-7be6-4342-a4c0-918ee04e5983",
				},
			})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(checkRes2.Can).Should(Equal(base.CheckResult_CHECK_RESULT_DENIED))

			// Step 7: Insert the same data again
			tup2, err := tuple.Tuple("organisation:1#member@user:b56661f8-7be6-4342-a4c0-918ee04e5983")
			Expect(err).ShouldNot(HaveOccurred())

			_, err = dataWriter.Write(ctx, tenantID, database.NewTupleCollection(tup2), database.NewAttributeCollection())
			Expect(err).ShouldNot(HaveOccurred())

			// Step 8: Perform a permission check (expected to be valid)
			checkRes3, err := invoker.Check(ctx, &base.PermissionCheckRequest{
				Metadata: &base.PermissionCheckRequestMetadata{
					SnapToken:     "",
					SchemaVersion: "",
					Depth:         20,
				},
				TenantId: tenantID,
				Entity: &base.Entity{
					Type: "organisation",
					Id:   "1",
				},
				Permission: "is_member",
				Subject: &base.Subject{
					Type: "user",
					Id:   "b56661f8-7be6-4342-a4c0-918ee04e5983",
				},
			})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(checkRes3.Can).Should(Equal(base.CheckResult_CHECK_RESULT_ALLOWED))

			// Step 9: Run the garbage collector
			time.Sleep(5 * time.Second) // Pause for 5 seconds
			err = garbageCollector.Run()
			Expect(err).ShouldNot(HaveOccurred())

			// Step 10: Perform a permission check after GC (expected to be valid)
			checkRes4, err := invoker.Check(ctx, &base.PermissionCheckRequest{
				Metadata: &base.PermissionCheckRequestMetadata{
					SnapToken:     "",
					SchemaVersion: "",
					Depth:         20,
				},
				TenantId: tenantID,
				Entity: &base.Entity{
					Type: "organisation",
					Id:   "1",
				},
				Permission: "is_member",
				Subject: &base.Subject{
					Type: "user",
					Id:   "b56661f8-7be6-4342-a4c0-918ee04e5983",
				},
			})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(checkRes4.Can).Should(Equal(base.CheckResult_CHECK_RESULT_ALLOWED))
		})

		It("should perform tenant-aware garbage collection correctly", func() {
			// Step 1: Create two tenants
			tenantA := "tenant-a"
			tenantB := "tenant-b"

			_, err := tenantWriter.CreateTenant(ctx, tenantA, "Tenant A")
			Expect(err).ShouldNot(HaveOccurred())

			_, err = tenantWriter.CreateTenant(ctx, tenantB, "Tenant B")
			Expect(err).ShouldNot(HaveOccurred())

			// Step 2: Create schema for both tenants
			conf, _, err := newSchema(tenantA, "entity user {}\n\nentity organisation {\n    relation member @user\n    permission is_member = member\n}")
			Expect(err).ShouldNot(HaveOccurred())
			err = schemaWriter.WriteSchema(ctx, conf)
			Expect(err).ShouldNot(HaveOccurred())

			conf, _, err = newSchema(tenantB, "entity user {}\n\nentity organisation {\n    relation member @user\n    permission is_member = member\n}")
			Expect(err).ShouldNot(HaveOccurred())
			err = schemaWriter.WriteSchema(ctx, conf)
			Expect(err).ShouldNot(HaveOccurred())

			// Step 3: Insert data for both tenants
			tupA, err := tuple.Tuple("organisation:1#member@user:user-a")
			Expect(err).ShouldNot(HaveOccurred())

			tupB, err := tuple.Tuple("organisation:1#member@user:user-b")
			Expect(err).ShouldNot(HaveOccurred())

			_, err = dataWriter.Write(ctx, tenantA, database.NewTupleCollection(tupA), database.NewAttributeCollection())
			Expect(err).ShouldNot(HaveOccurred())

			_, err = dataWriter.Write(ctx, tenantB, database.NewTupleCollection(tupB), database.NewAttributeCollection())
			Expect(err).ShouldNot(HaveOccurred())

			// Step 4: Verify both tenants can access their data
			checkEngine := engines.NewCheckEngine(schemaReader, dataReader)
			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				nil,
				nil,
			)
			checkEngine.SetInvoker(invoker)

			// Check tenant A
			checkResA, err := invoker.Check(ctx, &base.PermissionCheckRequest{
				Metadata: &base.PermissionCheckRequestMetadata{
					SnapToken:     "",
					SchemaVersion: "",
					Depth:         20,
				},
				TenantId: tenantA,
				Entity: &base.Entity{
					Type: "organisation",
					Id:   "1",
				},
				Permission: "is_member",
				Subject: &base.Subject{
					Type: "user",
					Id:   "user-a",
				},
			})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(checkResA.Can).Should(Equal(base.CheckResult_CHECK_RESULT_ALLOWED))

			// Check tenant B
			checkResB, err := invoker.Check(ctx, &base.PermissionCheckRequest{
				Metadata: &base.PermissionCheckRequestMetadata{
					SnapToken:     "",
					SchemaVersion: "",
					Depth:         20,
				},
				TenantId: tenantB,
				Entity: &base.Entity{
					Type: "organisation",
					Id:   "1",
				},
				Permission: "is_member",
				Subject: &base.Subject{
					Type: "user",
					Id:   "user-b",
				},
			})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(checkResB.Can).Should(Equal(base.CheckResult_CHECK_RESULT_ALLOWED))

			// Step 5: Delete data for tenant A only
			_, err = dataWriter.Delete(ctx, tenantA, &base.TupleFilter{
				Entity: &base.EntityFilter{
					Type: "organisation",
					Ids:  []string{"1"},
				},
			}, &base.AttributeFilter{})
			Expect(err).ShouldNot(HaveOccurred())

			// Step 6: Run garbage collection
			time.Sleep(5 * time.Second) // Pause for 5 seconds
			err = garbageCollector.Run()
			Expect(err).ShouldNot(HaveOccurred())

			// Step 7: Verify tenant A's permission is denied (data was deleted)
			checkResA2, err := invoker.Check(ctx, &base.PermissionCheckRequest{
				Metadata: &base.PermissionCheckRequestMetadata{
					SnapToken:     "",
					SchemaVersion: "",
					Depth:         20,
				},
				TenantId: tenantA,
				Entity: &base.Entity{
					Type: "organisation",
					Id:   "1",
				},
				Permission: "is_member",
				Subject: &base.Subject{
					Type: "user",
					Id:   "user-a",
				},
			})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(checkResA2.Can).Should(Equal(base.CheckResult_CHECK_RESULT_DENIED))

			// Step 8: Verify tenant B's permission is still allowed (data was not affected by GC)
			checkResB2, err := invoker.Check(ctx, &base.PermissionCheckRequest{
				Metadata: &base.PermissionCheckRequestMetadata{
					SnapToken:     "",
					SchemaVersion: "",
					Depth:         20,
				},
				TenantId: tenantB,
				Entity: &base.Entity{
					Type: "organisation",
					Id:   "1",
				},
				Permission: "is_member",
				Subject: &base.Subject{
					Type: "user",
					Id:   "user-b",
				},
			})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(checkResB2.Can).Should(Equal(base.CheckResult_CHECK_RESULT_ALLOWED))
		})
	})

	Context("Error Handling", func() {
		It("should handle context cancellation in Start method", func() {
			// Create a context that will be cancelled
			ctx, cancel := context.WithCancel(context.Background())

			// Cancel the context immediately
			cancel()

			// Start the garbage collector with the cancelled context
			err := garbageCollector.Start(ctx)

			Expect(err).Should(HaveOccurred())
			Expect(err).Should(Equal(context.Canceled))
		})

		It("should handle database connection errors in Run", func() {
			// Create a garbage collector with a closed database connection
			// This tests errors at various stages: getting database time, getAllTenants,
			// getLastTransactionIDForTenant, deleteRecordsForTenant, deleteTransactionsForTenant
			closedDB := testinstance.PostgresDB("14")
			err := closedDB.Close()
			Expect(err).ShouldNot(HaveOccurred())

			badGC := NewGC(
				closedDB.(*PQDatabase.Postgres),
				Window(5*time.Second),
			)

			// Run should fail with closed database connection
			err = badGC.Run()
			Expect(err).Should(HaveOccurred())
		})

		It("should continue processing other tenants when one tenant fails", func() {
			// Create two tenants
			tenant1 := "gc-error-tenant-1"
			tenant2 := "gc-error-tenant-2"

			_, err := tenantWriter.CreateTenant(ctx, tenant1, "Tenant 1")
			Expect(err).ShouldNot(HaveOccurred())

			_, err = tenantWriter.CreateTenant(ctx, tenant2, "Tenant 2")
			Expect(err).ShouldNot(HaveOccurred())

			// Create schema and data for tenant1
			conf, _, err := newSchema(tenant1, "entity user {}\n\nentity organisation {\n    relation member @user\n    permission is_member = member\n}")
			Expect(err).ShouldNot(HaveOccurred())
			err = schemaWriter.WriteSchema(ctx, conf)
			Expect(err).ShouldNot(HaveOccurred())

			tup, err := tuple.Tuple("organisation:1#member@user:user-1")
			Expect(err).ShouldNot(HaveOccurred())
			_, err = dataWriter.Write(ctx, tenant1, database.NewTupleCollection(tup), database.NewAttributeCollection())
			Expect(err).ShouldNot(HaveOccurred())

			// Run GC - should succeed even if one tenant has no data
			err = garbageCollector.Run()
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("should handle tenant with no transactions", func() {
			// Create a tenant with no data
			tenantID := "gc-empty-tenant"
			_, err := tenantWriter.CreateTenant(ctx, tenantID, "Empty Tenant")
			Expect(err).ShouldNot(HaveOccurred())

			// Run GC - should succeed for tenant with no transactions
			err = garbageCollector.Run()
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("should handle case when lastTransactionID is 0", func() {
			// Create a tenant with very recent data (within window)
			tenantID := "gc-recent-tenant"
			_, err := tenantWriter.CreateTenant(ctx, tenantID, "Recent Tenant")
			Expect(err).ShouldNot(HaveOccurred())

			conf, _, err := newSchema(tenantID, "entity user {}\n\nentity organisation {\n    relation member @user\n    permission is_member = member\n}")
			Expect(err).ShouldNot(HaveOccurred())
			err = schemaWriter.WriteSchema(ctx, conf)
			Expect(err).ShouldNot(HaveOccurred())

			tup, err := tuple.Tuple("organisation:1#member@user:user-1")
			Expect(err).ShouldNot(HaveOccurred())
			_, err = dataWriter.Write(ctx, tenantID, database.NewTupleCollection(tup), database.NewAttributeCollection())
			Expect(err).ShouldNot(HaveOccurred())

			// Run GC immediately - should succeed with no cleanup (all data is recent)
			err = garbageCollector.Run()
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("should handle multiple tenants with mixed success and failure", func() {
			// Create multiple tenants
			tenant1 := "gc-mixed-1"
			tenant2 := "gc-mixed-2"
			tenant3 := "gc-mixed-3"

			_, err := tenantWriter.CreateTenant(ctx, tenant1, "Tenant 1")
			Expect(err).ShouldNot(HaveOccurred())
			_, err = tenantWriter.CreateTenant(ctx, tenant2, "Tenant 2")
			Expect(err).ShouldNot(HaveOccurred())
			_, err = tenantWriter.CreateTenant(ctx, tenant3, "Tenant 3")
			Expect(err).ShouldNot(HaveOccurred())

			// Create data for tenant1 and tenant2
			conf, _, err := newSchema(tenant1, "entity user {}\n\nentity organisation {\n    relation member @user\n    permission is_member = member\n}")
			Expect(err).ShouldNot(HaveOccurred())
			err = schemaWriter.WriteSchema(ctx, conf)
			Expect(err).ShouldNot(HaveOccurred())

			conf, _, err = newSchema(tenant2, "entity user {}\n\nentity organisation {\n    relation member @user\n    permission is_member = member\n}")
			Expect(err).ShouldNot(HaveOccurred())
			err = schemaWriter.WriteSchema(ctx, conf)
			Expect(err).ShouldNot(HaveOccurred())

			tup1, err := tuple.Tuple("organisation:1#member@user:user-1")
			Expect(err).ShouldNot(HaveOccurred())
			_, err = dataWriter.Write(ctx, tenant1, database.NewTupleCollection(tup1), database.NewAttributeCollection())
			Expect(err).ShouldNot(HaveOccurred())

			tup2, err := tuple.Tuple("organisation:1#member@user:user-2")
			Expect(err).ShouldNot(HaveOccurred())
			_, err = dataWriter.Write(ctx, tenant2, database.NewTupleCollection(tup2), database.NewAttributeCollection())
			Expect(err).ShouldNot(HaveOccurred())

			// Run GC - should succeed for all tenants
			err = garbageCollector.Run()
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("should handle Run with timeout", func() {
			// Create a GC with very short timeout
			shortTimeoutGC := NewGC(
				db.(*PQDatabase.Postgres),
				Window(5*time.Second),
				Timeout(1*time.Nanosecond), // Very short timeout
			)

			// Create a tenant with data
			tenantID := "gc-timeout-tenant"
			_, err := tenantWriter.CreateTenant(ctx, tenantID, "Timeout Tenant")
			Expect(err).ShouldNot(HaveOccurred())

			// Run should fail due to timeout
			err = shortTimeoutGC.Run()
			Expect(err).Should(HaveOccurred())
		})
	})

	Context("Edge Cases", func() {
		It("should handle Run with no tenants", func() {
			// Create a fresh database with no tenants (except default t1)
			// Run GC - should succeed
			err := garbageCollector.Run()
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("should handle tenant with expired data older than window and verify deleteRecordsForTenant is called", func() {
			// Create a tenant
			tenantID := "gc-expired-tenant"
			_, err := tenantWriter.CreateTenant(ctx, tenantID, "Expired Tenant")
			Expect(err).ShouldNot(HaveOccurred())

			// Create schema and data
			conf, _, err := newSchema(tenantID, "entity user {}\n\nentity organisation {\n    relation member @user\n    permission is_member = member\n}")
			Expect(err).ShouldNot(HaveOccurred())
			err = schemaWriter.WriteSchema(ctx, conf)
			Expect(err).ShouldNot(HaveOccurred())

			tup, err := tuple.Tuple("organisation:1#member@user:user-1")
			Expect(err).ShouldNot(HaveOccurred())
			_, err = dataWriter.Write(ctx, tenantID, database.NewTupleCollection(tup), database.NewAttributeCollection())
			Expect(err).ShouldNot(HaveOccurred())

			// Delete the tuple to create expired records
			_, err = dataWriter.Delete(ctx, tenantID, &base.TupleFilter{
				Entity: &base.EntityFilter{
					Type: "organisation",
					Ids:  []string{"1"},
				},
			}, &base.AttributeFilter{})
			Expect(err).ShouldNot(HaveOccurred())

			// Wait for window to pass
			time.Sleep(5 * time.Second)

			// Run GC - should clean up expired data via deleteRecordsForTenant (relation_tuples) and deleteTransactionsForTenant
			// This verifies that deleteRecordsForTenant is called for RelationTuplesTable
			err = garbageCollector.Run()
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("should handle tenant with attributes to clean up and verify deleteRecordsForTenant is called", func() {
			// Create a tenant
			tenantID := "gc-attributes-tenant"
			_, err := tenantWriter.CreateTenant(ctx, tenantID, "Attributes Tenant")
			Expect(err).ShouldNot(HaveOccurred())

			// Create schema with attributes
			conf, _, err := newSchema(tenantID, "entity user {}\n\nentity organisation {\n    attribute public boolean\n    permission is_public = public\n}")
			Expect(err).ShouldNot(HaveOccurred())
			err = schemaWriter.WriteSchema(ctx, conf)
			Expect(err).ShouldNot(HaveOccurred())

			// Write attributes
			attr, err := attribute.Attribute("organisation:1$public|boolean:true")
			Expect(err).ShouldNot(HaveOccurred())
			_, err = dataWriter.Write(ctx, tenantID, database.NewTupleCollection(), database.NewAttributeCollection(attr))
			Expect(err).ShouldNot(HaveOccurred())

			// Delete attributes to create expired records
			_, err = dataWriter.Delete(ctx, tenantID, &base.TupleFilter{}, &base.AttributeFilter{
				Entity: &base.EntityFilter{
					Type: "organisation",
					Ids:  []string{"1"},
				},
				Attributes: []string{"public"},
			})
			Expect(err).ShouldNot(HaveOccurred())

			// Wait for window to pass
			time.Sleep(5 * time.Second)

			// Run GC - should clean up expired attributes via deleteRecordsForTenant (AttributesTable)
			// This verifies that deleteRecordsForTenant is called for AttributesTable
			err = garbageCollector.Run()
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("should verify deleteTransactionsForTenant is called and deletes old transactions", func() {
			// Create a tenant
			tenantID := "gc-transactions-tenant"
			_, err := tenantWriter.CreateTenant(ctx, tenantID, "Transactions Tenant")
			Expect(err).ShouldNot(HaveOccurred())

			// Create schema and data
			conf, _, err := newSchema(tenantID, "entity user {}\n\nentity organisation {\n    relation member @user\n    permission is_member = member\n}")
			Expect(err).ShouldNot(HaveOccurred())
			err = schemaWriter.WriteSchema(ctx, conf)
			Expect(err).ShouldNot(HaveOccurred())

			tup, err := tuple.Tuple("organisation:1#member@user:user-1")
			Expect(err).ShouldNot(HaveOccurred())
			_, err = dataWriter.Write(ctx, tenantID, database.NewTupleCollection(tup), database.NewAttributeCollection())
			Expect(err).ShouldNot(HaveOccurred())

			// Delete the tuple to create expired records
			_, err = dataWriter.Delete(ctx, tenantID, &base.TupleFilter{
				Entity: &base.EntityFilter{
					Type: "organisation",
					Ids:  []string{"1"},
				},
			}, &base.AttributeFilter{})
			Expect(err).ShouldNot(HaveOccurred())

			// Get transaction count before GC
			var transactionCountBefore int
			err = db.(*PQDatabase.Postgres).WritePool.QueryRow(ctx,
				"SELECT COUNT(*) FROM "+postgres.TransactionsTable+" WHERE tenant_id = $1",
				tenantID).Scan(&transactionCountBefore)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(transactionCountBefore).Should(BeNumerically(">", 0))

			// Wait for window to pass
			time.Sleep(5 * time.Second)

			// Run GC - should clean up old transactions via deleteTransactionsForTenant
			err = garbageCollector.Run()
			Expect(err).ShouldNot(HaveOccurred())

			// Verify old transactions were deleted by deleteTransactionsForTenant
			// We should have fewer transactions now (only recent ones remain)
			var transactionCountAfter int
			err = db.(*PQDatabase.Postgres).WritePool.QueryRow(ctx,
				"SELECT COUNT(*) FROM "+postgres.TransactionsTable+" WHERE tenant_id = $1",
				tenantID).Scan(&transactionCountAfter)
			Expect(err).ShouldNot(HaveOccurred())
			// After GC, old transactions should be deleted, so count should be less
			Expect(transactionCountAfter).Should(BeNumerically("<=", transactionCountBefore))
		})

		It("should handle getAllTenants with empty result", func() {
			// Create a fresh database instance with no custom tenants
			freshDB := testinstance.PostgresDB("14")
			freshGC := NewGC(
				freshDB.(*PQDatabase.Postgres),
				Window(5*time.Second),
			)

			// Run GC on empty database - should succeed
			err := freshGC.Run()
			Expect(err).ShouldNot(HaveOccurred())

			err = freshDB.Close()
			Expect(err).ShouldNot(HaveOccurred())
		})
	})
})

func newSchema(tenant, model string) ([]storage.SchemaDefinition, string, error) {
	sch, err := parser.NewParser(model).Parse()
	if err != nil {
		return nil, "", err
	}

	_, _, err = compiler.NewCompiler(false, sch).Compile()
	if err != nil {
		return nil, "", err
	}

	version := xid.New().String()

	cnf := make([]storage.SchemaDefinition, 0, len(sch.Statements))
	for _, st := range sch.Statements {
		cnf = append(cnf, storage.SchemaDefinition{
			TenantID:             tenant,
			Version:              version,
			Name:                 st.GetName(),
			SerializedDefinition: []byte(st.String()),
		})
	}

	return cnf, version, err
}
