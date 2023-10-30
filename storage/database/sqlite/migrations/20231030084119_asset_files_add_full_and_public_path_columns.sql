-- +goose Up
-- +goose StatementBegin
ALTER TABLE asset_files ADD COLUMN full_path TEXT NOT NULL;
ALTER TABLE asset_files ADD COLUMN public_path TEXT NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE asset_files DROP COLUMN full_path;
ALTER TABLE asset_files DROP COLUMN public_path;
-- +goose StatementEnd

