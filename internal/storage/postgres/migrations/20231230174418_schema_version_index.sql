-- +goose NO TRANSACTION
-- +goose Up
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_schema_version ON schema_definitions (version);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX CONCURRENTLY IF EXISTS idx_schema_version;
-- +goose StatementEnd
