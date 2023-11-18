package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/RobinThrift/stuff/entities"
	"github.com/RobinThrift/stuff/storage/database"
	"github.com/RobinThrift/stuff/storage/database/sqlite/models"
	"github.com/RobinThrift/stuff/storage/database/sqlite/types"
	"github.com/aarondl/opt/omit"
	"github.com/aarondl/opt/omitnull"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/sqlite"
	"github.com/stephenafamo/bob/dialect/sqlite/dialect"
	"github.com/stephenafamo/bob/dialect/sqlite/im"
	"github.com/stephenafamo/bob/dialect/sqlite/sm"
)

var ErrAssetNotFound = errors.New("asset not found")

type AssetRepo struct{}

func (ar *AssetRepo) Get(ctx context.Context, exec bob.Executor, query database.GetAssetQuery) (*entities.Asset, error) {
	qmods := make([]bob.Mod[*dialect.SelectQuery], 0, 3)

	if query.IncludeParts {
		qmods = append(qmods, models.ThenLoadAssetAssetParts())
	}

	if query.IncludePurchases {
		qmods = append(qmods, models.ThenLoadAssetAssetPurchases())
	}

	if query.IncludeParent {
		qmods = append(qmods, models.PreloadAssetParentAsset())
	}

	if query.IncludeFiles {
		qmods = append(qmods, models.ThenLoadAssetAssetFiles())
	}

	switch {
	case query.Tag != "" && query.ID != 0:
		qmods = append(qmods, sqlite.WhereOr(
			models.SelectWhere.Assets.Tag.EQ(query.Tag),
			models.SelectWhere.Assets.ID.EQ(query.ID),
		))
	case query.Tag != "":
		qmods = append(qmods, models.SelectWhere.Assets.Tag.EQ(query.Tag))
	case query.ID != 0:
		qmods = append(qmods, models.SelectWhere.Assets.ID.EQ(query.ID))
	}

	asset, err := models.Assets.Query(ctx, exec, qmods...).One()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAssetNotFound
		}
		return nil, fmt.Errorf("error getting assets: %w", err)
	}

	var children models.AssetSlice
	if query.IncludeChildren {
		children, err = models.Assets.Query(ctx, exec,
			sm.Columns(models.Assets.Columns().Only("id", "name", "tag")),
			models.SelectWhere.Assets.ParentAssetID.EQ(asset.ID),
		).All()
		if err != nil {
			return nil, fmt.Errorf("error getting asset children: %w", err)
		}
	}

	return mapDBModelToAsset(asset, children), nil
}

func (ar *AssetRepo) List(ctx context.Context, exec bob.Executor, query database.ListAssetsQuery) (*entities.ListPage[*entities.Asset], error) {
	var err error
	var assets models.AssetSlice
	var count int64
	if len(query.SearchFields) != 0 || query.SearchRaw != "" {
		assets, count, err = searchAssets(ctx, exec, query)
	} else {
		assets, count, err = listAssets(ctx, exec, query)
	}
	if err != nil {
		return nil, err
	}

	numPages, pageSize := calcNumPages(query.PageSize, count)
	page := &entities.ListPage[*entities.Asset]{
		Items:    make([]*entities.Asset, 0, len(assets)),
		Total:    int(count),
		Page:     query.Page,
		PageSize: pageSize,
		NumPages: numPages,
	}

	for i := range assets {
		page.Items = append(page.Items, mapDBModelToAsset(assets[i], nil))
	}

	return page, nil
}

func (ar *AssetRepo) Create(ctx context.Context, exec bob.Executor, asset *entities.Asset) error {
	setter := mapAssetToSetter(asset)

	inserted, err := models.Assets.Insert(ctx, exec, setter)
	if err != nil {
		return err
	}

	err = createParts(ctx, exec, inserted.ID, asset.Parts)
	if err != nil {
		return err
	}

	asset.ID = inserted.ID

	err = createPurchases(ctx, exec, asset, asset.Purchases)
	if err != nil {
		return err
	}
	return nil
}

func (ar *AssetRepo) Update(ctx context.Context, exec bob.Executor, asset *entities.Asset) error {
	setter := mapAssetToSetter(asset)
	setter.UpdatedAt = omit.From(types.NewSQLiteDatetime(time.Now()))
	setter.CreatedBy = omit.FromPtr[int64](nil)

	_, err := models.Assets.UpdateQ(ctx, exec, models.UpdateWhere.Assets.ID.EQ(asset.ID), setter).Exec()
	if err != nil {
		return err
	}

	err = deleteParts(ctx, exec, asset.ID)
	if err != nil {
		return err
	}

	err = createParts(ctx, exec, asset.ID, asset.Parts)
	if err != nil {
		return err
	}

	err = deletePurchases(ctx, exec, asset.ID)
	if err != nil {
		return err
	}

	err = createPurchases(ctx, exec, asset, asset.Purchases)
	if err != nil {
		return err
	}

	return nil
}

