-- +goose Up
-- +goose StatementBegin
DROP VIEW custom_attr_names;
CREATE VIEW custom_attr_names AS SELECT j.value->>'name' as name FROM assets, json_each(custom_attrs) j WHERE custom_attrs IS NOT NULL GROUP BY name;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP VIEW custom_attr_names;
CREATE VIEW custom_attr_names AS SELECT j.key as name, j.type as type FROM assets, json_each(custom_attrs) j WHERE custom_attrs IS NOT NULL GROUP BY j.key;
-- +goose StatementEnd
