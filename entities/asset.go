package entities

import (
	"fmt"
	"time"
)

type MonetaryAmount int64

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

type AssetType string

const (
	AssetTypeAsset      AssetType = "ASSET"
	AssetTypeComponent  AssetType = "COMPONENT"
	AssetTypeConsumable AssetType = "CONSUMABLE"
)

type Asset struct {
	ID   int64     `form:"-"`
	Type AssetType `form:"type"`

	ParentAssetID int64    `form:"parent_asset_id"`
	Parent        *Asset   `form:"-"`
	Children      []*Asset `form:"-"`

	Status Status `form:"status"`

	Tag           string       `form:"tag"`
	Name          string       `form:"name"`
	Category      string       `form:"category"`
	Model         string       `form:"model"`
	ModelNo       string       `form:"model_no"`
	SerialNo      string       `form:"serial_no"`
	Manufacturer  string       `form:"manufacturer"`
	Notes         string       `form:"notes"`
	ImageURL      string       `form:"-"`
	ThumbnailURL  string       `form:"-"`
	WarrantyUntil time.Time    `form:"warranty_until,omitempty"`
	Quantity      uint64       `form:"quantity"`
	QuantityUnit  string       `form:"quantity_unit"`
	CustomAttrs   []CustomAttr `form:"custom_attrs"`

	CheckedOutTo int64  `form:"checked_out_to"`
	Location     string `form:"location"`
	PositionCode string `form:"position_code"`

	Purchases []*Purchase `form:"purchases"`

	PartsTotalCounter int     `form:"parts_total_counter"`
	Parts             []*Part `form:"parts"`

	Files []*File `form:"-"`

	MetaInfo MetaInfo `form:"-"`
}

type CustomAttr struct {
	Name  string `form:"name" json:"name,omitempty"`
	Value any    `form:"value" json:"value,omitempty"`
}

type Purchase struct {
	Supplier string         `form:"supplier,omitempty"`
	OrderNo  string         `form:"order_no,omitempty"`
	Date     time.Time      `form:"order_date,omitempty"`
	Amount   MonetaryAmount `form:"amount,omitempty"`
	Currency string         `form:"currency,omitempty"`
}

type MetaInfo struct {
	CreatedBy int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Part struct {
	ID      int64 `from:"id"`
	AssetID int64 `form:"asset_id"`

	Tag          string `form:"tag"`
	Name         string `form:"name"`
	Location     string `form:"location"`
	PositionCode string `form:"position_code"`
	Notes        string `form:"notes"`

	CreatedBy int64     `form:"-"`
	CreatedAt time.Time `form:"-"`
	UpdatedAt time.Time `form:"-"`
}

type ListAssetsQuery struct {
	Search *ListAssetsQuerySearch

	IDs []int64

	Page     int
	PageSize int

	OrderBy  string
	OrderDir string

	AssetType AssetType
}

type ListAssetsQuerySearch struct {
	Raw string

	Fields map[string]string
}

type ListCategoriesQuery struct {
	Search   string
	Page     int
	PageSize int
}

type ListCustomAttrNamesQuery struct {
	Search   string
	Page     int
	PageSize int
}

type ListLocationsQuery struct {
	Search   string
	Page     int
	PageSize int
}

type ListPositionCodesQuery struct {
	Search   string
	Page     int
	PageSize int
}
