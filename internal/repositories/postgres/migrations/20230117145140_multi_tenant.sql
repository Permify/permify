-- +goose Up
CREATE TABLE IF NOT EXISTS tenants (
   id VARCHAR PRIMARY KEY,
   name VARCHAR NOT NULL,
   created_at TIMESTAMP DEFAULT (now() AT TIME ZONE 'UTC') NOT NULL
);

INSERT INTO tenants (id, name) VALUES ('1', 'initial');

ALTER TABLE relation_tuples
    DROP CONSTRAINT IF EXISTS uq_relation_tuple,
    DROP CONSTRAINT IF EXISTS uq_relation_tuple_not_expired,
    ADD COLUMN IF NOT EXISTS tenant_id VARCHAR NOT NULL DEFAULT '1',
    ADD CONSTRAINT uq_relation_tuple UNIQUE (tenant_id, entity_type, entity_id, relation, subject_type, subject_id, subject_relation, created_tx_id, expired_tx_id),
    ADD CONSTRAINT uq_relation_tuple_not_expired UNIQUE (tenant_id, entity_type, entity_id, relation, subject_type, subject_id, subject_relation, expired_tx_id);

ALTER TABLE schema_definitions
    DROP CONSTRAINT IF EXISTS pk_schema_definition,
    ADD COLUMN IF NOT EXISTS tenant_id VARCHAR NOT NULL DEFAULT '1',
    ADD CONSTRAINT pk_schema_definition PRIMARY KEY (tenant_id, entity_type, version);

-- +goose Down
DROP TABLE tenants;

ALTER TABLE relation_tuples
    DROP CONSTRAINT IF EXISTS uq_relation_tuple,
    DROP CONSTRAINT IF EXISTS uq_relation_tuple_not_expired,
    DROP COLUMN IF EXISTS tenant_id,
    ADD CONSTRAINT uq_relation_tuple UNIQUE (entity_type, entity_id, relation, subject_type, subject_id, subject_relation, created_tx_id, expired_tx_id),
    ADD CONSTRAINT uq_relation_tuple_not_expired UNIQUE (entity_type, entity_id, relation, subject_type, subject_id, subject_relation, expired_tx_id);

ALTER TABLE schema_definitions
    DROP CONSTRAINT IF EXISTS pk_schema_definition,
    DROP COLUMN IF EXISTS tenant_id,
    ADD CONSTRAINT pk_schema_definition PRIMARY KEY (entity_type, version);
