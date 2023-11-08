-- +goose Up
-- +goose StatementBegin
CREATE VIEW models AS SELECT model, model_no FROM assets WHERE model IS NOT NULL OR model_no IS NOT NULL GROUP BY model, model_no;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP VIEW models;
-- +goose StatementEnd
