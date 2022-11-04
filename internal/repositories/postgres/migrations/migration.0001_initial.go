package migrations

// InitialRelationTuplesMigration -
const InitialRelationTuplesMigration = `create table if not exists relation_tuples (
 	id SERIAL not null,
    entity_type varchar not null,
    entity_id varchar not null,
    relation varchar not null,
    subject_type varchar not null,
    subject_id varchar not null,
    subject_relation varchar not null,
    created_tx_id xid8 default (pg_current_xact_id()),
    expired_tx_id xid8 default ('0'),
	constraint pk_relation_tuple primary key (id),
    constraint uq_relation_tuple unique (entity_type, entity_id, relation, subject_type, subject_id, subject_relation, created_tx_id, expired_tx_id),
    constraint uq_relation_tuple_not_expired unique (entity_type, entity_id, relation, subject_type, subject_id, subject_relation, expired_tx_id));`

// InitialSchemaDefinitionsMigration -
const InitialSchemaDefinitionsMigration = `create table if not exists schema_definitions (
 	entity_type varchar not null,
    serialized_definition BYTEA not null,
	version CHAR(20) NOT NULL,
    constraint pk_schema_definition primary key (entity_type, version));`

// InitialTransactionsMigration -
const InitialTransactionsMigration = `create table if not exists transactions (
    id xid8 default (pg_current_xact_id()) not null,
    snapshot pg_snapshot default (pg_current_snapshot()) not null,
    timestamp timestamp default (now() at time zone 'UTC') not null,
    constraint pk_transaction primary key (id));`

// CreateRelationTupleUserSetIndexIfNotExistMigration -
const CreateRelationTupleUserSetIndexIfNotExistMigration = `create index if not exists idx_tuples_subject ON relation_tuples (subject_type, subject_id, subject_relation, entity_type, relation);`

// CreateRelationTupleUserSetRelationIndexIfNotExistMigration -
const CreateRelationTupleUserSetRelationIndexIfNotExistMigration = `create index if not exists idx_tuples_subject_relation ON relation_tuples (subject_type, subject_relation, entity_type, relation);`
