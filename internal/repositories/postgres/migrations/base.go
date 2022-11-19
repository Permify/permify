package migrations

// GetMigrations - Get migrations for Postgresql database
func GetMigrations() []string {
	return []string{
		InitialRelationTuplesMigration,
		InitialSchemaDefinitionsMigration,
		InitialTransactionsMigration,
		CreateRelationTupleUserSetIndexIfNotExistMigration,
		CreateRelationTupleUserSetRelationIndexIfNotExistMigration,
	}
}
