package utils_test

import (
	"testing"

	"github.com/Masterminds/squirrel"
	"github.com/stretchr/testify/assert"

	"github.com/Permify/permify/internal/storage/postgres/utils"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

func TestFilterQueryForSelectBuilder(t *testing.T) {
	sl := squirrel.Select("*").From("test_table")

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

	sl = utils.FilterQueryForSelectBuilder(sl, filter)

	expectedSql := "SELECT * FROM test_table WHERE entity_id IN (?,?) AND entity_type = ? AND relation = ? AND subject_id IN (?,?) AND subject_relation = ? AND subject_type = ?"
	expectedArgs := []interface{}{"1", "2", "entity_type", "relation", "3", "4", "subject_relation", "subject_type"}

	sql, args, _ := sl.ToSql()
	assert.Equal(t, expectedSql, sql)
	assert.Equal(t, expectedArgs, args)
}
