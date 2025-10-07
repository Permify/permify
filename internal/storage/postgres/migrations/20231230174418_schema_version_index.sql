-- Required for CONCURRENTLY operations
-- +goose NO TRANSACTION 

-- Migration to add schema version index
-- +goose Up 

-- Create index on schema version column for faster schema lookups
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_schema_version ON schema_definitions (version); -- Improves query performance for version-based schema queries

-- Rollback schema version index
-- +goose Down 

-- Remove schema version index during rollback
DROP INDEX CONCURRENTLY IF EXISTS idx_schema_version; -- Cleanup index on migration rollback
