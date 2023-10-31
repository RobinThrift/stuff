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
	"github.com/stephenafamo/bob/dialect/sqlite/im"
	"github.com/stephenafamo/bob/dialect/sqlite/sm"
	bmods "github.com/stephenafamo/bob/mods"
)

type RepoSQLite struct{}

func (ar *RepoSQLite) Get(ctx context.Context, exec bob.Executor, idOrTag string) (*Asset, error) {
	mods := []bob.Mod[*dialect.SelectQuery]{
		models.ThenLoadAssetAssetParts(),
		models.ThenLoadAssetAssetPurchases(),
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

func (ar *RepoSQLite) GetWithFiles(ctx context.Context, exec bob.Executor, idOrTag string) (*Asset, error) {
	mods := []bob.Mod[*dialect.SelectQuery]{
		models.ThenLoadAssetAssetFiles(),
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

	return mapDBModelToAsset(asset, nil), nil
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
		Type:          omit.From(string(asset.Type)),
		ParentAssetID: omitnullInt64(asset.ParentAssetID),
		Status:        omit.From(string(asset.Status)),
		Tag:           omitnullStr(asset.Tag),
		Name:          omit.From(asset.Name),
		Category:      omit.From(asset.Category),
		Model:         omitnullStr(asset.Model),
		ModelNo:       omitnullStr(asset.ModelNo),
		SerialNo:      omitnullStr(asset.SerialNo),
		Manufacturer:  omitnullStr(asset.Manufacturer),
		Notes:         omitnullStr(asset.Notes),
		ImageURL:      omitnullStr(asset.ImageURL),
		ThumbnailURL:  omitnullStr(asset.ThumbnailURL),
		WarrantyUntil: omitnullTime(asset.WarrantyUntil),
		CustomAttrs:   omitnullCustomAttrs(asset.CustomAttrs),
		Quantity:      omit.From(asset.Quantity),
		QuantityUnit:  omitnullStr(asset.QuantityUnit),
		CheckedOutTo:  omitnullInt64(asset.CheckedOutTo),
		Location:      omitnullStr(asset.Location),
		PositionCode:  omitnullStr(asset.PositionCode),
		CreatedBy:     omit.From(asset.MetaInfo.CreatedBy),
	}

	inserted, err := models.Assets.Insert(ctx, exec, model)
	if err != nil {
		return nil, err
	}

	purchases := make([]*models.AssetPurchaseSetter, 0, len(asset.Purchases))
	for _, p := range asset.Purchases {
		purchases = append(purchases, &models.AssetPurchaseSetter{
			AssetID:   omit.From(inserted.ID),
			Supplier:  omitnullStr(p.Supplier),
			OrderNo:   omitnullStr(p.OrderNo),
			OrderDate: omitnullTime(p.Date),
			Amount:    omitnullInt64(int64(p.Amount)),
			Currency:  omitnullStr(p.Currency),
			CreatedBy: omit.From(asset.MetaInfo.CreatedBy),
		})
	}

	insertedPurchases, err := models.AssetPurchases.InsertMany(ctx, exec, purchases...)
	if err != nil {
		return nil, err
	}

	inserted.R.AssetPurchases = insertedPurchases

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
		PartsTotalCounter: omit.From(int64(asset.PartsTotalCounter)),
		UpdatedAt:         omit.From(types.NewSQLiteDatetime(time.Now())),
	}

	err := model.Update(ctx, exec, setter)
	if err != nil {
		return nil, err
	}

	_, err = models.AssetPurchases.DeleteQ(ctx, exec, models.DeleteWhere.AssetPurchases.AssetID.EQ(model.ID)).Exec()
	if err != nil {
		return nil, err
	}

	purchases := make([]bob.Mod[*dialect.InsertQuery], 0, len(asset.Purchases)+1)
	purchases = append(purchases,
		im.IntoAs(models.TableNames.AssetPurchases, models.TableNames.AssetPurchases,
			models.ColumnNames.AssetPurchases.AssetID,
			models.ColumnNames.AssetPurchases.Supplier,
			models.ColumnNames.AssetPurchases.OrderNo,
			models.ColumnNames.AssetPurchases.OrderDate,
			models.ColumnNames.AssetPurchases.Amount,
			models.ColumnNames.AssetPurchases.Currency,
			models.ColumnNames.AssetPurchases.CreatedBy,
			models.ColumnNames.AssetPurchases.UpdatedAt,
		),
	)
	for _, p := range asset.Purchases {
		purchases = append(purchases, models.AssetPurchaseSetter{
			AssetID:   omit.From(model.ID),
			Supplier:  omitnullStr(p.Supplier),
			OrderNo:   omitnullStr(p.OrderNo),
			OrderDate: omitnullTime(p.Date),
			Amount:    omitnullInt64(int64(p.Amount)),
			Currency:  omitnullStr(p.Currency),
			CreatedBy: omit.From(asset.MetaInfo.CreatedBy),
			UpdatedAt: omit.From(types.NewSQLiteDatetime(time.Now())),
		}.Insert())
	}

	insertedPurchases, err := models.AssetPurchases.InsertQ(ctx, exec, purchases...).All()
	if err != nil {
		return nil, err
	}

	model.R.AssetPurchases = insertedPurchases

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

func (ar *RepoSQLite) CreateFiles(ctx context.Context, exec bob.Executor, files []*File) error {
	setter := make([]*models.AssetFileSetter, 0, len(files))
	for _, f := range files {
		setter = append(setter, &models.AssetFileSetter{
			AssetID:    omit.From(f.AssetID),
			Name:       omit.From(f.Name),
			Filetype:   omit.From(f.Filetype),
			Sha256:     omit.From(f.Sha256),
			SizeBytes:  omit.From(f.SizeBytes),
			CreatedBy:  omit.From(f.SizeBytes),
			FullPath:   omit.From(f.FullPath),
			PublicPath: omit.From(f.PublicPath),
		})
	}

	_, err := models.AssetFiles.InsertMany(ctx, exec, setter...)
	return err
}

func (ar *RepoSQLite) GetFile(ctx context.Context, exec bob.Executor, id int64) (*File, error) {
	file, err := models.AssetFiles.Query(ctx, exec, models.SelectWhere.AssetFiles.ID.EQ(id)).One()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: %d", ErrFileNotFound, id)
		}
		return nil, fmt.Errorf("error getting asset file: %w", err)
	}

	return &File{
		ID:         file.ID,
		AssetID:    file.AssetID,
		PublicPath: file.PublicPath,
		FullPath:   file.FullPath,
		Name:       file.Name,
		Filetype:   file.Filetype,
		Sha256:     file.Sha256,
		SizeBytes:  file.SizeBytes,
		CreatedBy:  file.CreatedBy,
		CreatedAt:  file.CreatedAt.Time,
		UpdatedAt:  file.UpdatedAt.Time,
	}, nil
}

