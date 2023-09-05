package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/aarondl/opt/null"
	"github.com/aarondl/opt/omit"
	"github.com/aarondl/opt/omitnull"
	"github.com/kodeshack/stuff/storage/database"
	"github.com/kodeshack/stuff/storage/database/sqlite/models"
	"github.com/kodeshack/stuff/storage/database/sqlite/types"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/sqlite"
	"github.com/stephenafamo/bob/dialect/sqlite/dialect"
	"github.com/stephenafamo/bob/dialect/sqlite/sm"
	bmods "github.com/stephenafamo/bob/mods"
)

type AssetRepo struct{}

func (ar *AssetRepo) GetAsset(ctx context.Context, exec bob.Executor, id int64) (*database.Asset, error) {
	asset, err := models.FindAsset(ctx, exec, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: %d", database.ErrAssetNotFound, id)
		}
		return nil, fmt.Errorf("error getting assets: %w", err)
	}

	return &database.Asset{
		ID:               asset.ID,
		ParentAssetID:    asset.ParentAssetID.GetOrZero(),
		Status:           asset.Status,
		Tag:              asset.Tag.GetOrZero(),
		Name:             asset.Name,
		Category:         asset.Category,
		Model:            asset.Model.GetOrZero(),
		ModelNo:          asset.ModelNo.GetOrZero(),
		SerialNo:         asset.SerialNo.GetOrZero(),
		Manufacturer:     asset.Manufacturer.GetOrZero(),
		Notes:            asset.Notes.GetOrZero(),
		ImageURL:         asset.ImageURL.GetOrZero(),
		ThumbnailURL:     asset.ThumbnailURL.GetOrZero(),
		WarrantyUntil:    asset.WarrantyUntil.GetOrZero().Time,
		CustomAttrs:      asset.CustomAttrs.GetOrZero().JSON,
		CheckedOutTo:     asset.CheckedOutTo.GetOrZero(),
		Location:         asset.Location.GetOrZero(),
		PositionCode:     asset.PositionCode.GetOrZero(),
		PurchaseSupplier: asset.PurchaseSupplier.GetOrZero(),
		PurchaseOrderNo:  asset.PurchaseOrderNo.GetOrZero(),
		PurchaseDate:     asset.PurchaseDate.GetOrZero().Time,
		PurchaseAmount:   int(asset.PurchaseAmount.GetOrZero()),
		PurchaseCurrency: asset.PurchaseCurrency.GetOrZero(),
		CreatedBy:        asset.CreatedBy,
		CreatedAt:        asset.CreatedAt.Time,
		UpdatedAt:        asset.UpdatedAt.Time,
	}, nil
}

