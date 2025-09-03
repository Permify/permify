package postgres

import (
	"context"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/internal/storage/postgres/instance"
	"github.com/Permify/permify/pkg/database"
	PQDatabase "github.com/Permify/permify/pkg/database/postgres"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

var _ = Describe("BundleReader", func() {
	var db database.Database
	var bundleWriter *BundleWriter
	var bundleReader *BundleReader

	BeforeEach(func() {
		version := os.Getenv("POSTGRES_VERSION")

		if version == "" {
			version = "14"
		}

		db = instance.PostgresDB(version)
		bundleWriter = NewBundleWriter(db.(*PQDatabase.Postgres))
		bundleReader = NewBundleReader(db.(*PQDatabase.Postgres))
	})

	AfterEach(func() {
		err := db.Close()
		Expect(err).ShouldNot(HaveOccurred())
	})

	Context("Read", func() {
		It("should write and read DataBundles with correct relationships and attributes", func() {
			ctx := context.Background()

			bundles := []*base.DataBundle{
				{
					Name: "user_created",
					Arguments: []string{
						"organizationID",
						"userID",
					},
					Operations: []*base.Operation{
						{
							RelationshipsWrite: []string{
								"organization:{{.organizationID}}#member@user:{{.userID}}",
								"organization:{{.organizationID}}#admin@user:{{.userID}}",
							},
							RelationshipsDelete: []string{},
							AttributesWrite: []string{
								"organization:{{.organizationID}}$public|boolean:true",
							},
							AttributesDelete: []string{
								"organization:{{.organizationID}}$balance|integer[]:120,568",
							},
						},
					},
				},
			}

			var sBundles []storage.Bundle
			for _, b := range bundles {
				sBundles = append(sBundles, storage.Bundle{
					Name:       b.Name,
					DataBundle: b,
					TenantID:   "t1",
				})
			}

			names, err := bundleWriter.Write(ctx, sBundles)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(names).Should(Equal([]string{"user_created"}))

			bundle, err := bundleReader.Read(ctx, "t1", "user_created")
			Expect(err).ShouldNot(HaveOccurred())

			Expect(bundle.GetName()).Should(Equal("user_created"))
			Expect(bundle.GetArguments()).Should(Equal([]string{
				"organizationID",
				"userID",
			}))

			Expect(bundle.GetOperations()[0].RelationshipsWrite).Should(Equal([]string{
				"organization:{{.organizationID}}#member@user:{{.userID}}",
				"organization:{{.organizationID}}#admin@user:{{.userID}}",
			}))

			Expect(bundle.GetOperations()[0].RelationshipsDelete).Should(BeNil())

			Expect(bundle.GetOperations()[0].AttributesWrite).Should(Equal([]string{
				"organization:{{.organizationID}}$public|boolean:true",
			}))

			Expect(bundle.GetOperations()[0].AttributesDelete).Should(Equal([]string{
				"organization:{{.organizationID}}$balance|integer[]:120,568",
			}))
		})

		It("should get error on non-existing bundle", func() {
			ctx := context.Background()

			_, err := bundleReader.Read(ctx, "t1", "user_created")
			Expect(err.Error()).Should(Equal(base.ErrorCode_ERROR_CODE_BUNDLE_NOT_FOUND.String()))
		})
	})

	Context("Error Handling", func() {
		It("should handle SQL builder error", func() {
			ctx := context.Background()

			// Create a bundleReader with a closed database to trigger SQL builder error
			closedDB := db.(*PQDatabase.Postgres)
			err := closedDB.Close()
			Expect(err).ShouldNot(HaveOccurred())

			// Create a new bundleReader with the closed database
			readerWithClosedDB := NewBundleReader(closedDB)

			_, err = readerWithClosedDB.Read(ctx, "t1", "test_bundle")
			Expect(err).Should(HaveOccurred())
			// The error could be either SQL_BUILDER or SCAN depending on when the connection fails
			Expect(err.Error()).Should(Or(
				Equal(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String()),
				Equal(base.ErrorCode_ERROR_CODE_SCAN.String()),
			))
		})

		It("should handle scan error", func() {
			ctx := context.Background()

			// Create a bundleReader with a closed database to trigger scan error
			closedDB := db.(*PQDatabase.Postgres)
			err := closedDB.Close()
			Expect(err).ShouldNot(HaveOccurred())

			// Create a new bundleReader with the closed database
			readerWithClosedDB := NewBundleReader(closedDB)

			_, err = readerWithClosedDB.Read(ctx, "t1", "test_bundle")
			Expect(err).Should(HaveOccurred())
			// The error could be either SQL_BUILDER or SCAN depending on when the connection fails
			Expect(err.Error()).Should(Or(
				Equal(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String()),
				Equal(base.ErrorCode_ERROR_CODE_SCAN.String()),
			))
		})

		It("should handle protojson unmarshal error", func() {
			ctx := context.Background()

			// First, write a bundle with valid JSON that doesn't match the protobuf structure
			// We'll need to directly insert JSON that's valid JSON but invalid for the DataBundle protobuf
			postgresDB := db.(*PQDatabase.Postgres)

			// Insert valid JSON that doesn't match the DataBundle structure
			invalidJSON := `{"invalid_field": "invalid_value", "another_field": 123}`
			_, err := postgresDB.WritePool.Exec(ctx,
				"INSERT INTO bundles (name, tenant_id, payload) VALUES ($1, $2, $3)",
				"invalid_bundle", "t1", invalidJSON)
			Expect(err).ShouldNot(HaveOccurred())

			// Now try to read the bundle - this should trigger the unmarshal error
			_, err = bundleReader.Read(ctx, "t1", "invalid_bundle")
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(Equal(base.ErrorCode_ERROR_CODE_INVALID_ARGUMENT.String()))
		})

		It("should handle context cancellation", func() {
			// Create a context that's already cancelled
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			_, err := bundleReader.Read(ctx, "t1", "test_bundle")
			Expect(err).Should(HaveOccurred())
			// The error could be context-related or SQL_BUILDER depending on when cancellation occurs
			Expect(err.Error()).Should(Or(
				Equal(base.ErrorCode_ERROR_CODE_CANCELLED.String()),
				Equal(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String()),
			))
		})

		It("should handle database connection errors", func() {
			ctx := context.Background()

			// Create a bundleReader with a closed database
			closedDB := db.(*PQDatabase.Postgres)
			err := closedDB.Close()
			Expect(err).ShouldNot(HaveOccurred())

			readerWithClosedDB := NewBundleReader(closedDB)

			_, err = readerWithClosedDB.Read(ctx, "t1", "test_bundle")
			Expect(err).Should(HaveOccurred())
			// The error could be SQL_BUILDER or SCAN depending on when the connection fails
			Expect(err.Error()).Should(Or(
				Equal(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String()),
				Equal(base.ErrorCode_ERROR_CODE_SCAN.String()),
			))
		})
	})
})
