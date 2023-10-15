package assets

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"time"

	"github.com/RobinThrift/stuff/storage/database"
	"github.com/RobinThrift/stuff/storage/database/sqlite/models"
	"github.com/RobinThrift/stuff/storage/database/sqlite/types"
	"github.com/aarondl/opt/null"
	"github.com/aarondl/opt/omit"
	"github.com/aarondl/opt/omitnull"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/sqlite"
	"github.com/stephenafamo/bob/dialect/sqlite/dialect"
	"github.com/stephenafamo/bob/dialect/sqlite/sm"
	bmods "github.com/stephenafamo/bob/mods"
)

type RepoSQLite struct{}

func (ar *RepoSQLite) Get(ctx context.Context, exec bob.Executor, idOrTag string) (*Asset, error) {
	mods := []bob.Mod[*dialect.SelectQuery]{
		models.ThenLoadAssetAssetParts(),
		models.PreloadAssetParentAsset(),
	}

	if id, err := strconv.ParseInt(idOrTag, 10, 64); err == nil {
		mods = append(mods, sqlite.WhereOr(models.SelectWhere.Assets.ID.EQ(id), models.SelectWhere.Assets.Tag.EQ(idOrTag)))
	} else {
		mods = append(mods, models.SelectWhere.Assets.Tag.EQ(idOrTag))
	}

	asset, err := models.Assets.Query(ctx, exec, mods...).One()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: %s", ErrAssetNotFound, idOrTag)
		}
		return nil, fmt.Errorf("error getting assets: %w", err)
	}

	children, err := models.Assets.Query(ctx, exec,
		sm.Columns(models.Assets.Columns().Only("id", "name", "tag")),
		models.SelectWhere.Assets.ParentAssetID.EQ(asset.ID),
	).All()
	if err != nil {
		return nil, fmt.Errorf("error getting asset children: %w", err)
	}

	return mapDBModelToAsset(asset, children), nil
}

func (ar *RepoSQLite) List(ctx context.Context, exec bob.Executor, query ListAssetsQuery) (*AssetListPage, error) {
	var err error
	var assets models.AssetSlice
	var count int64
	if query.Search != nil && query.Search.Raw != "" {
		assets, count, err = ar.listUsingFTS(ctx, exec, query)
	} else {
		assets, count, err = ar.list(ctx, exec, query)
	}

	if err != nil {
		return nil, err
	}

	numPages := 1
	if query.PageSize > 0 {
		numPages = int(math.Ceil(float64(count) / float64(query.PageSize)))
	}

	page := &AssetListPage{
		Assets:   make([]*Asset, 0, len(assets)),
		Total:    int(count),
		Page:     query.Page,
		PageSize: query.PageSize,
		NumPages: numPages,
	}

	for i := range assets {
		page.Assets = append(page.Assets, mapDBModelToAsset(assets[i], nil))
	}

	return page, nil
}

func (ar *RepoSQLite) list(ctx context.Context, exec bob.Executor, query ListAssetsQuery) (models.AssetSlice, int64, error) {
	limit := query.PageSize
	offset := limit * query.Page

	mods := make([]bob.Mod[*dialect.SelectQuery], 0, 3)

	if query.AssetType != "" {
		mods = append(mods, models.SelectWhere.Assets.Type.EQ(string(query.AssetType)))
	}

	if query.OrderBy != "" {
		if query.OrderDir == "" {
			query.OrderDir = database.OrderASC
		}

		mods = append(mods, bmods.OrderBy[*dialect.SelectQuery]{
			Expression: query.OrderBy,
			Direction:  query.OrderDir,
		})
	}

	count, err := models.Assets.Query(ctx, exec, mods...).Count()
	if err != nil {
		return nil, 0, fmt.Errorf("error counting assets: %w", err)
	}

	if limit > 0 {
		mods = append(mods, sm.Limit(limit))
	}

	if offset > 0 {
		mods = append(mods, sm.Offset(offset))
	}

	assets, err := models.Assets.Query(ctx, exec, mods...).All()
	if err != nil {
		return nil, 0, fmt.Errorf("error getting assets: %w", err)
	}

	return assets, count, nil
}