func (ar *AssetRepo) ListAssets(ctx context.Context, exec bob.Executor, query database.ListAssetsQuery) (*database.AssetList, error) {
	if query.Limit == 0 {
		query.Limit = 50
	}

	mods := []bob.Mod[*dialect.SelectQuery]{
		sm.Limit(query.Limit),
		sm.Offset(query.Offset),
		models.SelectWhere.Assets.ParentAssetID.IsNull(),
	}

	if query.OrderBy != "" {
		if query.OrderDir == "" {
			query.OrderDir = "ASC"
		}

		mods = append(mods, bmods.OrderBy[*dialect.SelectQuery]{
			Expression: query.OrderBy,
			Direction:  query.OrderDir,
		})
	}

	assets, err := models.Assets.Query(ctx, exec, mods...).All()
	if err != nil {
		return nil, fmt.Errorf("error getting assets: %w", err)
	}

	count, err := models.Assets.Query(ctx, exec, mods...).Count()
	if err != nil {
		return nil, fmt.Errorf("error counting assets: %w", err)
	}

	assetList := &database.AssetList{
		Assets: make([]*database.Asset, 0, len(assets)),
		Total:  int(count),
	}

	for i := range assets {
		assetList.Assets = append(assetList.Assets, &database.Asset{
			ID:               assets[i].ID,
			ParentAssetID:    assets[i].ParentAssetID.GetOrZero(),
			Status:           assets[i].Status,
			Tag:              assets[i].Tag.GetOrZero(),
			Name:             assets[i].Name,
			Category:         assets[i].Category,
			Model:            assets[i].Model.GetOrZero(),
			ModelNo:          assets[i].ModelNo.GetOrZero(),
			SerialNo:         assets[i].SerialNo.GetOrZero(),
			Manufacturer:     assets[i].Manufacturer.GetOrZero(),
			Notes:            assets[i].Notes.GetOrZero(),
			ImageURL:         assets[i].ImageURL.GetOrZero(),
			ThumbnailURL:     assets[i].ThumbnailURL.GetOrZero(),
			WarrantyUntil:    assets[i].WarrantyUntil.GetOrZero().Time,
			CustomAttrs:      assets[i].CustomAttrs.GetOrZero().JSON,
			CheckedOutTo:     assets[i].CheckedOutTo.GetOrZero(),
			Location:         assets[i].Location.GetOrZero(),
			PositionCode:     assets[i].PositionCode.GetOrZero(),
			PurchaseSupplier: assets[i].PurchaseSupplier.GetOrZero(),
			PurchaseOrderNo:  assets[i].PurchaseOrderNo.GetOrZero(),
			PurchaseDate:     assets[i].PurchaseDate.GetOrZero().Time,
			PurchaseAmount:   int(assets[i].PurchaseAmount.GetOrZero()),
			PurchaseCurrency: assets[i].PurchaseCurrency.GetOrZero(),
			CreatedBy:        assets[i].CreatedBy,
			CreatedAt:        assets[i].CreatedAt.Time,
			UpdatedAt:        assets[i].UpdatedAt.Time,
		})
	}

	return assetList, nil
}

func (ar *AssetRepo) Search(ctx context.Context, exec bob.Executor, query database.SearchAssetsQuery) (*database.AssetList, error) {
	exec = bob.Debug(exec)

	if query.Limit == 0 {
		query.Limit = 50
	}

	entries, err := models.AssetsFTS.Query(ctx, exec, sm.Where(sqlite.Quote(models.TableNames.AssetsFTS).EQ(sqlite.Quote(query.Search))), sm.Limit(query.Limit),
		sm.Offset(query.Offset),
	).All()
	if err != nil {
		return nil, fmt.Errorf("error searching assets: %w", err)
	}

	ids := make([]int64, 0, len(entries))

	for _, entry := range entries {
		id, err := strconv.ParseInt(entry.ID.GetOrZero(), 10, 64)
		if err != nil {
			continue
		}

		ids = append(ids, id)
	}

	assets, err := models.Assets.Query(ctx, exec, models.SelectWhere.Assets.ID.In(ids...)).All()
	if err != nil {
		return nil, fmt.Errorf("error getting assets: %w", err)
	}

	assetList := &database.AssetList{
		Assets: make([]*database.Asset, 0, len(assets)),
		Total:  len(entries),
	}

	for i := range assets {
		assetList.Assets = append(assetList.Assets, &database.Asset{
			ID:               assets[i].ID,
			ParentAssetID:    assets[i].ParentAssetID.GetOrZero(),
			Status:           assets[i].Status,
			Tag:              assets[i].Tag.GetOrZero(),
			Name:             assets[i].Name,
			Category:         assets[i].Category,
			Model:            assets[i].Model.GetOrZero(),
			ModelNo:          assets[i].ModelNo.GetOrZero(),
			SerialNo:         assets[i].SerialNo.GetOrZero(),
			Manufacturer:     assets[i].Manufacturer.GetOrZero(),
			Notes:            assets[i].Notes.GetOrZero(),
			ImageURL:         assets[i].ImageURL.GetOrZero(),
			ThumbnailURL:     assets[i].ThumbnailURL.GetOrZero(),
			WarrantyUntil:    assets[i].WarrantyUntil.GetOrZero().Time,
			CustomAttrs:      assets[i].CustomAttrs.GetOrZero().JSON,
			CheckedOutTo:     assets[i].CheckedOutTo.GetOrZero(),
			Location:         assets[i].Location.GetOrZero(),
			PositionCode:     assets[i].PositionCode.GetOrZero(),
			PurchaseSupplier: assets[i].PurchaseSupplier.GetOrZero(),
			PurchaseOrderNo:  assets[i].PurchaseOrderNo.GetOrZero(),
			PurchaseDate:     assets[i].PurchaseDate.GetOrZero().Time,
			PurchaseAmount:   int(assets[i].PurchaseAmount.GetOrZero()),
			PurchaseCurrency: assets[i].PurchaseCurrency.GetOrZero(),
			CreatedBy:        assets[i].CreatedBy,
			CreatedAt:        assets[i].CreatedAt.Time,
			UpdatedAt:        assets[i].UpdatedAt.Time,
		})
	}

	return assetList, nil
}

