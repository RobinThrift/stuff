-- +goose Up
-- +goose StatementBegin
CREATE TABLE asset_parts_new (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    asset_id    INTEGER NOT NULL REFERENCES assets(id) ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED,

    tag            TEXT NOT NULL,
    name           TEXT NOT NULL,
    location       TEXT DEFAULT NULL,
    position_code  TEXT DEFAULT NULL,
    notes          TEXT DEFAULT NULL,

    created_by INTEGER NOT NULL REFERENCES users(id),
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%SZ', CURRENT_TIMESTAMP)),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%SZ', CURRENT_TIMESTAMP))
);

INSERT INTO asset_parts_new (
	id, asset_id, tag, name, location, position_code, notes, created_by, created_at, updated_at
) SELECT
	id, asset_id, tag, name, location, position_code, notes, created_by, created_at, updated_at
FROM asset_parts;

DROP VIEW locations;
DROP VIEW position_codes;
DROP TABLE asset_parts;

ALTER TABLE asset_parts_new RENAME TO asset_parts;

CREATE VIEW locations AS SELECT location as loc_name FROM assets WHERE location IS NOT NULL UNION SELECT location as loc_name FROM asset_parts WHERE location IS NOT NULL GROUP BY loc_name;

CREATE VIEW position_codes AS SELECT position_code as pos_code FROM assets WHERE position_code IS NOT NULL UNION SELECT position_code as pos_code FROM asset_parts WHERE position_code IS NOT NULL GROUP BY pos_code;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'cannot undo, also no need';
-- +goose StatementEnd
