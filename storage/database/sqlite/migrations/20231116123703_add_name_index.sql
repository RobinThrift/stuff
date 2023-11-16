-- +goose Up
-- +goose StatementBegin
CREATE INDEX asset_name_idx ON assets(name COLLATE NOCASE);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX asset_name_idx;
-- +goose StatementEnd