func (ar *AssetRepo) Delete(ctx context.Context, exec bob.Executor, id int64) error {
	_, err := models.Assets.DeleteQ(ctx, exec, models.DeleteWhere.Assets.ID.EQ(id)).Exec()
	if err != nil {
		return err
	}

	err = deletePurchases(ctx, exec, id)
	if err != nil {
		return err
	}

	return deleteParts(ctx, exec, id)
}

func listAssets(ctx context.Context, exec bob.Executor, query database.ListAssetsQuery) (models.AssetSlice, int64, error) {
	limit := query.PageSize
	offset := limit * query.Page

	qmods := make([]bob.Mod[*dialect.SelectQuery], 0, 3)

	if len(query.IDs) != 0 {
		qmods = append(qmods, models.SelectWhere.Assets.ID.In(query.IDs...))
	}

	count, err := models.Assets.Query(ctx, exec, qmods...).Count()
	if err != nil {
		return nil, 0, fmt.Errorf("error counting assets: %w", err)
	}

	if query.IncludeParts {
		qmods = append(qmods, models.ThenLoadAssetAssetParts())
	}

	if query.IncludePurchases {
		qmods = append(qmods, models.ThenLoadAssetAssetPurchases())
	}

	if query.IncludeParent {
		qmods = append(qmods, models.PreloadAssetParentAsset())
	}

	if query.IncludeFiles {
		qmods = append(qmods, models.PreloadAssetFileAsset())
	}

	if query.AssetType != "" {
		qmods = append(qmods, models.SelectWhere.Assets.Type.EQ(query.AssetType))
	}

	if query.OrderBy != "" {
		if query.OrderDir == "" {
			query.OrderDir = database.OrderASC
		}

		qmods = append(qmods, orderByClause(models.TableNames.Assets, query.OrderBy, query.OrderDir))
	}

	if limit > 0 {
		qmods = append(qmods, sm.Limit(limit))
	}

	if offset > 0 {
		qmods = append(qmods, sm.Offset(offset))
	}

	assets, err := models.Assets.Query(ctx, exec, qmods...).All()
	if err != nil {
		return nil, 0, fmt.Errorf("error getting assets: %w", wrapSqliteErr(err))
	}

	return assets, count, nil
}

var removeSpecialChars = regexp.MustCompile(`[^\w* ]`)

func searchAssets(ctx context.Context, exec bob.Executor, query database.ListAssetsQuery) (models.AssetSlice, int64, error) {
	limit := query.PageSize
	offset := limit * query.Page

	mods := make([]bob.Mod[*dialect.SelectQuery], 0, 2)

	if limit > 0 {
		mods = append(mods, sm.Limit(limit))
	}

	if offset > 0 {
		mods = append(mods, sm.Offset(offset))
	}

	if len(query.SearchFields) == 0 {
		value := removeSpecialChars.ReplaceAllString(query.SearchRaw, "")
		if value[len(value)-1] != '*' {
			value += "*"
		}

		mods = append(mods,
			sm.Where(sqlite.Quote(models.TableNames.AssetsFTS).EQ(sqlite.Quote(value))),
		)
	}

	for field, value := range query.SearchFields {
		column, ok := isAssetsFTSColumn(field)
		if !ok {
			continue
		}

		value = removeSpecialChars.ReplaceAllString(value, "")

		mods = append(mods, sm.Where(sqlite.Raw(column+" MATCH ?", value)))
	}

	entries, err := models.AssetsFTS.Query(ctx, exec, mods...).All()
	if err != nil {
		return nil, 0, fmt.Errorf("error searching assets: %w", err)
	}

	if len(entries) == 0 {
		return nil, 0, nil
	}

	ids := make([]int64, 0, len(entries))

	for _, entry := range entries {
		id, err := strconv.ParseInt(entry.ID.GetOrZero(), 10, 64)
		if err != nil {
			continue
		}

		ids = append(ids, id)
	}

	return listAssets(ctx, exec, database.ListAssetsQuery{
		IDs:              ids,
		Page:             query.Page,
		PageSize:         query.PageSize,
		OrderBy:          query.OrderBy,
		OrderDir:         query.OrderDir,
		AssetType:        query.AssetType,
		IncludePurchases: query.IncludePurchases,
		IncludeParts:     query.IncludeParts,
		IncludeFiles:     query.IncludeFiles,
		IncludeParent:    query.IncludeParent,
		IncludeChildren:  query.IncludeChildren,
	})
}

