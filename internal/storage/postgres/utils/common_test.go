package utils_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Masterminds/squirrel"

	"github.com/Permify/permify/internal/storage/postgres/utils"
)

var _ = Describe("Common", func() {
	Context("TestSnapshotQuery", func() {
		It("Case 1", func() {
			sl := squirrel.Select("column").From("table")
			revision := uint64(42)

			query := utils.SnapshotQuery(sl, revision)
			sql, _, err := query.ToSql()
			Expect(err).ShouldNot(HaveOccurred())

			expectedSQL := "SELECT column FROM table WHERE (pg_visible_in_snapshot(created_tx_id, (select snapshot from transactions where id = '42'::xid8)) = true OR created_tx_id = '42'::xid8) AND ((pg_visible_in_snapshot(expired_tx_id, (select snapshot from transactions where id = '42'::xid8)) = false OR expired_tx_id = '0'::xid8) AND expired_tx_id <> '42'::xid8)"
			Expect(sql).Should(Equal(expectedSQL))
		})
	})

	Context("TestGarbageCollectQuery", func() {
		It("Case 1", func() {
			query := utils.GenerateGCQuery("relation_tuples", 100)
			sql, _, err := query.ToSql()
			Expect(err).ShouldNot(HaveOccurred())

			expectedSQL := "DELETE FROM relation_tuples WHERE expired_tx_id <> '0'::xid8 AND expired_tx_id < '100'::xid8"
			Expect(expectedSQL).Should(Equal(sql))
		})
	})
})
