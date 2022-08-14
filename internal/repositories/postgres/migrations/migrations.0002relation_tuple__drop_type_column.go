package migrations

// DropRelationTupleTypeColumnIfExistMigration -
func DropRelationTupleTypeColumnIfExistMigration() string {
	return `ALTER TABLE relation_tuple
DROP COLUMN IF EXISTS type;`
}
