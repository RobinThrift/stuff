-- +goose Up
-- +goose StatementBegin
DROP VIEW locations;
CREATE VIEW locations AS SELECT location as loc_name FROM assets WHERE location IS NOT NULL UNION SELECT location as loc_name FROM asset_parts WHERE location IS NOT NULL GROUP BY loc_name;;

DROP VIEW position_codes;
CREATE VIEW position_codes AS SELECT position_code as pos_code FROM assets WHERE position_code IS NOT NULL UNION SELECT position_code as pos_code FROM asset_parts WHERE position_code IS NOT NULL GROUP BY pos_code;;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP VIEW locations;
CREATE VIEW locations AS SELECT location as name FROM assets WHERE location IS NOT NULL GROUP BY name;

DROP VIEW position_codes;
CREATE VIEW position_codes AS SELECT position_code as code FROM assets WHERE position_code IS NOT NULL GROUP BY code;
-- +goose StatementEnd
