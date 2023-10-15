-- +goose Up
-- +goose StatementBegin
ALTER TABLE assets ADD COLUMN type TEXT CHECK(type IN ('ASSET', 'COMPONENT', 'CONSUMABLE')) NOT NULL DEFAULT 'ASSET';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE assets DROP COLUMN type;
-- +goose StatementEnd
