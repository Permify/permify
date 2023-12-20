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

	var dataWriter *DataWriter

	BeforeEach(func() {
		database, err := memory.New(migrations.Schema)
		Expect(err).ShouldNot(HaveOccurred())
		db = database

		dataWriter = NewDataWriter(db)
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
