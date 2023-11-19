package control

import (
	"context"
	"errors"
	"fmt"
	"path"

	"github.com/RobinThrift/stuff/entities"
	"github.com/RobinThrift/stuff/storage/database"
	"github.com/RobinThrift/stuff/storage/database/sqlite"
	"github.com/stephenafamo/bob"
)

var ErrAssetNotFound = errors.New("asset not found")
var ErrAssetMissingTag = errors.New("asset is missing a tag")
var ErrDeleteAsset = errors.New("error deleting asset")

type AssetControl struct {
	db *database.Database

	tags  *TagControl
	files *FileControl

	repo AssetRepo
}

type AssetRepo interface {
	Get(ctx context.Context, exec bob.Executor, query database.GetAssetQuery) (*entities.Asset, error)
	List(ctx context.Context, exec bob.Executor, query database.ListAssetsQuery) (*entities.ListPage[*entities.Asset], error)
	Create(ctx context.Context, exec bob.Executor, asset *entities.Asset) error
	Update(ctx context.Context, exec bob.Executor, asset *entities.Asset) error
	Delete(ctx context.Context, exec bob.Executor, id int64) error
}

func NewAssetControl(db *database.Database, tags *TagControl, files *FileControl, repo AssetRepo) *AssetControl {
	return &AssetControl{db: db, tags: tags, files: files, repo: repo}
}

type GetAssetQuery struct {
	ID               int64
	Tag              string
	IncludePurchases bool
	IncludeParts     bool
	IncludeFiles     bool
	IncludeParent    bool
	IncludeChildren  bool
}

func (ac *AssetControl) Get(ctx context.Context, query GetAssetQuery) (*entities.Asset, error) {
	return database.InTransaction(ctx, ac.db, func(ctx context.Context, tx database.Executor) (*entities.Asset, error) {
		asset, err := ac.repo.Get(ctx, tx, database.GetAssetQuery{
			ID:               query.ID,
			Tag:              query.Tag,
			IncludePurchases: query.IncludePurchases,
			IncludeParts:     query.IncludeParts,
			IncludeFiles:     query.IncludeFiles,
			IncludeParent:    query.IncludeParent,
			IncludeChildren:  query.IncludeChildren,
		})
		if err != nil {
			if errors.Is(err, sqlite.ErrAssetNotFound) {
				return nil, fmt.Errorf("%w: %d/%s", ErrAssetNotFound, query.ID, query.Tag)
			}
			return nil, err
		}

		return asset, nil
	})
}

type ListAssetsQuery struct {
	SearchRaw    string
	SearchFields map[string]string

	IDs []int64

	Page     int
	PageSize int

	OrderBy  string
	OrderDir string

	AssetType entities.AssetType

	IncludeParts bool
}

func (ac *AssetControl) List(ctx context.Context, query ListAssetsQuery) (*entities.ListPage[*entities.Asset], error) {
	return database.InTransaction(ctx, ac.db, func(ctx context.Context, tx database.Executor) (*entities.ListPage[*entities.Asset], error) {
		return ac.repo.List(ctx, tx, database.ListAssetsQuery{
			SearchRaw:    query.SearchRaw,
			SearchFields: query.SearchFields,
			IDs:          query.IDs,
			Page:         query.Page,
			PageSize:     query.PageSize,
			OrderBy:      query.OrderBy,
			OrderDir:     query.OrderDir,
			AssetType:    string(query.AssetType),
			IncludeParts: query.IncludeParts,
		})
	})
}

type CreateAssetCmd struct {
	Asset *entities.Asset
	Image *entities.File
}

func (ac *AssetControl) Create(ctx context.Context, cmd CreateAssetCmd) (*entities.Asset, error) {
	if cmd.Asset.Tag == "" {
		return nil, ErrAssetMissingTag
	}

	return database.InTransaction(ctx, ac.db, func(ctx context.Context, tx database.Executor) (*entities.Asset, error) {
		return ac.create(ctx, tx, cmd)
	})
}