func (ar *AssetRepo) CreateAsset(ctx context.Context, exec bob.Executor, asset *database.Asset) (*database.Asset, error) {
	model := &models.AssetSetter{
		ParentAssetID:    omitnullInt64(asset.ParentAssetID),
		Status:           omit.From(asset.Status),
		Tag:              omitnullStr(asset.Tag),
		Name:             omit.From(asset.Name),
		Category:         omit.From(asset.Category),
		Model:            omitnullStr(asset.Model),
		ModelNo:          omitnullStr(asset.ModelNo),
		SerialNo:         omitnullStr(asset.SerialNo),
		Manufacturer:     omitnullStr(asset.Manufacturer),
		Notes:            omitnullStr(asset.Notes),
		ImageURL:         omitnullStr(asset.ImageURL),
		ThumbnailURL:     omitnullStr(asset.ThumbnailURL),
		WarrantyUntil:    omitnullTime(asset.WarrantyUntil),
		CustomAttrs:      omitnullCustomAttrs(asset.CustomAttrs),
		CheckedOutTo:     omitnullInt64(asset.CheckedOutTo),
		Location:         omitnullStr(asset.Location),
		PositionCode:     omitnullStr(asset.PositionCode),
		PurchaseSupplier: omitnullStr(asset.PurchaseSupplier),
		PurchaseOrderNo:  omitnullStr(asset.PurchaseOrderNo),
		PurchaseDate:     omitnullTime(asset.PurchaseDate),
		PurchaseAmount:   omitnullInt64(int64(asset.PurchaseAmount)),
		PurchaseCurrency: omitnullStr(asset.PurchaseCurrency),
		CreatedBy:        omit.From(asset.CreatedBy),
	}

	inserted, err := models.Assets.Insert(ctx, exec, model)
	if err != nil {
		return nil, err
	}

	return &database.Asset{
		ID:               inserted.ID,
		ParentAssetID:    inserted.ParentAssetID.GetOrZero(),
		Status:           inserted.Status,
		Tag:              inserted.Tag.GetOrZero(),
		Name:             inserted.Name,
		Category:         inserted.Category,
		Model:            inserted.Model.GetOrZero(),
		ModelNo:          inserted.ModelNo.GetOrZero(),
		SerialNo:         inserted.SerialNo.GetOrZero(),
		Manufacturer:     inserted.Manufacturer.GetOrZero(),
		Notes:            inserted.Notes.GetOrZero(),
		ImageURL:         inserted.ImageURL.GetOrZero(),
		ThumbnailURL:     inserted.ThumbnailURL.GetOrZero(),
		WarrantyUntil:    inserted.WarrantyUntil.GetOrZero().Time,
		CustomAttrs:      inserted.CustomAttrs.GetOrZero().JSON,
		CheckedOutTo:     inserted.CheckedOutTo.GetOrZero(),
		Location:         inserted.Location.GetOrZero(),
		PositionCode:     inserted.PositionCode.GetOrZero(),
		PurchaseSupplier: inserted.PurchaseSupplier.GetOrZero(),
		PurchaseOrderNo:  inserted.PurchaseOrderNo.GetOrZero(),
		PurchaseDate:     inserted.PurchaseDate.GetOrZero().Time,
		PurchaseAmount:   int(inserted.PurchaseAmount.GetOrZero()),
		PurchaseCurrency: inserted.PurchaseCurrency.GetOrZero(),
		CreatedBy:        inserted.CreatedBy,
		CreatedAt:        inserted.CreatedAt.Time,
		UpdatedAt:        inserted.UpdatedAt.Time,
	}, nil
}

