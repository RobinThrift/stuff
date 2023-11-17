-- +goose Up
-- +goose StatementBegin
CREATE TABLE user_preferences (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id    INTEGER NOT NULL,
    key        TEXT NOT NULL,
    value      BLOB NOT NULL,

    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%SZ', CURRENT_TIMESTAMP)),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%SZ', CURRENT_TIMESTAMP)),

    FOREIGN KEY(user_id) REFERENCES users(id)
);
CREATE UNIQUE INDEX user_preferences_key_user_id_idx ON user_preferences(user_id, key);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX user_preferences_key_user_id_idx;
DROP TABLE user_preferences;
-- +goose StatementEnd
