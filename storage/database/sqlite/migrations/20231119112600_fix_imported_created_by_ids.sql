-- +goose Up
-- +goose StatementBegin
UPDATE assets SET created_by = 1 WHERE created_by = 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'no need to undo';
-- +goose StatementEnd
