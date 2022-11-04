package migrations

// GetMigrations -
func GetMigrations() []string {
	return []string{
		InitialRelationTuplesMigration,
		InitialSchemaDefinitionsMigration,
		InitialTransactionsMigration,
		CreateRelationTupleUserSetIndexIfNotExistMigration,
		CreateRelationTupleUserSetRelationIndexIfNotExistMigration,
	}
}
