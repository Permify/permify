package memory

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/internal/storage/memory/migrations"
	"github.com/Permify/permify/pkg/database/memory"
)

var _ = Describe("TenantWriter", func() {
	var db *memory.Memory

	var tenantWriter *TenantWriter

	BeforeEach(func() {
		database, err := memory.New(migrations.Schema)
		Expect(err).ShouldNot(HaveOccurred())
		db = database

		tenantWriter = NewTenantWriter(db)
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
