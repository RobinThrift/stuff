package assets

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/aarondl/opt/null"
	"github.com/aarondl/opt/omit"
	"github.com/aarondl/opt/omitnull"
	"github.com/kodeshack/stuff/storage/database/sqlite/models"
	"github.com/kodeshack/stuff/storage/database/sqlite/types"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/sqlite"
	"github.com/stephenafamo/bob/dialect/sqlite/dialect"
	"github.com/stephenafamo/bob/dialect/sqlite/sm"
	bmods "github.com/stephenafamo/bob/mods"
)

type RepoSQLite struct{}

func (ar *RepoSQLite) Get(ctx context.Context, exec bob.Executor, id int64) (*Asset, error) {
	asset, err := models.FindAsset(ctx, exec, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: %d", ErrAssetNotFound, id)
		}
		return nil, fmt.Errorf("error getting assets: %w", err)
	}

	return mapDBModelToAsset(asset), nil
}

func (ar *RepoSQLite) List(ctx context.Context, exec bob.Executor, query ListAssetsQuery) (*AssetListPage, error) {
	limit := query.PageSize

	if limit == 0 {
		limit = 50
	}

	if limit > 100 {
		limit = 100
	}

	offset := limit * query.Page

	mods := []bob.Mod[*dialect.SelectQuery]{
		sm.Limit(limit),
		sm.Offset(offset),
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

	var err error
	var assets models.AssetSlice
	var count int64

	if query.Search != nil && query.Search.Raw != "" {
		assets, count, err = ar.listAssetsFromFTSTable(ctx, exec, query.Search, mods)
		if err != nil {
			return nil, err
		}
	} else {
		mods = append(mods, models.SelectWhere.Assets.ParentAssetID.IsNull())

		count, err = models.Assets.Query(ctx, exec, mods...).Count()
		if err != nil {
			return nil, fmt.Errorf("error counting assets: %w", err)
		}

		assets, err = models.Assets.Query(ctx, exec, mods...).All()
		if err != nil {
			return nil, fmt.Errorf("error getting assets: %w", err)
		}
	}

	page := &AssetListPage{
		Assets:   make([]*Asset, 0, len(assets)),
		Total:    int(count),
		Page:     query.Page,
		PageSize: query.PageSize,
		NumPages: int(count) / query.PageSize,
	}

	for i := range assets {
		page.Assets = append(page.Assets, mapDBModelToAsset(assets[i]))
	}

	return page, nil
}

var removeSpecialChars = regexp.MustCompile(`[^\w* ]`)

func (ar *RepoSQLite) listAssetsFromFTSTable(ctx context.Context, exec bob.Executor, search *ListAssetsQuerySearch, mods []bob.Mod[*dialect.SelectQuery]) (models.AssetSlice, int64, error) {
	if len(search.Fields) == 0 {
		value := removeSpecialChars.ReplaceAllString(search.Raw, "")
		mods = append(mods,
			sm.Where(sqlite.Quote(models.TableNames.AssetsFTS).EQ(sqlite.Quote(value))),
		)
	}

	for field, value := range search.Fields {
		column, ok := isAssetsFTSColumn(field)
		if !ok {
			continue
		}

		value = removeSpecialChars.ReplaceAllString(value, "")

		mods = append(mods, sm.Where(sqlite.Raw(column+" MATCH ?", value)))
	}

	count, err := models.AssetsFTS.Query(ctx, exec, mods...).Count()
	if err != nil {
		return nil, 0, fmt.Errorf("error counting searched assets: %w", err)
	}

	entries, err := models.AssetsFTS.Query(ctx, exec, mods...).All()
	if err != nil {
		return nil, 0, fmt.Errorf("error searching assets: %w", err)
	}

	ids := make([]int64, 0, len(entries))

	for _, entry := range entries {
		id, err := strconv.ParseInt(entry.ID.GetOrZero(), 10, 64)
		if err != nil {
			continue
		}

		ids = append(ids, id)
	}

	assets, err := models.Assets.Query(ctx, exec, models.SelectWhere.Assets.ID.In(ids...), models.SelectWhere.Assets.ParentAssetID.IsNull()).All()
	if err != nil {
		return nil, 0, fmt.Errorf("error getting assets: %w", err)
	}

	return assets, count, nil
}

func (ar *RepoSQLite) Create(ctx context.Context, exec bob.Executor, asset *Asset) (*Asset, error) {
	model := &models.AssetSetter{
		ParentAssetID:    omitnullInt64(asset.ParentAssetID),
		Status:           omit.From(string(asset.Status)),
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
		PurchaseSupplier: omitnullStr(asset.PurchaseInfo.Supplier),
		PurchaseOrderNo:  omitnullStr(asset.PurchaseInfo.OrderNo),
		PurchaseDate:     omitnullTime(asset.PurchaseInfo.Date),
		PurchaseAmount:   omitnullInt64(int64(asset.PurchaseInfo.Amount)),
		PurchaseCurrency: omitnullStr(asset.PurchaseInfo.Currency),
		CreatedBy:        omit.From(asset.MetaInfo.CreatedBy),
	}

	inserted, err := models.Assets.Insert(ctx, exec, model)
	if err != nil {
		return nil, err
	}

	return mapDBModelToAsset(inserted), nil
}

func (ar *RepoSQLite) Update(ctx context.Context, exec bob.Executor, asset *Asset) (*Asset, error) {
	model := &models.Asset{
		ID:               asset.ID,
		ParentAssetID:    nullInt64(asset.ParentAssetID),
		Status:           string(asset.Status),
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
		PurchaseSupplier: nullStr(asset.PurchaseInfo.Supplier),
		PurchaseOrderNo:  nullStr(asset.PurchaseInfo.OrderNo),
		PurchaseDate:     nullTime(asset.PurchaseInfo.Date),
		PurchaseAmount:   nullInt64(int64(asset.PurchaseInfo.Amount)),
		PurchaseCurrency: nullStr(asset.PurchaseInfo.Currency),
		CreatedBy:        asset.MetaInfo.CreatedBy,
	}

	setter := &models.AssetSetter{
		ParentAssetID:    omitnullInt64(asset.ParentAssetID),
		Status:           omit.From(string(asset.Status)),
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
		PurchaseSupplier: omitnullStr(asset.PurchaseInfo.Supplier),
		PurchaseOrderNo:  omitnullStr(asset.PurchaseInfo.OrderNo),
		PurchaseDate:     omitnullTime(asset.PurchaseInfo.Date),
		PurchaseAmount:   omitnullInt64(int64(asset.PurchaseInfo.Amount)),
		PurchaseCurrency: omitnullStr(asset.PurchaseInfo.Currency),
		UpdatedAt:        omit.From(types.NewSQLiteDatetime(time.Now())),
	}

	err := model.Update(ctx, exec, setter)
	if err != nil {
		return nil, err
	}

	return mapDBModelToAsset(model), nil
}

func (ar *RepoSQLite) Delete(ctx context.Context, exec bob.Executor, id int64) error {
	_, err := models.Assets.DeleteQ(ctx, exec, models.DeleteWhere.Assets.ID.EQ(id)).Exec()
	return err
}

func (ar *RepoSQLite) ListCategories(ctx context.Context, exec bob.Executor, query ListCategoriesQuery) ([]Category, error) {
	limit := query.PageSize

	if limit == 0 {
		limit = 50
	}

	if limit > 100 {
		limit = 100
	}

	offset := limit * query.Page

	mods := []bob.Mod[*dialect.SelectQuery]{
		sm.Limit(limit),
		sm.Offset(offset),
	}

	if query.Search != "" {
		mods = append(mods, models.SelectWhere.Categories.Name.Like("%"+query.Search+"%"))
	}

	categories, err := models.Categories.Query(ctx, exec, mods...).All()
	if err != nil {
		return nil, err
	}

	cats := make([]Category, 0, len(categories))

	for _, c := range categories {
		if c.Name.IsSet() {
			cats = append(cats, Category{Name: c.Name.GetOrZero()})
		}
	}

	return cats, nil
}

func mapDBModelToAsset(model *models.Asset) *Asset {
	return &Asset{
		ID:            model.ID,
		ParentAssetID: model.ParentAssetID.GetOrZero(),
		Status:        Status(model.Status),
		Tag:           model.Tag.GetOrZero(),
		Name:          model.Name,
		Category:      model.Category,
		Model:         model.Model.GetOrZero(),
		ModelNo:       model.ModelNo.GetOrZero(),
		SerialNo:      model.SerialNo.GetOrZero(),
		Manufacturer:  model.Manufacturer.GetOrZero(),
		Notes:         model.Notes.GetOrZero(),
		ImageURL:      model.ImageURL.GetOrZero(),
		ThumbnailURL:  model.ThumbnailURL.GetOrZero(),
		WarrantyUntil: model.WarrantyUntil.GetOrZero().Time,
		CustomAttrs:   model.CustomAttrs.GetOrZero().JSON,
		CheckedOutTo:  model.CheckedOutTo.GetOrZero(),
		Location:      model.Location.GetOrZero(),
		PositionCode:  model.PositionCode.GetOrZero(),
		PurchaseInfo: PurchaseInfo{
			Supplier: model.PurchaseSupplier.GetOrZero(),
			OrderNo:  model.PurchaseOrderNo.GetOrZero(),
			Date:     model.PurchaseDate.GetOrZero().Time,
			Amount:   MonetaryAmount(model.PurchaseAmount.GetOrZero()),
			Currency: model.PurchaseCurrency.GetOrZero(),
		},

		MetaInfo: MetaInfo{
			CreatedBy: model.CreatedBy,
			CreatedAt: model.CreatedAt.Time,
			UpdatedAt: model.UpdatedAt.Time,
		},
	}
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

func isAssetsFTSColumn(s string) (string, bool) {
	switch s {
	case models.ColumnNames.AssetsFTS.Tag:
		return models.ColumnNames.AssetsFTS.Tag, true
	case models.ColumnNames.AssetsFTS.Name:
		return models.ColumnNames.AssetsFTS.Name, true
	case models.ColumnNames.AssetsFTS.Category:
		return models.ColumnNames.AssetsFTS.Category, true
	case models.ColumnNames.AssetsFTS.Model:
		return models.ColumnNames.AssetsFTS.Model, true
	case models.ColumnNames.AssetsFTS.ModelNo, "modelno":
		return models.ColumnNames.AssetsFTS.ModelNo, true
	case models.ColumnNames.AssetsFTS.SerialNo, "serial", "serialno":
		return models.ColumnNames.AssetsFTS.SerialNo, true
	case models.ColumnNames.AssetsFTS.Manufacturer:
		return models.ColumnNames.AssetsFTS.Manufacturer, true
	case models.ColumnNames.AssetsFTS.Notes:
		return models.ColumnNames.AssetsFTS.Notes, true
	}
	return "", false
}
