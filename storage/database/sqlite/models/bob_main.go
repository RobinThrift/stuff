// Code generated by BobGen sqlite v0.22.0. DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package models

import (
	"context"

	"github.com/stephenafamo/bob/dialect/sqlite"
	"github.com/stephenafamo/bob/dialect/sqlite/dialect"
)

var TableNames = struct {
	AssetFiles       string
	Assets           string
	LocalAuthUsers   string
	Tags             string
	Users            string
	CustomAttrNames  string
	Manufacturers    string
	StatusNames      string
	StorageLocations string
	Suppliers        string
}{
	AssetFiles:       "asset_files",
	Assets:           "assets",
	LocalAuthUsers:   "local_auth_users",
	Tags:             "tags",
	Users:            "users",
	CustomAttrNames:  "custom_attr_names",
	Manufacturers:    "manufacturers",
	StatusNames:      "status_names",
	StorageLocations: "storage_locations",
	Suppliers:        "suppliers",
}

var ColumnNames = struct {
	AssetFiles       assetFileColumnNames
	Assets           assetColumnNames
	LocalAuthUsers   localAuthUserColumnNames
	Tags             tagColumnNames
	Users            userColumnNames
	CustomAttrNames  customAttrNameColumnNames
	Manufacturers    manufacturerColumnNames
	StatusNames      statusNameColumnNames
	StorageLocations storageLocationColumnNames
	Suppliers        supplierColumnNames
}{
	AssetFiles: assetFileColumnNames{
		ID:        "id",
		AssetID:   "asset_id",
		Name:      "name",
		Sha256:    "sha256",
		SizeBytes: "size_bytes",
		CreatedBy: "created_by",
		CreatedAt: "created_at",
		UpdatedAt: "updated_at",
	},
	Assets: assetColumnNames{
		ID:               "id",
		ParentAssetID:    "parent_asset_id",
		Status:           "status",
		Name:             "name",
		SerialNo:         "serial_no",
		ModelNo:          "model_no",
		Manufacturer:     "manufacturer",
		Notes:            "notes",
		ImageURL:         "image_url",
		ThumbnailURL:     "thumbnail_url",
		WarrantyUntil:    "warranty_until",
		CustomAttrs:      "custom_attrs",
		TagID:            "tag_id",
		CheckedOutTo:     "checked_out_to",
		StorageLocation:  "storage_location",
		StorageShelf:     "storage_shelf",
		PurchaseSupplier: "purchase_supplier",
		PurchaseOrderNo:  "purchase_order_no",
		PurchaseDate:     "purchase_date",
		PurchaseAmount:   "purchase_amount",
		PurchaseCurrency: "purchase_currency",
		CreatedBy:        "created_by",
		CreatedAt:        "created_at",
		UpdatedAt:        "updated_at",
	},
	LocalAuthUsers: localAuthUserColumnNames{
		ID:                     "id",
		Username:               "username",
		Algorithm:              "algorithm",
		Params:                 "params",
		Salt:                   "salt",
		Password:               "password",
		RequiresPasswordChange: "requires_password_change",
		CreatedAt:              "created_at",
		UpdatedAt:              "updated_at",
	},
	Tags: tagColumnNames{
		ID:        "id",
		Tag:       "tag",
		CreatedAt: "created_at",
		UpdatedAt: "updated_at",
	},
	Users: userColumnNames{
		ID:          "id",
		Username:    "username",
		DisplayName: "display_name",
		IsAdmin:     "is_admin",
		AuthRef:     "auth_ref",
		CreatedAt:   "created_at",
		UpdatedAt:   "updated_at",
	},
	CustomAttrNames: customAttrNameColumnNames{
		Name: "name",
		Type: "type",
	},
	Manufacturers: manufacturerColumnNames{
		Name: "name",
	},
	StatusNames: statusNameColumnNames{
		Name: "name",
	},
	StorageLocations: storageLocationColumnNames{
		Name: "name",
	},
	Suppliers: supplierColumnNames{
		Name: "name",
	},
}

var (
	SelectWhere = Where[*dialect.SelectQuery]()
	InsertWhere = Where[*dialect.InsertQuery]()
	UpdateWhere = Where[*dialect.UpdateQuery]()
	DeleteWhere = Where[*dialect.DeleteQuery]()
)

func Where[Q sqlite.Filterable]() struct {
	AssetFiles       assetFileWhere[Q]
	Assets           assetWhere[Q]
	LocalAuthUsers   localAuthUserWhere[Q]
	Tags             tagWhere[Q]
	Users            userWhere[Q]
	CustomAttrNames  customAttrNameWhere[Q]
	Manufacturers    manufacturerWhere[Q]
	StatusNames      statusNameWhere[Q]
	StorageLocations storageLocationWhere[Q]
	Suppliers        supplierWhere[Q]
} {
	return struct {
		AssetFiles       assetFileWhere[Q]
		Assets           assetWhere[Q]
		LocalAuthUsers   localAuthUserWhere[Q]
		Tags             tagWhere[Q]
		Users            userWhere[Q]
		CustomAttrNames  customAttrNameWhere[Q]
		Manufacturers    manufacturerWhere[Q]
		StatusNames      statusNameWhere[Q]
		StorageLocations storageLocationWhere[Q]
		Suppliers        supplierWhere[Q]
	}{
		AssetFiles:       AssetFileWhere[Q](),
		Assets:           AssetWhere[Q](),
		LocalAuthUsers:   LocalAuthUserWhere[Q](),
		Tags:             TagWhere[Q](),
		Users:            UserWhere[Q](),
		CustomAttrNames:  CustomAttrNameWhere[Q](),
		Manufacturers:    ManufacturerWhere[Q](),
		StatusNames:      StatusNameWhere[Q](),
		StorageLocations: StorageLocationWhere[Q](),
		Suppliers:        SupplierWhere[Q](),
	}
}

var (
	SelectJoins = getJoins[*dialect.SelectQuery]
	UpdateJoins = getJoins[*dialect.UpdateQuery]
)

type joinSet[Q any] struct {
	InnerJoin Q
	LeftJoin  Q
	RightJoin Q
}

type joins[Q dialect.Joinable] struct {
	AssetFiles joinSet[assetFileRelationshipJoins[Q]]
	Assets     joinSet[assetRelationshipJoins[Q]]
	Tags       joinSet[tagRelationshipJoins[Q]]
	Users      joinSet[userRelationshipJoins[Q]]
}

func getJoins[Q dialect.Joinable](ctx context.Context) joins[Q] {
	return joins[Q]{
		AssetFiles: assetFilesJoin[Q](ctx),
		Assets:     assetsJoin[Q](ctx),
		Tags:       tagsJoin[Q](ctx),
		Users:      usersJoin[Q](ctx),
	}
}
