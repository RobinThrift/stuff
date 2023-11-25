package sqlite

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/RobinThrift/stuff/entities"
	"github.com/RobinThrift/stuff/storage/database"
	nanoid "github.com/matoous/go-nanoid/v2"
	"github.com/stephenafamo/bob"
	"github.com/stretchr/testify/assert"
)

func TestAssetRepo_CRUD(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	ar, exec := newTestAssetRepo(t) // nolint: varnamelen // this is fine for the test

	asset := newTestAsset(t)

	err := ar.Create(ctx, exec, asset)
	assert.NoError(t, err)

	created, err := ar.Get(ctx, exec, database.GetAssetQuery{ID: 1, IncludePurchases: true, IncludeParts: true})
	assert.NoError(t, err)

	asset.Children = []*entities.Asset{}
	asset.Files = []*entities.File{}

	asset.ID = created.ID
	asset.MetaInfo.CreatedAt = created.MetaInfo.CreatedAt
	asset.MetaInfo.UpdatedAt = created.MetaInfo.UpdatedAt

	asset.Parts[0].ID = created.Parts[0].ID
	asset.Parts[0].AssetID = created.ID
	asset.Parts[0].CreatedAt = created.Parts[0].CreatedAt
	asset.Parts[0].UpdatedAt = created.Parts[0].UpdatedAt

	asset.Parts[1].ID = created.Parts[1].ID
	asset.Parts[1].AssetID = created.ID
	asset.Parts[1].CreatedAt = created.Parts[1].CreatedAt
	asset.Parts[1].UpdatedAt = created.Parts[1].UpdatedAt

	assert.Equal(t, asset, created)

	updated := updateAsset(created)

	err = ar.Update(ctx, exec, updated)
	assert.NoError(t, err)

	fetched, err := ar.Get(ctx, exec, database.GetAssetQuery{ID: 1, IncludePurchases: true, IncludeParts: true})
	assert.NoError(t, err)

	updated.MetaInfo.UpdatedAt = fetched.MetaInfo.UpdatedAt

	updated.Parts[0].ID = fetched.Parts[0].ID
	updated.Parts[0].CreatedAt = fetched.Parts[0].CreatedAt
	updated.Parts[0].UpdatedAt = fetched.Parts[0].UpdatedAt

	updated.Parts[1].ID = fetched.Parts[1].ID
	updated.Parts[1].CreatedAt = fetched.Parts[1].CreatedAt
	updated.Parts[1].UpdatedAt = fetched.Parts[1].UpdatedAt

	assert.Equal(t, updated.Parts[0], fetched.Parts[0])

	err = ar.Delete(ctx, exec, fetched.ID)
	assert.NoError(t, err)
	_, err = ar.Get(ctx, exec, database.GetAssetQuery{ID: 1})
	assert.ErrorIs(t, err, ErrAssetNotFound)
}

func TestAssetRepo_ListPagination(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	repo, exec := newTestAssetRepo(t)

	types := []entities.AssetType{entities.AssetTypeAsset, entities.AssetTypeComponent, entities.AssetTypeConsumable}

	for i := 0; i < 100; i++ {
		asset := newTestAsset(t)
		asset.Name = fmt.Sprintf("%s %d", asset.Name, i)
		asset.Type = types[i%3]
		err := repo.Create(ctx, exec, asset)
		assert.NoError(t, err)
	}

	type exp struct {
		len      int
		total    int
		page     int
		pageSize int
		numPages int
	}

	tt := []struct {
		name string
		q    database.ListAssetsQuery
		exp  exp
	}{
		{
			"List all",
			database.ListAssetsQuery{},
			exp{
				len:      100,
				total:    100,
				page:     0,
				pageSize: 100,
				numPages: 1,
			},
		},
		{
			"Page size 25",
			database.ListAssetsQuery{PageSize: 25},
			exp{
				len:      25,
				total:    100,
				page:     0,
				pageSize: 25,
				numPages: 4,
			},
		},
		{
			"All filtered",
			database.ListAssetsQuery{AssetType: string(entities.AssetTypeAsset)},
			exp{
				len:      34,
				total:    34,
				page:     0,
				pageSize: 34,
				numPages: 1,
			},
		},
		{
			"Filtered by FTS",
			database.ListAssetsQuery{SearchRaw: "name:Test Asset 1*", SearchFields: map[string]string{"name": "Test Asset 1*"}},
			exp{
				len:      11,
				total:    11,
				page:     0,
				pageSize: 11,
				numPages: 1,
			},
		},
	}

	for _, tt := range tt {
		t.Run(tt.name, func(t *testing.T) {
			list, err := repo.List(ctx, exec, tt.q)
			assert.NoError(t, err)

			assert.Len(t, list.Items, tt.exp.len, "len of items")
			assert.Equal(t, tt.exp.total, list.Total, "total num of items")
			assert.Equal(t, tt.exp.page, list.Page, "page num")
			assert.Equal(t, tt.exp.pageSize, list.PageSize, "page size")
			assert.Equal(t, tt.exp.numPages, list.NumPages, "num of pages")
		})
	}
}

