package migrations

// CreateEntityConfigVersionField -
func CreateEntityConfigVersionField() string {
	return `ALTER TABLE entity_config
ADD COLUMN IF NOT EXISTS version CHAR(20) NOT NULL;`
}

// CreateEntityConfigChangePrimaryKey -
func CreateEntityConfigChangePrimaryKey() string {
	return `ALTER TABLE entity_config
DROP CONSTRAINT IF EXISTS pk_entity_config CASCADE,
ADD CONSTRAINT pk_entity_config PRIMARY KEY(entity, version);`
}
