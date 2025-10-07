-- Required for CONCURRENTLY operations
-- +goose NO TRANSACTION 

-- Migration to add tenant-scoped indexes
-- +goose Up 

-- Create composite index on transactions table for tenant-scoped queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_transactions_tenant_id ON transactions (tenant_id, id); -- Improves performance for tenant-specific transaction lookups

-- Create composite index on schema_definitions for tenant and version lookups
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_schema_tenant_version ON schema_definitions (tenant_id, version); -- Optimizes schema version queries per tenant

-- Remove old single-column version index as it's replaced by tenant_version composite index
DROP INDEX CONCURRENTLY IF EXISTS idx_schema_version; -- No longer needed after composite index creation

-- Rollback tenant-scoped indexes
-- +goose Down 

-- Recreate the original version index for rollback
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_schema_version ON schema_definitions (version); -- Restore original index

-- Remove tenant-scoped transactions index
DROP INDEX CONCURRENTLY IF EXISTS idx_transactions_tenant_id; -- Rollback tenant optimization

-- Remove tenant-scoped schema version index
DROP INDEX CONCURRENTLY IF EXISTS idx_schema_tenant_version; -- Rollback schema optimization
