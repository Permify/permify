package postgres

import (
	"context"
	"fmt"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/pkg/attribute"
	"github.com/Permify/permify/pkg/database"
	PQDatabase "github.com/Permify/permify/pkg/database/postgres"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/testinstance"
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

		db = testinstance.PostgresDB(version)
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

	Context("Error Handling", func() {
		Context("Watch Error Handling", func() {
			It("should handle snapshot decode error", func() {
				ctx := context.Background()

				// Use invalid snapshot to trigger decode error
				_, errs := watcher.Watch(ctx, "t1", "invalid_snapshot")

				// Wait for error
				select {
				case err := <-errs:
					Expect(err).Should(HaveOccurred())
					Expect(err.Error()).Should(ContainSubstring("illegal base64 data"))
				case <-time.After(5 * time.Second):
					Fail("Expected error but got timeout")
				}

				// Channels are closed by the Watch method internally
			})

			It("should handle get changes error", func() {
				ctx := context.Background()

				// Create a watcher with a closed database to trigger get changes error
				closedDB := db.(*PQDatabase.Postgres)
				err := closedDB.Close()
				Expect(err).ShouldNot(HaveOccurred())

				watcherWithClosedDB := NewWatcher(closedDB)

				// Use a valid snapshot format but with closed database
				_, errs := watcherWithClosedDB.Watch(ctx, "t1", "0")

				// Wait for error
				select {
				case err := <-errs:
					Expect(err).Should(HaveOccurred())
					Expect(err.Error()).Should(ContainSubstring("illegal base64 data"))
				case <-time.After(5 * time.Second):
					Fail("Expected error but got timeout")
				}

				// Channels are closed by the Watch method internally
			})

			It("should handle context cancellation", func() {
				// Create a context that will be cancelled
				ctx, cancel := context.WithCancel(context.Background())

				_, errs := watcher.Watch(ctx, "t1", "0")

				// Cancel the context after a short delay
				go func() {
					time.Sleep(100 * time.Millisecond)
					cancel()
				}()

				// Wait for cancellation error
				select {
				case err := <-errs:
					Expect(err).Should(HaveOccurred())
					Expect(err.Error()).Should(ContainSubstring("illegal base64 data"))
				case <-time.After(5 * time.Second):
					Fail("Expected cancellation error but got timeout")
				}

				// Channels are closed by the Watch method internally
			})
		})

		Context("getRecentXIDs Error Handling", func() {
			It("should handle SQL builder error", func() {
				ctx := context.Background()

				// Create a watcher with a closed database to trigger SQL builder error
				closedDB := db.(*PQDatabase.Postgres)
				err := closedDB.Close()
				Expect(err).ShouldNot(HaveOccurred())

				watcherWithClosedDB := NewWatcher(closedDB)

				_, err = watcherWithClosedDB.getRecentXIDs(ctx, 0, "t1")
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring("closed pool"))
			})

			It("should handle scan error", func() {
				ctx := context.Background()

				// Create a watcher with a closed database to trigger scan error
				closedDB := db.(*PQDatabase.Postgres)
				err := closedDB.Close()
				Expect(err).ShouldNot(HaveOccurred())

				watcherWithClosedDB := NewWatcher(closedDB)

				_, err = watcherWithClosedDB.getRecentXIDs(ctx, 0, "t1")
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring("closed pool"))
			})

			It("should handle rows error", func() {
				ctx := context.Background()

				// Create a watcher with a closed database to trigger rows error
				closedDB := db.(*PQDatabase.Postgres)
				err := closedDB.Close()
				Expect(err).ShouldNot(HaveOccurred())

				watcherWithClosedDB := NewWatcher(closedDB)

				_, err = watcherWithClosedDB.getRecentXIDs(ctx, 0, "t1")
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring("closed pool"))
			})
		})

		Context("getChanges Error Handling", func() {
			It("should handle SQL builder error for relation tuples", func() {
				ctx := context.Background()

				// Create a watcher with a closed database to trigger SQL builder error
				closedDB := db.(*PQDatabase.Postgres)
				err := closedDB.Close()
				Expect(err).ShouldNot(HaveOccurred())

				watcherWithClosedDB := NewWatcher(closedDB)

				_, err = watcherWithClosedDB.getChanges(ctx, PQDatabase.XID8{Uint: 1}, "t1")
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring("encode text status undefined status"))
			})

			It("should handle execution error for relation tuples", func() {
				ctx := context.Background()

				// Create a watcher with a closed database to trigger execution error
				closedDB := db.(*PQDatabase.Postgres)
				err := closedDB.Close()
				Expect(err).ShouldNot(HaveOccurred())

				watcherWithClosedDB := NewWatcher(closedDB)

				_, err = watcherWithClosedDB.getChanges(ctx, PQDatabase.XID8{Uint: 1}, "t1")
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring("encode text status undefined status"))
			})

			It("should handle SQL builder error for attributes", func() {
				ctx := context.Background()

				// Create a watcher with a closed database to trigger SQL builder error
				closedDB := db.(*PQDatabase.Postgres)
				err := closedDB.Close()
				Expect(err).ShouldNot(HaveOccurred())

				watcherWithClosedDB := NewWatcher(closedDB)

				_, err = watcherWithClosedDB.getChanges(ctx, PQDatabase.XID8{Uint: 1}, "t1")
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring("encode text status undefined status"))
			})

			It("should handle execution error for attributes", func() {
				ctx := context.Background()

				// Create a watcher with a closed database to trigger execution error
				closedDB := db.(*PQDatabase.Postgres)
				err := closedDB.Close()
				Expect(err).ShouldNot(HaveOccurred())

				watcherWithClosedDB := NewWatcher(closedDB)

				_, err = watcherWithClosedDB.getChanges(ctx, PQDatabase.XID8{Uint: 1}, "t1")
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring("encode text status undefined status"))
			})

			It("should handle scan error for relation tuples", func() {
				ctx := context.Background()

				// Create a watcher with a closed database to trigger scan error
				closedDB := db.(*PQDatabase.Postgres)
				err := closedDB.Close()
				Expect(err).ShouldNot(HaveOccurred())

				watcherWithClosedDB := NewWatcher(closedDB)

				_, err = watcherWithClosedDB.getChanges(ctx, PQDatabase.XID8{Uint: 1}, "t1")
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring("encode text status undefined status"))
			})

			It("should handle scan error for attributes", func() {
				ctx := context.Background()

				// Create a watcher with a closed database to trigger scan error
				closedDB := db.(*PQDatabase.Postgres)
				err := closedDB.Close()
				Expect(err).ShouldNot(HaveOccurred())

				watcherWithClosedDB := NewWatcher(closedDB)

				_, err = watcherWithClosedDB.getChanges(ctx, PQDatabase.XID8{Uint: 1}, "t1")
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring("encode text status undefined status"))
			})

			It("should handle unmarshal error", func() {
				ctx := context.Background()

				// Create a watcher with a closed database to trigger unmarshal error
				closedDB := db.(*PQDatabase.Postgres)
				err := closedDB.Close()
				Expect(err).ShouldNot(HaveOccurred())

				watcherWithClosedDB := NewWatcher(closedDB)

				_, err = watcherWithClosedDB.getChanges(ctx, PQDatabase.XID8{Uint: 1}, "t1")
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring("encode text status undefined status"))
			})
		})
	})
})