func (ar *RepoSQLite) listUsingFTS(ctx context.Context, exec bob.Executor, query ListAssetsQuery) (models.AssetSlice, int64, error) {
	limit := query.PageSize
	offset := limit * query.Page

	mods := make([]bob.Mod[*dialect.SelectQuery], 0, 2)

	if limit > 0 {
		mods = append(mods, sm.Limit(limit))
	}

	if offset > 0 {
		mods = append(mods, sm.Offset(offset))
	}

	assets, count, err := ar.listAssetsFromFTSTable(ctx, exec, query, mods)
	if err != nil {
		return nil, 0, err
	}

	return assets, count, nil
}

func (ar *RepoSQLite) ListForExport(ctx context.Context, exec bob.Executor, query ListAssetsQuery) (*AssetListPage, error) {
	limit := query.PageSize
	offset := limit * query.Page

	mods := make([]bob.Mod[*dialect.SelectQuery], 0, 4)

	if limit > 0 {
		mods = append(mods, sm.Limit(limit))
	}

	if offset > 0 {
		mods = append(mods, sm.Offset(offset))
	}

	if query.OrderBy != "" {
		if query.OrderDir == "" {
			query.OrderDir = database.OrderASC
		}

		mods = append(mods, bmods.OrderBy[*dialect.SelectQuery]{
			Expression: query.OrderBy,
			Direction:  query.OrderDir,
		})
	}

	if len(query.IDs) != 0 {
		mods = append(mods, models.SelectWhere.Assets.ID.In(query.IDs...))
	}

	count, err := models.Assets.Query(ctx, exec, mods...).Count()
	if err != nil {
		return nil, fmt.Errorf("error counting assets: %w", err)
	}

	mods = append(mods, models.ThenLoadAssetAssetParts())

	assets, err := models.Assets.Query(ctx, exec, mods...).All()
	if err != nil {
		return nil, fmt.Errorf("error getting assets: %w", err)
	}

	numPages := 0
	if limit > 0 {
		numPages = int(count) / limit
	}

	page := &AssetListPage{
		Assets:   make([]*Asset, 0, len(assets)),
		Total:    int(count),
		Page:     query.Page,
		PageSize: query.PageSize,
		NumPages: numPages,
	}

	for i := range assets {
		page.Assets = append(page.Assets, mapDBModelToAsset(assets[i], nil))
	}

	return page, nil
}

var removeSpecialChars = regexp.MustCompile(`[^\w* ]`)