func newTestAssetRepo(t *testing.T) (*AssetRepo, bob.Executor) {
	db, err := NewSQLiteDB(&Config{File: ":memory:", Timeout: time.Millisecond * 500})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err = db.Close(); err != nil {
			t.Error(err)
		}
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err = db.ExecContext(ctx, "PRAGMA foreign_keys = 0")
	if err != nil {
		t.Fatal(err)
	}

	err = RunMigrations(ctx, db)
	if err != nil {
		t.Fatal(err)
	}

	return &AssetRepo{}, bob.NewDB(db)
}

func newTestAsset(t *testing.T) *entities.Asset {
	tag, err := nanoid.Generate("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ", 6)
	if err != nil {
		t.Fatal(err)
	}

	asset := &entities.Asset{
		Type:          entities.AssetTypeConsumable,
		Status:        entities.StatusInUse,
		Tag:           tag,
		Name:          "Test Asset",
		Category:      "Regular Assets",
		Model:         "Asset Model",
		ModelNo:       "98765",
		SerialNo:      "45678",
		Manufacturer:  "Asset Maker",
		Notes:         "Notes about asset",
		ImageURL:      "/path/to/image",
		ThumbnailURL:  "/path/to/thumbnail",
		WarrantyUntil: randTime(),
		Quantity:      3,
		QuantityUnit:  "pc",
		CustomAttrs: []entities.CustomAttr{
			{Name: "Attr1", Value: "value1"},
			{Name: "Attr2", Value: 3.0},
		},
		Location:          "Where it was put",
		PositionCode:      "yes",
		Purchases:         []*entities.Purchase{},
		Parts:             []*entities.Part{},
		PartsTotalCounter: 1,
		MetaInfo:          entities.MetaInfo{CreatedBy: 1},
	}

	asset.Parts = []*entities.Part{newTestAssetPart(asset)}
	asset.Parts = append(asset.Parts, newTestAssetPart(asset))

	asset.Purchases = append(asset.Purchases, newTestPurchase(asset))
	asset.Purchases = append(asset.Purchases, newTestPurchase(asset))

	return asset
}

func updateAsset(asset *entities.Asset) *entities.Asset {
	updated := *asset

	updated.Status = entities.StatusInStorage
	updated.Name = asset.Name + " changed" // nolint const
	updated.Category += " changed"
	updated.Model += " changed"
	updated.ModelNo += " changed"
	updated.SerialNo += " changed"
	updated.Manufacturer += " changed"
	updated.Notes += " changed"
	updated.ImageURL += " changed"
	updated.ThumbnailURL += " changed"
	updated.WarrantyUntil = asset.WarrantyUntil.Add(time.Hour * 24)
	updated.Quantity++
	updated.QuantityUnit = "pieces"
	updated.Location = "somehwere else"
	updated.PositionCode = "nope"

	updated.Parts = updated.Parts[1:]
	updated.PartsTotalCounter = len(updated.Parts)
	updated.Parts = append(updated.Parts, newTestAssetPart(&updated))

	updated.Purchases = updated.Purchases[1:]
	updated.Purchases = append(updated.Purchases, newTestPurchase(&updated))

	return &updated
}

func newTestAssetPart(asset *entities.Asset) *entities.Part {
	char := 65 + asset.PartsTotalCounter
	asset.PartsTotalCounter = len(asset.Parts) + 1
	return &entities.Part{
		AssetID:      asset.ID,
		Tag:          fmt.Sprintf("%s-%v", asset.Tag, char),
		Name:         fmt.Sprintf("Part %v", char),
		Location:     fmt.Sprintf("%s - part %v", asset.Location, char),
		PositionCode: fmt.Sprintf("%s - part %v", asset.PositionCode, char),
		Notes:        fmt.Sprintf("this is part %s-%v", asset.Tag, char),
		CreatedBy:    asset.MetaInfo.CreatedBy,
	}
}

func newTestPurchase(asset *entities.Asset) *entities.Purchase {
	char := 65 + len(asset.Purchases)
	return &entities.Purchase{
		Supplier: fmt.Sprintf("Shop %v", char),
		OrderNo:  fmt.Sprintf("#%v-%v", char, randMonetaryAmount(10000, 200000)),
		Date:     randTime(),
		Amount:   randMonetaryAmount(1000, 100000),
		Currency: randFrom([]string{"GBP", "EUR"}),
	}
}

func randMonetaryAmount(min int, max int) entities.MonetaryAmount {
	return entities.MonetaryAmount(rand.Intn(max-min) + min)
}

func randFrom[T any](items []T) T {
	i := rand.Intn(len(items))
	return items[i]
}

func randTime() time.Time {
	return time.Date(
		2023,
		randFrom([]time.Month{time.January, time.February, time.March, time.April}),
		rand.Intn(8)+8, rand.Intn(8)+8, rand.Intn(60),
		0, 0, time.UTC,
	)
}
