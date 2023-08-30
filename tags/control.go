package tags

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/kodeshack/stuff/storage/database"
	nanoid "github.com/matoous/go-nanoid/v2"
	"github.com/segmentio/ksuid"
	"github.com/stephenafamo/bob"
)

type Control struct {
	Algorithm string
	DB        *database.Database
	TagRepo   TagRepo
}

type TagRepo interface {
	Get(ctx context.Context, exec bob.Executor, tag string) (*database.Tag, error)
	Create(ctx context.Context, exec bob.Executor, tag *database.Tag) (*database.Tag, error)
	NextSequential(ctx context.Context, exec bob.Executor) (int64, error)
	Delete(ctx context.Context, exec bob.Executor, tag string) error
}

func (c *Control) Generate(ctx context.Context) (string, error) {
	switch c.Algorithm {
	case "nanoid":
		return nanoid.Generate("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ", 6)
	case "ksuid":
		return ksuid.New().String(), nil
	case "uuid":
		return uuid.NewString(), nil
	case "sequential":
		return c.generateSequential(ctx)
	}

	return "", fmt.Errorf("unknown tag algorithm: %s", c.Algorithm)
}

func (c *Control) generateSequential(ctx context.Context) (string, error) {
	next, err := database.InTransaction(ctx, c.DB, func(ctx context.Context, tx bob.Tx) (int64, error) {
		return c.TagRepo.NextSequential(ctx, tx)
	})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%0.6d", next), nil
}

func (c *Control) CreateIfNotExists(ctx context.Context, tag string) (*Tag, error) {
	return database.InTransaction(ctx, c.DB, func(ctx context.Context, tx bob.Tx) (*Tag, error) {
		found, err := c.TagRepo.Get(ctx, tx, tag)
		if err != nil {
			if !errors.Is(err, database.ErrTagNotFound) {
				return nil, err
			}
		}

		if found != nil {
			return &Tag{
				ID:        found.ID,
				Tag:       found.Tag,
				CreatedAt: found.CreatedAt,
				UpdatedAt: found.UpdatedAt,
			}, nil
		}

		created, err := c.TagRepo.Create(ctx, tx, &database.Tag{Tag: tag})
		if err != nil {
			return nil, err
		}

		return &Tag{
			ID:        created.ID,
			Tag:       created.Tag,
			CreatedAt: created.CreatedAt,
			UpdatedAt: created.UpdatedAt,
		}, nil
	})
}

func (c *Control) Delete(ctx context.Context, exec bob.Executor, tag string) error {
	return c.DB.InTransaction(ctx, func(ctx context.Context, tx bob.Tx) error {
		return c.TagRepo.Delete(ctx, tx, tag)
	})

}
