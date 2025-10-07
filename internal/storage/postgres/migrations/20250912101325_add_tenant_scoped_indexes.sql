-- +goose NO TRANSACTION
-- +goose Up
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_transactions_tenant_id ON transactions (tenant_id, id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_schema_tenant_version ON schema_definitions (tenant_id, version);
DROP INDEX CONCURRENTLY IF EXISTS idx_schema_version;

-- +goose Down
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_schema_version ON schema_definitions (version);
DROP INDEX CONCURRENTLY IF EXISTS idx_transactions_tenant_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_schema_tenant_version;
