-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_schema_version ON schema_definitions (version);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_schema_version;
-- +goose StatementEnd
