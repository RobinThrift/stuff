package assets

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/kodeshack/stuff/storage/database"
	"github.com/kodeshack/stuff/tags"
	"github.com/stephenafamo/bob"
)

type Control struct {
	DB        *database.Database
	AssetRepo AssetRepo
	TagCtrl   *tags.Control

	FileDir string
}

type AssetRepo interface {
	GetAsset(ctx context.Context, exec bob.Executor, id int64) (*database.Asset, error)
	ListAssets(ctx context.Context, exec bob.Executor, query database.ListAssetsQuery) (*database.AssetList, error)
	CreateAsset(ctx context.Context, exec bob.Executor, asset *database.Asset) (*database.Asset, error)
	UpdateAsset(ctx context.Context, exec bob.Executor, asset *database.Asset) (*database.Asset, error)
	DeleteAsset(ctx context.Context, exec bob.Executor, id int64) error

	ListCategories(ctx context.Context, exec bob.Executor) ([]string, error)
}

type listAssetsQuery struct {
	offset   int
	limit    int
	orderBy  string
	orderDir string
}

func (c *Control) generateTag(ctx context.Context) (string, error) {
	return c.TagCtrl.GetNext(ctx)
}

func (c *Control) getAsset(ctx context.Context, id int64) (*Asset, error) {
	return database.InTransaction(ctx, c.DB, func(ctx context.Context, tx bob.Tx) (*Asset, error) {
		asset, err := c.AssetRepo.GetAsset(ctx, tx, id)
		if err != nil {
			return nil, err
		}

		return &Asset{
			ID:            asset.ID,
			ParentAssetID: asset.ParentAssetID,
			Status:        Status(asset.Status),

			Name:          asset.Name,
			Category:      asset.Category,
			SerialNo:      asset.SerialNo,
			Model:         asset.Model,
			ModelNo:       asset.ModelNo,
			Manufacturer:  asset.Manufacturer,
			Notes:         asset.Notes,
			ImageURL:      asset.ImageURL,
			ThumbnailURL:  asset.ThumbnailURL,
			WarrantyUntil: asset.WarrantyUntil,
			CustomAttrs:   asset.CustomAttrs,
			Tag:           asset.Tag,
			CheckedOutTo:  asset.CheckedOutTo,
			Location:      asset.Location,
			PositionCode:  asset.PositionCode,

			PurchaseInfo: PurchaseInfo{
				Supplier: asset.PurchaseSupplier,
				OrderNo:  asset.PurchaseOrderNo,
				Date:     asset.PurchaseDate,
				Amount:   MonetaryAmount(asset.PurchaseAmount),
				Currency: asset.PurchaseCurrency,
			},

			MetaInfo: MetaInfo{
				CreatedBy: asset.CreatedBy,
				CreatedAt: asset.CreatedAt,
				UpdatedAt: asset.UpdatedAt,
			},
		}, nil
	})
}

func (c *Control) listAssets(ctx context.Context, query listAssetsQuery) (*AssetList, error) {
	assets, err := c.AssetRepo.ListAssets(ctx, c.DB, database.ListAssetsQuery{
		Offset:   query.offset,
		Limit:    query.limit,
		OrderBy:  query.orderBy,
		OrderDir: query.orderDir,
	})

	if err != nil {
		return nil, err
	}

	assetList := &AssetList{
		Assets: make([]*Asset, 0, len(assets.Assets)),
		Total:  assets.Total,
	}

	for i := range assets.Assets {
		assetList.Assets = append(assetList.Assets, &Asset{
			ID:            assets.Assets[i].ID,
			ParentAssetID: assets.Assets[i].ParentAssetID,
			Status:        Status(assets.Assets[i].Status),

			Name:          assets.Assets[i].Name,
			Category:      assets.Assets[i].Category,
			SerialNo:      assets.Assets[i].SerialNo,
			Model:         assets.Assets[i].Model,
			ModelNo:       assets.Assets[i].ModelNo,
			Manufacturer:  assets.Assets[i].Manufacturer,
			Notes:         assets.Assets[i].Notes,
			ImageURL:      assets.Assets[i].ImageURL,
			ThumbnailURL:  assets.Assets[i].ThumbnailURL,
			WarrantyUntil: assets.Assets[i].WarrantyUntil,
			CustomAttrs:   assets.Assets[i].CustomAttrs,
			Tag:           assets.Assets[i].Tag,
			CheckedOutTo:  assets.Assets[i].CheckedOutTo,
			Location:      assets.Assets[i].Location,
			PositionCode:  assets.Assets[i].PositionCode,

			PurchaseInfo: PurchaseInfo{
				Supplier: assets.Assets[i].PurchaseSupplier,
				OrderNo:  assets.Assets[i].PurchaseOrderNo,
				Date:     assets.Assets[i].PurchaseDate,
				Amount:   MonetaryAmount(assets.Assets[i].PurchaseAmount),
				Currency: assets.Assets[i].PurchaseCurrency,
			},

			MetaInfo: MetaInfo{
				CreatedBy: assets.Assets[i].CreatedBy,
				CreatedAt: assets.Assets[i].CreatedAt,
				UpdatedAt: assets.Assets[i].UpdatedAt,
			},
		})
	}

	return assetList, nil
}

