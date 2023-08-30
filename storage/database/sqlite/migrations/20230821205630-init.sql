-- +migrate Up
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
    tag          TEXT NOT NULL,
    created_at   TEXT NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%SZ', CURRENT_TIMESTAMP)),
    updated_at   TEXT NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%SZ', CURRENT_TIMESTAMP))
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

    created_by INTEGER NOT NULL,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%SZ', CURRENT_TIMESTAMP)),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%SZ', CURRENT_TIMESTAMP)),

    FOREIGN KEY(parent_asset_id) REFERENCES assets(id),
    FOREIGN KEY(tag) REFERENCES tags(tag),
    FOREIGN KEY(checked_out_to) REFERENCES users(id),
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

    FOREIGN KEY(asset_id) REFERENCES asset(id),
    FOREIGN KEY(created_by) REFERENCES users(id)
);

CREATE VIEW categories AS SELECT category as name FROM assets GROUP BY name;
CREATE VIEW manufacturers AS SELECT manufacturer as name FROM assets WHERE manufacturer IS NOT NULL GROUP BY name;
CREATE VIEW suppliers AS SELECT purchase_supplier as name FROM assets WHERE purchase_supplier IS NOT NULL GROUP BY name;
CREATE VIEW locations AS SELECT location as name FROM assets WHERE location IS NOT NULL GROUP BY name;
CREATE VIEW position_codes AS SELECT position_code as code FROM assets WHERE position_code IS NOT NULL GROUP BY code;
CREATE VIEW custom_attr_names AS SELECT j.key as name, j.type as type FROM assets, json_each(custom_attrs) j WHERE custom_attrs IS NOT NULL GROUP BY j.key;

-- +migrate Down
DROP VIEW custom_attr_names;
DROP VIEW position_codes;
DROP VIEW locations;
DROP VIEW suppliers;
DROP VIEW manufacturers;
DROP VIEW categories;

DROP TABLE asset_files;
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
