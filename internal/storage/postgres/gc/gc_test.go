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
	"github.com/Permify/permify/internal/storage/postgres/instance"
	"github.com/Permify/permify/pkg/database"
	PQDatabase "github.com/Permify/permify/pkg/database/postgres"
	"github.com/Permify/permify/pkg/dsl/compiler"
	"github.com/Permify/permify/pkg/dsl/parser"
	base "github.com/Permify/permify/pkg/pb/base/v1"
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

		db = instance.PostgresDB(version)
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