func createPurchases(ctx context.Context, exec bob.Executor, asset *entities.Asset, purchases []*entities.Purchase) error {
	if len(purchases) == 0 {
		return nil
	}

	inserts := make([]bob.Mod[*dialect.InsertQuery], 0, len(purchases)+1)

	inserts = append(inserts,
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

	for _, p := range purchases {
		inserts = append(inserts, models.AssetPurchaseSetter{
			AssetID:   omit.From(asset.ID),
			Supplier:  omitnullStr(p.Supplier),
			OrderNo:   omitnullStr(p.OrderNo),
			OrderDate: omitnullTime(p.Date),
			Amount:    omitnullInt64(int64(p.Amount)),
			Currency:  omitnullStr(p.Currency),
			CreatedBy: omit.From(asset.MetaInfo.CreatedBy),
			UpdatedAt: omit.From(types.NewSQLiteDatetime(time.Now())),
		}.Insert())
	}

	_, err := models.AssetPurchases.InsertQ(ctx, exec, inserts...).Exec()
	return err
}

func deletePurchases(ctx context.Context, exec bob.Executor, assetID int64) error {
	_, err := models.AssetPurchases.DeleteQ(ctx, exec, models.DeleteWhere.AssetPurchases.AssetID.EQ(assetID)).Exec()
	return err
}

func createParts(ctx context.Context, exec bob.Executor, assetID int64, parts []*entities.Part) error {
	if len(parts) == 0 {
		return nil
	}

	inserts := make([]bob.Mod[*dialect.InsertQuery], 0, len(parts)+1)

	inserts = append(inserts,
		im.Into(models.TableNames.AssetParts,
			models.ColumnNames.AssetParts.AssetID,
			models.ColumnNames.AssetParts.Tag,
			models.ColumnNames.AssetParts.Name,
			models.ColumnNames.AssetParts.Location,
			models.ColumnNames.AssetParts.PositionCode,
			models.ColumnNames.AssetParts.Notes,
			models.ColumnNames.AssetParts.CreatedBy,
		),
	)

	for _, part := range parts {
		inserts = append(inserts, models.AssetPartSetter{
			AssetID:      omit.From(assetID),
			Tag:          omit.From(part.Tag),
			Name:         omit.From(part.Name),
			Location:     omitnullStr(part.Location),
			PositionCode: omitnullStr(part.PositionCode),
			Notes:        omitnullStr(part.Notes),
			CreatedBy:    omit.From(part.CreatedBy),
		}.Insert())
	}

	_, err := models.AssetParts.InsertQ(ctx, exec, inserts...).Exec()
	return err
}

func deleteParts(ctx context.Context, exec bob.Executor, assetID int64) error {
	_, err := models.AssetParts.DeleteQ(ctx, exec, models.DeleteWhere.AssetParts.AssetID.EQ(assetID)).Exec()
	return err
}

func mapDBModelToAsset(model *models.Asset, children []*models.Asset) *entities.Asset {
	if model == nil {
		return nil
	}

	purchases := make([]*entities.Purchase, 0, len(model.R.AssetPurchases))
	for _, p := range model.R.AssetPurchases {
		purchases = append(purchases, &entities.Purchase{
			Supplier: p.Supplier.GetOrZero(),
			OrderNo:  p.OrderNo.GetOrZero(),
			Date:     p.OrderDate.GetOrZero().Time,
			Amount:   entities.MonetaryAmount(p.Amount.GetOrZero()),
			Currency: p.Currency.GetOrZero(),
		})
	}

	files := make([]*entities.File, 0, len(model.R.AssetFiles))
	for _, f := range model.R.AssetFiles {
		files = append(files, &entities.File{
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

	parts := make([]*entities.Part, 0, len(model.R.AssetParts))
	for _, p := range model.R.AssetParts {
		parts = append(parts, &entities.Part{
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

	mappedChildren := make([]*entities.Asset, 0, len(children))
	for _, child := range children {
		mappedChildren = append(mappedChildren, mapDBModelToAsset(child, nil))
	}

	return &entities.Asset{
		ID:            model.ID,
		Type:          entities.AssetType(model.Type),
		ParentAssetID: model.ParentAssetID.GetOrZero(),
		Parent:        mapDBModelToAsset(model.R.ParentAsset, nil),
		Status:        entities.Status(model.Status),
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
		Parts:             parts,

		Files: files,

		Children: mappedChildren,

		MetaInfo: entities.MetaInfo{
			CreatedBy: model.CreatedBy,
			CreatedAt: model.CreatedAt.Time,
			UpdatedAt: model.UpdatedAt.Time,
		},
	}
}

func mapAssetToSetter(asset *entities.Asset) *models.AssetSetter {
	return &models.AssetSetter{
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
		CheckedOutTo:      omitnullInt64(asset.CheckedOutTo),
		Location:          omitnullStr(asset.Location),
		PositionCode:      omitnullStr(asset.PositionCode),
		PartsTotalCounter: omit.From(int64(len(asset.Parts))),
		CreatedBy:         omit.From(asset.MetaInfo.CreatedBy),
		Type:              omit.From(string(asset.Type)),
		Quantity:          omit.From(asset.Quantity),
		QuantityUnit:      omitnullStr(asset.QuantityUnit),
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

func omitnullCustomAttrs(cas []entities.CustomAttr) omitnull.Val[types.SQLiteJSON[[]map[string]any]] {
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

func unmarshalCustomAttrs(encoded []map[string]any) []entities.CustomAttr {
	cas := make([]entities.CustomAttr, 0, len(encoded))
	for _, ca := range encoded {
		cas = append(cas, entities.CustomAttr{
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
