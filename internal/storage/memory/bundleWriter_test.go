package memory

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/internal/storage/memory/migrations"
	"github.com/Permify/permify/pkg/database/memory"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

var _ = Describe("BundleWriter", func() {
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

	Context("Write", func() {
		It("should write and read DataBundles with correct relationships and attributes", func() {
			ctx := context.Background()

			bundles1 := []*base.DataBundle{
				{
					Name: "user_created",
					Arguments: []string{
						"organizationID",
						"companyID",
						"userID",
					},
					Operations: []*base.Operation{
						{
							RelationshipsWrite: []string{
								"organization:{{.organizationID}}#member@company:{{.companyID}}#admin",
								"organization:{{.organizationID}}#member@user:{{.userID}}",
								"organization:{{.organizationID}}#admin@user:{{.userID}}",
							},
							RelationshipsDelete: []string{
								"company:{{.companyID}}#admin@user:{{.userID}}",
							},
							AttributesWrite: []string{
								"organization:{{.organizationID}}$public|boolean:true",
							},
							AttributesDelete: []string{
								"organization:{{.organizationID}}$balance|double:120.900",
							},
						},
					},
				},
			}

			var sBundles1 []storage.Bundle
			for _, b := range bundles1 {
				sBundles1 = append(sBundles1, storage.Bundle{
					Name:       b.Name,
					DataBundle: b,
					TenantID:   "t1",
				})
			}

			names1, err := bundleWriter.Write(ctx, sBundles1)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(names1).Should(Equal([]string{"user_created"}))

			bundle1, err := bundleReader.Read(ctx, "t1", "user_created")
			Expect(err).ShouldNot(HaveOccurred())

			Expect(bundle1.GetName()).Should(Equal("user_created"))
			Expect(bundle1.GetArguments()).Should(Equal([]string{
				"organizationID",
				"companyID",
				"userID",
			}))

			Expect(bundle1.GetOperations()[0].RelationshipsWrite).Should(Equal([]string{
				"organization:{{.organizationID}}#member@company:{{.companyID}}#admin",
				"organization:{{.organizationID}}#member@user:{{.userID}}",
				"organization:{{.organizationID}}#admin@user:{{.userID}}",
			}))

			Expect(bundle1.GetOperations()[0].RelationshipsDelete).Should(Equal([]string{
				"company:{{.companyID}}#admin@user:{{.userID}}",
			}))

			Expect(bundle1.GetOperations()[0].AttributesWrite).Should(Equal([]string{
				"organization:{{.organizationID}}$public|boolean:true",
			}))

			Expect(bundle1.GetOperations()[0].AttributesDelete).Should(Equal([]string{
				"organization:{{.organizationID}}$balance|double:120.900",
			}))

			bundles2 := []*base.DataBundle{
				{
					Name: "user_created",
					Arguments: []string{
						"organizationID",
						"companyID",
						"userID",
					},
					Operations: []*base.Operation{
						{
							RelationshipsWrite: []string{
								"organization:{{.organizationID}}#member@company:{{.companyID}}#admin",
								"organization:{{.organizationID}}#admin@user:{{.userID}}",
							},
							RelationshipsDelete: []string{
								"company:{{.companyID}}#admin@user:{{.userID}}",
							},
							AttributesWrite:  []string{},
							AttributesDelete: []string{},
						},
					},
				},
			}

			var sBundles2 []storage.Bundle
			for _, b := range bundles2 {
				sBundles2 = append(sBundles2, storage.Bundle{
					Name:       b.Name,
					DataBundle: b,
					TenantID:   "t1",
				})
			}

			names2, err := bundleWriter.Write(ctx, sBundles2)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(names2).Should(Equal([]string{"user_created"}))

			bundle2, err := bundleReader.Read(ctx, "t1", "user_created")
			Expect(err).ShouldNot(HaveOccurred())

			Expect(bundle2.GetName()).Should(Equal("user_created"))
			Expect(bundle2.GetArguments()).Should(Equal([]string{
				"organizationID",
				"companyID",
				"userID",
			}))

			Expect(bundle2.GetOperations()[0].RelationshipsWrite).Should(Equal([]string{
				"organization:{{.organizationID}}#member@company:{{.companyID}}#admin",
				"organization:{{.organizationID}}#admin@user:{{.userID}}",
			}))

			Expect(bundle2.GetOperations()[0].RelationshipsDelete).Should(Equal([]string{
				"company:{{.companyID}}#admin@user:{{.userID}}",
			}))

			Expect(bundle2.GetOperations()[0].AttributesWrite).Should(BeEmpty())

			Expect(bundle2.GetOperations()[0].AttributesDelete).Should(BeEmpty())
		})
	})

	Context("Delete", func() {
		It("should delete DataBundles Correctly", func() {
			ctx := context.Background()

			bundles := []*base.DataBundle{
				{
					Name: "user_created",
					Arguments: []string{
						"organizationID",
						"companyID",
						"userID",
					},
					Operations: []*base.Operation{
						{
							RelationshipsWrite: []string{
								"organization:{{.organizationID}}#member@company:{{.companyID}}#admin",
								"organization:{{.organizationID}}#member@user:{{.userID}}",
								"organization:{{.organizationID}}#admin@user:{{.userID}}",
							},
							RelationshipsDelete: []string{
								"company:{{.companyID}}#admin@user:{{.userID}}",
							},
							AttributesWrite: []string{
								"organization:{{.organizationID}}$public|boolean:true",
							},
							AttributesDelete: []string{
								"organization:{{.organizationID}}$balance|double:120.900",
							},
						},
					},
				},
				{
					Name: "user_deleted",
					Arguments: []string{
						"organizationID",
						"companyID",
					},
					Operations: []*base.Operation{
						{
							RelationshipsWrite: []string{
								"organization:{{.organizationID}}#member@company:{{.companyID}}#admin",
							},
							RelationshipsDelete: []string{},
							AttributesWrite:     []string{},
							AttributesDelete:    []string{},
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
			Expect(names).Should(Equal([]string{"user_created", "user_deleted"}))

			_, err = bundleReader.Read(ctx, "t1", "user_created")
			Expect(err).ShouldNot(HaveOccurred())

			err = bundleWriter.Delete(ctx, "t1", "user_created")
			Expect(err).ShouldNot(HaveOccurred())

			_, err = bundleReader.Read(ctx, "t1", "user_created")
			Expect(err.Error()).Should(Equal(base.ErrorCode_ERROR_CODE_BUNDLE_NOT_FOUND.String()))

			_, err = bundleReader.Read(ctx, "t1", "user_deleted")
			Expect(err).ShouldNot(HaveOccurred())
		})
	})
})
