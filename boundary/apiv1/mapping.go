//lint:file-ignore SA1019 Ignore because generated code produces these
package apiv1

import (
	"fmt"
	"strings"
	"time"

	"github.com/RobinThrift/stuff/entities"
	"github.com/deepmap/oapi-codegen/pkg/types"
)

func mapCreateAssetBodyToAsset(asset *CreateAssetJSONRequestBody) *entities.Asset {
	purchases := make([]*entities.Purchase, 0, len(asset.Purchases))
	for _, p := range asset.Purchases {
		purchases = append(purchases, &entities.Purchase{
			Supplier: valFromPtr(p.Supplier),
			OrderNo:  valFromPtr(p.OrderNo),
			Date:     valFromPtr(p.Date).Time,
			Amount:   entities.MonetaryAmount(valFromPtr(p.Amount)),
			Currency: valFromPtr(p.Currency),
		})
	}

	customAttrs := make([]entities.CustomAttr, 0, len(asset.CustomAttrs))
	for _, ca := range asset.CustomAttrs {
		customAttrs = append(customAttrs, entities.CustomAttr{Name: ca.Name, Value: ca.Value})
	}

	parts := make([]*entities.Part, 0, len(asset.Parts))
	for _, part := range asset.Parts {
		parts = append(parts, &entities.Part{
			ID:           int64(part.Id),
			AssetID:      int64(part.AssetID),
			Tag:          part.Tag,
			Name:         part.Name,
			Location:     valFromPtr(part.Location),
			PositionCode: valFromPtr(part.PositionCode),
			Notes:        valFromPtr(part.Notes),
			CreatedBy:    int64(part.CreatedBy),
			CreatedAt:    part.CreatedAt.Time,
			UpdatedAt:    part.UpdatedAt.Time,
		})
	}

	return &entities.Asset{
		Type:              entities.AssetType(asset.Type),
		ParentAssetID:     int64(valFromPtr[int](asset.ParentAssetID)),
		Parent:            &entities.Asset{},
		Children:          []*entities.Asset{},
		Status:            entities.Status(asset.Status),
		Tag:               valFromPtr(asset.Tag),
		Name:              asset.Name,
		Category:          valFromPtr(asset.Category),
		Model:             valFromPtr(asset.Model),
		ModelNo:           valFromPtr(asset.ModelNo),
		SerialNo:          valFromPtr(asset.SerialNo),
		Manufacturer:      valFromPtr(asset.Manufacturer),
		Notes:             valFromPtr(asset.Notes),
		WarrantyUntil:     valFromPtr(asset.WarrantyUntil).Time,
		Quantity:          uint64(valFromPtr(asset.Quantity)),
		QuantityUnit:      valFromPtr(asset.QuantityUnit),
		CustomAttrs:       customAttrs,
		Location:          valFromPtr(asset.Location),
		PositionCode:      valFromPtr(asset.PositionCode),
		Purchases:         purchases,
		PartsTotalCounter: valFromPtr(asset.PartsTotalCount),
		Parts:             parts,
		Files:             []*entities.File{},
		MetaInfo:          entities.MetaInfo{},
	}
}

func mapUpdateIntoAsset(asset *entities.Asset, update *UpdateAssetJSONRequestBody) {
	asset.Purchases = make([]*entities.Purchase, 0, len(update.Purchases))
	for _, p := range update.Purchases {
		asset.Purchases = append(asset.Purchases, &entities.Purchase{
			Supplier: valFromPtr(p.Supplier),
			OrderNo:  valFromPtr(p.OrderNo),
			Date:     valFromPtr(p.Date).Time,
			Amount:   entities.MonetaryAmount(valFromPtr(p.Amount)),
			Currency: valFromPtr(p.Currency),
		})
	}

	asset.CustomAttrs = make([]entities.CustomAttr, 0, len(asset.CustomAttrs))
	for _, ca := range asset.CustomAttrs {
		asset.CustomAttrs = append(asset.CustomAttrs, entities.CustomAttr{Name: ca.Name, Value: ca.Value})
	}

	asset.Parts = make([]*entities.Part, 0, len(update.Parts))
	for _, part := range update.Parts {
		asset.Parts = append(asset.Parts, &entities.Part{
			ID:           int64(part.Id),
			AssetID:      int64(part.AssetID),
			Tag:          part.Tag,
			Name:         part.Name,
			Location:     valFromPtr(part.Location),
			PositionCode: valFromPtr(part.PositionCode),
			Notes:        valFromPtr(part.Notes),
			CreatedBy:    int64(part.CreatedBy),
			CreatedAt:    part.CreatedAt.Time,
			UpdatedAt:    part.UpdatedAt.Time,
		})
	}

	asset.Type = entities.AssetType(update.Type)
	asset.Tag = valFromPtr(update.Tag)
	asset.ParentAssetID = int64(valFromPtr[int](update.ParentAssetID))
	asset.Status = entities.Status(update.Status)
	asset.Name = update.Name
	asset.Category = valFromPtr(update.Category)
	asset.Model = valFromPtr(update.Model)
	asset.ModelNo = valFromPtr(update.ModelNo)
	asset.SerialNo = valFromPtr(update.SerialNo)
	asset.Manufacturer = valFromPtr(update.Manufacturer)
	asset.Notes = valFromPtr(update.Notes)
	asset.WarrantyUntil = valFromPtr(update.WarrantyUntil).Time
	asset.Location = valFromPtr(update.Location)
	asset.PositionCode = valFromPtr(update.PositionCode)
	asset.Quantity = uint64(valFromPtr(update.Quantity))
	asset.QuantityUnit = valFromPtr(update.QuantityUnit)
}