func (ar *AssetRepo) UpdateAsset(ctx context.Context, exec bob.Executor, asset *database.Asset) (*database.Asset, error) {
	model := &models.Asset{
		ID:               asset.ID,
		ParentAssetID:    nullInt64(asset.ParentAssetID),
		Status:           asset.Status,
		Tag:              nullStr(asset.Tag),
		Name:             asset.Name,
		Category:         asset.Category,
		Model:            nullStr(asset.Model),
		ModelNo:          nullStr(asset.ModelNo),
		SerialNo:         nullStr(asset.SerialNo),
		Manufacturer:     nullStr(asset.Manufacturer),
		Notes:            nullStr(asset.Notes),
		ImageURL:         nullStr(asset.ImageURL),
		ThumbnailURL:     nullStr(asset.ThumbnailURL),
		WarrantyUntil:    nullTime(asset.WarrantyUntil),
		CustomAttrs:      nullCustomAttrs(asset.CustomAttrs),
		CheckedOutTo:     nullInt64(asset.CheckedOutTo),
		Location:         nullStr(asset.Location),
		PositionCode:     nullStr(asset.PositionCode),
		PurchaseSupplier: nullStr(asset.PurchaseSupplier),
		PurchaseOrderNo:  nullStr(asset.PurchaseOrderNo),
		PurchaseDate:     nullTime(asset.PurchaseDate),
		PurchaseAmount:   nullInt64(int64(asset.PurchaseAmount)),
		PurchaseCurrency: nullStr(asset.PurchaseCurrency),
		CreatedBy:        asset.CreatedBy,
	}

	setter := &models.AssetSetter{
		ParentAssetID:    omitnullInt64(asset.ParentAssetID),
		Status:           omit.From(asset.Status),
		Tag:              omitnullStr(asset.Tag),
		Name:             omit.From(asset.Name),
		Category:         omit.From(asset.Category),
		Model:            omitnullStr(asset.Model),
		ModelNo:          omitnullStr(asset.ModelNo),
		SerialNo:         omitnullStr(asset.SerialNo),
		Manufacturer:     omitnullStr(asset.Manufacturer),
		Notes:            omitnullStr(asset.Notes),
		ImageURL:         omitnullStr(asset.ImageURL),
		ThumbnailURL:     omitnullStr(asset.ThumbnailURL),
		WarrantyUntil:    omitnullTime(asset.WarrantyUntil),
		CustomAttrs:      omitnullCustomAttrs(asset.CustomAttrs),
		CheckedOutTo:     omitnullInt64(asset.CheckedOutTo),
		Location:         omitnullStr(asset.Location),
		PositionCode:     omitnullStr(asset.PositionCode),
		PurchaseSupplier: omitnullStr(asset.PurchaseSupplier),
		PurchaseOrderNo:  omitnullStr(asset.PurchaseOrderNo),
		PurchaseDate:     omitnullTime(asset.PurchaseDate),
		PurchaseAmount:   omitnullInt64(int64(asset.PurchaseAmount)),
		PurchaseCurrency: omitnullStr(asset.PurchaseCurrency),
		UpdatedAt:        omit.From(types.NewSQLiteDatetime(time.Now())),
	}

	err := model.Update(ctx, exec, setter)
	if err != nil {
		return nil, err
	}

	return &database.Asset{
		ID:               model.ID,
		ParentAssetID:    model.ParentAssetID.GetOrZero(),
		Status:           model.Status,
		Tag:              model.Tag.GetOrZero(),
		Name:             model.Name,
		Category:         model.Category,
		Model:            model.Model.GetOrZero(),
		ModelNo:          model.ModelNo.GetOrZero(),
		SerialNo:         model.SerialNo.GetOrZero(),
		Manufacturer:     model.Manufacturer.GetOrZero(),
		Notes:            model.Notes.GetOrZero(),
		ImageURL:         model.ImageURL.GetOrZero(),
		ThumbnailURL:     model.ThumbnailURL.GetOrZero(),
		WarrantyUntil:    model.WarrantyUntil.GetOrZero().Time,
		CustomAttrs:      model.CustomAttrs.GetOrZero().JSON,
		CheckedOutTo:     model.CheckedOutTo.GetOrZero(),
		Location:         model.Location.GetOrZero(),
		PositionCode:     model.PositionCode.GetOrZero(),
		PurchaseSupplier: model.PurchaseSupplier.GetOrZero(),
		PurchaseOrderNo:  model.PurchaseOrderNo.GetOrZero(),
		PurchaseDate:     model.PurchaseDate.GetOrZero().Time,
		PurchaseAmount:   int(model.PurchaseAmount.GetOrZero()),
		PurchaseCurrency: model.PurchaseCurrency.GetOrZero(),
		CreatedBy:        model.CreatedBy,
		CreatedAt:        model.CreatedAt.Time,
		UpdatedAt:        model.UpdatedAt.Time,
	}, nil
}

