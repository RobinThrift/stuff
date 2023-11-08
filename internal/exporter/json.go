package exporter

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/RobinThrift/stuff/entities"
)

func ExportAssetsAsJSON(w io.Writer, assets []*entities.Asset) error {
	encoder := json.NewEncoder(w)

	forExport := make([]*Asset, 0, len(assets))

	for _, asset := range assets {
		customAttrs := make([]CustomAttr, 0, len(asset.CustomAttrs))
		for _, ca := range asset.CustomAttrs {
			customAttrs = append(customAttrs, CustomAttr(ca))
		}

		forExport = append(forExport, &Asset{
			ID:              int(asset.ID),
			ParentAssetID:   int(asset.ParentAssetID),
			Category:        (asset.Category),
			CheckedOutTo:    int(asset.CheckedOutTo),
			CreatedAt:       asset.MetaInfo.CreatedAt,
			CreatedBy:       int(asset.MetaInfo.CreatedBy),
			CustomAttrs:     customAttrs,
			ImageURL:        asset.ImageURL,
			Location:        (asset.Location),
			Manufacturer:    (asset.Manufacturer),
			Model:           (asset.Model),
			ModelNo:         (asset.ModelNo),
			Name:            asset.Name,
			Notes:           (asset.Notes),
			PartsTotalCount: asset.PartsTotalCounter,
			PositionCode:    (asset.PositionCode),
			Quantity:        int(asset.Quantity),
			QuantityUnit:    asset.QuantityUnit,
			SerialNo:        (asset.SerialNo),
			Status:          string(asset.Status),
			Tag:             asset.Tag,
			ThumbnailURL:    asset.ThumbnailURL,
			Type:            string(asset.Type),
			UpdatedAt:       asset.MetaInfo.UpdatedAt,
			WarrantyUntil:   (asset.WarrantyUntil),
		})
	}

	err := encoder.Encode(forExport)
	if err != nil {
		return fmt.Errorf("error exporting assets as JSON: %w", err)
	}

	return nil
}

type Asset struct {
	ID              int          `json:"id"`
	ParentAssetID   int          `json:"parentAssetID"`
	Category        string       `json:"category,omitempty"`
	CheckedOutTo    int          `json:"checkedOutTo,omitempty"`
	CreatedAt       time.Time    `json:"createdAt"`
	CreatedBy       int          `json:"createdBy"`
	CustomAttrs     []CustomAttr `json:"customAttrs"`
	ImageURL        string       `json:"imageURL,omitempty"`
	Location        string       `json:"location,omitempty"`
	Manufacturer    string       `json:"manufacturer,omitempty"`
	Model           string       `json:"model,omitempty"`
	ModelNo         string       `json:"modelNo,omitempty"`
	Name            string       `json:"name"`
	Notes           string       `json:"notes,omitempty"`
	PartsTotalCount int          `json:"partsTotalCount,omitempty"`
	PositionCode    string       `json:"positionCode,omitempty"`
	Quantity        int          `json:"quantity,omitempty"`
	QuantityUnit    string       `json:"quantityUnit,omitempty"`
	SerialNo        string       `json:"serialNo,omitempty"`
	Status          string       `json:"status"`
	Tag             string       `json:"tag"`
	ThumbnailURL    string       `json:"thumbnailURL,omitempty"`
	Type            string       `json:"type"`
	UpdatedAt       time.Time    `json:"updatedAt"`
	WarrantyUntil   time.Time    `json:"warrantyUntil,omitempty"`
}

type CustomAttr struct {
	Name  string `json:"name"`
	Value any    `json:"value"`
}