func mapAssetToAPI(asset *entities.Asset) Asset {
	purchases := make([]Purchase, 0, len(asset.Purchases))
	for _, p := range asset.Purchases {
		purchases = append(purchases, Purchase{
			Supplier: ptrFromVal(p.Supplier),
			OrderNo:  ptrFromVal(p.OrderNo),
			Date:     timeToDate(p.Date),
			Amount:   ptrFromVal(int(p.Amount)),
			Currency: ptrFromVal(p.Currency),
		})
	}

	customAttrs := make([]CustomAttr, 0, len(asset.CustomAttrs))
	for _, ca := range asset.CustomAttrs {
		customAttrs = append(customAttrs, CustomAttr(ca))
	}

	parts := make([]AssetPart, 0, len(asset.Parts))
	for _, part := range asset.Parts {
		parts = append(parts, AssetPart{
			Id:           int(part.ID),
			AssetID:      int(part.AssetID),
			Tag:          part.Tag,
			Name:         part.Name,
			Location:     ptrFromVal(part.Location),
			PositionCode: ptrFromVal(part.PositionCode),
			Notes:        ptrFromVal(part.Notes),
			CreatedBy:    int(part.CreatedBy),
			CreatedAt:    valFromPtr(timeToDate(part.CreatedAt)),
			UpdatedAt:    valFromPtr(timeToDate(part.UpdatedAt)),
		})
	}

	files := make([]AssetFile, 0, len(asset.Files))
	for _, file := range asset.Files {
		files = append(files, AssetFile{
			AssetID:    int(file.AssetID),
			Id:         int(file.ID),
			Name:       file.Name,
			Filetype:   file.Filetype,
			PublicPath: file.PublicPath,
			Sha256:     fmt.Sprintf("%x", file.Sha256),
			SizeBytes:  int(file.SizeBytes),
			CreatedBy:  int(file.CreatedBy),
			CreatedAt:  valFromPtr(timeToDate(file.CreatedAt)),
			UpdatedAt:  valFromPtr(timeToDate(file.UpdatedAt)),
		})
	}

	children := make([]Asset, 0, len(asset.Children))
	for _, c := range asset.Children {
		children = append(children, mapAssetToAPI(c))
	}

	return Asset{
		Id:            int(asset.ID),
		ParentAssetID: ptrFromVal(int(asset.ParentAssetID)),
		Tag:           asset.Tag,
		Status:        AssetStatus(asset.Status),
		Name:          asset.Name,
		Category:      ptrFromVal(asset.Category),
		Model:         ptrFromVal(asset.Model),
		ModelNo:       ptrFromVal(asset.ModelNo),
		SerialNo:      ptrFromVal(asset.SerialNo),
		Manufacturer:  ptrFromVal(asset.Manufacturer),
		Notes:         ptrFromVal(asset.Notes),
		WarrantyUntil: timeToDate(asset.WarrantyUntil),
		CustomAttrs:   customAttrs,
		Parts:         parts,
		Location:      ptrFromVal(asset.Location),
		PositionCode:  ptrFromVal(asset.PositionCode),
		Purchases:     purchases,
		Files:         files,
		Children:      &children,
		CreatedBy:     int(asset.MetaInfo.CreatedBy),
		CreatedAt:     types.Date{Time: asset.MetaInfo.CreatedAt},
		UpdatedAt:     types.Date{Time: asset.MetaInfo.UpdatedAt},
	}
}

func mapTagToAPI(tag *entities.Tag) Tag {
	return Tag{
		Id:        int(tag.ID),
		Tag:       tag.Tag,
		InUse:     tag.InUse,
		CreatedAt: types.Date{Time: tag.CreatedAt},
		UpdatedAt: types.Date{Time: tag.UpdatedAt},
	}
}

func ptrFromVal[T comparable](v T) *T {
	var zero T
	if v == zero {
		return nil
	}

	return &v
}

func valFromPtr[T any](v *T) T {
	var zero T
	if v == nil {
		return zero
	}

	return *v
}

func timeToDate(t time.Time) *types.Date {
	if t.IsZero() {
		return nil
	}

	return &types.Date{Time: t}
}

func decodeSearchQuery(queryPtr *string) map[string]string {
	if queryPtr == nil {
		return nil
	}

	queryStr := strings.TrimPrefix(*queryPtr, "*")

	fields := map[string]string{}

	lastWordEnd := 0
	lastNameEnd := 0
	name := ""
	value := ""
	for i := 0; i < len(queryStr)-1; i++ {
		switch queryStr[i] {
		case ':':
			value = queryStr[lastNameEnd:lastWordEnd]
			if name != "" {
				fields[strings.ToLower(name)] = value
			}
			if queryStr[lastWordEnd] == ' ' {
				lastWordEnd += 1
			}
			name = queryStr[lastWordEnd:i]
			lastNameEnd = i + 1
			if i+1 < len(queryStr) && queryStr[i+1] == ' ' {
				lastNameEnd = i + 2
			}
		case ' ':
			lastWordEnd = i
		}
	}

	if name != "" {
		value = queryStr[lastNameEnd:]
		if value != "" {
			fields[strings.ToLower(name)] = queryStr[lastNameEnd:]
		}
	}

	return fields
}
