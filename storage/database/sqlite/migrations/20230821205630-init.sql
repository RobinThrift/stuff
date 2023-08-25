-- +migrate Up
CREATE TABLE local_auth_users (
    id                       INTEGER PRIMARY KEY,
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
    id              INTEGER PRIMARY KEY,

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
    id           INTEGER PRIMARY KEY,
    tag          TEXT NOT NULL,
    created_at   TEXT NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%SZ', CURRENT_TIMESTAMP)),
    updated_at   TEXT NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%SZ', CURRENT_TIMESTAMP))
);
CREATE UNIQUE INDEX unique_tag ON tags(tag);

CREATE TABLE assets (
    id              INTEGER PRIMARY KEY,
    parent_asset_id INTEGER DEFAULT NULL,
    status          TEXT NOT NULL,

    name           TEXT NOT NULL,
    serial_no      TEXT DEFAULT NULL,
    model_no       TEXT DEFAULT NULL,
    manufacturer   TEXT DEFAULT NULL,
    notes          TEXT DEFAULT NULL,
    image_url      TEXT DEFAULT NULL,
    thumbnail_url  TEXT DEFAULT NULL,
    warranty_until TEXT DEFAULT NULL,
    custom_attrs   TEXT DEFAULT NULL,

    tag_id           INTEGER DEFAULT NULL,
    checked_out_to   INTEGER DEFAULT NULL,
    storage_location TEXT DEFAULT NULL,
    storage_shelf    TEXT DEFAULT NULL,

    purchase_supplier TEXT DEFAULT NULL,
    purchase_order_no TEXT DEFAULT NULL,
    purchase_date     TEXT DEFAULT NULL,
    purchase_amount   TEXT DEFAULT NULL,
    purchase_currency TEXT DEFAULT NULL,

    created_by INTEGER NOT NULL,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%SZ', CURRENT_TIMESTAMP)),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%SZ', CURRENT_TIMESTAMP)),

    FOREIGN KEY(parent_asset_id) REFERENCES assets(id),
    FOREIGN KEY(tag_id) REFERENCES tags(id),
    FOREIGN KEY(checked_out_to) REFERENCES users(id),
    FOREIGN KEY(created_by) REFERENCES users(id)
);

CREATE TABLE asset_files (
    id          INTEGER PRIMARY KEY,
    asset_id    INTEGER NOT NULL,

    name        TEXT NOT NULL,
    sha256      BLOB NOT NULL,
    size_bytes  INT NOT NULL,

    created_by INTEGER NOT NULL,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%SZ', CURRENT_TIMESTAMP)),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%SZ', CURRENT_TIMESTAMP)),

    FOREIGN KEY(asset_id) REFERENCES asset(id),
    FOREIGN KEY(created_by) REFERENCES users(id)
);

CREATE VIEW status_names AS SELECT status as name FROM assets GROUP BY name;
CREATE VIEW manufacturers AS SELECT manufacturer as name FROM assets WHERE manufacturer IS NOT NULL GROUP BY name;
CREATE VIEW suppliers AS SELECT purchase_supplier as name FROM assets WHERE purchase_supplier IS NOT NULL GROUP BY name;
CREATE VIEW storage_locations AS SELECT storage_location as name FROM assets WHERE storage_location IS NOT NULL GROUP BY name;
CREATE VIEW custom_attr_names AS SELECT j.key as name, j.type as type FROM assets, json_each(custom_attrs) j WHERE custom_attrs IS NOT NULL GROUP BY j.key;

-- +migrate Down
DROP VIEW custom_attr_names;
DROP VIEW storage_location;
DROP VIEW suppliers;
DROP VIEW manufacturers;
DROP VIEW status_names;

DROP TABLE asset_files;
DROP TABLE assets;

DROP INDEX unique_tag;
DROP TABLE tags;

DROP INDEX unique_usernames;
DROP INDEX unique_auth_ref;
DROP TABLE users;

DROP INDEX unique_usernames_auth_method_local;
DROP TABLE local_auth_users;