func (ar *RepoSQLite) listAssetsFromFTSTable(ctx context.Context, exec bob.Executor, query ListAssetsQuery, mods []bob.Mod[*dialect.SelectQuery]) (models.AssetSlice, int64, error) {
	if len(query.Search.Fields) == 0 {
		value := removeSpecialChars.ReplaceAllString(query.Search.Raw, "")
		mods = append(mods,
			sm.Where(sqlite.Quote(models.TableNames.AssetsFTS).EQ(sqlite.Quote(value))),
		)
	}

	for field, value := range query.Search.Fields {
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

	assetQueryMods := make([]bob.Mod[*dialect.SelectQuery], 0, 3)

	if query.AssetType != "" {
		assetQueryMods = append(assetQueryMods, models.SelectWhere.Assets.Tag.EQ(string(query.AssetType)))
	}

	assetQueryMods = append(assetQueryMods,
		models.SelectWhere.Assets.ID.In(ids...),
		models.SelectWhere.Assets.ParentAssetID.IsNull(),
	)

	assets, err := models.Assets.Query(ctx, exec, assetQueryMods...).All()
	if err != nil {
		return nil, 0, fmt.Errorf("error getting assets: %w", err)
	}

	return assets, count, nil
}

func (ar *RepoSQLite) Create(ctx context.Context, exec bob.Executor, asset *Asset) (*Asset, error) {
	model := &models.AssetSetter{
		Type:             omit.From(string(asset.Type)),
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
		Quantity:         omit.From(asset.Quantity),
		QuantityUnit:     omitnullStr(asset.QuantityUnit),
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

	return mapDBModelToAsset(inserted, nil), nil
}

func (ar *RepoSQLite) Update(ctx context.Context, exec bob.Executor, asset *Asset) (*Asset, error) {
	model := &models.Asset{
		ID:                asset.ID,
		Type:              string(asset.Type),
		ParentAssetID:     nullInt64(asset.ParentAssetID),
		Status:            string(asset.Status),
		Tag:               nullStr(asset.Tag),
		Name:              asset.Name,
		Category:          asset.Category,
		Model:             nullStr(asset.Model),
		ModelNo:           nullStr(asset.ModelNo),
		SerialNo:          nullStr(asset.SerialNo),
		Manufacturer:      nullStr(asset.Manufacturer),
		Notes:             nullStr(asset.Notes),
		ImageURL:          nullStr(asset.ImageURL),
		ThumbnailURL:      nullStr(asset.ThumbnailURL),
		WarrantyUntil:     nullTime(asset.WarrantyUntil),
		CustomAttrs:       nullCustomAttrs(asset.CustomAttrs),
		Quantity:          asset.Quantity,
		QuantityUnit:      nullStr(asset.QuantityUnit),
		CheckedOutTo:      nullInt64(asset.CheckedOutTo),
		Location:          nullStr(asset.Location),
		PositionCode:      nullStr(asset.PositionCode),
		PurchaseSupplier:  nullStr(asset.PurchaseInfo.Supplier),
		PurchaseOrderNo:   nullStr(asset.PurchaseInfo.OrderNo),
		PurchaseDate:      nullTime(asset.PurchaseInfo.Date),
		PurchaseAmount:    nullInt64(int64(asset.PurchaseInfo.Amount)),
		PurchaseCurrency:  nullStr(asset.PurchaseInfo.Currency),
		PartsTotalCounter: int64(asset.PartsTotalCounter),
		CreatedBy:         asset.MetaInfo.CreatedBy,
	}

	setter := &models.AssetSetter{
		Type:              omit.From(string(asset.Type)),
		ParentAssetID:     omitnullInt64(asset.ParentAssetID),
		Status:            omit.From(string(asset.Status)),
		Tag:               omitnullStr(asset.Tag),
		Name:              omit.From(asset.Name),
		Category:          omit.From(asset.Category),
		Model:             omitnullStr(asset.Model),
		ModelNo:           omitnullStr(asset.ModelNo),
		SerialNo:          omitnullStr(asset.SerialNo),
		Manufacturer:      omitnullStr(asset.Manufacturer),
		Notes:             omitnullStr(asset.Notes),
		ImageURL:          omitnullStr(asset.ImageURL),
		ThumbnailURL:      omitnullStr(asset.ThumbnailURL),
		WarrantyUntil:     omitnullTime(asset.WarrantyUntil),
		CustomAttrs:       omitnullCustomAttrs(asset.CustomAttrs),
		Quantity:          omit.From(asset.Quantity),
		QuantityUnit:      omitnullStr(asset.QuantityUnit),
		CheckedOutTo:      omitnullInt64(asset.CheckedOutTo),
		Location:          omitnullStr(asset.Location),
		PositionCode:      omitnullStr(asset.PositionCode),
		PurchaseSupplier:  omitnullStr(asset.PurchaseInfo.Supplier),
		PurchaseOrderNo:   omitnullStr(asset.PurchaseInfo.OrderNo),
		PurchaseDate:      omitnullTime(asset.PurchaseInfo.Date),
		PurchaseAmount:    omitnullInt64(int64(asset.PurchaseInfo.Amount)),
		PurchaseCurrency:  omitnullStr(asset.PurchaseInfo.Currency),
		PartsTotalCounter: omit.From(int64(asset.PartsTotalCounter)),
		UpdatedAt:         omit.From(types.NewSQLiteDatetime(time.Now())),
	}

	err := model.Update(ctx, exec, setter)
	if err != nil {
		return nil, err
	}

	return mapDBModelToAsset(model, nil), nil
}

func (ar *RepoSQLite) Delete(ctx context.Context, exec bob.Executor, id int64) error {
	_, err := models.Assets.DeleteQ(ctx, exec, models.DeleteWhere.Assets.ID.EQ(id)).Exec()
	if err != nil {
		return err
	}

	return ar.DeleteParts(ctx, exec, id)
}

func (ar *RepoSQLite) CreateParts(ctx context.Context, exec bob.Executor, parts []*Part) error {
	setters := make([]*models.AssetPartSetter, 0, len(parts))
	for _, part := range parts {
		setters = append(setters, &models.AssetPartSetter{
			AssetID:      omit.From(part.AssetID),
			Tag:          omit.From(part.Tag),
			Name:         omit.From(part.Name),
			Notes:        omitnullStr(part.Notes),
			Location:     omitnullStr(part.Location),
			PositionCode: omitnullStr(part.PositionCode),
			CreatedBy:    omit.From(part.CreatedBy),
		})
	}

	_, err := models.AssetParts.InsertMany(ctx, exec, setters...)
	if err != nil {
		return err
	}

	return nil
}

func (ar *RepoSQLite) DeleteParts(ctx context.Context, exec bob.Executor, assetID int64) error {
	_, err := models.AssetParts.DeleteQ(ctx, exec, models.DeleteWhere.AssetParts.AssetID.EQ(assetID)).Exec()
	return err
}

func (ar *RepoSQLite) ListCategories(ctx context.Context, exec bob.Executor, query ListCategoriesQuery) ([]Category, error) {
	limit := query.PageSize
	if limit == 0 {
		limit = 25
	}
	offset := limit * query.Page

	mods := []bob.Mod[*dialect.SelectQuery]{
		sm.Limit(limit),
		sm.Offset(offset),
		sm.Distinct(),
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

func mapDBModelToAsset(model *models.Asset, children []*models.Asset) *Asset {
	asset := &Asset{
		ID:            model.ID,
		Type:          AssetType(model.Type),
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
		Quantity:      model.Quantity,
		QuantityUnit:  model.QuantityUnit.GetOrZero(),
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

		PartsTotalCounter: int(model.PartsTotalCounter),

		MetaInfo: MetaInfo{
			CreatedBy: model.CreatedBy,
			CreatedAt: model.CreatedAt.Time,
			UpdatedAt: model.UpdatedAt.Time,
		},
	}

	asset.Parts = make([]*Part, 0, len(model.R.AssetParts))
	for _, p := range model.R.AssetParts {
		asset.Parts = append(asset.Parts, &Part{
			ID:           p.ID,
			AssetID:      p.AssetID,
			Tag:          p.Tag,
			Name:         p.Name,
			Notes:        p.Notes.GetOrZero(),
			Location:     p.Location.GetOrZero(),
			PositionCode: p.PositionCode.GetOrZero(),
			CreatedBy:    p.CreatedBy,
			CreatedAt:    p.CreatedAt.Time,
			UpdatedAt:    p.UpdatedAt.Time,
		})
	}

	if len(children) != 0 {
		asset.Children = make([]*Asset, 0, len(children))
		for _, child := range children {
			asset.Children = append(asset.Children, mapDBModelToAsset(child, nil))
		}
	}

	if model.R.ParentAsset != nil {
		asset.Parent = mapDBModelToAsset(model.R.ParentAsset, nil)
	}

	return asset
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
