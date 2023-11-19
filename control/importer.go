package control

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"net/url"
	"path"

	"github.com/RobinThrift/stuff/entities"
	"github.com/RobinThrift/stuff/internal/importer"
	"github.com/RobinThrift/stuff/storage/database"
)

type ImporterCtrl struct {
	config ImporterCtrlConfig
	db     *database.Database
	assets *AssetControl
	tags   *TagControl
}

type ImporterCtrlConfig struct {
	DefaultCurrency string
}

func NewImporterCtrl(config ImporterCtrlConfig, db *database.Database, assets *AssetControl, tags *TagControl) *ImporterCtrl {
	return &ImporterCtrl{config: config, db: db, assets: assets, tags: tags}
}

type ImportCmd struct {
	ImportUserID     int64
	IgnoreDuplicates bool
	Format           string
	SnipeITURL       string
	SnipeITAPIKey    string
}

func (ic *ImporterCtrl) Import(r *http.Request, cmd ImportCmd) (map[string]string, error) {
	var assets []*entities.Asset
	var validationErrs map[string]string
	var err error

	switch cmd.Format {
	case "snipeit_json_export":
		assets, validationErrs, err = importer.ImportFromSnipeITJSONExport(r)
	case "snipeit_api":
		assets, err = importer.ImportFromSnipeITAPI(r.Context(), cmd.SnipeITURL, cmd.SnipeITAPIKey)
	default:
		errMsg := fmt.Sprintf("unknown format '%s'", cmd.Format)
		slog.ErrorContext(r.Context(), "error importing assets", "error", errMsg)
		return map[string]string{"format": errMsg}, nil
	}

	if err != nil {
		return validationErrs, err
	}

	err = ic.db.InTransaction(r.Context(), func(ctx context.Context, tx database.Executor) error {
		return ic.createAssets(ctx, assets, cmd)
	})
	if err != nil {
		return nil, err
	}

	return validationErrs, nil
}

func (ic *ImporterCtrl) createAssets(ctx context.Context, assets []*entities.Asset, cmd ImportCmd) error {
	for i := range assets {
		tag, err := ic.tags.Get(ctx, assets[i].Tag)
		if err != nil {
			if !errors.Is(err, entities.ErrTagNotFound) {
				return err
			}
		}

		if tag != nil && tag.InUse && !cmd.IgnoreDuplicates {
			return fmt.Errorf("asset with tag '%s' already exists", assets[i].Tag)
		}

		err = ic.createAsset(ctx, assets[i], cmd)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ic *ImporterCtrl) createAsset(ctx context.Context, asset *entities.Asset, cmd ImportCmd) error {
	for j := range asset.Purchases {
		if asset.Purchases[j].Amount != 0 && asset.Purchases[j].Currency == "" {
			asset.Purchases[j].Currency = ic.config.DefaultCurrency
		}
	}

	if asset.Tag == "" {
		tag, err := ic.tags.GetNext(ctx)
		if err != nil {
			return err
		}
		asset.Tag = tag
	}

	if _, err := ic.tags.CreateIfNotExists(ctx, asset.Tag); err != nil {
		return err
	}

	var imgFile *entities.File
	if imgURL, err := url.Parse(asset.ImageURL); asset.ImageURL != "" && err == nil {
		imgFile, err = downloadImage(ctx, imgURL)
		if err != nil {
			return err
		}

		asset.ImageURL = imgFile.PublicPath
		asset.ThumbnailURL = asset.ImageURL
	} else {
		asset.ImageURL = ""
	}

	asset.MetaInfo.CreatedBy = cmd.ImportUserID

	_, err := ic.assets.Create(ctx, CreateAssetCmd{
		Asset: asset,
		Image: imgFile,
	})
	if err != nil {
		return err
	}

	return nil
}

func downloadImage(ctx context.Context, url *url.URL) (file *entities.File, err error) {
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

	var content bytes.Buffer

	_, err = io.Copy(&content, res.Body)
	if err != nil {
		return nil, err
	}

	origFileName := path.Base(url.Path)

	return &entities.File{
		Reader:   &content,
		Name:     origFileName,
		Filetype: contentType,
	}, nil
}

var imgAllowList = []string{"image/png", "image/jpeg", "image/webp"}

func checkContentTypeAllowed(ct string, allowlist []string) error {
	mt, _, err := mime.ParseMediaType(ct)
	if err != nil {
		return err
	}

	allowed := false
	for _, m := range allowlist {
		if mt == m {
			allowed = true
			break
		}
	}

	if !allowed {
		return fmt.Errorf("file type not allowed: %s", mt)
	}

	return nil

}
