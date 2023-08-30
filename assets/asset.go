package assets

import (
	"fmt"
	"io"
	"time"
)

type MonetaryAmount int

func (c MonetaryAmount) Format(decimalSeparator string) string {
	fraction := c % 100
	base := (c - fraction) / 100
	return fmt.Sprintf("%d%s%0.2d", base, decimalSeparator, fraction)
}

type Status string

const (
	StatusInStorage Status = "IN_STORAGE"
	StatusInUse     Status = "IN_USE"
	StatusArchived  Status = "ARCHIVED"
)

type Asset struct {
	ID            int64  `form:"-"`
	ParentAssetID int64  `form:"parent_asset_id"`
	Status        Status `form:"status"`

	Tag           string         `form:"tag"`
	Name          string         `form:"name"`
	Category      string         `form:"category"`
	Model         string         `form:"model"`
	ModelNo       string         `form:"model_no"`
	SerialNo      string         `form:"serial_no"`
	Manufacturer  string         `form:"manufacturer"`
	Notes         string         `form:"notes"`
	ImageURL      string         `form:"-"`
	ThumbnailURL  string         `form:"-"`
	WarrantyUntil time.Time      `form:"warranty_until,omitempty"`
	CustomAttrs   map[string]any `form:"custom_attrs"`

	CheckedOutTo int64  `form:"checked_out_to"`
	Location     string `form:"location"`
	PositionCode string `form:"position_code"`

	PurchaseInfo PurchaseInfo `form:"purchase"`

	MetaInfo MetaInfo `form:"-"`
}

type PurchaseInfo struct {
	Supplier string         `form:"supplier"`
	OrderNo  string         `form:"order_no"`
	Date     time.Time      `form:"date,omitempty"`
	Amount   MonetaryAmount `form:"amount"`
	Currency string         `form:"currency"`
}

type MetaInfo struct {
	CreatedBy int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type File struct {
	ID      int64
	AssetID int64

	Name      string
	Sha256    []byte
	SizeBytes int64

	CreatedBy int64
	CreatedAt time.Time
	UpdatedAt time.Time

	r io.Reader
}

type AssetList struct {
	Assets []*Asset
	Total  int
}
