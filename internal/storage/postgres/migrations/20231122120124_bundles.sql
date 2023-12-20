-- +goose Up
CREATE TABLE IF NOT EXISTS bundles
(
    name          VARCHAR NOT NULL,
    payload       jsonb   NOT NULL,
    tenant_id     VARCHAR NOT NULL,
    created_at TIMESTAMP DEFAULT (now() AT TIME ZONE 'UTC') NOT NULL,
    CONSTRAINT pk_bundle PRIMARY KEY (name, tenant_id)
);

-- +goose Down
DROP TABLE IF EXISTS bundles;
