package gc

import (
	"context"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/Permify/permify/internal/storage/postgres/instance"
	"github.com/Permify/permify/pkg/database"
	PQDatabase "github.com/Permify/permify/pkg/database/postgres"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
)

var _ = Describe("GarbageCollector", func() {
	var db database.Database
	var ctx context.Context
	var schemaClient base.SchemaClient
	var permissionClient base.PermissionClient
	var dataClient base.DataClient
	var tenantClient base.TenancyClient
	var garbageCollector *GC

	BeforeEach(func() {
		ctx = context.Background()
		version := os.Getenv("POSTGRES_VERSION")
		if version == "" {
			version = "14"
		}

		db = instance.PostgresDB(version)
		garbageCollector = NewGC(
			db.(*PQDatabase.Postgres),
			Window(5*time.Second),
		)
		// Initialize the permission client
		conn, err := grpc.DialContext(ctx, "permify:3478", grpc.WithTransportCredentials(insecure.NewCredentials()))
		Expect(err).ShouldNot(HaveOccurred())

		// Create a PermissionClient using the connection.
		schemaClient = base.NewSchemaClient(conn)
		permissionClient = base.NewPermissionClient(conn)
		dataClient = base.NewDataClient(conn)
		tenantClient = base.NewTenancyClient(conn)
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
			_, err := tenantClient.Create(ctx, &base.TenantCreateRequest{
				Id:   tenantID,
				Name: tenantName,
			})
			Expect(err).ShouldNot(HaveOccurred())

			// Step 2: Write schema
			schema_resp, err := schemaClient.Write(ctx, &base.SchemaWriteRequest{
				TenantId: tenantID,
				Schema:   "entity user {}\n\nentity organisation {\n    relation admin @user\n    relation member @user\n    relation billing @user\n    relation guest @user\n\n    permission is_member = admin or member or billing or guest\n}\n\nentity billing {\n    relation admin @organisation#admin\n    relation billing @organisation#billing\n\n    permission is_member = admin or billing\n    permission full_access = is_member\n    permission read_only = is_member\n}\n\nentity project {\n    relation admin @organisation#admin\n    relation member @organisation#member\n\n    permission is_member = admin or member\n    permission full_access = admin\n    permission read_only = is_member\n}\n\nentity instance {\n    relation project @project\n\n    permission full_access = project.full_access\n    permission read_only = project.read_only\n}",
			})
			Expect(err).ShouldNot(HaveOccurred())
			schemaVersion := schema_resp.SchemaVersion

			// Step 3: Insert data
			tup1, err := tuple.Tuple("organisation:1#member@user:b56661f8-7be6-4342-a4c0-918ee04e5983")
			Expect(err).ShouldNot(HaveOccurred())

			_, err = dataClient.Write(ctx, &base.DataWriteRequest{
				TenantId: tenantID,
				Metadata: &base.DataWriteRequestMetadata{
					SchemaVersion: schemaVersion,
				},
				Tuples: []*base.Tuple{
					tup1,
				},
				Attributes: []*base.Attribute{},
			})
			Expect(err).ShouldNot(HaveOccurred())

			// Step 4: Perform a permission check (expected to be valid)
			checkRes1, err := permissionClient.Check(ctx, &base.PermissionCheckRequest{
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
			_, err = dataClient.Delete(ctx, &base.DataDeleteRequest{
				TenantId: tenantID,
				TupleFilter: &base.TupleFilter{
					Entity: &base.EntityFilter{
						Type: "organisation",
						Ids:  []string{"1"},
					}},
				AttributeFilter: &base.AttributeFilter{},
			})
			Expect(err).ShouldNot(HaveOccurred())

			// Step 6: Perform a permission check (expected to be invalid)
			checkRes2, err := permissionClient.Check(ctx, &base.PermissionCheckRequest{
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

			_, err = dataClient.Write(ctx, &base.DataWriteRequest{
				TenantId: tenantID,
				Metadata: &base.DataWriteRequestMetadata{
					SchemaVersion: schemaVersion,
				},
				Tuples: []*base.Tuple{
					tup2,
				},
				Attributes: []*base.Attribute{},
			})
			Expect(err).ShouldNot(HaveOccurred())

			// Step 8: Perform a permission check (expected to be valid)
			checkRes3, err := permissionClient.Check(ctx, &base.PermissionCheckRequest{
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
			checkRes4, err := permissionClient.Check(ctx, &base.PermissionCheckRequest{
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
	})
})
