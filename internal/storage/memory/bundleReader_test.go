package memory

import (
	"context"

	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/internal/storage/memory/migrations"
	"github.com/Permify/permify/pkg/database/memory"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	base "github.com/Permify/permify/pkg/pb/base/v1"
)

var _ = Describe("BundleReader", func() {
	var db *memory.Memory
	var bundleWriter *BundleWriter
	var bundleReader *BundleReader

	BeforeEach(func() {
		database, err := memory.New(migrations.Schema)
		Expect(err).ShouldNot(HaveOccurred())
		db = database
		bundleWriter = NewBundleWriter(db)
		bundleReader = NewBundleReader(db)
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

			Expect(bundle.GetOperations()[0].RelationshipsDelete).Should(BeEmpty())

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
})
