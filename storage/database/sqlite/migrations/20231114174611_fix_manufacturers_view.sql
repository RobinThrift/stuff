-- +goose Up
-- +goose StatementBegin
DROP VIEW manufacturers;
CREATE VIEW manufacturers AS SELECT manufacturer FROM assets WHERE manufacturer IS NOT NULL GROUP BY manufacturer;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP VIEW manufacturers;
CREATE VIEW manufacturers AS SELECT manufacturer as name FROM assets WHERE manufacturer IS NOT NULL GROUP BY name;
-- +goose StatementEnd
