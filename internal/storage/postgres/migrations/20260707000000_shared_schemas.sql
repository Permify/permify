-- +goose Up

-- Table for shared schema definitions, keyed by shared_schema_id instead of tenant_id.
CREATE TABLE IF NOT EXISTS shared_schema_definitions (
    name                  VARCHAR  NOT NULL,
    serialized_definition BYTEA    NOT NULL,
    version               CHAR(20) NOT NULL,
    shared_schema_id      VARCHAR  NOT NULL,
    CONSTRAINT pk_shared_schema_definition PRIMARY KEY (shared_schema_id, name, version)
);

CREATE INDEX IF NOT EXISTS idx_shared_schema_version
    ON shared_schema_definitions (shared_schema_id, version);

-- Add shared_schema_id column to tenants table. NULL means tenant uses per-tenant schemas.
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS shared_schema_id VARCHAR DEFAULT NULL;

-- +goose Down
ALTER TABLE tenants DROP COLUMN IF EXISTS shared_schema_id;
DROP TABLE IF EXISTS shared_schema_definitions;
