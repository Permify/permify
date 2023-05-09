-- +goose NO TRANSACTION
-- +goose Up
DROP INDEX CONCURRENTLY IF EXISTS idx_tuples_subject;
DROP INDEX CONCURRENTLY IF EXISTS idx_tuples_subject_relation;

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_tuples_subject ON relation_tuples (tenant_id, subject_type, subject_id, subject_relation, entity_type, relation);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_tuples_entity ON relation_tuples (tenant_id, entity_type, entity_id, relation);

-- +goose Down
DROP INDEX CONCURRENTLY IF EXISTS idx_tuples_subject;
DROP INDEX CONCURRENTLY IF EXISTS idx_tuples_entity;

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_tuples_subject ON relation_tuples (subject_type, subject_id, subject_relation, entity_type, relation);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_tuples_subject_relation ON relation_tuples (subject_type, subject_relation, entity_type, relation);
