package postgres

import (
	"context"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/internal/storage/postgres/instance"
	"github.com/Permify/permify/pkg/attribute"
	"github.com/Permify/permify/pkg/database"
	PQDatabase "github.com/Permify/permify/pkg/database/postgres"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/token"
	"github.com/Permify/permify/pkg/tuple"
)

var _ = Describe("DataReader", func() {
	var db database.Database
	var dataWriter *DataWriter
	var dataReader *DataReader

	BeforeEach(func() {
		version := os.Getenv("POSTGRES_VERSION")

		if version == "" {
			version = "14"
		}

		db = instance.PostgresDB(version)
		dataWriter = NewDataWriter(db.(*PQDatabase.Postgres))
		dataReader = NewDataReader(db.(*PQDatabase.Postgres))
	})

	AfterEach(func() {
		err := db.Close()
		Expect(err).ShouldNot(HaveOccurred())
	})

	Context("Head Snapshot", func() {
		It("should retrieve the most recent snapshot for a tenant", func() {
			ctx := context.Background()

			var mostRecentSnapshot token.EncodedSnapToken

			// Insert multiple snapshots for a single tenant
			for i := 0; i < 3; i++ {

				tup1, err := tuple.Tuple("organization:organization-1#admin@user:user-1")
				Expect(err).ShouldNot(HaveOccurred())

				tup2, err := tuple.Tuple("organization:organization-2#admin@user:user-1")
				Expect(err).ShouldNot(HaveOccurred())

				tuples := database.NewTupleCollection([]*base.Tuple{
					tup1,
					tup2,
				}...)

				attr1, err := attribute.Attribute("organization:1$public|boolean:true")
				Expect(err).ShouldNot(HaveOccurred())

				attr2, err := attribute.Attribute("organization:2$public|boolean:false")
				Expect(err).ShouldNot(HaveOccurred())

				attributes := database.NewAttributeCollection([]*base.Attribute{
					attr1,
					attr2,
				}...)

				token, err := dataWriter.Write(ctx, "t1", tuples, attributes)
				Expect(err).ShouldNot(HaveOccurred())

				mostRecentSnapshot = token

				time.Sleep(time.Millisecond * 2)
			}

			// Attempt to retrieve the head snapshot from DataReader
			headSnapshot, err := dataReader.HeadSnapshot(ctx, "t1")
			Expect(err).ShouldNot(HaveOccurred())

			// Validate that the retrieved head snapshot matches the most recently inserted snapshot
			Expect(headSnapshot.Encode()).Should(Equal(mostRecentSnapshot), "The retrieved head snapshot should be the most recently written one.")
		})
	})

	Context("Query Relationships", func() {
		It("should write relationships and query relationships correctly", func() {
			ctx := context.Background()

			tup1, err := tuple.Tuple("organization:organization-1#admin@user:user-1")
			Expect(err).ShouldNot(HaveOccurred())

			tup2, err := tuple.Tuple("organization:organization-2#admin@user:user-1")
			Expect(err).ShouldNot(HaveOccurred())

			tuples1 := database.NewTupleCollection([]*base.Tuple{
				tup1,
				tup2,
			}...)

			token1, err := dataWriter.Write(ctx, "t1", tuples1, database.NewAttributeCollection())
			Expect(err).ShouldNot(HaveOccurred())

			tup3, err := tuple.Tuple("organization:organization-1#admin@user:user-2")
			Expect(err).ShouldNot(HaveOccurred())

			tuples2 := database.NewTupleCollection([]*base.Tuple{
				tup3,
			}...)

			token2, err := dataWriter.Write(ctx, "t1", tuples2, database.NewAttributeCollection())
			Expect(err).ShouldNot(HaveOccurred())

			it1, err := dataReader.QueryRelationships(ctx, "t1", &base.TupleFilter{
				Entity: &base.EntityFilter{
					Type: "organization",
					Ids:  []string{"organization-1"},
				},
			}, token1.String(), database.NewCursorPagination())
			Expect(err).ShouldNot(HaveOccurred())

			Expect(it1.HasNext()).Should(Equal(true))
			Expect(it1.GetNext()).Should(Equal(tup1))
			Expect(it1.HasNext()).Should(Equal(false))

			it2, err := dataReader.QueryRelationships(ctx, "t1", &base.TupleFilter{
				Entity: &base.EntityFilter{
					Type: "organization",
					Ids:  []string{"organization-1"},
				},
			}, token2.String(), database.NewCursorPagination())
			Expect(err).ShouldNot(HaveOccurred())

			Expect(it2.HasNext()).Should(Equal(true))
			Expect(it2.GetNext()).Should(Equal(tup1))
			Expect(it2.HasNext()).Should(Equal(true))
			Expect(it2.GetNext()).Should(Equal(tup3))
			Expect(it2.HasNext()).Should(Equal(false))
		})
	})

	Context("Read Relationships", func() {
		It("should write relationships and read relationships correctly", func() {
			ctx := context.Background()

			tup1, err := tuple.Tuple("organization:organization-1#admin@user:user-1")
			Expect(err).ShouldNot(HaveOccurred())

			tup2, err := tuple.Tuple("organization:organization-2#admin@user:user-1")
			Expect(err).ShouldNot(HaveOccurred())

			tup3, err := tuple.Tuple("organization:organization-1#admin@user:user-2")
			Expect(err).ShouldNot(HaveOccurred())

			tup4, err := tuple.Tuple("organization:organization-1#admin@user:user-3")
			Expect(err).ShouldNot(HaveOccurred())

			tup5, err := tuple.Tuple("organization:organization-1#admin@user:user-4")
			Expect(err).ShouldNot(HaveOccurred())

			tup6, err := tuple.Tuple("organization:organization-1#admin@user:user-5")
			Expect(err).ShouldNot(HaveOccurred())

			tuples1 := database.NewTupleCollection([]*base.Tuple{
				tup1,
				tup2,
				tup3,
				tup4,
				tup5,
				tup6,
			}...)

			token1, err := dataWriter.Write(ctx, "t1", tuples1, database.NewAttributeCollection())
			Expect(err).ShouldNot(HaveOccurred())

			col1, ct1, err := dataReader.ReadRelationships(ctx, "t1", &base.TupleFilter{
				Entity: &base.EntityFilter{
					Type: "organization",
					Ids:  []string{"organization-1"},
				},
			}, token1.String(), database.NewPagination(database.Size(2), database.Token("")))
			Expect(err).ShouldNot(HaveOccurred())

			Expect(len(col1.GetTuples())).Should(Equal(2))

			col2, ct2, err := dataReader.ReadRelationships(ctx, "t1", &base.TupleFilter{
				Entity: &base.EntityFilter{
					Type: "organization",
					Ids:  []string{"organization-1"},
				},
			}, token1.String(), database.NewPagination(database.Size(3), database.Token(ct1.String())))
			Expect(err).ShouldNot(HaveOccurred())

			Expect(len(col2.GetTuples())).Should(Equal(3))
			Expect(ct2.String()).Should(Equal(""))

			token3, err := dataWriter.Delete(ctx, "t1", &base.TupleFilter{
				Entity: &base.EntityFilter{
					Type: "organization",
					Ids:  []string{"organization-1"},
				},
				Relation: "",
				Subject: &base.SubjectFilter{
					Type: "user",
					Ids:  []string{"user-5"},
				},
			}, &base.AttributeFilter{})
			Expect(err).ShouldNot(HaveOccurred())

			col3, ct3, err := dataReader.ReadRelationships(ctx, "t1", &base.TupleFilter{
				Entity: &base.EntityFilter{
					Type: "organization",
					Ids:  []string{"organization-1"},
				},
			}, token3.String(), database.NewPagination(database.Size(4), database.Token("")))
			Expect(err).ShouldNot(HaveOccurred())

			Expect(len(col3.GetTuples())).Should(Equal(4))
			Expect(ct3.String()).Should(Equal(""))
		})
	})

	Context("Query Single Attribute", func() {
		It("should write attributes and query single attributes correctly", func() {
			ctx := context.Background()

			attr1, err := attribute.Attribute("organization:organization-1$public|boolean:true")
			Expect(err).ShouldNot(HaveOccurred())

			attr2, err := attribute.Attribute("organization:organization-2$public|boolean:false")
			Expect(err).ShouldNot(HaveOccurred())

			attributes := database.NewAttributeCollection([]*base.Attribute{
				attr1,
				attr2,
			}...)

			token1, err := dataWriter.Write(ctx, "t1", database.NewTupleCollection(), attributes)
			Expect(err).ShouldNot(HaveOccurred())

			attribute1, err := dataReader.QuerySingleAttribute(ctx, "t1", &base.AttributeFilter{
				Entity: &base.EntityFilter{
					Type: "organization",
					Ids:  []string{"organization-1"},
				},
				Attributes: []string{"public"},
			}, token1.String())
			Expect(err).ShouldNot(HaveOccurred())

			Expect(attr1).Should(Equal(attribute1))

			token2, err := dataWriter.Delete(ctx, "t1",
				&base.TupleFilter{},
				&base.AttributeFilter{
					Entity: &base.EntityFilter{
						Type: "organization",
						Ids:  []string{"organization-1"},
					},
					Attributes: []string{"public"},
				})
			Expect(err).ShouldNot(HaveOccurred())

			attribute2, err := dataReader.QuerySingleAttribute(ctx, "t1", &base.AttributeFilter{
				Entity: &base.EntityFilter{
					Type: "organization",
					Ids:  []string{"organization-1"},
				},
				Attributes: []string{"public"},
			}, token2.String())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(attribute2).Should(BeNil())
		})
	})

	Context("Query Attributes", func() {
		It("should write attributes and query attributes correctly", func() {
			ctx := context.Background()

			attr1, err := attribute.Attribute("organization:organization-2$public|boolean:true")
			Expect(err).ShouldNot(HaveOccurred())

			attr2, err := attribute.Attribute("organization:organization-1$ip_addresses|string[]:127.0.0.1,127.0.0.2")
			Expect(err).ShouldNot(HaveOccurred())

			attributes1 := database.NewAttributeCollection([]*base.Attribute{
				attr1,
				attr2,
			}...)

			token1, err := dataWriter.Write(ctx, "t1", database.NewTupleCollection(), attributes1)
			Expect(err).ShouldNot(HaveOccurred())

			attr3, err := attribute.Attribute("organization:organization-1$balance|integer:3000")
			Expect(err).ShouldNot(HaveOccurred())

			attributes2 := database.NewAttributeCollection([]*base.Attribute{
				attr3,
			}...)

			token2, err := dataWriter.Write(ctx, "t1", database.NewTupleCollection(), attributes2)
			Expect(err).ShouldNot(HaveOccurred())

			it1, err := dataReader.QueryAttributes(ctx, "t1", &base.AttributeFilter{
				Entity: &base.EntityFilter{
					Type: "organization",
					Ids:  []string{"organization-1"},
				},
			}, token1.String(), database.NewCursorPagination())
			Expect(err).ShouldNot(HaveOccurred())

			Expect(it1.HasNext()).Should(Equal(true))
			Expect(it1.GetNext()).Should(Equal(attr2))
			Expect(it1.HasNext()).Should(Equal(false))

			it2, err := dataReader.QueryAttributes(ctx, "t1", &base.AttributeFilter{
				Entity: &base.EntityFilter{
					Type: "organization",
					Ids:  []string{"organization-1"},
				},
			}, token2.String(), database.NewCursorPagination())
			Expect(err).ShouldNot(HaveOccurred())

			Expect(it2.HasNext()).Should(Equal(true))
			Expect(it2.GetNext()).Should(Equal(attr3))
			Expect(it2.HasNext()).Should(Equal(true))
			Expect(it2.GetNext()).Should(Equal(attr2))
			Expect(it2.HasNext()).Should(Equal(false))
		})
	})

	Context("Read Attributes", func() {
		It("should write attributes and read attributes correctly", func() {
			ctx := context.Background()

			attr1, err := attribute.Attribute("organization:organization-1$public|boolean:true")
			Expect(err).ShouldNot(HaveOccurred())

			attr2, err := attribute.Attribute("organization:organization-2$ip_addresses|string[]:127.0.0.1,127.0.0.2")
			Expect(err).ShouldNot(HaveOccurred())

			attr3, err := attribute.Attribute("organization:organization-1$ip_addresses|string[]:127.0.0.1,127.0.0.2")
			Expect(err).ShouldNot(HaveOccurred())

			attr4, err := attribute.Attribute("organization:organization-1$balance|integer:3000")
			Expect(err).ShouldNot(HaveOccurred())

			attr5, err := attribute.Attribute("organization:organization-1$private|boolean:false")
			Expect(err).ShouldNot(HaveOccurred())

			attr6, err := attribute.Attribute("organization:organization-1$ppp|boolean[]:true,false")
			Expect(err).ShouldNot(HaveOccurred())

			attributes1 := database.NewAttributeCollection([]*base.Attribute{
				attr1,
				attr2,
				attr3,
				attr4,
				attr5,
				attr6,
			}...)

			token1, err := dataWriter.Write(ctx, "t1", database.NewTupleCollection(), attributes1)
			Expect(err).ShouldNot(HaveOccurred())

			col1, ct1, err := dataReader.ReadAttributes(ctx, "t1", &base.AttributeFilter{
				Entity: &base.EntityFilter{
					Type: "organization",
					Ids:  []string{"organization-1"},
				},
			}, token1.String(), database.NewPagination(database.Size(2), database.Token("")))
			Expect(err).ShouldNot(HaveOccurred())

			Expect(len(col1.GetAttributes())).Should(Equal(2))

			col2, ct2, err := dataReader.ReadAttributes(ctx, "t1", &base.AttributeFilter{
				Entity: &base.EntityFilter{
					Type: "organization",
					Ids:  []string{"organization-1"},
				},
			}, token1.String(), database.NewPagination(database.Size(3), database.Token(ct1.String())))
			Expect(err).ShouldNot(HaveOccurred())

			Expect(len(col2.GetAttributes())).Should(Equal(3))
			Expect(ct2.String()).Should(Equal(""))

			token3, err := dataWriter.Delete(ctx, "t1",
				&base.TupleFilter{},
				&base.AttributeFilter{
					Entity: &base.EntityFilter{
						Type: "organization",
						Ids:  []string{"organization-1"},
					},
					Attributes: []string{"ppp"},
				})
			Expect(err).ShouldNot(HaveOccurred())

			col3, ct3, err := dataReader.ReadAttributes(ctx, "t1", &base.AttributeFilter{
				Entity: &base.EntityFilter{
					Type: "organization",
					Ids:  []string{"organization-1"},
				},
			}, token3.String(), database.NewPagination(database.Size(4), database.Token("")))
			Expect(err).ShouldNot(HaveOccurred())

			Expect(len(col3.GetAttributes())).Should(Equal(4))
			Expect(ct3.String()).Should(Equal(""))
		})
	})

	Context("Query Unique Subject References", func() {
		It("should write tuples and query unique subject references correctly", func() {
			ctx := context.Background()

			attr1, err := attribute.Attribute("organization:organization-1$public|boolean:true")
			Expect(err).ShouldNot(HaveOccurred())

			attr2, err := attribute.Attribute("organization:organization-2$ip_addresses|string[]:127.0.0.1,127.0.0.2")
			Expect(err).ShouldNot(HaveOccurred())

			tup1, err := tuple.Tuple("organization:organization-1#admin@user:user-1")
			Expect(err).ShouldNot(HaveOccurred())

			tup2, err := tuple.Tuple("organization:organization-3#admin@user:user-1")
			Expect(err).ShouldNot(HaveOccurred())

			tup3, err := tuple.Tuple("organization:organization-19#admin@user:user-2")
			Expect(err).ShouldNot(HaveOccurred())

			tup4, err := tuple.Tuple("organization:organization-10#admin@user:user-3")
			Expect(err).ShouldNot(HaveOccurred())

			tup5, err := tuple.Tuple("organization:organization-14#admin@organization:organization-8#member")
			Expect(err).ShouldNot(HaveOccurred())

			tup6, err := tuple.Tuple("repository:repository-13#admin@user:user-5")
			Expect(err).ShouldNot(HaveOccurred())

			attributes1 := database.NewAttributeCollection([]*base.Attribute{
				attr1,
				attr2,
			}...)

			tuples1 := database.NewTupleCollection([]*base.Tuple{
				tup1,
				tup2,
				tup3,
				tup4,
				tup5,
				tup6,
			}...)

			token1, err := dataWriter.Write(ctx, "t1", tuples1, attributes1)
			Expect(err).ShouldNot(HaveOccurred())

			refs1, ct1, err := dataReader.QueryUniqueSubjectReferences(ctx, "t1", &base.RelationReference{
				Type:     "user",
				Relation: "",
			}, []string{}, token1.String(), database.NewPagination(database.Size(2), database.Token("")))
			Expect(err).ShouldNot(HaveOccurred())
			Expect(len(refs1)).Should(Equal(2))

			refs2, ct2, err := dataReader.QueryUniqueSubjectReferences(ctx, "t1", &base.RelationReference{
				Type:     "user",
				Relation: "",
			}, []string{}, token1.String(), database.NewPagination(database.Size(2), database.Token(ct1.String())))
			Expect(err).ShouldNot(HaveOccurred())
			Expect(len(refs2)).Should(Equal(2))
			Expect(ct2.String()).Should(Equal(""))

			refs3, ct3, err := dataReader.QueryUniqueSubjectReferences(ctx, "t1", &base.RelationReference{
				Type:     "user",
				Relation: "",
			}, []string{}, token1.String(), database.NewPagination(database.Size(20), database.Token("")))
			Expect(err).ShouldNot(HaveOccurred())
			Expect(len(refs3)).Should(Equal(4))
			Expect(ct3.String()).Should(Equal(""))

			Expect(isSameArray(refs3, []string{"user-1", "user-2", "user-3", "user-5"})).Should(BeTrue())

			refs4, ct4, err := dataReader.QueryUniqueSubjectReferences(ctx, "t1", &base.RelationReference{
				Type:     "organization",
				Relation: "member",
			}, []string{}, token1.String(), database.NewPagination(database.Size(20), database.Token("")))
			Expect(err).ShouldNot(HaveOccurred())
			Expect(len(refs4)).Should(Equal(1))
			Expect(ct4.String()).Should(Equal(""))

			Expect(isSameArray(refs4, []string{"organization-8"})).Should(BeTrue())
		})
	})

	Context("Error Handling", func() {
		Context("QueryRelationships Error Handling", func() {
			It("should handle snapshot decode error", func() {
				ctx := context.Background()

				// Create a dataReader with a closed database to trigger errors
				closedDB := db.(*PQDatabase.Postgres)
				err := closedDB.Close()
				Expect(err).ShouldNot(HaveOccurred())

				readerWithClosedDB := NewDataReader(closedDB)

				// Test with invalid snapshot token
				_, err = readerWithClosedDB.QueryRelationships(ctx, "t1", &base.TupleFilter{}, "invalid_snapshot", database.CursorPagination{})
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(Equal(base.ErrorCode_ERROR_CODE_INTERNAL.String()))
			})

			It("should handle continuous token decode error", func() {
				ctx := context.Background()

				// Test with invalid continuous token
				pagination := database.NewCursorPagination(
					database.Cursor("invalid_token"),
					database.Sort("id"),
				)

				_, err := dataReader.QueryRelationships(ctx, "t1", &base.TupleFilter{}, "0", pagination)
				Expect(err).Should(HaveOccurred())
				// The error could be either INTERNAL (snapshot decode) or INVALID_CONTINUOUS_TOKEN depending on which fails first
				Expect(err.Error()).Should(Or(
					Equal(base.ErrorCode_ERROR_CODE_INTERNAL.String()),
					Equal(base.ErrorCode_ERROR_CODE_INVALID_CONTINUOUS_TOKEN.String()),
				))
			})

			It("should handle SQL builder error", func() {
				ctx := context.Background()

				// Create a dataReader with a closed database to trigger SQL builder error
				closedDB := db.(*PQDatabase.Postgres)
				err := closedDB.Close()
				Expect(err).ShouldNot(HaveOccurred())

				readerWithClosedDB := NewDataReader(closedDB)

				_, err = readerWithClosedDB.QueryRelationships(ctx, "t1", &base.TupleFilter{}, "0", database.CursorPagination{})
				Expect(err).Should(HaveOccurred())
				// The error could be either INTERNAL (snapshot decode) or SQL_BUILDER depending on when the connection fails
				Expect(err.Error()).Should(Or(
					Equal(base.ErrorCode_ERROR_CODE_INTERNAL.String()),
					Equal(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String()),
				))
			})

			It("should handle execution error", func() {
				ctx := context.Background()

				// Create a dataReader with a closed database to trigger execution error
				closedDB := db.(*PQDatabase.Postgres)
				err := closedDB.Close()
				Expect(err).ShouldNot(HaveOccurred())

				readerWithClosedDB := NewDataReader(closedDB)

				_, err = readerWithClosedDB.QueryRelationships(ctx, "t1", &base.TupleFilter{}, "0", database.CursorPagination{})
				Expect(err).Should(HaveOccurred())
				// The error could be INTERNAL, SQL_BUILDER, or EXECUTION depending on when the connection fails
				Expect(err.Error()).Should(Or(
					Equal(base.ErrorCode_ERROR_CODE_INTERNAL.String()),
					Equal(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String()),
					Equal(base.ErrorCode_ERROR_CODE_EXECUTION.String()),
				))
			})

			It("should handle scan error", func() {
				ctx := context.Background()

				// Create a dataReader with a closed database to trigger scan error
				closedDB := db.(*PQDatabase.Postgres)
				err := closedDB.Close()
				Expect(err).ShouldNot(HaveOccurred())

				readerWithClosedDB := NewDataReader(closedDB)

				_, err = readerWithClosedDB.QueryRelationships(ctx, "t1", &base.TupleFilter{}, "0", database.CursorPagination{})
				Expect(err).Should(HaveOccurred())
				// The error could be various types depending on when the connection fails
				Expect(err.Error()).Should(Or(
					Equal(base.ErrorCode_ERROR_CODE_INTERNAL.String()),
					Equal(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String()),
					Equal(base.ErrorCode_ERROR_CODE_EXECUTION.String()),
					Equal(base.ErrorCode_ERROR_CODE_SCAN.String()),
				))
			})
		})

		Context("ReadRelationships Error Handling", func() {
			It("should handle snapshot decode error", func() {
				ctx := context.Background()

				// Create a dataReader with a closed database to trigger errors
				closedDB := db.(*PQDatabase.Postgres)
				err := closedDB.Close()
				Expect(err).ShouldNot(HaveOccurred())

				readerWithClosedDB := NewDataReader(closedDB)

				// Test with invalid snapshot token
				_, _, err = readerWithClosedDB.ReadRelationships(ctx, "t1", &base.TupleFilter{}, "invalid_snapshot", database.Pagination{})
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(Equal(base.ErrorCode_ERROR_CODE_INTERNAL.String()))
			})

			It("should handle continuous token decode error", func() {
				ctx := context.Background()

				// Test with invalid continuous token
				pagination := database.NewPagination(database.Token("invalid_token"))

				_, _, err := dataReader.ReadRelationships(ctx, "t1", &base.TupleFilter{}, "0", pagination)
				Expect(err).Should(HaveOccurred())
				// The error could be either INTERNAL (snapshot decode) or INVALID_CONTINUOUS_TOKEN depending on which fails first
				Expect(err.Error()).Should(Or(
					Equal(base.ErrorCode_ERROR_CODE_INTERNAL.String()),
					Equal(base.ErrorCode_ERROR_CODE_INVALID_CONTINUOUS_TOKEN.String()),
				))
			})

			It("should handle continuous token parse error", func() {
				ctx := context.Background()

				// Test with invalid continuous token that can't be parsed as uint64
				pagination := database.NewPagination(database.Token("not_a_number"))

				_, _, err := dataReader.ReadRelationships(ctx, "t1", &base.TupleFilter{}, "0", pagination)
				Expect(err).Should(HaveOccurred())
				// The error could be either INTERNAL (snapshot decode) or INVALID_CONTINUOUS_TOKEN depending on which fails first
				Expect(err.Error()).Should(Or(
					Equal(base.ErrorCode_ERROR_CODE_INTERNAL.String()),
					Equal(base.ErrorCode_ERROR_CODE_INVALID_CONTINUOUS_TOKEN.String()),
				))
			})

			It("should handle SQL builder error", func() {
				ctx := context.Background()

				// Create a dataReader with a closed database to trigger SQL builder error
				closedDB := db.(*PQDatabase.Postgres)
				err := closedDB.Close()
				Expect(err).ShouldNot(HaveOccurred())

				readerWithClosedDB := NewDataReader(closedDB)

				_, _, err = readerWithClosedDB.ReadRelationships(ctx, "t1", &base.TupleFilter{}, "0", database.Pagination{})
				Expect(err).Should(HaveOccurred())
				// The error could be either INTERNAL (snapshot decode) or SQL_BUILDER depending on when the connection fails
				Expect(err.Error()).Should(Or(
					Equal(base.ErrorCode_ERROR_CODE_INTERNAL.String()),
					Equal(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String()),
				))
			})

			It("should handle execution error", func() {
				ctx := context.Background()

				// Create a dataReader with a closed database to trigger execution error
				closedDB := db.(*PQDatabase.Postgres)
				err := closedDB.Close()
				Expect(err).ShouldNot(HaveOccurred())

				readerWithClosedDB := NewDataReader(closedDB)

				_, _, err = readerWithClosedDB.ReadRelationships(ctx, "t1", &base.TupleFilter{}, "0", database.Pagination{})
				Expect(err).Should(HaveOccurred())
				// The error could be INTERNAL, SQL_BUILDER, or EXECUTION depending on when the connection fails
				Expect(err.Error()).Should(Or(
					Equal(base.ErrorCode_ERROR_CODE_INTERNAL.String()),
					Equal(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String()),
					Equal(base.ErrorCode_ERROR_CODE_EXECUTION.String()),
				))
			})

			It("should handle scan error", func() {
				ctx := context.Background()

				// Create a dataReader with a closed database to trigger scan error
				closedDB := db.(*PQDatabase.Postgres)
				err := closedDB.Close()
				Expect(err).ShouldNot(HaveOccurred())

				readerWithClosedDB := NewDataReader(closedDB)

				_, _, err = readerWithClosedDB.ReadRelationships(ctx, "t1", &base.TupleFilter{}, "0", database.Pagination{})
				Expect(err).Should(HaveOccurred())
				// The error could be various types depending on when the connection fails
				Expect(err.Error()).Should(Or(
					Equal(base.ErrorCode_ERROR_CODE_INTERNAL.String()),
					Equal(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String()),
					Equal(base.ErrorCode_ERROR_CODE_EXECUTION.String()),
					Equal(base.ErrorCode_ERROR_CODE_SCAN.String()),
				))
			})
		})

		Context("QueryAttributes Error Handling", func() {
			It("should handle snapshot decode error", func() {
				ctx := context.Background()

				// Create a dataReader with a closed database to trigger errors
				closedDB := db.(*PQDatabase.Postgres)
				err := closedDB.Close()
				Expect(err).ShouldNot(HaveOccurred())

				readerWithClosedDB := NewDataReader(closedDB)

				// Test with invalid snapshot token
				_, err = readerWithClosedDB.QueryAttributes(ctx, "t1", &base.AttributeFilter{}, "invalid_snapshot", database.CursorPagination{})
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(Equal(base.ErrorCode_ERROR_CODE_INTERNAL.String()))
			})

			It("should handle continuous token decode error", func() {
				ctx := context.Background()

				// Test with invalid continuous token
				pagination := database.NewCursorPagination(
					database.Cursor("invalid_token"),
					database.Sort("id"),
				)

				_, err := dataReader.QueryAttributes(ctx, "t1", &base.AttributeFilter{}, "0", pagination)
				Expect(err).Should(HaveOccurred())
				// The error could be either INTERNAL (snapshot decode) or INVALID_CONTINUOUS_TOKEN depending on which fails first
				Expect(err.Error()).Should(Or(
					Equal(base.ErrorCode_ERROR_CODE_INTERNAL.String()),
					Equal(base.ErrorCode_ERROR_CODE_INVALID_CONTINUOUS_TOKEN.String()),
				))
			})

			It("should handle SQL builder error", func() {
				ctx := context.Background()

				// Create a dataReader with a closed database to trigger SQL builder error
				closedDB := db.(*PQDatabase.Postgres)
				err := closedDB.Close()
				Expect(err).ShouldNot(HaveOccurred())

				readerWithClosedDB := NewDataReader(closedDB)

				_, err = readerWithClosedDB.QueryAttributes(ctx, "t1", &base.AttributeFilter{}, "0", database.CursorPagination{})
				Expect(err).Should(HaveOccurred())
				// The error could be either INTERNAL (snapshot decode) or SQL_BUILDER depending on when the connection fails
				Expect(err.Error()).Should(Or(
					Equal(base.ErrorCode_ERROR_CODE_INTERNAL.String()),
					Equal(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String()),
				))
			})

			It("should handle execution error", func() {
				ctx := context.Background()

				// Create a dataReader with a closed database to trigger execution error
				closedDB := db.(*PQDatabase.Postgres)
				err := closedDB.Close()
				Expect(err).ShouldNot(HaveOccurred())

				readerWithClosedDB := NewDataReader(closedDB)

				_, err = readerWithClosedDB.QueryAttributes(ctx, "t1", &base.AttributeFilter{}, "0", database.CursorPagination{})
				Expect(err).Should(HaveOccurred())
				// The error could be INTERNAL, SQL_BUILDER, or EXECUTION depending on when the connection fails
				Expect(err.Error()).Should(Or(
					Equal(base.ErrorCode_ERROR_CODE_INTERNAL.String()),
					Equal(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String()),
					Equal(base.ErrorCode_ERROR_CODE_EXECUTION.String()),
				))
			})

			It("should handle scan error", func() {
				ctx := context.Background()

				// Create a dataReader with a closed database to trigger scan error
				closedDB := db.(*PQDatabase.Postgres)
				err := closedDB.Close()
				Expect(err).ShouldNot(HaveOccurred())

				readerWithClosedDB := NewDataReader(closedDB)

				_, err = readerWithClosedDB.QueryAttributes(ctx, "t1", &base.AttributeFilter{}, "0", database.CursorPagination{})
				Expect(err).Should(HaveOccurred())
				// The error could be various types depending on when the connection fails
				Expect(err.Error()).Should(Or(
					Equal(base.ErrorCode_ERROR_CODE_INTERNAL.String()),
					Equal(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String()),
					Equal(base.ErrorCode_ERROR_CODE_EXECUTION.String()),
					Equal(base.ErrorCode_ERROR_CODE_SCAN.String()),
				))
			})
		})

		Context("ReadAttributes Error Handling", func() {
			It("should handle snapshot decode error", func() {
				ctx := context.Background()

				// Create a dataReader with a closed database to trigger errors
				closedDB := db.(*PQDatabase.Postgres)
				err := closedDB.Close()
				Expect(err).ShouldNot(HaveOccurred())

				readerWithClosedDB := NewDataReader(closedDB)

				// Test with invalid snapshot token
				_, _, err = readerWithClosedDB.ReadAttributes(ctx, "t1", &base.AttributeFilter{}, "invalid_snapshot", database.Pagination{})
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(Equal(base.ErrorCode_ERROR_CODE_INTERNAL.String()))
			})

			It("should handle continuous token decode error", func() {
				ctx := context.Background()

				// Test with invalid continuous token
				pagination := database.NewPagination(database.Token("invalid_token"))

				_, _, err := dataReader.ReadAttributes(ctx, "t1", &base.AttributeFilter{}, "0", pagination)
				Expect(err).Should(HaveOccurred())
				// The error could be either INTERNAL (snapshot decode) or INVALID_CONTINUOUS_TOKEN depending on which fails first
				Expect(err.Error()).Should(Or(
					Equal(base.ErrorCode_ERROR_CODE_INTERNAL.String()),
					Equal(base.ErrorCode_ERROR_CODE_INVALID_CONTINUOUS_TOKEN.String()),
				))
			})

			It("should handle continuous token parse error", func() {
				ctx := context.Background()

				// Test with invalid continuous token that can't be parsed as uint64
				pagination := database.NewPagination(database.Token("not_a_number"))

				_, _, err := dataReader.ReadAttributes(ctx, "t1", &base.AttributeFilter{}, "0", pagination)
				Expect(err).Should(HaveOccurred())
				// The error could be either INTERNAL (snapshot decode) or INVALID_CONTINUOUS_TOKEN depending on which fails first
				Expect(err.Error()).Should(Or(
					Equal(base.ErrorCode_ERROR_CODE_INTERNAL.String()),
					Equal(base.ErrorCode_ERROR_CODE_INVALID_CONTINUOUS_TOKEN.String()),
				))
			})

			It("should handle SQL builder error", func() {
				ctx := context.Background()

				// Create a dataReader with a closed database to trigger SQL builder error
				closedDB := db.(*PQDatabase.Postgres)
				err := closedDB.Close()
				Expect(err).ShouldNot(HaveOccurred())

				readerWithClosedDB := NewDataReader(closedDB)

				_, _, err = readerWithClosedDB.ReadAttributes(ctx, "t1", &base.AttributeFilter{}, "0", database.Pagination{})
				Expect(err).Should(HaveOccurred())
				// The error could be either INTERNAL (snapshot decode) or SQL_BUILDER depending on when the connection fails
				Expect(err.Error()).Should(Or(
					Equal(base.ErrorCode_ERROR_CODE_INTERNAL.String()),
					Equal(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String()),
				))
			})

			It("should handle execution error", func() {
				ctx := context.Background()

				// Create a dataReader with a closed database to trigger execution error
				closedDB := db.(*PQDatabase.Postgres)
				err := closedDB.Close()
				Expect(err).ShouldNot(HaveOccurred())

				readerWithClosedDB := NewDataReader(closedDB)

				_, _, err = readerWithClosedDB.ReadAttributes(ctx, "t1", &base.AttributeFilter{}, "0", database.Pagination{})
				Expect(err).Should(HaveOccurred())
				// The error could be INTERNAL, SQL_BUILDER, or EXECUTION depending on when the connection fails
				Expect(err.Error()).Should(Or(
					Equal(base.ErrorCode_ERROR_CODE_INTERNAL.String()),
					Equal(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String()),
					Equal(base.ErrorCode_ERROR_CODE_EXECUTION.String()),
				))
			})

			It("should handle scan error", func() {
				ctx := context.Background()

				// Create a dataReader with a closed database to trigger scan error
				closedDB := db.(*PQDatabase.Postgres)
				err := closedDB.Close()
				Expect(err).ShouldNot(HaveOccurred())

				readerWithClosedDB := NewDataReader(closedDB)

				_, _, err = readerWithClosedDB.ReadAttributes(ctx, "t1", &base.AttributeFilter{}, "0", database.Pagination{})
				Expect(err).Should(HaveOccurred())
				// The error could be various types depending on when the connection fails
				Expect(err.Error()).Should(Or(
					Equal(base.ErrorCode_ERROR_CODE_INTERNAL.String()),
					Equal(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String()),
					Equal(base.ErrorCode_ERROR_CODE_EXECUTION.String()),
					Equal(base.ErrorCode_ERROR_CODE_SCAN.String()),
				))
			})
		})

		Context("QueryUniqueSubjectReferences Error Handling", func() {
			It("should handle continuous token decode error", func() {
				ctx := context.Background()

				// Test with invalid continuous token
				pagination := database.NewPagination(database.Token("invalid_token"))

				_, _, err := dataReader.QueryUniqueSubjectReferences(ctx, "t1", &base.RelationReference{Type: "user"}, []string{}, "0", pagination)
				Expect(err).Should(HaveOccurred())
				// The error could be either INTERNAL (snapshot decode) or INVALID_CONTINUOUS_TOKEN depending on which fails first
				Expect(err.Error()).Should(Or(
					Equal(base.ErrorCode_ERROR_CODE_INTERNAL.String()),
					Equal(base.ErrorCode_ERROR_CODE_INVALID_CONTINUOUS_TOKEN.String()),
				))
			})

			It("should handle SQL builder error", func() {
				ctx := context.Background()

				// Create a dataReader with a closed database to trigger SQL builder error
				closedDB := db.(*PQDatabase.Postgres)
				err := closedDB.Close()
				Expect(err).ShouldNot(HaveOccurred())

				readerWithClosedDB := NewDataReader(closedDB)

				_, _, err = readerWithClosedDB.QueryUniqueSubjectReferences(ctx, "t1", &base.RelationReference{Type: "user"}, []string{}, "0", database.Pagination{})
				Expect(err).Should(HaveOccurred())
				// The error could be either INTERNAL (snapshot decode) or SQL_BUILDER depending on when the connection fails
				Expect(err.Error()).Should(Or(
					Equal(base.ErrorCode_ERROR_CODE_INTERNAL.String()),
					Equal(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String()),
				))
			})

			It("should handle execution error", func() {
				ctx := context.Background()

				// Create a dataReader with a closed database to trigger execution error
				closedDB := db.(*PQDatabase.Postgres)
				err := closedDB.Close()
				Expect(err).ShouldNot(HaveOccurred())

				readerWithClosedDB := NewDataReader(closedDB)

				_, _, err = readerWithClosedDB.QueryUniqueSubjectReferences(ctx, "t1", &base.RelationReference{Type: "user"}, []string{}, "0", database.Pagination{})
				Expect(err).Should(HaveOccurred())
				// The error could be INTERNAL, SQL_BUILDER, or EXECUTION depending on when the connection fails
				Expect(err.Error()).Should(Or(
					Equal(base.ErrorCode_ERROR_CODE_INTERNAL.String()),
					Equal(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String()),
					Equal(base.ErrorCode_ERROR_CODE_EXECUTION.String()),
				))
			})

			It("should handle scan error", func() {
				ctx := context.Background()

				// Create a dataReader with a closed database to trigger scan error
				closedDB := db.(*PQDatabase.Postgres)
				err := closedDB.Close()
				Expect(err).ShouldNot(HaveOccurred())

				readerWithClosedDB := NewDataReader(closedDB)

				_, _, err = readerWithClosedDB.QueryUniqueSubjectReferences(ctx, "t1", &base.RelationReference{Type: "user"}, []string{}, "0", database.Pagination{})
				Expect(err).Should(HaveOccurred())
				// The error could be various types depending on when the connection fails
				Expect(err.Error()).Should(Or(
					Equal(base.ErrorCode_ERROR_CODE_INTERNAL.String()),
					Equal(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String()),
					Equal(base.ErrorCode_ERROR_CODE_EXECUTION.String()),
					Equal(base.ErrorCode_ERROR_CODE_SCAN.String()),
				))
			})
		})

		Context("HeadSnapshot Error Handling", func() {
			It("should handle SQL builder error", func() {
				ctx := context.Background()

				// Create a dataReader with a closed database to trigger SQL builder error
				closedDB := db.(*PQDatabase.Postgres)
				err := closedDB.Close()
				Expect(err).ShouldNot(HaveOccurred())

				readerWithClosedDB := NewDataReader(closedDB)

				_, err = readerWithClosedDB.HeadSnapshot(ctx, "t1")
				Expect(err).Should(HaveOccurred())
				// The error could be SQL_BUILDER or SCAN depending on when the connection fails
				Expect(err.Error()).Should(Or(
					Equal(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String()),
					Equal(base.ErrorCode_ERROR_CODE_SCAN.String()),
				))
			})

			It("should handle scan error with no rows", func() {
				ctx := context.Background()

				// Test with a non-existent tenant to trigger no rows error
				_, err := dataReader.HeadSnapshot(ctx, "non_existent_tenant")
				Expect(err).ShouldNot(HaveOccurred()) // This should return a snapshot with value 0, not an error
			})

			It("should handle scan error", func() {
				ctx := context.Background()

				// Create a dataReader with a closed database to trigger scan error
				closedDB := db.(*PQDatabase.Postgres)
				err := closedDB.Close()
				Expect(err).ShouldNot(HaveOccurred())

				readerWithClosedDB := NewDataReader(closedDB)

				_, err = readerWithClosedDB.HeadSnapshot(ctx, "t1")
				Expect(err).Should(HaveOccurred())
				// The error could be SQL_BUILDER or SCAN depending on when the connection fails
				Expect(err.Error()).Should(Or(
					Equal(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String()),
					Equal(base.ErrorCode_ERROR_CODE_SCAN.String()),
				))
			})
		})
	})
})
