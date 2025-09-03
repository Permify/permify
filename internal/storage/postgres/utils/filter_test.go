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

	Context("Single ID/Attribute Tests", func() {
		Context("TuplesFilterQueryForSelectBuilder - Single IDs", func() {
			It("should handle single entity ID", func() {
				sl := squirrel.Select("*").From("relation_tuples")

				filter := &base.TupleFilter{
					Entity: &base.EntityFilter{
						Type: "entity_type",
						Ids:  []string{"single_id"},
					},
				}

				sl = utils.TuplesFilterQueryForSelectBuilder(sl, filter)

				expectedSql := "SELECT * FROM relation_tuples WHERE entity_id = ? AND entity_type = ?"
				expectedArgs := []interface{}{"single_id", "entity_type"}

				sql, args, err := sl.ToSql()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(args).Should(Equal(expectedArgs))
				Expect(sql).Should(Equal(expectedSql))
			})

			It("should handle single subject ID", func() {
				sl := squirrel.Select("*").From("relation_tuples")

				filter := &base.TupleFilter{
					Subject: &base.SubjectFilter{
						Type: "subject_type",
						Ids:  []string{"single_subject_id"},
					},
				}

				sl = utils.TuplesFilterQueryForSelectBuilder(sl, filter)

				expectedSql := "SELECT * FROM relation_tuples WHERE subject_id = ? AND subject_type = ?"
				expectedArgs := []interface{}{"single_subject_id", "subject_type"}

				sql, args, err := sl.ToSql()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(args).Should(Equal(expectedArgs))
				Expect(sql).Should(Equal(expectedSql))
			})

			It("should handle single entity ID and single subject ID", func() {
				sl := squirrel.Select("*").From("relation_tuples")

				filter := &base.TupleFilter{
					Entity: &base.EntityFilter{
						Type: "entity_type",
						Ids:  []string{"single_entity_id"},
					},
					Subject: &base.SubjectFilter{
						Type: "subject_type",
						Ids:  []string{"single_subject_id"},
					},
				}

				sl = utils.TuplesFilterQueryForSelectBuilder(sl, filter)

				expectedSql := "SELECT * FROM relation_tuples WHERE entity_id = ? AND entity_type = ? AND subject_id = ? AND subject_type = ?"
				expectedArgs := []interface{}{"single_entity_id", "entity_type", "single_subject_id", "subject_type"}

				sql, args, err := sl.ToSql()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(args).Should(Equal(expectedArgs))
				Expect(sql).Should(Equal(expectedSql))
			})
		})

		Context("AttributesFilterQueryForSelectBuilder - Single IDs/Attributes", func() {
			It("should handle single entity ID", func() {
				sl := squirrel.Select("*").From("attributes")

				filter := &base.AttributeFilter{
					Entity: &base.EntityFilter{
						Type: "entity_type",
						Ids:  []string{"single_id"},
					},
				}

				sl = utils.AttributesFilterQueryForSelectBuilder(sl, filter)

				expectedSql := "SELECT * FROM attributes WHERE entity_id = ? AND entity_type = ?"
				expectedArgs := []interface{}{"single_id", "entity_type"}

				sql, args, err := sl.ToSql()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(args).Should(Equal(expectedArgs))
				Expect(sql).Should(Equal(expectedSql))
			})

			It("should handle single attribute", func() {
				sl := squirrel.Select("*").From("attributes")

				filter := &base.AttributeFilter{
					Attributes: []string{"single_attribute"},
				}

				sl = utils.AttributesFilterQueryForSelectBuilder(sl, filter)

				expectedSql := "SELECT * FROM attributes WHERE attribute = ?"
				expectedArgs := []interface{}{"single_attribute"}

				sql, args, err := sl.ToSql()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(args).Should(Equal(expectedArgs))
				Expect(sql).Should(Equal(expectedSql))
			})

			It("should handle single entity ID and single attribute", func() {
				sl := squirrel.Select("*").From("attributes")

				filter := &base.AttributeFilter{
					Entity: &base.EntityFilter{
						Type: "entity_type",
						Ids:  []string{"single_entity_id"},
					},
					Attributes: []string{"single_attribute"},
				}

				sl = utils.AttributesFilterQueryForSelectBuilder(sl, filter)

				expectedSql := "SELECT * FROM attributes WHERE attribute = ? AND entity_id = ? AND entity_type = ?"
				expectedArgs := []interface{}{"single_attribute", "single_entity_id", "entity_type"}

				sql, args, err := sl.ToSql()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(args).Should(Equal(expectedArgs))
				Expect(sql).Should(Equal(expectedSql))
			})
		})

		Context("TuplesFilterQueryForUpdateBuilder - Single IDs", func() {
			It("should handle single entity ID", func() {
				sl := squirrel.Update("relation_tuples").Set("expired_tx_id", 1000)

				filter := &base.TupleFilter{
					Entity: &base.EntityFilter{
						Type: "entity_type",
						Ids:  []string{"single_id"},
					},
				}

				sl = utils.TuplesFilterQueryForUpdateBuilder(sl, filter)

				expectedSql := "UPDATE relation_tuples SET expired_tx_id = ? WHERE entity_id = ? AND entity_type = ?"

				sql, _, err := sl.ToSql()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(sql).Should(Equal(expectedSql))
			})

			It("should handle single subject ID", func() {
				sl := squirrel.Update("relation_tuples").Set("expired_tx_id", 1000)

				filter := &base.TupleFilter{
					Subject: &base.SubjectFilter{
						Type: "subject_type",
						Ids:  []string{"single_subject_id"},
					},
				}

				sl = utils.TuplesFilterQueryForUpdateBuilder(sl, filter)

				expectedSql := "UPDATE relation_tuples SET expired_tx_id = ? WHERE subject_id = ? AND subject_type = ?"

				sql, _, err := sl.ToSql()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(sql).Should(Equal(expectedSql))
			})

			It("should handle single entity ID and single subject ID", func() {
				sl := squirrel.Update("relation_tuples").Set("expired_tx_id", 1000)

				filter := &base.TupleFilter{
					Entity: &base.EntityFilter{
						Type: "entity_type",
						Ids:  []string{"single_entity_id"},
					},
					Subject: &base.SubjectFilter{
						Type: "subject_type",
						Ids:  []string{"single_subject_id"},
					},
				}

				sl = utils.TuplesFilterQueryForUpdateBuilder(sl, filter)

				expectedSql := "UPDATE relation_tuples SET expired_tx_id = ? WHERE entity_id = ? AND entity_type = ? AND subject_id = ? AND subject_type = ?"

				sql, _, err := sl.ToSql()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(sql).Should(Equal(expectedSql))
			})
		})

		Context("AttributesFilterQueryForUpdateBuilder - Single IDs/Attributes", func() {
			It("should handle single entity ID", func() {
				sl := squirrel.Update("attributes").Set("expired_tx_id", 1000)

				filter := &base.AttributeFilter{
					Entity: &base.EntityFilter{
						Type: "entity_type",
						Ids:  []string{"single_id"},
					},
				}

				sl = utils.AttributesFilterQueryForUpdateBuilder(sl, filter)

				expectedSql := "UPDATE attributes SET expired_tx_id = ? WHERE entity_id = ? AND entity_type = ?"

				sql, _, err := sl.ToSql()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(sql).Should(Equal(expectedSql))
			})

			It("should handle single attribute", func() {
				sl := squirrel.Update("attributes").Set("expired_tx_id", 1000)

				filter := &base.AttributeFilter{
					Attributes: []string{"single_attribute"},
				}

				sl = utils.AttributesFilterQueryForUpdateBuilder(sl, filter)

				expectedSql := "UPDATE attributes SET expired_tx_id = ? WHERE attribute = ?"

				sql, _, err := sl.ToSql()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(sql).Should(Equal(expectedSql))
			})

			It("should handle single entity ID and single attribute", func() {
				sl := squirrel.Update("attributes").Set("expired_tx_id", 1000)

				filter := &base.AttributeFilter{
					Entity: &base.EntityFilter{
						Type: "entity_type",
						Ids:  []string{"single_entity_id"},
					},
					Attributes: []string{"single_attribute"},
				}

				sl = utils.AttributesFilterQueryForUpdateBuilder(sl, filter)

				expectedSql := "UPDATE attributes SET expired_tx_id = ? WHERE attribute = ? AND entity_id = ? AND entity_type = ?"

				sql, _, err := sl.ToSql()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(sql).Should(Equal(expectedSql))
			})
		})
	})

	Context("Edge Cases and Empty Filters", func() {
		Context("TuplesFilterQueryForSelectBuilder - Edge Cases", func() {
			It("should handle empty entity IDs array", func() {
				sl := squirrel.Select("*").From("relation_tuples")

				filter := &base.TupleFilter{
					Entity: &base.EntityFilter{
						Type: "entity_type",
						Ids:  []string{},
					},
				}

				sl = utils.TuplesFilterQueryForSelectBuilder(sl, filter)

				expectedSql := "SELECT * FROM relation_tuples WHERE entity_type = ?"
				expectedArgs := []interface{}{"entity_type"}

				sql, args, err := sl.ToSql()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(args).Should(Equal(expectedArgs))
				Expect(sql).Should(Equal(expectedSql))
			})

			It("should handle empty subject IDs array", func() {
				sl := squirrel.Select("*").From("relation_tuples")

				filter := &base.TupleFilter{
					Subject: &base.SubjectFilter{
						Type: "subject_type",
						Ids:  []string{},
					},
				}

				sl = utils.TuplesFilterQueryForSelectBuilder(sl, filter)

				expectedSql := "SELECT * FROM relation_tuples WHERE subject_type = ?"
				expectedArgs := []interface{}{"subject_type"}

				sql, args, err := sl.ToSql()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(args).Should(Equal(expectedArgs))
				Expect(sql).Should(Equal(expectedSql))
			})

			It("should handle nil entity IDs", func() {
				sl := squirrel.Select("*").From("relation_tuples")

				filter := &base.TupleFilter{
					Entity: &base.EntityFilter{
						Type: "entity_type",
						Ids:  nil,
					},
				}

				sl = utils.TuplesFilterQueryForSelectBuilder(sl, filter)

				expectedSql := "SELECT * FROM relation_tuples WHERE entity_type = ?"
				expectedArgs := []interface{}{"entity_type"}

				sql, args, err := sl.ToSql()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(args).Should(Equal(expectedArgs))
				Expect(sql).Should(Equal(expectedSql))
			})

			It("should handle nil subject IDs", func() {
				sl := squirrel.Select("*").From("relation_tuples")

				filter := &base.TupleFilter{
					Subject: &base.SubjectFilter{
						Type: "subject_type",
						Ids:  nil,
					},
				}

				sl = utils.TuplesFilterQueryForSelectBuilder(sl, filter)

				expectedSql := "SELECT * FROM relation_tuples WHERE subject_type = ?"
				expectedArgs := []interface{}{"subject_type"}

				sql, args, err := sl.ToSql()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(args).Should(Equal(expectedArgs))
				Expect(sql).Should(Equal(expectedSql))
			})
		})

		Context("AttributesFilterQueryForSelectBuilder - Edge Cases", func() {
			It("should handle empty entity IDs array", func() {
				sl := squirrel.Select("*").From("attributes")

				filter := &base.AttributeFilter{
					Entity: &base.EntityFilter{
						Type: "entity_type",
						Ids:  []string{},
					},
				}

				sl = utils.AttributesFilterQueryForSelectBuilder(sl, filter)

				expectedSql := "SELECT * FROM attributes WHERE entity_type = ?"
				expectedArgs := []interface{}{"entity_type"}

				sql, args, err := sl.ToSql()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(args).Should(Equal(expectedArgs))
				Expect(sql).Should(Equal(expectedSql))
			})

			It("should handle empty attributes array", func() {
				sl := squirrel.Select("*").From("attributes")

				filter := &base.AttributeFilter{
					Attributes: []string{},
				}

				sl = utils.AttributesFilterQueryForSelectBuilder(sl, filter)

				expectedSql := "SELECT * FROM attributes"

				sql, _, err := sl.ToSql()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(sql).Should(Equal(expectedSql))
			})

			It("should handle nil entity IDs", func() {
				sl := squirrel.Select("*").From("attributes")

				filter := &base.AttributeFilter{
					Entity: &base.EntityFilter{
						Type: "entity_type",
						Ids:  nil,
					},
				}

				sl = utils.AttributesFilterQueryForSelectBuilder(sl, filter)

				expectedSql := "SELECT * FROM attributes WHERE entity_type = ?"
				expectedArgs := []interface{}{"entity_type"}

				sql, args, err := sl.ToSql()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(args).Should(Equal(expectedArgs))
				Expect(sql).Should(Equal(expectedSql))
			})

			It("should handle nil attributes", func() {
				sl := squirrel.Select("*").From("attributes")

				filter := &base.AttributeFilter{
					Attributes: nil,
				}

				sl = utils.AttributesFilterQueryForSelectBuilder(sl, filter)

				expectedSql := "SELECT * FROM attributes"

				sql, _, err := sl.ToSql()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(sql).Should(Equal(expectedSql))
			})
		})

		Context("TuplesFilterQueryForUpdateBuilder - Edge Cases", func() {
			It("should handle empty entity IDs array", func() {
				sl := squirrel.Update("relation_tuples").Set("expired_tx_id", 1000)

				filter := &base.TupleFilter{
					Entity: &base.EntityFilter{
						Type: "entity_type",
						Ids:  []string{},
					},
				}

				sl = utils.TuplesFilterQueryForUpdateBuilder(sl, filter)

				expectedSql := "UPDATE relation_tuples SET expired_tx_id = ? WHERE entity_type = ?"

				sql, _, err := sl.ToSql()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(sql).Should(Equal(expectedSql))
			})

			It("should handle empty subject IDs array", func() {
				sl := squirrel.Update("relation_tuples").Set("expired_tx_id", 1000)

				filter := &base.TupleFilter{
					Subject: &base.SubjectFilter{
						Type: "subject_type",
						Ids:  []string{},
					},
				}

				sl = utils.TuplesFilterQueryForUpdateBuilder(sl, filter)

				expectedSql := "UPDATE relation_tuples SET expired_tx_id = ? WHERE subject_type = ?"

				sql, _, err := sl.ToSql()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(sql).Should(Equal(expectedSql))
			})
		})

		Context("AttributesFilterQueryForUpdateBuilder - Edge Cases", func() {
			It("should handle empty entity IDs array", func() {
				sl := squirrel.Update("attributes").Set("expired_tx_id", 1000)

				filter := &base.AttributeFilter{
					Entity: &base.EntityFilter{
						Type: "entity_type",
						Ids:  []string{},
					},
				}

				sl = utils.AttributesFilterQueryForUpdateBuilder(sl, filter)

				expectedSql := "UPDATE attributes SET expired_tx_id = ? WHERE entity_type = ?"

				sql, _, err := sl.ToSql()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(sql).Should(Equal(expectedSql))
			})

			It("should handle empty attributes array", func() {
				sl := squirrel.Update("attributes").Set("expired_tx_id", 1000)

				filter := &base.AttributeFilter{
					Attributes: []string{},
				}

				sl = utils.AttributesFilterQueryForUpdateBuilder(sl, filter)

				expectedSql := "UPDATE attributes SET expired_tx_id = ?"

				sql, _, err := sl.ToSql()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(sql).Should(Equal(expectedSql))
			})
		})
	})
})
