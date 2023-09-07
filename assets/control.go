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
	Get(ctx context.Context, exec bob.Executor, id int64) (*Asset, error)
	List(ctx context.Context, exec bob.Executor, query ListAssetsQuery) (*AssetListPage, error)
	Search(ctx context.Context, exec bob.Executor, query ListAssetsQuery) (*AssetListPage, error)
	Create(ctx context.Context, exec bob.Executor, asset *Asset) (*Asset, error)
	Update(ctx context.Context, exec bob.Executor, asset *Asset) (*Asset, error)
	Delete(ctx context.Context, exec bob.Executor, id int64) error

	ListCategories(ctx context.Context, exec bob.Executor) ([]string, error)
}

type ListAssetsQuery struct {
	Search string

	Page     int
	PageSize int

	OrderBy  string
	OrderDir string
}

func (c *Control) generateTag(ctx context.Context) (string, error) {
	return c.TagCtrl.GetNext(ctx)
}

func (c *Control) getAsset(ctx context.Context, id int64) (*Asset, error) {
	return database.InTransaction(ctx, c.DB, func(ctx context.Context, tx bob.Tx) (*Asset, error) {
		return c.AssetRepo.Get(ctx, tx, id)
	})
}

func (c *Control) listAssets(ctx context.Context, query ListAssetsQuery) (*AssetListPage, error) {
	return database.InTransaction(ctx, c.DB, func(ctx context.Context, tx bob.Tx) (*AssetListPage, error) {
		return c.AssetRepo.List(ctx, tx, query)
	})
}

func (c *Control) searchAssets(ctx context.Context, query ListAssetsQuery) (*AssetListPage, error) {
	return database.InTransaction(ctx, c.DB, func(ctx context.Context, tx bob.Tx) (*AssetListPage, error) {
		return c.AssetRepo.Search(ctx, tx, query)
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

		asset.ImageURL = fileURL
		asset.ThumbnailURL = fileURL

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
