package memory

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/internal/storage/memory/migrations"
	"github.com/Permify/permify/pkg/database"
	"github.com/Permify/permify/pkg/database/memory"
	base "github.com/Permify/permify/pkg/pb/base/v1"
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
})