func (c *Control) createAsset(ctx context.Context, asset *Asset, file *File) (*Asset, error) {
	fileURL := asset.ImageURL

	if file != nil {
		filename, _, err := c.handleFileUpload(file.Name, file.r)
		if err != nil {
			return nil, err
		}

		fileURL = "/assets/files/" + filename
	}

	created, err := database.InTransaction(ctx, c.DB, func(ctx context.Context, tx bob.Tx) (*database.Asset, error) {
		_, err := c.TagCtrl.CreateIfNotExists(ctx, asset.Tag)
		if err != nil {
			return nil, err
		}

		return c.AssetRepo.CreateAsset(ctx, tx, &database.Asset{
			ParentAssetID:    asset.ParentAssetID,
			Status:           string(asset.Status),
			Tag:              asset.Tag,
			Name:             asset.Name,
			Category:         asset.Category,
			Model:            asset.Model,
			ModelNo:          asset.ModelNo,
			SerialNo:         asset.SerialNo,
			Manufacturer:     asset.Manufacturer,
			Notes:            asset.Notes,
			ImageURL:         fileURL,
			ThumbnailURL:     fileURL,
			WarrantyUntil:    asset.WarrantyUntil,
			CustomAttrs:      asset.CustomAttrs,
			CheckedOutTo:     asset.CheckedOutTo,
			Location:         asset.Location,
			PositionCode:     asset.PositionCode,
			PurchaseSupplier: asset.PurchaseInfo.Supplier,
			PurchaseOrderNo:  asset.PurchaseInfo.OrderNo,
			PurchaseDate:     asset.PurchaseInfo.Date,
			PurchaseAmount:   int(asset.PurchaseInfo.Amount),
			PurchaseCurrency: asset.PurchaseInfo.Currency,
			CreatedBy:        asset.MetaInfo.CreatedBy,
		})
	})
	if err != nil {
		return nil, err
	}

	return &Asset{
		ID:            created.ID,
		ParentAssetID: created.ParentAssetID,
		Status:        Status(created.Status),

		Name:          created.Name,
		SerialNo:      created.SerialNo,
		Model:         created.Model,
		ModelNo:       created.ModelNo,
		Manufacturer:  created.Manufacturer,
		Notes:         created.Notes,
		ImageURL:      created.ImageURL,
		ThumbnailURL:  created.ThumbnailURL,
		WarrantyUntil: created.WarrantyUntil,
		CustomAttrs:   created.CustomAttrs,
		Tag:           created.Tag,
		CheckedOutTo:  created.CheckedOutTo,
		Location:      created.Location,
		PositionCode:  created.PositionCode,

		PurchaseInfo: PurchaseInfo{
			Supplier: created.PurchaseSupplier,
			OrderNo:  created.PurchaseOrderNo,
			Date:     created.PurchaseDate,
			Amount:   MonetaryAmount(created.PurchaseAmount),
			Currency: created.PurchaseCurrency,
		},

		MetaInfo: MetaInfo{
			CreatedBy: created.CreatedBy,
			CreatedAt: created.CreatedAt,
			UpdatedAt: created.UpdatedAt,
		},
	}, nil
}

