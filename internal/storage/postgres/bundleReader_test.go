package postgres

import (
	"context"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

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

		db = postgresDB(version)
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

			names, err := bundleWriter.Write(ctx, "t1", bundles)

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
	})
})
