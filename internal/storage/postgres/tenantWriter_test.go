package postgres

import (
	"context"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/internal/storage/postgres/instance"
	"github.com/Permify/permify/pkg/database"
	PQDatabase "github.com/Permify/permify/pkg/database/postgres"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

var _ = Describe("TenantWriter", func() {
	var db database.Database
	var tenantWriter *TenantWriter

	BeforeEach(func() {
		version := os.Getenv("POSTGRES_VERSION")

		if version == "" {
			version = "14"
		}

		db = instance.PostgresDB(version)
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
	})

	Context("Delete Tenant", func() {
		It("should delete tenant", func() {
			ctx := context.Background()

			tenant, err := tenantWriter.CreateTenant(ctx, "test_id_1", "test name 1")
			Expect(err).ShouldNot(HaveOccurred())

			Expect(tenant.Id).Should(Equal("test_id_1"))
			Expect(tenant.Name).Should(Equal("test name 1"))

			tenant, err = tenantWriter.DeleteTenant(ctx, "test_id_1")
			Expect(err).ShouldNot(HaveOccurred())

			Expect(tenant.Id).Should(Equal("test_id_1"))
			Expect(tenant.Name).Should(Equal("test name 1"))
		})
	})
})
