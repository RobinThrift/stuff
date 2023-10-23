-- +goose Up
-- +goose StatementBegin
CREATE TABLE asset_purchases (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    asset_id    INTEGER NOT NULL,

    supplier   TEXT DEFAULT NULL,
    order_no   TEXT DEFAULT NULL,
    order_date TEXT DEFAULT NULL,
    amount     INT  DEFAULT NULL,
    currency   TEXT DEFAULT NULL,

    created_by INTEGER NOT NULL,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%SZ', CURRENT_TIMESTAMP)),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%SZ', CURRENT_TIMESTAMP)),

    FOREIGN KEY(asset_id) REFERENCES assets(id),
    FOREIGN KEY(created_by) REFERENCES users(id)
);

INSERT INTO asset_purchases(asset_id, supplier, order_no, order_date, amount, currency, created_by)
    SELECT
        assets.id as asset_id,
        assets.purchase_supplier as supplier,
        assets.purchase_order_no as order_no,
        assets.purchase_date as order_date,
        assets.purchase_amount as amount,
        assets.purchase_currency as currency,
        assets.created_by as created_by
    FROM assets
    WHERE
        assets.purchase_supplier NOT NULL OR
        assets.purchase_order_no NOT NULL OR
        assets.purchase_date NOT NULL OR
        assets.purchase_amount NOT NULL;

DROP VIEW suppliers;
CREATE VIEW suppliers AS SELECT DISTINCT supplier as name FROM asset_purchases WHERE supplier IS NOT NULL GROUP BY name;

ALTER TABLE assets DROP COLUMN purchase_supplier;
ALTER TABLE assets DROP COLUMN purchase_order_no;
ALTER TABLE assets DROP COLUMN purchase_date;
ALTER TABLE assets DROP COLUMN purchase_amount;
ALTER TABLE assets DROP COLUMN purchase_currency;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE assets ADD COLUMN purchase_supplier TEXT DEFAULT NULL;
ALTER TABLE assets ADD COLUMN purchase_order_no TEXT DEFAULT NULL;
ALTER TABLE assets ADD COLUMN purchase_date     TEXT DEFAULT NULL;
ALTER TABLE assets ADD COLUMN purchase_amount   INT  DEFAULT NULL;
ALTER TABLE assets ADD COLUMN purchase_currency TEXT DEFAULT NULL;

UPDATE assets SET
    purchase_supplier = ap.supplier,
    purchase_order_no = ap.order_no,
    purchase_date = ap.order_date,
    purchase_amount = ap.amount,
    purchase_currency = ap.currency
FROM assets as a JOIN asset_purchases as ap ON a.id = ap.asset_id
WHERE assets.id = ap.asset_id;

DROP TABLE asset_purchases;

DROP VIEW suppliers;
CREATE VIEW suppliers AS SELECT purchase_supplier as name FROM assets WHERE purchase_supplier IS NOT NULL GROUP BY name;
-- +goose StatementEnd
