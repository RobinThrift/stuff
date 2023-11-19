package control

import (
	"context"
	"errors"
	"fmt"

	"github.com/RobinThrift/stuff/entities"
	"github.com/RobinThrift/stuff/storage/database"
	"github.com/google/uuid"
	nanoid "github.com/matoous/go-nanoid/v2"
	"github.com/segmentio/ksuid"
	"github.com/stephenafamo/bob"
)

var ErrInvalidTag = errors.New("invalid tag")

type TagControl struct {
	algorithm string
	db        *database.Database
	repo      TagRepo
}

type TagRepo interface {
	List(ctx context.Context, exec bob.Executor, query database.ListTagsQuery) (*entities.ListPage[*entities.Tag], error)
	GetUnused(ctx context.Context, exec bob.Executor) (*entities.Tag, error)
	Get(ctx context.Context, exec bob.Executor, tag string) (*entities.Tag, error)
	Create(ctx context.Context, exec bob.Executor, tag *entities.Tag) error
	MarkTagUsed(ctx context.Context, exec bob.Executor, tag string) error
	MarkTagUnused(ctx context.Context, exec bob.Executor, tag string) error
	Delete(ctx context.Context, exec bob.Executor, tag string) error
	NextSequential(ctx context.Context, exec bob.Executor) (int64, error)
}

func NewTagControl(db *database.Database, algorithm string, repo TagRepo) *TagControl {
	return &TagControl{
		algorithm: algorithm,
		db:        db,
		repo:      repo,
	}
}

type ListTagsQuery struct {
	Search   string
	InUse    *bool
	Page     int
	PageSize int
	OrderBy  string
	OrderDir string
}

func (tc *TagControl) List(ctx context.Context, query ListTagsQuery) (*entities.ListPage[*entities.Tag], error) {
	return database.InTransaction(ctx, tc.db, func(ctx context.Context, tx database.Executor) (*entities.ListPage[*entities.Tag], error) {
		return tc.repo.List(ctx, tx, database.ListTagsQuery(query))
	})
}

func (tc *TagControl) GetNext(ctx context.Context) (string, error) {
	return database.InTransaction(ctx, tc.db, func(ctx context.Context, tx database.Executor) (string, error) {
		unused, err := tc.repo.GetUnused(ctx, tx)
		if err != nil {
			return "", err
		}

		if unused != nil {
			return unused.Tag, nil
		}

		switch tc.algorithm {
		case "nanoid":
			return nanoid.Generate("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ", 6)
		case "ksuid":
			return ksuid.New().String(), nil
		case "uuid":
			return uuid.NewString(), nil
		case "sequential":
			return tc.generateSequential(ctx)
		}

		return "", fmt.Errorf("unknown tag algorithm: %s", tc.algorithm)
	})
}

func (tc *TagControl) Get(ctx context.Context, tag string) (*entities.Tag, error) {
	return database.InTransaction(ctx, tc.db, func(ctx context.Context, tx database.Executor) (*entities.Tag, error) {
		tag, err := tc.repo.Get(ctx, tx, tag)
		if err != nil {
			if !errors.Is(err, entities.ErrTagNotFound) {
				return nil, err
			}
		}

		return tag, nil
	})
}

func (tc *TagControl) CreateIfNotExists(ctx context.Context, tag string) (*entities.Tag, error) {
	if tag == "" {
		return nil, fmt.Errorf("%w: '%s'", ErrInvalidTag, tag)
	}

	return database.InTransaction(ctx, tc.db, func(ctx context.Context, tx database.Executor) (*entities.Tag, error) {
		found, err := tc.repo.Get(ctx, tx, tag)
		if err != nil {
			if !errors.Is(err, entities.ErrTagNotFound) {
				return nil, err
			}
		}

		if found != nil {
			err = tc.repo.MarkTagUsed(ctx, tx, tag)
			if err != nil {
				return nil, err
			}

			return &entities.Tag{
				ID:        found.ID,
				Tag:       found.Tag,
				InUse:     true,
				CreatedAt: found.CreatedAt,
				UpdatedAt: found.UpdatedAt,
			}, nil
		}

		err = tc.repo.Create(ctx, tx, &entities.Tag{Tag: tag, InUse: true})
		if err != nil {
			return nil, err
		}

		created, err := tc.repo.Get(ctx, tx, tag)
		if err != nil {
			return nil, err
		}

		return &entities.Tag{
			ID:        created.ID,
			Tag:       created.Tag,
			InUse:     true,
			CreatedAt: created.CreatedAt,
			UpdatedAt: created.UpdatedAt,
		}, nil
	})
}

func (tc *TagControl) MarkTagUnused(ctx context.Context, tag string) error {
	return tc.db.InTransaction(ctx, func(ctx context.Context, tx database.Executor) error {
		return tc.repo.MarkTagUnused(ctx, tx, tag)
	})
}

func (tc *TagControl) Delete(ctx context.Context, tag string) error {
	return tc.db.InTransaction(ctx, func(ctx context.Context, tx database.Executor) error {
		return tc.repo.Delete(ctx, tx, tag)
	})
}

func (tc *TagControl) generateSequential(ctx context.Context) (string, error) {
	next, err := database.InTransaction(ctx, tc.db, func(ctx context.Context, tx database.Executor) (int64, error) {
		return tc.repo.NextSequential(ctx, tx)
	})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%0.6d", next), nil
}
