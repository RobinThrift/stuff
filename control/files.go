package control

import (
	"context"
	"errors"
	"fmt"

	"github.com/RobinThrift/stuff/entities"
	"github.com/RobinThrift/stuff/storage/database"
	"github.com/RobinThrift/stuff/storage/database/sqlite"
	"github.com/stephenafamo/bob"
)

var ErrFileNotFound = errors.New("file not found")

type FileControl struct {
	db *database.Database

	repo FileRepo

	blobs FileBlobs
}

type FileRepo interface {
	Get(ctx context.Context, exec bob.Executor, id int64) (*entities.File, error)
	GetByPublicPath(ctx context.Context, exec bob.Executor, publicPath string) (*entities.File, error)
	List(ctx context.Context, exec bob.Executor, query database.ListFilesQuery) (*entities.ListPage[*entities.File], error)
	Create(ctx context.Context, exec bob.Executor, file *entities.File) (int64, error)
	Delete(ctx context.Context, exec bob.Executor, ids []int64) error
}

type FileBlobs interface {
	WriteFile(*entities.File) error
	RemoveFile(*entities.File) error
}

func NewFileControl(db *database.Database, repo FileRepo, blobs FileBlobs) *FileControl {
	return &FileControl{
		db:    db,
		repo:  repo,
		blobs: blobs,
	}
}

func (fc *FileControl) Get(ctx context.Context, id int64) (*entities.File, error) {
	return database.InTransaction(ctx, fc.db, func(ctx context.Context, tx database.Executor) (*entities.File, error) {
		file, err := fc.repo.Get(ctx, tx, id)
		if err != nil {
			if errors.Is(err, sqlite.ErrFileNotFound) {
				return nil, fmt.Errorf("%w: %d", ErrFileNotFound, id)
			}
			return nil, err
		}
		return file, nil
	})
}

type ListFilesQuery struct {
	AssetID  int64
	Page     int
	PageSize int
}

func (fc *FileControl) List(ctx context.Context, query ListFilesQuery) (*entities.ListPage[*entities.File], error) {
	return database.InTransaction(ctx, fc.db, func(ctx context.Context, tx database.Executor) (*entities.ListPage[*entities.File], error) {
		return fc.repo.List(ctx, tx, database.ListFilesQuery{
			AssetID:  query.AssetID,
			Page:     query.Page,
			PageSize: query.PageSize,
		})
	})
}

func (fc *FileControl) WriteFile(ctx context.Context, file *entities.File) (*entities.File, error) {
	return database.InTransaction(ctx, fc.db, func(ctx context.Context, tx database.Executor) (*entities.File, error) {
		return fc.writeFile(ctx, tx, file)
	})
}

func (fc *FileControl) writeFile(ctx context.Context, exec bob.Executor, file *entities.File) (*entities.File, error) {
	err := fc.blobs.WriteFile(file)
	if err != nil {
		return nil, err
	}

	createdID, err := fc.repo.Create(ctx, exec, file)
	if err != nil {
		return nil, errors.Join(err, fc.blobs.RemoveFile(file))
	}

	created, err := fc.repo.Get(ctx, exec, createdID)
	if err != nil {
		return nil, errors.Join(err, fc.blobs.RemoveFile(file))
	}

	return created, nil
}

func (fc *FileControl) DeleteByPublicPath(ctx context.Context, publicPath string) error {
	return fc.db.InTransaction(ctx, func(ctx context.Context, tx database.Executor) error {
		file, err := fc.repo.GetByPublicPath(ctx, tx, publicPath)
		if err != nil {
			if errors.Is(err, sqlite.ErrFileNotFound) {
				return nil
			}
			return err
		}

		return fc.Delete(ctx, file.ID)
	})
}

func (fc *FileControl) Delete(ctx context.Context, id int64) error {
	return fc.db.InTransaction(ctx, func(ctx context.Context, tx database.Executor) error {
		file, err := fc.repo.Get(ctx, tx, id)
		if err != nil {
			return err
		}

		list, err := fc.repo.List(ctx, tx, database.ListFilesQuery{Hashes: [][]byte{file.Sha256}})
		if err != nil {
			return err
		}

		if len(list.Items) == 1 {
			err = fc.blobs.RemoveFile(file)
			if err != nil {
				return err
			}
		}

		return fc.repo.Delete(ctx, tx, []int64{id})
	})
}

func (fc *FileControl) DeleteAllForAsset(ctx context.Context, assetID int64) error {
	return fc.db.InTransaction(ctx, func(ctx context.Context, tx database.Executor) error {
		return fc.deleteAllForAsset(ctx, tx, assetID)
	})
}

func (fc *FileControl) deleteAllForAsset(ctx context.Context, exec bob.Executor, assetID int64) error {
	list, err := fc.repo.List(ctx, exec, database.ListFilesQuery{AssetID: assetID})
	if err != nil {
		return err
	}
	files := list.Items

	filesToDelete := make(map[string]*entities.File, len(files))

	fileIDs := make([]int64, 0, len(filesToDelete))
	hashes := make([][]byte, len(files))
	for i := range files {
		fileIDs = append(fileIDs, files[i].ID)
		hashes[i] = files[i].Sha256
		filesToDelete[fmt.Sprintf("%x", files[i].Sha256)] = files[i]
	}

	err = fc.repo.Delete(ctx, exec, fileIDs)
	if err != nil {
		return err
	}

	list, err = fc.repo.List(ctx, exec, database.ListFilesQuery{Hashes: hashes})
	if err != nil {
		return err
	}
	filesForHashes := list.Items

	for i := range filesForHashes {
		hash := fmt.Sprintf("%x", filesForHashes[i].Sha256)
		delete(filesToDelete, hash)
	}

	for _, file := range filesToDelete {
		err = fc.blobs.RemoveFile(file)
		if err != nil {
			return err
		}
	}

	return nil
}
