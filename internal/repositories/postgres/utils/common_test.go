package utils_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/Permify/permify/internal/repositories/postgres/utils"

	"github.com/stretchr/testify/assert"
)

func TestSnapshotQuery(t *testing.T) {
	sl := squirrel.Select("column").From("table")
	revision := uint64(42)

	query := utils.SnapshotQuery(sl, revision)
	sql, _, err := query.ToSql()

	assert.NoError(t, err)
	expectedSQL := "SELECT column FROM table WHERE (pg_visible_in_snapshot(created_tx_id, (select snapshot from transactions where id = '42'::xid8)) = true OR created_tx_id = '42'::xid8) AND ((pg_visible_in_snapshot(expired_tx_id, (select snapshot from transactions where id = '42'::xid8)) = false OR expired_tx_id = '0'::xid8) AND expired_tx_id <> '42'::xid8)"
	assert.Equal(t, expectedSQL, sql)
}

func TestGarbageCollectQuery(t *testing.T) {
	window := 24 * time.Hour
	tenantID := "testTenant"

	query := utils.GarbageCollectQuery(window, tenantID)
	sql, _, err := query.ToSql()

	assert.NoError(t, err)

	now := time.Now()
	arg := fmt.Sprintf("'%s'", now.Add(-window).Format(time.RFC3339))
	expectedSQL := "DELETE FROM relation_tuples WHERE created_tx_id IN (SELECT id FROM transactions WHERE timestamp < " + arg + ") AND ((expired_tx_id = '0'::xid8 OR expired_tx_id IN (SELECT id FROM transactions WHERE timestamp < " + arg + ")) AND tenant_id = 'testTenant')"
	assert.Equal(t, expectedSQL, sql)
}