func (ac *AssetControl) create(ctx context.Context, exec bob.Executor, cmd CreateAssetCmd) (*entities.Asset, error) {
	var err error

	_, err = ac.tags.CreateIfNotExists(ctx, cmd.Asset.Tag)
	if err != nil {
		return nil, err
	}

	err = ac.repo.Create(ctx, exec, cmd.Asset)
	if err != nil {
		return nil, err
	}

	if cmd.Image != nil {
		cmd.Image.AssetID = cmd.Asset.ID
		cmd.Image.Name = cmd.Asset.Tag + "_image" + path.Ext(cmd.Image.Name)
		cmd.Image, err = ac.files.WriteFile(ctx, cmd.Image)
		if err != nil {
			return nil, err
		}

		cmd.Asset.ImageURL = cmd.Image.PublicPath
		cmd.Asset.ThumbnailURL = cmd.Image.PublicPath
	}

	err = ac.repo.Update(ctx, exec, cmd.Asset)
	if err != nil {
		return nil, err
	}

	return ac.repo.Get(ctx, exec, database.GetAssetQuery{Tag: cmd.Asset.Tag, IncludePurchases: true, IncludeParts: true})
}

type UpdateAssetCmd struct {
	Asset *entities.Asset
	Image *entities.File
}

func (ac *AssetControl) Update(ctx context.Context, cmd UpdateAssetCmd) (*entities.Asset, error) {
	return database.InTransaction(ctx, ac.db, func(ctx context.Context, tx database.Executor) (*entities.Asset, error) {
		return ac.update(ctx, tx, cmd)
	})
}

func (ac *AssetControl) update(ctx context.Context, exec bob.Executor, cmd UpdateAssetCmd) (*entities.Asset, error) {
	imgURL := cmd.Asset.ImageURL
	thmbURL := cmd.Asset.ThumbnailURL

	var err error
	if cmd.Image != nil {
		cmd.Image.AssetID = cmd.Asset.ID
		cmd.Image.Name = cmd.Asset.Tag + "_image" + path.Ext(cmd.Image.Name)
		cmd.Image, err = ac.files.WriteFile(ctx, cmd.Image)
		if err != nil {
			return nil, fmt.Errorf("error writing image file for asset %v: %w", cmd.Asset.Tag, err)
		}

		cmd.Asset.ImageURL = cmd.Image.PublicPath
		cmd.Asset.ThumbnailURL = cmd.Image.PublicPath
	}

	err = ac.repo.Update(ctx, exec, cmd.Asset)
	if err != nil {
		return nil, fmt.Errorf("error updating asset %s in database: %w", cmd.Asset.Tag, err)
	}

	if imgURL != "" && cmd.Image != nil {
		err := ac.files.DeleteByPublicPath(ctx, imgURL)
		if err != nil {
			return nil, fmt.Errorf("error deleting old image for asset %s: %w", cmd.Asset.Tag, err)
		}

		if imgURL != thmbURL {
			err := ac.files.DeleteByPublicPath(ctx, thmbURL)
			if err != nil {
				return nil, fmt.Errorf("error deleting old thumbnail for asset %s: %w", cmd.Asset.Tag, err)
			}
		}
	}

	return ac.repo.Get(ctx, exec, database.GetAssetQuery{ID: cmd.Asset.ID, IncludePurchases: true, IncludeParts: true, IncludeChildren: true})
}

func (ac *AssetControl) Delete(ctx context.Context, asset *entities.Asset) error {
	return ac.db.InTransaction(ctx, func(ctx context.Context, tx database.Executor) error {
		return ac.delete(ctx, tx, asset)
	})
}

func (ac *AssetControl) delete(ctx context.Context, exec bob.Executor, asset *entities.Asset) error {
	err := ac.tags.MarkTagUnused(ctx, asset.Tag)
	if err != nil {
		return fmt.Errorf("%w: error marking tag as unused: %w", ErrDeleteAsset, err)
	}

	err = ac.files.DeleteAllForAsset(ctx, asset.ID)
	if err != nil {
		return fmt.Errorf("%w: error deleting asset files: %w", ErrDeleteAsset, err)
	}

	err = ac.repo.Delete(ctx, exec, asset.ID)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrDeleteAsset, err)
	}

	return nil
}
