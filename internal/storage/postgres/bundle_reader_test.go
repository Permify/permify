package postgres

import (
	"context"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/pkg/database"
	PQDatabase "github.com/Permify/permify/pkg/database/postgres"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/testinstance"
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

		db = testinstance.PostgresDB(version)
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
		It("should handle protojson unmarshal error", func() {
			ctx := context.Background()

			// First, write a bundle with valid JSON that doesn't match the protobuf structure
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
	})
})
