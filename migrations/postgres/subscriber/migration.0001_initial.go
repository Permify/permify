package subscriber

// CreateRelationTupleMigration -
func CreateRelationTupleMigration() string {
	return `create table if not exists relation_tuple (
 	id SERIAL NOT NULL,
    entity VARCHAR NOT NULL,
    object_id VARCHAR NOT NULL,
    relation VARCHAR NOT NULL,
    userset_entity VARCHAR NOT NULL,
    userset_object_id VARCHAR NOT NULL,
    userset_relation VARCHAR NOT NULL,
    type VARCHAR NOT NULL DEFAULT 'auto',
    commit_time TIMESTAMP WITHOUT TIME ZONE DEFAULT now() NOT NULL,
	CONSTRAINT pk_relation_tuple PRIMARY KEY (id),
    CONSTRAINT uq_relation_tuple UNIQUE (entity, object_id, relation, userset_entity, userset_object_id, userset_relation)
	);`
}

// CreateEntityConfigMigration -
func CreateEntityConfigMigration() string {
	return `create table if not exists entity_config (
 	entity VARCHAR NOT NULL,
    serialized_config BYTEA NOT NULL,
    commit_time TIMESTAMP WITHOUT TIME ZONE DEFAULT now() NOT NULL,
  	CONSTRAINT pk_entity_config PRIMARY KEY (entity)
	);`
}
