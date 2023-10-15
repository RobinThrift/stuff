-- +goose Up
-- +goose StatementBegin
ALTER TABLE assets ADD COLUMN quantity UNSIGNED BIG INT NOT NULL DEFAULT 0;
ALTER TABLE assets ADD COLUMN quantity_unit TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE assets DROP COLUMN quantity;
ALTER TABLE assets DROP COLUMN quantity_unit;
-- +goose StatementEnd
