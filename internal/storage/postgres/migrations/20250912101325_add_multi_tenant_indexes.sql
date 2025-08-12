-- +goose NO TRANSACTION
-- +goose Up
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_transactions_tenant_id ON transactions (tenant_id, id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_schema_tenant_version ON schema_definitions (tenant_id, version);

-- +goose Down
DROP INDEX CONCURRENTLY IF EXISTS idx_transactions_tenant_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_schema_tenant_version;
