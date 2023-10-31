package assets

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

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
	GetWithFiles(ctx context.Context, exec bob.Executor, idOrTag string) (*Asset, error)
	List(ctx context.Context, exec bob.Executor, query ListAssetsQuery) (*AssetListPage, error)
	ListForExport(ctx context.Context, exec bob.Executor, query ListAssetsQuery) (*AssetListPage, error)
	Create(ctx context.Context, exec bob.Executor, asset *Asset) (*Asset, error)
	Update(ctx context.Context, exec bob.Executor, asset *Asset) (*Asset, error)
	Delete(ctx context.Context, exec bob.Executor, id int64) error

	CreateParts(ctx context.Context, exec bob.Executor, parts []*Part) error
	DeleteParts(ctx context.Context, exec bob.Executor, assetID int64) error

	CreateFiles(ctx context.Context, exec bob.Executor, files []*File) error
	GetFile(ctx context.Context, exec bob.Executor, id int64) (*File, error)
	FileExists(ctx context.Context, exec bob.Executor, hash []byte) (bool, error)
	DeleteFile(ctx context.Context, exec bob.Executor, id int64) error

	ListCategories(ctx context.Context, exec bob.Executor, query ListCategoriesQuery) ([]Category, error)
	ListCustomAttrNames(ctx context.Context, exec bob.Executor, query ListCustomAttrNamesQuery) ([]string, error)
	ListLocations(ctx context.Context, exec bob.Executor, query ListLocationsQuery) ([]string, error)
	ListPositionCodes(ctx context.Context, exec bob.Executor, query ListPositionCodesQuery) ([]string, error)
}

func (c *Control) generateTag(ctx context.Context) (string, error) {
	return c.TagCtrl.GetNext(ctx)
}

func (c *Control) getAsset(ctx context.Context, idOrTag string) (*Asset, error) {
	return database.InTransaction(ctx, c.DB, func(ctx context.Context, tx bob.Tx) (*Asset, error) {
		return c.AssetRepo.Get(ctx, tx, idOrTag)
	})
}

