-- +goose Up
CREATE TABLE IF NOT EXISTS relation_tuples (
    id               SERIAL  NOT NULL,
    entity_type      VARCHAR NOT NULL,
    entity_id        VARCHAR NOT NULL,
    relation         VARCHAR NOT NULL,
    subject_type     VARCHAR NOT NULL,
    subject_id       VARCHAR NOT NULL,
    subject_relation VARCHAR NOT NULL,
    created_tx_id    xid8 DEFAULT (pg_current_xact_id()),
    expired_tx_id    xid8 DEFAULT ('0'),
    CONSTRAINT pk_relation_tuple PRIMARY KEY (id),
    CONSTRAINT uq_relation_tuple UNIQUE (entity_type, entity_id, relation, subject_type, subject_id, subject_relation, created_tx_id, expired_tx_id),
    CONSTRAINT uq_relation_tuple_not_expired UNIQUE (entity_type, entity_id, relation, subject_type, subject_id, subject_relation, expired_tx_id)
);

CREATE TABLE IF NOT EXISTS schema_definitions (
    entity_type           VARCHAR NOT NULL,
    serialized_definition BYTEA    NOT NULL,
    version               CHAR(20) NOT NULL,
    CONSTRAINT pk_schema_definition PRIMARY KEY (entity_type, version)
);

CREATE TABLE IF NOT EXISTS transactions (
    id        xid8        DEFAULT (pg_current_xact_id())     NOT NULL,
    snapshot  pg_snapshot DEFAULT (pg_current_snapshot())    NOT NULL,
    timestamp TIMESTAMP   DEFAULT (now() AT TIME ZONE 'UTC') NOT NULL,
    CONSTRAINT pk_transaction PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_tuples_subject ON relation_tuples (subject_type, subject_id, subject_relation, entity_type, relation);
CREATE INDEX IF NOT EXISTS idx_tuples_subject_relation ON relation_tuples (subject_type, subject_relation, entity_type, relation);

-- +goose Down
DROP TABLE IF EXISTS relation_tuples;
DROP TABLE IF EXISTS schema_definitions;
DROP TABLE IF EXISTS transactions;