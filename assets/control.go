package assets

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"

	"github.com/RobinThrift/stuff/storage/database"
	"github.com/RobinThrift/stuff/tags"
	"github.com/stephenafamo/bob"
)

type Control struct {
	DB        *database.Database
	AssetRepo AssetRepo
	TagCtrl   *tags.Control

	FileDir string
	TmpDir  string
}

type AssetRepo interface {
	Get(ctx context.Context, exec bob.Executor, idOrTag string) (*Asset, error)
	List(ctx context.Context, exec bob.Executor, query ListAssetsQuery) (*AssetListPage, error)
	ListForExport(ctx context.Context, exec bob.Executor, query ListAssetsQuery) (*AssetListPage, error)
	Create(ctx context.Context, exec bob.Executor, asset *Asset) (*Asset, error)
	Update(ctx context.Context, exec bob.Executor, asset *Asset) (*Asset, error)
	Delete(ctx context.Context, exec bob.Executor, id int64) error

	CreateParts(ctx context.Context, exec bob.Executor, parts []*Part) error
	DeleteParts(ctx context.Context, exec bob.Executor, assetID int64) error

	ListCategories(ctx context.Context, exec bob.Executor, query ListCategoriesQuery) ([]Category, error)
}

func (c *Control) generateTag(ctx context.Context) (string, error) {
	return c.TagCtrl.GetNext(ctx)
}

func (c *Control) getAsset(ctx context.Context, idOrTag string) (*Asset, error) {
	return database.InTransaction(ctx, c.DB, func(ctx context.Context, tx bob.Tx) (*Asset, error) {
		return c.AssetRepo.Get(ctx, tx, idOrTag)
	})
}

func (c *Control) listAssets(ctx context.Context, query ListAssetsQuery) (*AssetListPage, error) {
	return database.InTransaction(ctx, c.DB, func(ctx context.Context, tx bob.Tx) (*AssetListPage, error) {
		return c.AssetRepo.List(ctx, tx, query)
	})
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

	created, err := database.InTransaction(ctx, c.DB, func(ctx context.Context, tx bob.Tx) (*Asset, error) {
		_, err := c.TagCtrl.CreateIfNotExists(ctx, asset.Tag)
		if err != nil {
			return nil, err
		}

		asset.PartsTotalCounter = len(asset.Parts)
		asset.ImageURL = fileURL
		asset.ThumbnailURL = fileURL

		return c.AssetRepo.Create(ctx, tx, asset)
	})
	if err != nil {
		return nil, err
	}

	return created, nil
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

	updated, err := database.InTransaction(ctx, c.DB, func(ctx context.Context, tx bob.Tx) (*Asset, error) {
		_, err := c.TagCtrl.CreateIfNotExists(ctx, asset.Tag)
		if err != nil {
			return nil, err
		}

		err = c.AssetRepo.DeleteParts(ctx, tx, asset.ID)
		if err != nil {
			return nil, err
		}

		for i := range asset.Parts {
			asset.Parts[i].AssetID = asset.ID
		}

		err = c.AssetRepo.CreateParts(ctx, tx, asset.Parts)
		if err != nil {
			return nil, err
		}

		asset.ImageURL = fileURL
		asset.ThumbnailURL = fileURL
		asset.PartsTotalCounter = len(asset.Parts)

		return c.AssetRepo.Update(ctx, tx, asset)
	})
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (c *Control) deleteAsset(ctx context.Context, asset *Asset) (err error) {
	return c.DB.InTransaction(ctx, func(ctx context.Context, tx bob.Tx) error {
		err := c.TagCtrl.MarkTagUnused(ctx, tx, asset.Tag)
		if err != nil {
			return err
		}

		return c.AssetRepo.Delete(ctx, tx, asset.ID)
	})
}

func (c *Control) listCategories(ctx context.Context, query ListCategoriesQuery) ([]Category, error) {
	return database.InTransaction(ctx, c.DB, func(ctx context.Context, tx bob.Tx) ([]Category, error) {
		return c.AssetRepo.ListCategories(ctx, tx, query)
	})
}

type getLabelSheetsQuery struct {
	baseURL *url.URL
	ids     []int64
	sheet   *Sheet
}

func (c *Control) getLabelSheets(ctx context.Context, query getLabelSheetsQuery) ([]byte, error) {
	assets, err := c.AssetRepo.ListForExport(ctx, c.DB, ListAssetsQuery{IDs: query.ids})
	if err != nil {
		return nil, err
	}

	labels := make([]Label, 0, len(assets.Assets))
	for _, a := range assets.Assets {
		l, err := a.Labels(query.baseURL, 200)
		if err != nil {
			return nil, err
		}
		labels = append(labels, l...)
	}

	query.sheet.Labels = labels

	return query.sheet.Generate()
}

func (c *Control) handleFileUpload(origFileName string, r io.Reader) (filename string, hash string, err error) { //nolint: unparam // will fix soon
	err = ensureDirExists(c.FileDir)
	if err != nil {
		return "", "", err
	}

	fh, err := os.CreateTemp(c.TmpDir, origFileName)
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