func (c *Control) getAssetWithFiles(ctx context.Context, idOrTag string) (*Asset, error) {
	return database.InTransaction(ctx, c.DB, func(ctx context.Context, tx bob.Tx) (*Asset, error) {
		return c.AssetRepo.GetWithFiles(ctx, tx, idOrTag)
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
		err := c.handleFileUpload(file)
		if err != nil {
			return nil, err
		}

		fileURL = file.PublicPath
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

func (c *Control) createAssets(ctx context.Context, assets []*Asset, ignoreDuplicates bool) error {
	return c.DB.InTransaction(ctx, func(ctx context.Context, tx bob.Tx) error {
		for i := range assets {
			tag, err := c.TagCtrl.Get(ctx, assets[i].Tag)
			if err != nil {
				if !errors.Is(err, tags.ErrTagNotFound) {
					return err
				}
			}

			if tag != nil && tag.InUse && !ignoreDuplicates {
				return fmt.Errorf("asset with tag '%s' already exists", assets[i].Tag)
			}

			_, err = c.TagCtrl.CreateIfNotExists(ctx, assets[i].Tag)
			if err != nil {
				return err
			}

			if imgURL, err := url.Parse(assets[i].ImageURL); assets[i].ImageURL != "" && err == nil {
				file, err := c.downloadImage(ctx, imgURL)
				if err != nil {
					return err
				}

				assets[i].ImageURL = file.PublicPath
				assets[i].ThumbnailURL = assets[i].ImageURL
			} else {
				assets[i].ImageURL = ""
			}

			_, err = c.AssetRepo.Create(ctx, tx, assets[i])
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (c *Control) updateAsset(ctx context.Context, asset *Asset, file *File) (*Asset, error) {
	fileURL := asset.ImageURL

	if file != nil {
		err := c.handleFileUpload(file)
		if err != nil {
			return nil, err
		}

		fileURL = file.PublicPath
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

func (c *Control) addAssetFiles(ctx context.Context, files []*File) error {
	for _, f := range files {
		err := c.handleFileUpload(f)
		if err != nil {
			return err
		}
	}

	return c.DB.InTransaction(ctx, func(ctx context.Context, tx bob.Tx) error {
		return c.AssetRepo.CreateFiles(ctx, tx, files)
	})
}

func (c *Control) getFile(ctx context.Context, id int64) (*File, error) {
	return database.InTransaction(ctx, c.DB, func(ctx context.Context, tx bob.Tx) (*File, error) {
		return c.AssetRepo.GetFile(ctx, tx, id)
	})
}

func (c *Control) deleteFile(ctx context.Context, id int64) error {
	return c.DB.InTransaction(ctx, func(ctx context.Context, tx bob.Tx) error {
		file, err := c.AssetRepo.GetFile(ctx, tx, id)
		if err != nil {
			return err
		}

		err = c.AssetRepo.DeleteFile(ctx, tx, id)
		if err != nil {
			return err
		}

		referenceExists, err := c.AssetRepo.FileExists(ctx, tx, file.Sha256)
		if err != nil {
			return err
		}

		if !referenceExists {
			return removeFile(c.FileDir, file)
		}

		return nil
	})
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

func (c *Control) listCustomAttrNames(ctx context.Context, query ListCustomAttrNamesQuery) ([]string, error) {
	return database.InTransaction(ctx, c.DB, func(ctx context.Context, tx bob.Tx) ([]string, error) {
		return c.AssetRepo.ListCustomAttrNames(ctx, tx, query)
	})
}

func (c *Control) listLocations(ctx context.Context, query ListLocationsQuery) ([]string, error) {
	return database.InTransaction(ctx, c.DB, func(ctx context.Context, tx bob.Tx) ([]string, error) {
		return c.AssetRepo.ListLocations(ctx, tx, query)
	})
}

func (c *Control) listPositionCodes(ctx context.Context, query ListPositionCodesQuery) ([]string, error) {
	return database.InTransaction(ctx, c.DB, func(ctx context.Context, tx bob.Tx) ([]string, error) {
		return c.AssetRepo.ListPositionCodes(ctx, tx, query)
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

func (c *Control) handleFileUpload(file *File) (err error) {
	err = ensureDirExists(c.FileDir)
	if err != nil {
		return err
	}

	fhandle, err := os.CreateTemp(c.TmpDir, file.Name)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			err = errors.Join(err, fhandle.Close(), os.Remove(fhandle.Name()))
		}
	}()

	h := sha256.New()

	tee := io.TeeReader(file.r, h)

	file.SizeBytes, err = io.Copy(fhandle, tee)
	if err != nil {
		return err
	}

	err = fhandle.Close()
	if err != nil {
		return err
	}

	ext := path.Ext(file.Name)
	file.Sha256 = h.Sum(nil)
	for _, b := range file.Sha256 {
		file.FullPath = file.FullPath + "/" + fmt.Sprintf("%x", b)
	}

	file.FullPath += ext

	file.PublicPath = "/assets/files" + file.FullPath
	file.FullPath = path.Join(c.FileDir, file.FullPath)

	err = ensureDirExists(path.Dir(file.FullPath))
	if err != nil {
		return err
	}

	err = os.Rename(fhandle.Name(), file.FullPath)
	if err != nil {
		return err
	}

	return nil
}

func (c *Control) downloadImage(ctx context.Context, url *url.URL) (file *File, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		err = errors.Join(err, res.Body.Close())
	}()

	contentType := res.Header.Get("content-type")
	err = checkContentTypeAllowed(contentType, imgAllowList)
	if err != nil {
		return nil, err
	}

	origFileName := path.Base(url.Path)

	file = &File{
		Name:     origFileName,
		Filetype: contentType,
		r:        res.Body,
	}

	err = c.handleFileUpload(file)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func ensureDirExists(dir string) error {
	stat, err := os.Stat(dir)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}

	if stat == nil {
		return os.MkdirAll(dir, 0755)
	}

	if !stat.IsDir() {
		return fmt.Errorf("%s exists but is not a directory", dir)
	}

	return nil
}

func removeFile(rootDir string, file *File) error {
	rootDirIndex := strings.Index(file.FullPath, rootDir)
	if rootDirIndex == -1 {
		return fmt.Errorf("invalid file path for deletion: file path is not in configured file dir: %s", file.FullPath)
	}

	filename := path.Base(file.FullPath)

	err := os.Remove(file.FullPath)
	if err != nil {
		return err
	}

	dir := file.FullPath[:len(file.FullPath)-1-len(filename)]
	for dir != rootDir {
		isEmpty, err := isEmptyDir(dir)
		if err != nil {
			return err
		}

		if !isEmpty {
			return nil
		}

		err = os.RemoveAll(dir)
		if err != nil {
			return err
		}

		slashIndex := strings.LastIndex(dir, "/")
		if slashIndex == -1 {
			return nil
		}

		dir = dir[:slashIndex]
	}

	return nil
}

func isEmptyDir(dir string) (bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false, err
	}

	return len(entries) == 0, nil
}
