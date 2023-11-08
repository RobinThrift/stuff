package control

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/RobinThrift/stuff/entities"
	"github.com/RobinThrift/stuff/storage/blobs"
	"github.com/RobinThrift/stuff/storage/database"
	"github.com/RobinThrift/stuff/storage/database/sqlite"
	nanoid "github.com/matoous/go-nanoid/v2"
	"github.com/stephenafamo/bob"
	"github.com/stretchr/testify/assert"
)

func TestAssetControl_CRUD(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	assetCtrl := newTestAssetControl(t)
	asset := newTestAsset(t)

	imgFile := newTestFile(t, 0, 1)

	created, err := assetCtrl.Create(ctx, CreateAssetCmd{Asset: asset, Image: imgFile})
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
	fileExitsts(t, imgFile.FullPath)

	asset = updateAsset(created)
	imgFile = newTestFile(t, 2, 1)

	updated, err := assetCtrl.Update(ctx, UpdateAssetCmd{Asset: asset, Image: imgFile})
	assert.NoError(t, err)

	asset.ID = created.ID
	asset.MetaInfo.UpdatedAt = updated.MetaInfo.UpdatedAt

	asset.Parts[0].ID = updated.Parts[0].ID
	asset.Parts[0].AssetID = asset.ID
	asset.Parts[0].CreatedAt = updated.Parts[0].CreatedAt
	asset.Parts[0].UpdatedAt = updated.Parts[0].UpdatedAt

	asset.Parts[1].ID = updated.Parts[1].ID
	asset.Parts[1].AssetID = asset.ID
	asset.Parts[1].CreatedAt = updated.Parts[1].CreatedAt
	asset.Parts[1].UpdatedAt = updated.Parts[1].UpdatedAt

	assert.Equal(t, asset, updated)
	assert.Equal(t, asset.ImageURL, imgFile.PublicPath)
	fileExitsts(t, imgFile.FullPath)

	fetchedAfterUpdate, err := assetCtrl.Get(ctx, GetAssetQuery{ID: asset.ID, IncludeFiles: true})
	assert.NoError(t, err)
	assert.Len(t, fetchedAfterUpdate.Files, 1)

	err = assetCtrl.Delete(ctx, asset)
	assert.NoError(t, err)

	fetchedAfterDelete, err := assetCtrl.Get(ctx, GetAssetQuery{ID: asset.ID, IncludeFiles: true})
	assert.ErrorIs(t, err, ErrAssetNotFound)
	assert.Nil(t, fetchedAfterDelete)

	inUse := false
	tags, err := assetCtrl.tags.List(ctx, ListTagsQuery{InUse: &inUse})
	assert.NoError(t, err)
	assert.Len(t, tags.Items, 1)

	files, err := assetCtrl.files.List(ctx, ListFilesQuery{AssetID: asset.ID})
	assert.NoError(t, err)
	assert.Len(t, files.Items, 0)

	fileNotExitsts(t, imgFile.FullPath)
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
		MetaInfo: entities.MetaInfo{
			CreatedBy: 1,
		},
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

func randTime() time.Time {
	return time.Date(
		2023,
		randFrom([]time.Month{time.January, time.February, time.March, time.April}),
		rand.Intn(8)+8, rand.Intn(8)+8, rand.Intn(60),
		0, 0, time.UTC,
	)
}

func newTestAssetControl(t *testing.T) *AssetControl {
	db, err := sqlite.NewSQLiteDB(":memory:")
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

	err = sqlite.RunMigrations(ctx, db)
	if err != nil {
		t.Fatal(err)
	}

	database := &database.Database{DB: bob.NewDB(db)}

	return NewAssetControl(
		database,
		NewTagControl(database, "nanoid", &sqlite.TagRepo{}),
		NewFileControl(
			database,
			&sqlite.FileRepo{},
			&blobs.LocalFS{
				RootDir: t.TempDir(),
				TmpDir:  t.TempDir(),
			},
		),
		&sqlite.AssetRepo{},
	)
}
