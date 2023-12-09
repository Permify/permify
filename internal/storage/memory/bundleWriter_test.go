package memory

import (
	"context"

	"github.com/hashicorp/go-memdb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	memory "github.com/Permify/permify/pkg/database/memory"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

var _ = Describe("BundleWriter", func() {
	var Schema = &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			SchemaDefinitionsTable: {
				Name: SchemaDefinitionsTable,
				Indexes: map[string]*memdb.IndexSchema{
					"id": {
						Name:   "id",
						Unique: true,
						Indexer: &memdb.CompoundIndex{
							Indexes: []memdb.Indexer{
								&memdb.StringFieldIndex{Field: "TenantID"},
								&memdb.StringFieldIndex{Field: "Name"},
								&memdb.StringFieldIndex{Field: "Version"},
							},
						},
					},
					"version": {
						Name:   "version",
						Unique: false,
						Indexer: &memdb.CompoundIndex{
							Indexes: []memdb.Indexer{
								&memdb.StringFieldIndex{Field: "TenantID"},
								&memdb.StringFieldIndex{Field: "Version"},
							},
						},
					},
					"tenant": {
						Name:   "tenant",
						Unique: false,
						Indexer: &memdb.CompoundIndex{
							Indexes: []memdb.Indexer{
								&memdb.StringFieldIndex{Field: "TenantID"},
							},
						},
					},
				},
			},
			AttributesTable: {
				Name: AttributesTable,
				Indexes: map[string]*memdb.IndexSchema{
					"id": {
						Name:   "id",
						Unique: true,
						Indexer: &memdb.CompoundIndex{
							Indexes: []memdb.Indexer{
								&memdb.StringFieldIndex{Field: "TenantID"},
								&memdb.StringFieldIndex{Field: "EntityType"},
								&memdb.StringFieldIndex{Field: "EntityID"},
								&memdb.StringFieldIndex{Field: "Attribute"},
							},
						},
					},
					"entity-type-index": {
						Name:   "entity-type-index",
						Unique: false,
						Indexer: &memdb.CompoundIndex{
							Indexes: []memdb.Indexer{
								&memdb.StringFieldIndex{Field: "TenantID"},
								&memdb.StringFieldIndex{Field: "EntityType"},
							},
						},
					},
					"entity-type-and-attribute-index": {
						Name:   "entity-type-and-attribute-index",
						Unique: false,
						Indexer: &memdb.CompoundIndex{
							Indexes: []memdb.Indexer{
								&memdb.StringFieldIndex{Field: "TenantID"},
								&memdb.StringFieldIndex{Field: "EntityType"},
								&memdb.StringFieldIndex{Field: "Attribute"},
							},
						},
					},
				},
			},
			RelationTuplesTable: {
				Name: RelationTuplesTable,
				Indexes: map[string]*memdb.IndexSchema{
					"id": {
						Name:   "id",
						Unique: true,
						Indexer: &memdb.CompoundIndex{
							Indexes: []memdb.Indexer{
								&memdb.StringFieldIndex{Field: "TenantID"},
								&memdb.StringFieldIndex{Field: "EntityType"},
								&memdb.StringFieldIndex{Field: "EntityID"},
								&memdb.StringFieldIndex{Field: "Relation"},
								&memdb.StringFieldIndex{Field: "SubjectType"},
								&memdb.StringFieldIndex{Field: "SubjectID"},
								&memdb.StringFieldIndex{Field: "SubjectRelation"},
							},
							AllowMissing: true,
						},
					},
					"entity-index": {
						Name:   "entity-index",
						Unique: false,
						Indexer: &memdb.CompoundIndex{
							Indexes: []memdb.Indexer{
								&memdb.StringFieldIndex{Field: "TenantID"},
								&memdb.StringFieldIndex{Field: "EntityType"},
								&memdb.StringFieldIndex{Field: "EntityID"},
								&memdb.StringFieldIndex{Field: "Relation"},
							},
						},
					},
					"relation-index": {
						Name:   "relation-index",
						Unique: false,
						Indexer: &memdb.CompoundIndex{
							Indexes: []memdb.Indexer{
								&memdb.StringFieldIndex{Field: "TenantID"},
								&memdb.StringFieldIndex{Field: "EntityType"},
								&memdb.StringFieldIndex{Field: "Relation"},
								&memdb.StringFieldIndex{Field: "SubjectType"},
							},
						},
					},
					"entity-type-index": {
						Name:   "entity-type-index",
						Unique: false,
						Indexer: &memdb.CompoundIndex{
							Indexes: []memdb.Indexer{
								&memdb.StringFieldIndex{Field: "TenantID"},
								&memdb.StringFieldIndex{Field: "EntityType"},
							},
						},
					},
					"entity-type-and-relation-index": {
						Name:   "entity-type-and-relation-index",
						Unique: false,
						Indexer: &memdb.CompoundIndex{
							Indexes: []memdb.Indexer{
								&memdb.StringFieldIndex{Field: "TenantID"},
								&memdb.StringFieldIndex{Field: "EntityType"},
								&memdb.StringFieldIndex{Field: "Relation"},
							},
						},
					},
				},
			},
			TenantsTable: {
				Name: TenantsTable,
				Indexes: map[string]*memdb.IndexSchema{
					"id": {
						Name:   "id",
						Unique: true,
						Indexer: &memdb.CompoundIndex{
							Indexes: []memdb.Indexer{
								&memdb.StringFieldIndex{Field: "ID"},
							},
						},
					},
				},
			},
			BundlesTable: {
				Name: BundlesTable,
				Indexes: map[string]*memdb.IndexSchema{
					"id": {
						Name:   "id",
						Unique: true,
						Indexer: &memdb.StringFieldIndex{
							Field: "Name",
						},
					},
				},
			},
		},
	}

	var db *memory.Memory
	var bundleWriter *BundleWriter
	var bundleReader *BundleReader

	BeforeEach(func() {
		database, err := memory.New(Schema)
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

})