func (ar *AssetRepo) DeleteAsset(ctx context.Context, exec bob.Executor, id int64) error {
	_, err := models.Assets.DeleteQ(ctx, exec, models.DeleteWhere.Assets.ID.EQ(id)).Exec()
	return err
}

func (ar *AssetRepo) ListCategories(ctx context.Context, exec bob.Executor) ([]string, error) {
	categories, err := models.Categories.Query(ctx, exec).All()
	if err != nil {
		return nil, err
	}

	cats := make([]string, 0, len(categories))

	for _, c := range categories {
		if c.Name.IsSet() {
			cats = append(cats, c.Name.GetOrZero())
		}
	}

	return cats, nil
}

func omitnullStr(str string) omitnull.Val[string] {
	v := omitnull.From(str)
	if str == "" {
		v.Null()
	}

	return v
}

func omitnullInt64(i int64) omitnull.Val[int64] {
	v := omitnull.From(i)
	if i == 0 {
		v.Null()
	}

	return v
}

func omitnullTime(t time.Time) omitnull.Val[types.SQLiteDatetime] {
	st := types.NewSQLiteDatetime(t)
	v := omitnull.From(st)
	if !st.Valid {
		v.Null()
	}

	return v
}

func omitnullCustomAttrs(a map[string]any) omitnull.Val[types.SQLiteJSON[map[string]any]] {
	j := types.NewSQLiteJSON(a)
	v := omitnull.From(j)
	if len(a) == 0 {
		v.Null()
	}

	return v
}

func nullStr(str string) null.Val[string] {
	v := null.From(str)
	if str == "" {
		v.Null()
	}

	return v
}

func nullInt64(i int64) null.Val[int64] {
	v := null.From(i)
	if i == 0 {
		v.Null()
	}

	return v
}

func nullTime(t time.Time) null.Val[types.SQLiteDatetime] {
	st := types.NewSQLiteDatetime(t)
	v := null.From(st)
	if !st.Valid {
		v.Null()
	}

	return v
}

func nullCustomAttrs(a map[string]any) null.Val[types.SQLiteJSON[map[string]any]] {
	j := types.NewSQLiteJSON(a)
	v := null.From(j)
	if len(a) == 0 {
		v.Null()
	}

	return v
}