func (c *Control) updateAsset(ctx context.Context, asset *Asset, file *File) (*Asset, error) {
	fileURL := asset.ImageURL

	if file != nil {
		filename, _, err := c.handleFileUpload(file.Name, file.r)
		if err != nil {
			return nil, err
		}

		fileURL = "/assets/files/" + filename
	}

	updated, err := database.InTransaction(ctx, c.DB, func(ctx context.Context, tx bob.Tx) (*database.Asset, error) {
		_, err := c.TagCtrl.CreateIfNotExists(ctx, asset.Tag)
		if err != nil {
			return nil, err
		}

		return c.AssetRepo.UpdateAsset(ctx, tx, &database.Asset{
			ID:               asset.ID,
			ParentAssetID:    asset.ParentAssetID,
			Status:           string(asset.Status),
			Tag:              asset.Tag,
			Name:             asset.Name,
			Category:         asset.Category,
			Model:            asset.Model,
			ModelNo:          asset.ModelNo,
			SerialNo:         asset.SerialNo,
			Manufacturer:     asset.Manufacturer,
			Notes:            asset.Notes,
			ImageURL:         fileURL,
			ThumbnailURL:     fileURL,
			WarrantyUntil:    asset.WarrantyUntil,
			CustomAttrs:      asset.CustomAttrs,
			CheckedOutTo:     asset.CheckedOutTo,
			Location:         asset.Location,
			PositionCode:     asset.PositionCode,
			PurchaseSupplier: asset.PurchaseInfo.Supplier,
			PurchaseOrderNo:  asset.PurchaseInfo.OrderNo,
			PurchaseDate:     asset.PurchaseInfo.Date,
			PurchaseAmount:   int(asset.PurchaseInfo.Amount),
			PurchaseCurrency: asset.PurchaseInfo.Currency,
			CreatedBy:        asset.MetaInfo.CreatedBy,
		})
	})
	if err != nil {
		return nil, err
	}

	return &Asset{
		ID:            updated.ID,
		ParentAssetID: updated.ParentAssetID,
		Status:        Status(updated.Status),

		Name:          updated.Name,
		SerialNo:      updated.SerialNo,
		Model:         updated.Model,
		ModelNo:       updated.ModelNo,
		Manufacturer:  updated.Manufacturer,
		Notes:         updated.Notes,
		ImageURL:      updated.ImageURL,
		ThumbnailURL:  updated.ThumbnailURL,
		WarrantyUntil: updated.WarrantyUntil,
		CustomAttrs:   updated.CustomAttrs,
		Tag:           updated.Tag,
		CheckedOutTo:  updated.CheckedOutTo,
		Location:      updated.Location,
		PositionCode:  updated.PositionCode,

		PurchaseInfo: PurchaseInfo{
			Supplier: updated.PurchaseSupplier,
			OrderNo:  updated.PurchaseOrderNo,
			Date:     updated.PurchaseDate,
			Amount:   MonetaryAmount(updated.PurchaseAmount),
			Currency: updated.PurchaseCurrency,
		},

		MetaInfo: MetaInfo{
			CreatedBy: updated.CreatedBy,
			CreatedAt: updated.CreatedAt,
			UpdatedAt: updated.UpdatedAt,
		},
	}, nil
}

func (c *Control) deleteAsset(ctx context.Context, asset *Asset) (err error) {
	return c.DB.InTransaction(ctx, func(ctx context.Context, tx bob.Tx) error {
		err := c.TagCtrl.MarkTagUnused(ctx, tx, asset.Tag)
		if err != nil {
			return err
		}

		return c.AssetRepo.DeleteAsset(ctx, tx, asset.ID)
	})
}

func (c *Control) listCategories(ctx context.Context) ([]string, error) {
	return database.InTransaction(ctx, c.DB, func(ctx context.Context, tx bob.Tx) ([]string, error) {
		return c.AssetRepo.ListCategories(ctx, tx)
	})
}

func (c *Control) handleFileUpload(origFileName string, r io.Reader) (filename string, hash string, err error) {
	err = ensureDirExists(c.FileDir)
	if err != nil {
		return "", "", err
	}

	fh, err := os.CreateTemp("", origFileName)
	if err != nil {
		return "", "", err
	}

	defer func() {
		if err != nil {
			err = errors.Join(err, fh.Close(), os.Remove(fh.Name()))
		}
	}()

	h := sha256.New()

	tee := io.TeeReader(r, h)

	_, err = io.Copy(fh, tee)
	if err != nil {
		return "", "", err
	}

	err = fh.Close()
	if err != nil {
		return "", "", err
	}

	hash = fmt.Sprintf("%x", h.Sum(nil))
	filename = hash + path.Ext(origFileName)
	filepath := path.Join(c.FileDir, filename)

	err = os.Rename(fh.Name(), filepath)
	if err != nil {
		return "", "", err
	}

	return filename, hash, nil
}

func ensureDirExists(dir string) error {
	stat, err := os.Stat(dir)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}

	if stat == nil {
		return os.Mkdir(dir, 0755)
	}

	if !stat.IsDir() {
		return fmt.Errorf("%s exists but is not a directory", dir)
	}

	return nil
}
