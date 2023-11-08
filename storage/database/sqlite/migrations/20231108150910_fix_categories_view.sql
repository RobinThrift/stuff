-- +goose Up
-- +goose StatementBegin
DROP VIEW categories;
CREATE VIEW categories AS SELECT category as cat_name FROM assets GROUP BY cat_name;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP VIEW categories;
CREATE VIEW categories AS SELECT category as name FROM assets GROUP BY name;
-- +goose StatementEnd