func (ar *RepoSQLite) FileExists(ctx context.Context, exec bob.Executor, hash []byte) (bool, error) {
	file, err := models.AssetFiles.Query(ctx, exec, models.SelectWhere.AssetFiles.Sha256.EQ(hash)).One()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("error getting asset file: %w", err)
	}

	return file != nil, nil
}

func (ar *RepoSQLite) DeleteFile(ctx context.Context, exec bob.Executor, id int64) error {
	_, err := models.AssetFiles.DeleteQ(ctx, exec, models.DeleteWhere.AssetFiles.ID.EQ(id)).Exec()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: %d", ErrFileNotFound, id)
		}
		return fmt.Errorf("error getting asset file: %w", err)
	}

	return nil
}

func mapDBModelToAsset(model *models.Asset, children []*models.Asset) *Asset {
	purchases := make([]*Purchase, 0, len(model.R.AssetPurchases))
	for _, p := range model.R.AssetPurchases {
		purchases = append(purchases, &Purchase{
			Supplier: p.Supplier.GetOrZero(),
			OrderNo:  p.OrderNo.GetOrZero(),
			Date:     p.OrderDate.GetOrZero().Time,
			Amount:   MonetaryAmount(p.Amount.GetOrZero()),
			Currency: p.Currency.GetOrZero(),
		})
	}

	files := make([]*File, 0, len(model.R.AssetFiles))
	for _, f := range model.R.AssetFiles {
		files = append(files, &File{
			ID:         f.ID,
			AssetID:    f.AssetID,
			PublicPath: f.PublicPath,
			FullPath:   f.FullPath,
			Name:       f.Name,
			Filetype:   f.Filetype,
			Sha256:     f.Sha256,
			SizeBytes:  f.SizeBytes,
			CreatedBy:  f.CreatedBy,
			CreatedAt:  f.CreatedAt.Time,
			UpdatedAt:  f.UpdatedAt.Time,
		})
	}

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
		CustomAttrs:   unmarshalCustomAttrs(model.CustomAttrs.GetOrZero().JSON),
		Quantity:      model.Quantity,
		QuantityUnit:  model.QuantityUnit.GetOrZero(),
		CheckedOutTo:  model.CheckedOutTo.GetOrZero(),
		Location:      model.Location.GetOrZero(),
		PositionCode:  model.PositionCode.GetOrZero(),

		Purchases: purchases,

		PartsTotalCounter: int(model.PartsTotalCounter),

		Files: files,

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

func omitnullCustomAttrs(cas []CustomAttr) omitnull.Val[types.SQLiteJSON[[]map[string]any]] {
	encoded := make([]map[string]any, 0, len(cas))
	for _, ca := range cas {
		encoded = append(encoded, map[string]any{"name": ca.Name, "value": ca.Value})
	}

	j := types.NewSQLiteJSON(encoded)
	v := omitnull.From(j)
	if len(cas) == 0 {
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

func nullCustomAttrs(cas []CustomAttr) null.Val[types.SQLiteJSON[[]map[string]any]] {
	encoded := make([]map[string]any, 0, len(cas))
	for _, ca := range cas {
		encoded = append(encoded, map[string]any{"name": ca.Name, "value": ca.Value})
	}

	j := types.NewSQLiteJSON(encoded)
	v := null.From(j)
	if len(cas) == 0 {
		v.Null()
	}

	return v
}

func unmarshalCustomAttrs(encoded []map[string]any) []CustomAttr {
	cas := make([]CustomAttr, 0, len(encoded))
	for _, ca := range encoded {
		cas = append(cas, CustomAttr{
			Name:  ca["name"].(string),
			Value: ca["value"],
		})
	}

	return cas
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
