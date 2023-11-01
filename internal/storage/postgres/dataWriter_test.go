package postgres

import (
	"context"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/pkg/attribute"
	"github.com/Permify/permify/pkg/database"
	PQDatabase "github.com/Permify/permify/pkg/database/postgres"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
)

var _ = Describe("DataWriter", func() {
	var db database.Database
	var dataWriter *DataWriter
	var dataReader *DataReader

	BeforeEach(func() {
		version := os.Getenv("POSTGRES_VERSION")

		if version == "" {
			version = "14"
		}

		db = postgresDB(version)
		dataWriter = NewDataWriter(db.(*PQDatabase.Postgres))
		dataReader = NewDataReader(db.(*PQDatabase.Postgres))
	})

	AfterEach(func() {
		err := db.Close()
		Expect(err).ShouldNot(HaveOccurred())
	})

	Context("Write", func() {
		It("The test case verifies that an attribute's value for an entity can be updated and subsequently retrieved correctly using MVCC tokens", func() {
			ctx := context.Background()

			attr1, err := attribute.Attribute("organization:organization-1$public|boolean:true")
			Expect(err).ShouldNot(HaveOccurred())

			tup1, err := tuple.Tuple("organization:organization-1#admin@user:user-1")
			Expect(err).ShouldNot(HaveOccurred())

			attributes1 := database.NewAttributeCollection([]*base.Attribute{
				attr1,
			}...)

			tuples1 := database.NewTupleCollection([]*base.Tuple{
				tup1,
			}...)

			token1, err := dataWriter.Write(ctx, "t1", tuples1, attributes1)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(token1.String()).ShouldNot(BeNil())

			attrRes1, err := dataReader.QuerySingleAttribute(ctx, "t1", &base.AttributeFilter{
				Entity: &base.EntityFilter{
					Type: "organization",
					Ids:  []string{"organization-1"},
				},
				Attributes: []string{"public"},
			}, token1.String())
			Expect(err).ShouldNot(HaveOccurred())

			var msg1 base.BooleanValue
			err = attrRes1.GetValue().UnmarshalTo(&msg1)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(msg1.GetData()).Should(Equal(true))

			attr2, err := attribute.Attribute("organization:organization-1$public|boolean:false")
			Expect(err).ShouldNot(HaveOccurred())

			attributes2 := database.NewAttributeCollection([]*base.Attribute{
				attr2,
			}...)

			token2, err := dataWriter.Write(ctx, "t1", database.NewTupleCollection(), attributes2)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(token2.String()).ShouldNot(Equal(""))

			attrRes2, err := dataReader.QuerySingleAttribute(ctx, "t1", &base.AttributeFilter{
				Entity: &base.EntityFilter{
					Type: "organization",
					Ids:  []string{"organization-1"},
				},
				Attributes: []string{"public"},
			}, token2.String())
			Expect(err).ShouldNot(HaveOccurred())

			var msg2 base.BooleanValue
			err = attrRes2.GetValue().UnmarshalTo(&msg2)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(msg2.GetData()).Should(Equal(false))
		})

		It("should write attributes and tuples correctly", func() {
			ctx := context.Background()

			attr1, err := attribute.Attribute("organization:organization-1$public|boolean:true")
			Expect(err).ShouldNot(HaveOccurred())

			attr2, err := attribute.Attribute("organization:organization-2$ip_addresses|string[]:127.0.0.1,127.0.0.2")
			Expect(err).ShouldNot(HaveOccurred())

			attr3, err := attribute.Attribute("organization:organization-3$ip_addresses|double:234.344")
			Expect(err).ShouldNot(HaveOccurred())

			attr4, err := attribute.Attribute("organization:organization-16$balance|integer:3000")
			Expect(err).ShouldNot(HaveOccurred())

			attr5, err := attribute.Attribute("organization:organization-28$private|boolean:false")
			Expect(err).ShouldNot(HaveOccurred())

			attr6, err := attribute.Attribute("organization:organization-17$ppp|boolean[]:true,false")
			Expect(err).ShouldNot(HaveOccurred())

			attr7, err := attribute.Attribute("organization:organization-1$ip_addresses|integer[]:167,878")
			Expect(err).ShouldNot(HaveOccurred())

			tup1, err := tuple.Tuple("organization:organization-1#admin@user:user-1")
			Expect(err).ShouldNot(HaveOccurred())

			tup2, err := tuple.Tuple("organization:organization-28#admin@user:user-1")
			Expect(err).ShouldNot(HaveOccurred())

			tup3, err := tuple.Tuple("organization:organization-19#admin@user:user-2")
			Expect(err).ShouldNot(HaveOccurred())

			tup4, err := tuple.Tuple("organization:organization-10#admin@user:user-3")
			Expect(err).ShouldNot(HaveOccurred())

			tup5, err := tuple.Tuple("organization:organization-14#admin@user:user-4")
			Expect(err).ShouldNot(HaveOccurred())

			tup6, err := tuple.Tuple("repository:repository-13#admin@user:user-5")
			Expect(err).ShouldNot(HaveOccurred())

			attributes1 := database.NewAttributeCollection([]*base.Attribute{
				attr1,
				attr2,
				attr3,
				attr4,
				attr5,
				attr6,
				attr7,
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
			Expect(token1.String()).ShouldNot(Equal(""))
		})

		It("should write empty attributes and empty tuples correctly", func() {
			ctx := context.Background()
			token1, err := dataWriter.Write(ctx, "t1", database.NewTupleCollection(), database.NewAttributeCollection())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(token1.String()).ShouldNot(Equal(""))
		})
	})

	Context("Delete", func() {
		It("should delete, read relationships and read attributes correctly", func() {
			ctx := context.Background()

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

			token1, err := dataWriter.Write(ctx, "t1", tuples1, attributes1)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(token1.String()).ShouldNot(Equal(""))

			col1, ct1, err := dataReader.ReadRelationships(ctx, "t1", &base.TupleFilter{
				Entity: &base.EntityFilter{
					Type: "organization",
					Ids:  []string{"organization-1"},
				},
			}, token1.String(), database.NewPagination(database.Size(10), database.Token("")))
			Expect(err).ShouldNot(HaveOccurred())
			Expect(ct1.String()).Should(Equal(""))
			Expect(len(col1.GetTuples())).Should(Equal(3))

			col2, ct2, err := dataReader.ReadAttributes(ctx, "t1", &base.AttributeFilter{
				Entity: &base.EntityFilter{
					Type: "organization",
					Ids:  []string{"organization-1"},
				},
			}, token1.String(), database.NewPagination(database.Size(10), database.Token("")))
			Expect(err).ShouldNot(HaveOccurred())
			Expect(ct2.String()).Should(Equal(""))
			Expect(len(col2.GetAttributes())).Should(Equal(2))

			token2, err := dataWriter.Delete(ctx, "t1",
				&base.TupleFilter{
					Entity: &base.EntityFilter{
						Type: "organization",
						Ids:  []string{"organization-1"},
					},
					Relation: "admin",
					Subject: &base.SubjectFilter{
						Type: "user",
						Ids:  []string{"user-1"},
					},
				},
				&base.AttributeFilter{
					Entity: &base.EntityFilter{
						Type: "organization",
						Ids:  []string{"organization-1"},
					},
					Attributes: []string{"public"},
				})
			Expect(err).ShouldNot(HaveOccurred())

			col3, ct3, err := dataReader.ReadRelationships(ctx, "t1", &base.TupleFilter{
				Entity: &base.EntityFilter{
					Type: "organization",
					Ids:  []string{"organization-1"},
				},
			}, token2.String(), database.NewPagination(database.Size(10), database.Token("")))
			Expect(err).ShouldNot(HaveOccurred())
			Expect(ct3.String()).Should(Equal(""))
			Expect(len(col3.GetTuples())).Should(Equal(2))

			col4, ct5, err := dataReader.ReadAttributes(ctx, "t1", &base.AttributeFilter{
				Entity: &base.EntityFilter{
					Type: "organization",
					Ids:  []string{"organization-1"},
				},
			}, token2.String(), database.NewPagination(database.Size(10), database.Token("")))
			Expect(err).ShouldNot(HaveOccurred())
			Expect(ct5.String()).Should(Equal(""))
			Expect(len(col4.GetAttributes())).Should(Equal(1))
		})
	})
})
