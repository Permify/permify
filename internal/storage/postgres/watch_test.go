package postgres

import (
	"context"
	"fmt"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/internal/storage/postgres/instance"
	"github.com/Permify/permify/pkg/attribute"
	"github.com/Permify/permify/pkg/database"
	PQDatabase "github.com/Permify/permify/pkg/database/postgres"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
)

var _ = Describe("Watch", func() {
	var db database.Database
	var dataWriter *DataWriter
	var watcher *Watch

	BeforeEach(func() {
		version := os.Getenv("POSTGRES_VERSION")

		if version == "" {
			version = "14"
		}

		db = instance.PostgresDB(version)
		dataWriter = NewDataWriter(db.(*PQDatabase.Postgres))
		watcher = NewWatcher(db.(*PQDatabase.Postgres))
	})

	AfterEach(func() {
		err := db.Close()
		Expect(err).ShouldNot(HaveOccurred())
	})

	Context("Watch", func() {
		It("watch", func() {
			ctx := context.Background()

			tup1, err := tuple.Tuple("organization:organization-1#admin@user:user-1")
			Expect(err).ShouldNot(HaveOccurred())

			tuples := database.NewTupleCollection([]*base.Tuple{
				tup1,
			}...)

			token1, err := dataWriter.Write(ctx, "t1", tuples, database.NewAttributeCollection())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(token1.String()).ShouldNot(Equal(""))

			changes, errs := watcher.Watch(ctx, "t1", token1.String())

			time.Sleep(100 * time.Millisecond)

			go func() {
				defer GinkgoRecover()

				attr1, err := attribute.Attribute("organization:organization-1$public|boolean:true")
				Expect(err).ShouldNot(HaveOccurred())

				attr2, err := attribute.Attribute("organization:organization-1$ip_addresses|string[]:127.0.0.1,127.0.0.2")
				Expect(err).ShouldNot(HaveOccurred())

				attr3, err := attribute.Attribute("organization:organization-3$balance|double:234.344")
				Expect(err).ShouldNot(HaveOccurred())

				tup1, err := tuple.Tuple("organization:organization-1#admin@user:user-1")
				Expect(err).ShouldNot(HaveOccurred())

				tup2, err := tuple.Tuple("organization:organization-1#admin@user:user-4")
				Expect(err).ShouldNot(HaveOccurred())

				tup3, err := tuple.Tuple("organization:organization-1#admin@user:user-2")
				Expect(err).ShouldNot(HaveOccurred())

				attributes1 := database.NewAttributeCollection([]*base.Attribute{
					attr1,
					attr2,
					attr3,
				}...)

				tuples1 := database.NewTupleCollection([]*base.Tuple{
					tup1,
					tup2,
					tup3,
				}...)

				time.Sleep(time.Second)
				token2, err := dataWriter.Write(ctx, "t1", tuples1, attributes1)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(token2.String()).ShouldNot(Equal(""))
			}()

			select {
			case change := <-changes:
				// Test the received change
				Expect(change.DataChanges).ShouldNot(BeNil())
				// Additional assertions about the structure and content of 'change'
			case err := <-errs:
				// Handle and assert the error
				Expect(err).ShouldNot(HaveOccurred())
			case <-time.After(time.Second * 10):
				fmt.Println("test timed out")
			}
		})
	})
})
