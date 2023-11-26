-- +goose NO TRANSACTION
-- +goose Up
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_attributes_tenant_created_tx ON public.attributes (tenant_id, created_tx_id);
DROP INDEX CONCURRENTLY IF EXISTS idx_tuples_entity;

-- +goose Down
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_tuples_entity ON public.relation_tuples (tenant_id, entity_type, entity_id, relation);
DROP INDEX CONCURRENTLY IF EXISTS idx_attributes_tenant_created_tx;