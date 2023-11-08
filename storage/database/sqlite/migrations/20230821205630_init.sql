-- +goose Up

CREATE TABLE sessions (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    token      TEXT NOT NULL,
    data       BLOB NOT NULL,
    expires_at TEXT NOT NULL
);
CREATE UNIQUE INDEX unique_session_token ON sessions(token);

CREATE TABLE local_auth_users (
    id                       INTEGER PRIMARY KEY AUTOINCREMENT,
    username                 TEXT    NOT NULL,
    algorithm                TEXT    NOT NULL,
    params                   TEXT    NOT NULL,
    salt                     BLOB    NOT NULL,
    password                 BLOB    NOT NULL,
    requires_password_change BOOLEAN NOT NULL DEFAULT true,
    created_at               TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%SZ', CURRENT_TIMESTAMP)),
    updated_at               TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%SZ', CURRENT_TIMESTAMP))
);
CREATE UNIQUE INDEX unique_usernames_auth_method_local ON local_auth_users(username);

CREATE TABLE users (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,

    username        TEXT NOT NULL,
    display_name    TEXT NOT NULL,

    is_admin        BOOLEAN NOT NULL DEFAULT false,

    auth_ref        TEXT NOT NULL,

    created_at      TEXT NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%SZ', CURRENT_TIMESTAMP)),
    updated_at      TEXT NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%SZ', CURRENT_TIMESTAMP))
);
CREATE UNIQUE INDEX unique_usernames ON users(username);
CREATE UNIQUE INDEX unique_auth_ref ON users(auth_ref);

CREATE TABLE tags (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    tag          TEXT    NOT NULL,
    in_use       BOOLEAN NOT NULL DEFAULT false,
    created_at   TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%SZ', CURRENT_TIMESTAMP)),
    updated_at   TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%SZ', CURRENT_TIMESTAMP))
);
CREATE UNIQUE INDEX unique_tag ON tags(tag);

CREATE TABLE assets (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    parent_asset_id INTEGER DEFAULT NULL,
    status          TEXT CHECK(status IN ('IN_STORAGE', 'IN_USE', 'ARCHIVED')) NOT NULL DEFAULT 'IN_STORAGE',

    tag            TEXT DEFAULT NULL,
    name           TEXT NOT NULL,
    category       TEXT NOT NULL,
    model          TEXT DEFAULT NULL,
    model_no       TEXT DEFAULT NULL,
    serial_no      TEXT DEFAULT NULL,
    manufacturer   TEXT DEFAULT NULL,
    notes          TEXT DEFAULT NULL,
    image_url      TEXT DEFAULT NULL,
    thumbnail_url  TEXT DEFAULT NULL,
    warranty_until TEXT DEFAULT NULL,
    custom_attrs   TEXT DEFAULT NULL,

    checked_out_to INTEGER DEFAULT NULL,
    location       TEXT DEFAULT NULL,
    position_code  TEXT DEFAULT NULL,

    purchase_supplier TEXT DEFAULT NULL,
    purchase_order_no TEXT DEFAULT NULL,
    purchase_date     TEXT DEFAULT NULL,
    purchase_amount   INT  DEFAULT NULL,
    purchase_currency TEXT DEFAULT NULL,

	parts_total_counter INT NOT NULL DEFAULT 0,

    created_by INTEGER NOT NULL,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%SZ', CURRENT_TIMESTAMP)),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%SZ', CURRENT_TIMESTAMP)),

    FOREIGN KEY(parent_asset_id) REFERENCES assets(id),
    FOREIGN KEY(tag) REFERENCES tags(tag),
    FOREIGN KEY(checked_out_to) REFERENCES users(id),
    FOREIGN KEY(created_by) REFERENCES users(id)
);

CREATE TABLE asset_parts (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    asset_id    INTEGER NOT NULL,

    tag            TEXT NOT NULL,
    name           TEXT NOT NULL,
    location       TEXT DEFAULT NULL,
    position_code  TEXT DEFAULT NULL,
    notes          TEXT DEFAULT NULL,

    created_by INTEGER NOT NULL,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%SZ', CURRENT_TIMESTAMP)),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%SZ', CURRENT_TIMESTAMP)),

    FOREIGN KEY(asset_id) REFERENCES assets(id),
    FOREIGN KEY(tag) REFERENCES tags(tag),
    FOREIGN KEY(created_by) REFERENCES users(id)
);

