package memory

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/internal/storage/memory/migrations"
	memory "github.com/Permify/permify/pkg/database/memory"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

var _ = Describe("BundleWriter", func() {
	var db *memory.Memory
	var bundleWriter *BundleWriter
	var bundleReader *BundleReader
	var dataWriter *DataWriter

	BeforeEach(func() {
		database, err := memory.New(migrations.Schema)
		Expect(err).ShouldNot(HaveOccurred())
		db = database
		bundleWriter = NewBundleWriter(db)
		bundleReader = NewBundleReader(db)
		dataWriter = NewDataWriter(db)
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

			names1, err := bundleWriter.Write(ctx, "t1", bundles1)

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

			names2, err := bundleWriter.Write(ctx, "t1", bundles2)

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

			names, err := bundleWriter.Write(ctx, "t1", bundles)
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

	Context("RunBundle", func() {
		It("should run the bundle successfully and return an encoded snapshot token", func() {
			ctx := context.Background()

			// Create a valid DataBundle
			bundle := &base.DataBundle{
				Name: "test_bundle",
				Arguments: []string{
					"organizationID",
					"companyID",
					"userID",
				},
				Operations: []*base.Operation{
					{
						RelationshipsWrite: []string{
							"organization:1#member@company:1#admin",
							"organization:1#member@user:1",
							"organization:1#admin@user:1",
						},
						RelationshipsDelete: []string{
							"organization:1#member@company:1#admin",
							"organization:1#member@user:1",
							"organization:1#admin@user:1",
						},
						AttributesWrite: []string{
							"organization:1$public|boolean:true",
						},
						AttributesDelete: []string{
							"organization:1$public|boolean:true",
						},
					},
				},
			}

			// Call RunBundle with the real implementation of runOperation
			token, err := dataWriter.RunBundle(ctx, "t1", map[string]string{}, bundle)

			// Verify that the token is returned without errors
			Expect(err).ShouldNot(HaveOccurred())
			Expect(token).ShouldNot(BeNil())

			// Decode the token and verify its structure or properties if needed
			decodedToken, err := token.Decode()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(decodedToken).ShouldNot(BeNil())
			// Add additional assertions based on the expected structure or properties of the token
		})
	})
})
