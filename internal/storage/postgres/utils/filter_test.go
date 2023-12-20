package utils_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Masterminds/squirrel"

	"github.com/Permify/permify/internal/storage/postgres/utils"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

var _ = Describe("Filter", func() {
	Context("TestFilterQueryForSelectBuilder", func() {
		It("Case 1", func() {
			sl := squirrel.Select("*").From("relation_tuples")

			filter := &base.TupleFilter{
				Entity: &base.EntityFilter{
					Type: "entity_type",
					Ids:  []string{"1", "2"},
				},
				Relation: "relation",
				Subject: &base.SubjectFilter{
					Type:     "subject_type",
					Ids:      []string{"3", "4"},
					Relation: "subject_relation",
				},
			}

			sl = utils.TuplesFilterQueryForSelectBuilder(sl, filter)

			expectedSql := "SELECT * FROM relation_tuples WHERE entity_id IN (?,?) AND entity_type = ? AND relation = ? AND subject_id IN (?,?) AND subject_relation = ? AND subject_type = ?"
			expectedArgs := []interface{}{"1", "2", "entity_type", "relation", "3", "4", "subject_relation", "subject_type"}

			sql, args, err := sl.ToSql()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(args).Should(Equal(expectedArgs))
			Expect(sql).Should(Equal(expectedSql))
		})
	})

	Context("TestHalfEmptyFilterQueryForSelectBuilder", func() {
		It("Case 1", func() {
			sl := squirrel.Select("*").From("relation_tuples")

			filter := &base.TupleFilter{
				Entity: &base.EntityFilter{
					Type: "entity_type",
					Ids:  []string{"1", "2"},
				},
				Subject: &base.SubjectFilter{
					Type: "subject_type",
					Ids:  []string{"3", "4"},
				},
			}

			sl = utils.TuplesFilterQueryForSelectBuilder(sl, filter)

			expectedSql := "SELECT * FROM relation_tuples WHERE entity_id IN (?,?) AND entity_type = ? AND subject_id IN (?,?) AND subject_type = ?"
			expectedArgs := []interface{}{"1", "2", "entity_type", "3", "4", "subject_type"}

			sql, args, err := sl.ToSql()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(args).Should(Equal(expectedArgs))
			Expect(sql).Should(Equal(expectedSql))
		})
	})

	Context("TestEmptyFilterQueryForSelectBuilder", func() {
		It("Case 1", func() {
			sl := squirrel.Select("*").From("relation_tuples")

			filter := &base.TupleFilter{}

			sl = utils.TuplesFilterQueryForSelectBuilder(sl, filter)

			expectedSql := "SELECT * FROM relation_tuples"

			sql, _, err := sl.ToSql()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(sql).Should(Equal(expectedSql))

			filter = &base.TupleFilter{
				Entity: &base.EntityFilter{
					Type: "",
					Ids:  []string{},
				},
			}

			sl = utils.TuplesFilterQueryForSelectBuilder(sl, filter)

			expectedSql = "SELECT * FROM relation_tuples"

			sql, _, err = sl.ToSql()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(sql).Should(Equal(expectedSql))
		})
	})

	Context("TestEmptyFilterQueryForSelectBuilder", func() {
		It("Case 1", func() {
			sl := squirrel.Select("*").From("attributes")

			filter := &base.AttributeFilter{}

			sl = utils.AttributesFilterQueryForSelectBuilder(sl, filter)

			expectedSql := "SELECT * FROM attributes"

			sql, _, err := sl.ToSql()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(sql).Should(Equal(expectedSql))

			filter = &base.AttributeFilter{
				Entity: &base.EntityFilter{
					Type: "",
					Ids:  []string{},
				},
			}

			sl = utils.AttributesFilterQueryForSelectBuilder(sl, filter)

			expectedSql = "SELECT * FROM attributes"

			sql, _, err = sl.ToSql()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(sql).Should(Equal(expectedSql))
		})
	})

	Context("TuplesFilterQueryForUpdateBuilder", func() {
		It("Case 1", func() {
			sl := squirrel.Update("relation_tuples").Set("expired_tx_id", 1000)

			filter := &base.TupleFilter{}

			sl = utils.TuplesFilterQueryForUpdateBuilder(sl, filter)

			expectedSql := "UPDATE relation_tuples SET expired_tx_id = ?"

			sql, _, err := sl.ToSql()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(sql).Should(Equal(expectedSql))

			filter = &base.TupleFilter{
				Entity: &base.EntityFilter{
					Type: "",
					Ids:  []string{},
				},
			}

			sl = utils.TuplesFilterQueryForUpdateBuilder(sl, filter)

			expectedSql = "UPDATE relation_tuples SET expired_tx_id = ?"

			sql, _, err = sl.ToSql()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(sql).Should(Equal(expectedSql))
		})

		It("Case 2", func() {
			sl := squirrel.Update("relation_tuples").Set("expired_tx_id", 1000)

			filter := &base.TupleFilter{
				Entity: &base.EntityFilter{
					Type: "organization",
					Ids:  []string{"1", "2"},
				},
				Relation: "admin",
				Subject: &base.SubjectFilter{
					Type: "user",
					Ids:  []string{"4", "8"},
				},
			}

			sl = utils.TuplesFilterQueryForUpdateBuilder(sl, filter)

			expectedSql := "UPDATE relation_tuples SET expired_tx_id = ? WHERE entity_id IN (?,?) AND entity_type = ? AND relation = ? AND subject_id IN (?,?) AND subject_type = ?"

			sql, _, err := sl.ToSql()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(sql).Should(Equal(expectedSql))
		})
	})

	Context("AttributesFilterQueryForUpdateBuilder", func() {
		It("Case 1", func() {
			sl := squirrel.Update("attributes").Set("expired_tx_id", 1000)

			filter := &base.AttributeFilter{}

			sl = utils.AttributesFilterQueryForUpdateBuilder(sl, filter)

			expectedSql := "UPDATE attributes SET expired_tx_id = ?"

			sql, _, err := sl.ToSql()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(sql).Should(Equal(expectedSql))

			filter = &base.AttributeFilter{
				Entity: &base.EntityFilter{
					Type: "",
					Ids:  []string{},
				},
			}

			sl = utils.AttributesFilterQueryForUpdateBuilder(sl, filter)

			expectedSql = "UPDATE attributes SET expired_tx_id = ?"

			sql, _, err = sl.ToSql()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(sql).Should(Equal(expectedSql))
		})

		It("Case 2", func() {
			sl := squirrel.Update("attributes").Set("expired_tx_id", 1000)

			filter := &base.AttributeFilter{
				Entity: &base.EntityFilter{
					Type: "organization",
					Ids:  []string{"1", "2"},
				},
				Attributes: []string{"public", "balance"},
			}

			sl = utils.AttributesFilterQueryForUpdateBuilder(sl, filter)

			expectedSql := "UPDATE attributes SET expired_tx_id = ? WHERE attribute IN (?,?) AND entity_id IN (?,?) AND entity_type = ?"

			sql, _, err := sl.ToSql()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(sql).Should(Equal(expectedSql))
		})
	})
})
