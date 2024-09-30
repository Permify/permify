package postgres

import (
	"context"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/internal/storage/postgres/instance"
	"github.com/Permify/permify/pkg/database"
	PQDatabase "github.com/Permify/permify/pkg/database/postgres"
)

var _ = Describe("TenantReader", func() {
	var db database.Database
	var tenantWriter *TenantWriter
	var tenantReader *TenantReader

	BeforeEach(func() {
		version := os.Getenv("POSTGRES_VERSION")

		if version == "" {
			version = "14"
		}

		db = instance.PostgresDB(version)
		tenantWriter = NewTenantWriter(db.(*PQDatabase.Postgres))
		tenantReader = NewTenantReader(db.(*PQDatabase.Postgres))
	})

	AfterEach(func() {
		err := db.Close()
		Expect(err).ShouldNot(HaveOccurred())
	})

	Context("List Tenants", func() {
		It("should get tenants", func() {
			ctx := context.Background()

			_, err := tenantWriter.CreateTenant(ctx, "test_id_1", "test name 1")
			Expect(err).ShouldNot(HaveOccurred())

			_, err = tenantWriter.CreateTenant(ctx, "test_id_2", "test name 2")
			Expect(err).ShouldNot(HaveOccurred())

			_, err = tenantWriter.CreateTenant(ctx, "test_id_3", "test name 3")
			Expect(err).ShouldNot(HaveOccurred())

			_, err = tenantWriter.CreateTenant(ctx, "test_id_4", "test name 4")
			Expect(err).ShouldNot(HaveOccurred())

			_, err = tenantWriter.CreateTenant(ctx, "test_id_5", "test name 5")
			Expect(err).ShouldNot(HaveOccurred())

			_, err = tenantWriter.CreateTenant(ctx, "test_id_6", "test name 6")
			Expect(err).ShouldNot(HaveOccurred())

			col1, ct1, err := tenantReader.ListTenants(ctx, database.NewPagination(database.Size(3), database.Token("")))
			Expect(err).ShouldNot(HaveOccurred())
			Expect(len(col1)).Should(Equal(3))

			col2, ct2, err := tenantReader.ListTenants(ctx, database.NewPagination(database.Size(4), database.Token(ct1.String())))
			Expect(err).ShouldNot(HaveOccurred())
			Expect(len(col2)).Should(Equal(4))
			Expect(ct2.String()).Should(Equal(""))
		})
	})
})
