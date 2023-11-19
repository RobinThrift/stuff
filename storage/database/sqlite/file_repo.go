package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/RobinThrift/stuff/entities"
	"github.com/RobinThrift/stuff/storage/database"
	"github.com/RobinThrift/stuff/storage/database/sqlite/models"
	"github.com/aarondl/opt/omit"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/sqlite/dialect"
	"github.com/stephenafamo/bob/dialect/sqlite/sm"
)

var ErrFileNotFound = errors.New("file not found")
var ErrCreatingFile = errors.New("error creating file")

type FileRepo struct{}

func (fr *FileRepo) Create(ctx context.Context, exec bob.Executor, file *entities.File) (int64, error) {
	inserted, err := models.AssetFiles.Insert(ctx, exec, &models.AssetFileSetter{
		AssetID:    omit.From(file.AssetID),
		Name:       omit.From(file.Name),
		Filetype:   omit.From(file.Filetype),
		Sha256:     omit.From(file.Sha256),
		SizeBytes:  omit.From(file.SizeBytes),
		CreatedBy:  omit.From(file.CreatedBy),
		FullPath:   omit.From(file.FullPath),
		PublicPath: omit.From(file.PublicPath),
	})
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrCreatingFile, unwapSQLiteError(err))
	}
	return inserted.ID, nil
}

func (fr *FileRepo) Get(ctx context.Context, exec bob.Executor, id int64) (*entities.File, error) {
	file, err := models.AssetFiles.Query(ctx, exec, models.SelectWhere.AssetFiles.ID.EQ(id)).One()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: %d", ErrFileNotFound, id)
		}
		return nil, fmt.Errorf("error getting asset file: %w", err)
	}

	return &entities.File{
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

func (fr *FileRepo) GetByPublicPath(ctx context.Context, exec bob.Executor, publicPath string) (*entities.File, error) {
	file, err := models.AssetFiles.Query(ctx, exec, models.SelectWhere.AssetFiles.PublicPath.EQ(publicPath)).One()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: public path: %s", ErrFileNotFound, publicPath)
		}
		return nil, fmt.Errorf("error getting asset file: %w", err)
	}

	return &entities.File{
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

func (fr *FileRepo) List(ctx context.Context, exec bob.Executor, query database.ListFilesQuery) (*entities.ListPage[*entities.File], error) {
	limit := query.PageSize

	if limit == 0 {
		limit = 50
	}

	if limit > 100 {
		limit = 100
	}

	offset := limit * query.Page

	mods := []bob.Mod[*dialect.SelectQuery]{
		sm.Limit(limit),
		sm.Offset(offset),
	}

	if query.AssetID != 0 {
		mods = append(mods, models.SelectWhere.AssetFiles.AssetID.EQ(query.AssetID))
	}

	if len(query.Hashes) != 0 {
		mods = append(mods, models.SelectWhere.AssetFiles.Sha256.In(query.Hashes...))
	}

	files, err := models.AssetFiles.Query(ctx, exec, mods...).All()
	if err != nil {
		return nil, fmt.Errorf("error getting files: %w", err)
	}

	count, err := models.AssetFiles.Query(ctx, exec, mods...).Count()
	if err != nil {
		return nil, fmt.Errorf("error counting files: %w", err)
	}

	pageSize := query.PageSize
	if pageSize <= 0 {
		pageSize = 1
	}

	page := &entities.ListPage[*entities.File]{
		Items:    make([]*entities.File, 0, len(files)),
		Total:    int(count),
		Page:     query.Page,
		PageSize: query.PageSize,
		NumPages: int(count) / pageSize,
	}

	for i := range files {
		page.Items = append(page.Items, &entities.File{
			ID:         files[i].ID,
			AssetID:    files[i].AssetID,
			Name:       files[i].Name,
			Filetype:   files[i].Filetype,
			SizeBytes:  files[i].SizeBytes,
			PublicPath: files[i].PublicPath,
			FullPath:   files[i].FullPath,
			Sha256:     files[i].Sha256,
			CreatedBy:  files[i].CreatedBy,
			CreatedAt:  files[i].CreatedAt.Time,
			UpdatedAt:  files[i].UpdatedAt.Time,
		})
	}

	return page, nil
}

func (fr *FileRepo) Delete(ctx context.Context, exec bob.Executor, ids []int64) error {
	_, err := models.AssetFiles.DeleteQ(ctx, exec, models.DeleteWhere.AssetFiles.ID.In(ids...)).Exec()
	if err != nil {
		return fmt.Errorf("error deleting asset files: %w", err)
	}

	return nil
}
