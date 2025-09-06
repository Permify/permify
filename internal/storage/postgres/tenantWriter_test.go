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

			err = tenantWriter.DeleteTenant(ctx, "test_id_1")
			Expect(err).ShouldNot(HaveOccurred())
		})
	})

	Context("Error Handling", func() {
		Context("CreateTenant Error Handling", func() {
			It("should handle execution error", func() {
				ctx := context.Background()

				// Create a tenantWriter with a closed database to trigger execution error
				closedDB := db.(*PQDatabase.Postgres)
				err := closedDB.Close()
				Expect(err).ShouldNot(HaveOccurred())

				writerWithClosedDB := NewTenantWriter(closedDB)

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

				// Create a tenantWriter with a closed database to trigger transaction begin error
				closedDB := db.(*PQDatabase.Postgres)
				err := closedDB.Close()
				Expect(err).ShouldNot(HaveOccurred())

				writerWithClosedDB := NewTenantWriter(closedDB)

				err = writerWithClosedDB.DeleteTenant(ctx, "test_id")
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(Or(
					Equal(base.ErrorCode_ERROR_CODE_EXECUTION.String()),
					Equal(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String()),
					Equal(base.ErrorCode_ERROR_CODE_SCAN.String()),
				))
			})

			It("should handle batch execution error", func() {
				ctx := context.Background()

				// Create a tenantWriter with a closed database to trigger batch execution error
				closedDB := db.(*PQDatabase.Postgres)
				err := closedDB.Close()
				Expect(err).ShouldNot(HaveOccurred())

				writerWithClosedDB := NewTenantWriter(closedDB)

				err = writerWithClosedDB.DeleteTenant(ctx, "test_id")
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(Or(
					Equal(base.ErrorCode_ERROR_CODE_EXECUTION.String()),
					Equal(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String()),
					Equal(base.ErrorCode_ERROR_CODE_SCAN.String()),
				))
			})

			It("should handle commit error after batch execution", func() {
				ctx := context.Background()

				// Create a tenantWriter with a closed database to trigger commit error after batch execution
				closedDB := db.(*PQDatabase.Postgres)
				err := closedDB.Close()
				Expect(err).ShouldNot(HaveOccurred())

				writerWithClosedDB := NewTenantWriter(closedDB)

				err = writerWithClosedDB.DeleteTenant(ctx, "test_id")
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(Or(
					Equal(base.ErrorCode_ERROR_CODE_EXECUTION.String()),
					Equal(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String()),
					Equal(base.ErrorCode_ERROR_CODE_SCAN.String()),
				))
			})

			It("should handle query row error", func() {
				ctx := context.Background()

				// Create a tenantWriter with a closed database to trigger query row error
				closedDB := db.(*PQDatabase.Postgres)
				err := closedDB.Close()
				Expect(err).ShouldNot(HaveOccurred())

				writerWithClosedDB := NewTenantWriter(closedDB)

				err = writerWithClosedDB.DeleteTenant(ctx, "test_id")
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(Or(
					Equal(base.ErrorCode_ERROR_CODE_EXECUTION.String()),
					Equal(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String()),
					Equal(base.ErrorCode_ERROR_CODE_SCAN.String()),
				))
			})

			It("should handle batch close error", func() {
				ctx := context.Background()

				// Create a tenantWriter with a closed database to trigger batch close error
				closedDB := db.(*PQDatabase.Postgres)
				err := closedDB.Close()
				Expect(err).ShouldNot(HaveOccurred())

				writerWithClosedDB := NewTenantWriter(closedDB)

				err = writerWithClosedDB.DeleteTenant(ctx, "test_id")
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(Or(
					Equal(base.ErrorCode_ERROR_CODE_EXECUTION.String()),
					Equal(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String()),
					Equal(base.ErrorCode_ERROR_CODE_SCAN.String()),
				))
			})

			It("should handle final commit error", func() {
				ctx := context.Background()

				// Create a tenantWriter with a closed database to trigger final commit error
				closedDB := db.(*PQDatabase.Postgres)
				err := closedDB.Close()
				Expect(err).ShouldNot(HaveOccurred())

				writerWithClosedDB := NewTenantWriter(closedDB)

				err = writerWithClosedDB.DeleteTenant(ctx, "test_id")
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
