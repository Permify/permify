package migrations

// CreateRelationTupleUserSetIndexIfNotExistMigration -
func CreateRelationTupleUserSetIndexIfNotExistMigration() string {
	return `CREATE INDEX IF NOT EXISTS idx_relation_tuple_userset ON relation_tuple (userset_object_id, userset_entity, userset_relation, entity, relation);`
}

// CreateRelationTupleUserSetRelationIndexIfNotExistMigration -
func CreateRelationTupleUserSetRelationIndexIfNotExistMigration() string {
	return `CREATE INDEX IF NOT EXISTS idx_relation_tuple_userset_relation ON relation_tuple (userset_entity, userset_relation, entity, relation);`
}
