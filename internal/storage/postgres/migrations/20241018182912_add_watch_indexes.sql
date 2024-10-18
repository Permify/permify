-- +goose NO TRANSACTION
-- +goose Up
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_attributes_created_txid ON attributes (tenant_id, created_tx_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_attributes_expired_txid ON attributes (tenant_id, expired_tx_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_relation_tuples_created_txid ON relation_tuples (tenant_id, created_tx_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_relation_tuples_expired_txid ON relation_tuples (tenant_id, expired_tx_id);

-- +goose Down
DROP INDEX CONCURRENTLY IF EXISTS idx_attributes_created_txid;
DROP INDEX CONCURRENTLY IF EXISTS idx_attributes_expired_txid;
DROP INDEX CONCURRENTLY IF EXISTS idx_relation_tuples_created_txid;
DROP INDEX CONCURRENTLY IF EXISTS idx_relation_tuples_expired_txid;