CREATE TABLE asset_files (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    asset_id    INTEGER NOT NULL,

    name        TEXT NOT NULL,
    filetype    TEXT NOT NULL,
    sha256      BLOB NOT NULL,
    size_bytes  INT  NOT NULL,

    created_by INTEGER NOT NULL,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%SZ', CURRENT_TIMESTAMP)),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%SZ', CURRENT_TIMESTAMP)),

    FOREIGN KEY(asset_id) REFERENCES assets(id),
    FOREIGN KEY(created_by) REFERENCES users(id)
);

CREATE VIEW categories AS SELECT category as name FROM assets GROUP BY name;
CREATE VIEW manufacturers AS SELECT manufacturer as name FROM assets WHERE manufacturer IS NOT NULL GROUP BY name;
CREATE VIEW suppliers AS SELECT purchase_supplier as name FROM assets WHERE purchase_supplier IS NOT NULL GROUP BY name;
CREATE VIEW locations AS SELECT location as name FROM assets WHERE location IS NOT NULL GROUP BY name;
CREATE VIEW position_codes AS SELECT position_code as code FROM assets WHERE position_code IS NOT NULL GROUP BY code;
CREATE VIEW custom_attr_names AS SELECT j.key as name, j.type as type FROM assets, json_each(custom_attrs) j WHERE custom_attrs IS NOT NULL GROUP BY j.key;

CREATE VIRTUAL TABLE assets_fts USING fts5(id, name, tag, category, model, model_no, serial_no, manufacturer, notes, custom_attrs, content='assets', content_rowid='id');

-- +goose StatementBegin
CREATE TRIGGER assets_after_insert AFTER INSERT ON assets BEGIN
	INSERT INTO assets_fts(rowid, id, name, tag, category, model, model_no, serial_no, manufacturer, notes, custom_attrs) VALUES (
		new.id,
		new.id,
		coalesce(new.name, ""),
		coalesce(new.tag, ""),
		coalesce(new.category, ""),
		coalesce(new.model, ""),
		coalesce(new.model_no, ""),
		coalesce(new.serial_no, ""),
		coalesce(new.manufacturer, ""),
		coalesce(new.notes, ""),
		coalesce(new.custom_attrs, "")
	);
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER assets_after_delete AFTER DELETE ON assets BEGIN
	INSERT INTO assets_fts(assets_fts, rowid, id, name, tag, category, model, model_no, serial_no, manufacturer, notes, custom_attrs) VALUES (
		'delete',
		old.id,
		old.id,
		coalesce(old.name, ""),
		coalesce(old.tag, ""),
		coalesce(old.category, ""),
		coalesce(old.model, ""),
		coalesce(old.model_no, ""),
		coalesce(old.serial_no, ""),
		coalesce(old.manufacturer, ""),
		coalesce(old.notes, ""),
		coalesce(old.custom_attrs, "")
	);
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER assets_after_update AFTER UPDATE ON assets BEGIN
	INSERT INTO assets_fts(assets_fts, rowid, id, name, tag, category, model, model_no, serial_no, manufacturer, notes, custom_attrs) VALUES (
		'delete',
		old.id,
		old.id,
		coalesce(old.name, ""),
		coalesce(old.tag, ""),
		coalesce(old.category, ""),
		coalesce(old.model, ""),
		coalesce(old.model_no, ""),
		coalesce(old.serial_no, ""),
		coalesce(old.manufacturer, ""),
		coalesce(old.notes, ""),
		coalesce(old.custom_attrs, "")
	);

	INSERT INTO assets_fts(rowid, id, name, tag, category, model, model_no, serial_no, manufacturer, notes, custom_attrs) VALUES (
		new.id,
		new.id,
		coalesce(new.name, ""),
		coalesce(new.tag, ""),
		coalesce(new.category, ""),
		coalesce(new.model, ""),
		coalesce(new.model_no, ""),
		coalesce(new.serial_no, ""),
		coalesce(new.manufacturer, ""),
		coalesce(new.notes, ""),
		coalesce(new.custom_attrs, "")
	);
END;
-- +goose StatementEnd


-- +goose Down
DROP TRIGGER assets_after_update;
DROP TRIGGER assets_after_delete;
DROP TRIGGER assets_after_insert;
DROP TABLE assets_fts;

DROP VIEW custom_attr_names;
DROP VIEW position_codes;
DROP VIEW locations;
DROP VIEW suppliers;
DROP VIEW manufacturers;
DROP VIEW categories;

DROP TABLE asset_files;
DROP TABLE asset_parts;
DROP TABLE assets;

DROP INDEX unique_tag;
DROP TABLE tags;

DROP INDEX unique_usernames;
DROP INDEX unique_auth_ref;
DROP TABLE users;

DROP INDEX unique_usernames_auth_method_local;
DROP TABLE local_auth_users;

DROP INDEX unique_session_token;
DROP TABLE sessions;
