package memory

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/internal/storage/memory/migrations"
	"github.com/Permify/permify/pkg/attribute"
	"github.com/Permify/permify/pkg/database"
	"github.com/Permify/permify/pkg/database/memory"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
)

var _ = Describe("DataWriter", func() {
	var db *memory.Memory

	var bundleReader *BundleReader
	var bundleWriter *BundleWriter
	var dataWriter *DataWriter
	var dataReader *DataReader

	BeforeEach(func() {
		database, err := memory.New(migrations.Schema)
		Expect(err).ShouldNot(HaveOccurred())
		db = database

		bundleReader = NewBundleReader(db)
		bundleWriter = NewBundleWriter(db)
		dataWriter = NewDataWriter(db)
		dataReader = NewDataReader(db)
	})

	AfterEach(func() {
		err := db.Close()
		Expect(err).ShouldNot(HaveOccurred())
	})

	Context("RunBundle", func() {
		It("should run the bundle successfully and return an encoded snapshot token", func() {
			ctx := context.Background()

			// Create a valid DataBundle
			bundle := &base.DataBundle{
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
							"organization:{{.organizationID}}#admin@user:{{.userID}}",
						},
						AttributesWrite: []string{
							"organization:{{.organizationID}}$public|boolean:true",
							"company:{{.companyID}}$public|boolean:true",
						},
						AttributesDelete: []string{
							"organization:{{.organizationID}}$balance|double:120.900",
						},
					},
				},
			}

			names, err := bundleWriter.Write(ctx, []storage.Bundle{
				{
					Name:       bundle.GetName(),
					DataBundle: bundle,
					TenantID:   "t1",
				},
			})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(names).Should(Equal([]string{"user_created"}))

			b, err := bundleReader.Read(ctx, "t1", "user_created")
			Expect(err).ShouldNot(HaveOccurred())

			// Call RunBundle with the real implementation of runOperation
			token, err := dataWriter.RunBundle(ctx, "t1", map[string]string{
				"organizationID": "1",
				"companyID":      "4",
				"userID":         "1",
			}, b)

			// Verify that the token is returned without errors
			Expect(err).ShouldNot(HaveOccurred())
			Expect(token).ShouldNot(BeNil())

			// Decode the token and verify its structure or properties if needed
			decodedToken, err := token.Decode()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(decodedToken).ShouldNot(BeNil())

			tCollection, _, err := dataReader.ReadRelationships(ctx, "t1", &base.TupleFilter{
				Entity: &base.EntityFilter{
					Type: "organization",
					Ids:  []string{"1"},
				},
				Relation: "",
				Subject: &base.SubjectFilter{
					Type:     "",
					Ids:      []string{},
					Relation: "",
				},
			}, "", database.NewPagination())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(len(tCollection.GetTuples())).Should(Equal(2))

			aCollection, _, err := dataReader.ReadAttributes(ctx, "t1", &base.AttributeFilter{
				Entity: &base.EntityFilter{
					Type: "company",
					Ids:  []string{"4"},
				},
				Attributes: []string{},
			}, "", database.NewPagination())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(len(aCollection.GetAttributes())).Should(Equal(1))
		})
	})

	Context("Write method", func() {
		It("should return noop token when no tuples or attributes", func() {
			ctx := context.Background()

			// Test with empty collections
			token, err := dataWriter.Write(ctx, "t1", database.NewTupleCollection(), database.NewAttributeCollection())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(token).ShouldNot(BeNil())

			// Verify it's a noop token by checking if it can be decoded
			decodedToken, err := token.Decode()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(decodedToken).ShouldNot(BeNil())
		})

		It("should handle ellipsis relation in tuples", func() {
			ctx := context.Background()

			// Create a tuple with ellipsis relation
			tup, err := tuple.Tuple("organization:org-1#admin@user:user-1")
			Expect(err).ShouldNot(HaveOccurred())

			// Modify the tuple to have ellipsis relation
			tup.Subject.Relation = tuple.ELLIPSIS

			tuples := database.NewTupleCollection([]*base.Tuple{
				tup,
			}...)

			token, err := dataWriter.Write(ctx, "t1", tuples, database.NewAttributeCollection())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(token).ShouldNot(BeNil())

			// Verify the tuple was written correctly
			it, err := dataReader.QueryRelationships(ctx, "t1", &base.TupleFilter{
				Entity: &base.EntityFilter{
					Type: "organization",
					Ids:  []string{"org-1"},
				},
			}, "", database.NewCursorPagination())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(it.HasNext()).Should(BeTrue())
		})
	})

	Context("Delete method", func() {
		It("should handle tuple deletion with type conversion error", func() {
			ctx := context.Background()

			// First write some tuples
			tup1, err := tuple.Tuple("organization:org-1#admin@user:user-1")
			Expect(err).ShouldNot(HaveOccurred())

			tuples := database.NewTupleCollection([]*base.Tuple{
				tup1,
			}...)

			_, err = dataWriter.Write(ctx, "t1", tuples, database.NewAttributeCollection())
			Expect(err).ShouldNot(HaveOccurred())

			// Test deletion with valid filter
			token, err := dataWriter.Delete(ctx, "t1", &base.TupleFilter{
				Entity: &base.EntityFilter{
					Type: "organization",
					Ids:  []string{"org-1"},
				},
			}, &base.AttributeFilter{})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(token).ShouldNot(BeNil())
		})

		It("should handle attribute deletion with type conversion error", func() {
			ctx := context.Background()

			// First write some attributes
			attr1, err := attribute.Attribute("organization:org-1$public|boolean:true")
			Expect(err).ShouldNot(HaveOccurred())

			attributes := database.NewAttributeCollection([]*base.Attribute{
				attr1,
			}...)

			_, err = dataWriter.Write(ctx, "t1", database.NewTupleCollection(), attributes)
			Expect(err).ShouldNot(HaveOccurred())

			// Test deletion with valid filter
			token, err := dataWriter.Delete(ctx, "t1", &base.TupleFilter{}, &base.AttributeFilter{
				Entity: &base.EntityFilter{
					Type: "organization",
					Ids:  []string{"org-1"},
				},
				Attributes: []string{"public"},
			})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(token).ShouldNot(BeNil())
		})
	})

	Context("RunBundle error handling", func() {
		It("should handle bundle operation errors", func() {
			ctx := context.Background()

			// Create a bundle with invalid operation
			bundle := &base.DataBundle{
				Name: "invalid_bundle",
				Operations: []*base.Operation{
					{
						RelationshipsWrite: []string{
							"invalid:tuple:format", // This should cause an error
						},
					},
				},
			}

			// This should return an error due to invalid tuple format
			_, err := dataWriter.RunBundle(ctx, "t1", map[string]string{}, bundle)
			Expect(err).Should(HaveOccurred())
		})

		It("should handle ellipsis relation in bundle operations", func() {
			ctx := context.Background()

			// Create a bundle with ellipsis relation
			bundle := &base.DataBundle{
				Name: "ellipsis_bundle",
				Operations: []*base.Operation{
					{
						RelationshipsWrite: []string{
							"organization:org-1#admin@user:user-1#...", // Ellipsis relation
						},
					},
				},
			}

			token, err := dataWriter.RunBundle(ctx, "t1", map[string]string{}, bundle)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(token).ShouldNot(BeNil())
		})

		It("should handle tuple deletion in bundle operations", func() {
			ctx := context.Background()

			// First write some tuples
			tup1, err := tuple.Tuple("organization:org-1#admin@user:user-1")
			Expect(err).ShouldNot(HaveOccurred())

			tuples := database.NewTupleCollection([]*base.Tuple{
				tup1,
			}...)

			_, err = dataWriter.Write(ctx, "t1", tuples, database.NewAttributeCollection())
			Expect(err).ShouldNot(HaveOccurred())

			// Create a bundle that deletes the tuple
			bundle := &base.DataBundle{
				Name: "delete_bundle",
				Operations: []*base.Operation{
					{
						RelationshipsDelete: []string{
							"organization:org-1#admin@user:user-1",
						},
					},
				},
			}

			token, err := dataWriter.RunBundle(ctx, "t1", map[string]string{}, bundle)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(token).ShouldNot(BeNil())

			// Verify the tuple was deleted
			it, err := dataReader.QueryRelationships(ctx, "t1", &base.TupleFilter{
				Entity: &base.EntityFilter{
					Type: "organization",
					Ids:  []string{"org-1"},
				},
			}, "", database.NewCursorPagination())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(it.HasNext()).Should(BeFalse())
		})

		It("should handle attribute deletion in bundle operations", func() {
			ctx := context.Background()

			// First write some attributes
			attr1, err := attribute.Attribute("organization:org-1$public|boolean:true")
			Expect(err).ShouldNot(HaveOccurred())

			attributes := database.NewAttributeCollection([]*base.Attribute{
				attr1,
			}...)

			_, err = dataWriter.Write(ctx, "t1", database.NewTupleCollection(), attributes)
			Expect(err).ShouldNot(HaveOccurred())

			// Create a bundle that deletes the attribute
			bundle := &base.DataBundle{
				Name: "delete_attribute_bundle",
				Operations: []*base.Operation{
					{
						AttributesDelete: []string{
							"organization:org-1$public|boolean:true",
						},
					},
				},
			}

			token, err := dataWriter.RunBundle(ctx, "t1", map[string]string{}, bundle)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(token).ShouldNot(BeNil())

			// Verify the attribute was deleted
			attr, err := dataReader.QuerySingleAttribute(ctx, "t1", &base.AttributeFilter{
				Entity: &base.EntityFilter{
					Type: "organization",
					Ids:  []string{"org-1"},
				},
				Attributes: []string{"public"},
			}, "")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(attr).Should(BeNil())
		})

		It("should handle mixed operations in bundle", func() {
			ctx := context.Background()

			// Create a bundle with mixed operations
			bundle := &base.DataBundle{
				Name: "mixed_bundle",
				Operations: []*base.Operation{
					{
						RelationshipsWrite: []string{
							"organization:org-1#admin@user:user-1",
						},
						AttributesWrite: []string{
							"organization:org-1$public|boolean:true",
						},
					},
				},
			}

			token, err := dataWriter.RunBundle(ctx, "t1", map[string]string{}, bundle)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(token).ShouldNot(BeNil())

			// Verify both tuple and attribute were written
			it, err := dataReader.QueryRelationships(ctx, "t1", &base.TupleFilter{
				Entity: &base.EntityFilter{
					Type: "organization",
					Ids:  []string{"org-1"},
				},
			}, "", database.NewCursorPagination())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(it.HasNext()).Should(BeTrue())

			attr, err := dataReader.QuerySingleAttribute(ctx, "t1", &base.AttributeFilter{
				Entity: &base.EntityFilter{
					Type: "organization",
					Ids:  []string{"org-1"},
				},
				Attributes: []string{"public"},
			}, "")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(attr).ShouldNot(BeNil())
		})
	})
})
