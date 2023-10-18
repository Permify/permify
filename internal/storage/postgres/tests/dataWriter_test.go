//go:build !integration

package tests

import (
	"context"
	"database/sql"
	"regexp"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Masterminds/squirrel"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	postgres2 "github.com/Permify/permify/internal/storage/postgres"
	"github.com/Permify/permify/pkg/database"
	"github.com/Permify/permify/pkg/database/postgres"
	basev1 "github.com/Permify/permify/pkg/pb/base/v1"
)

var _ = Describe("RelationshipWriter", func() {
	var dataWriter *postgres2.DataWriter
	var mock sqlmock.Sqlmock

	BeforeEach(func() {
		var db *sql.DB
		var err error

		db, mock, err = sqlmock.New()
		Expect(err).ShouldNot(HaveOccurred())

		pg := &postgres.Postgres{
			DB:      db,
			Builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		}

		dataWriter = postgres2.NewDataWriter(pg)
	})

	AfterEach(func() {
		err := mock.ExpectationsWereMet()
		Expect(err).ShouldNot(HaveOccurred())
	})

	Context("Writes Relationships", func() {
		columns := []string{"entity_type", "entity_id", "relation", "subject_type", "subject_id", "subject_relation", "tenant_id"}

		It("Insert and throws no error", func() {
			mock.ExpectBegin()
			mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO relation_tuples (entity_type, entity_id, relation, subject_type, subject_id, subject_relation, tenant_id)
			VALUES ($1,$2,$3,$4,$5,$6,$7) ON CONFLICT DO NOTHING`)).
				WithArgs("organization", "abc", "admin", "subject-1", "sub-id", "admin", "noop").
				WillReturnRows(
					sqlmock.NewRows(columns).AddRow("organization", "abc", "admin", "subject-1", "sub-id", "admin", "noop"),
				)
			mock.ExpectCommit()
			tp := &database.TupleCollection{}
			tp.Add(&basev1.Tuple{
				Entity:   &basev1.Entity{Type: "organization", Id: "abc"},
				Relation: "admin",
				Subject:  &basev1.Subject{Type: "subject-1", Id: "sub-id", Relation: "admin-sub"},
			})
			_, err := dataWriter.Write(context.Background(), "noop", tp, &database.AttributeCollection{})

			Expect(err).ShouldNot(HaveOccurred())
		})

		It("Insert and compares", func() {
			mock.ExpectBegin()
			mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO relation_tuples (entity_type, entity_id, relation, subject_type, subject_id, subject_relation, tenant_id)
			VALUES ($1,$2,$3,$4,$5,$6,$7)`)).
				WithArgs("organization", "abc", "admin", "subject-1", "sub-id", "admin-sub", "noop").
				WillReturnRows(
					sqlmock.NewRows(columns).AddRow("organization", "abc", "admin", "subject-1", "sub-id", "admin-sub", "noop"),
				)
			mock.ExpectCommit()
			tp := &database.TupleCollection{}
			tp.Add(&basev1.Tuple{
				Entity: &basev1.Entity{
					Type: "organization",
					Id:   "abc",
				},
				Subject: &basev1.Subject{
					Type:     "subject-1",
					Id:       "sub-id",
					Relation: "admin-sub",
				},
				Relation: "admin",
			})
			_, err := dataWriter.Write(context.Background(), "noop", tp, &database.AttributeCollection{})

			Expect(err).ShouldNot(HaveOccurred())

			// TODO: can we write a helper function to fetch the recently inserted record? as we are just creating a mock! any comments? Will think about it!
		})
	})
})
