package database

import (
	"errors"
	"time"
)

var ErrAssetNotFound = errors.New("asset not found")

type Asset struct {
	ID            int64
	ParentAssetID int64
	Status        string

	Tag           string
	Name          string
	Category      string
	Model         string
	ModelNo       string
	SerialNo      string
	Manufacturer  string
	Notes         string
	ImageURL      string
	ThumbnailURL  string
	WarrantyUntil time.Time
	CustomAttrs   map[string]any

	CheckedOutTo int64
	Location     string
	PositionCode string

	PurchaseSupplier string
	PurchaseOrderNo  string
	PurchaseDate     time.Time
	PurchaseAmount   int
	PurchaseCurrency string

	CreatedBy int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type AssetFile struct {
	ID      int64
	AssetID int64

	Name      string
	Sha256    []byte
	SizeBytes int64

	CreatedBy int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type AssetList struct {
	Assets []*Asset
	Total  int
}

type ListAssetsQuery struct {
	Offset   int
	Limit    int
	OrderBy  string
	OrderDir string
}

var ErrUserNotFound = errors.New("user not found")

type User struct {
	ID       int64
	Username string

	DisplayName string
	IsAdmin     bool

	AuthRef string

	CreatedAt time.Time
	UpdatedAt time.Time
}

var ErrLocalAuthUserNotFound = errors.New("user for local auth not found")

type LocalAuthUser struct {
	ID                     int64
	Username               string
	Algorithm              string
	Params                 string
	Salt                   []byte
	Password               []byte
	RequiresPasswordChange bool
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

var ErrTagNotFound = errors.New("tag not found")

type ListTagsQuery struct {
	Offset   int
	Limit    int
	OrderBy  string
	OrderDir string
}

type Tag struct {
	ID        int64
	Tag       string
	InUse     bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type TagList struct {
	Tags  []*Tag
	Total int
}
