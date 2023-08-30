package sqlite

import (
	"context"
	"database/sql"
	"errors"

	"github.com/aarondl/opt/omit"
	"github.com/kodeshack/stuff/storage/database"
	"github.com/kodeshack/stuff/storage/database/sqlite/models"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/sqlite"
	"github.com/stephenafamo/bob/dialect/sqlite/sm"
	"github.com/stephenafamo/scan"
)

type TagRepo struct{}

func (*TagRepo) Get(ctx context.Context, exec bob.Executor, tag string) (*database.Tag, error) {
	model, err := models.Tags.Query(ctx, exec, models.SelectWhere.Tags.Tag.EQ(tag)).One()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, database.ErrTagNotFound
		}
		return nil, err
	}

	return &database.Tag{
		ID:        model.ID,
		Tag:       model.Tag,
		CreatedAt: model.CreatedAt.Time,
		UpdatedAt: model.UpdatedAt.Time,
	}, nil
}

func (*TagRepo) Create(ctx context.Context, exec bob.Executor, tag *database.Tag) (*database.Tag, error) {
	model := &models.TagSetter{
		Tag: omit.From(tag.Tag),
	}

	inserted, err := models.Tags.Insert(ctx, exec, model)
	if err != nil {
		return nil, err
	}

	return &database.Tag{
		ID:        inserted.ID,
		Tag:       inserted.Tag,
		CreatedAt: inserted.CreatedAt.Time,
		UpdatedAt: inserted.UpdatedAt.Time,
	}, nil
}

func (*TagRepo) Delete(ctx context.Context, exec bob.Executor, tag string) error {
	_, err := models.Tags.DeleteQ(ctx, exec, models.DeleteWhere.Tags.Tag.EQ(tag)).Exec()
	return err
}

func (*TagRepo) NextSequential(ctx context.Context, exec bob.Executor) (int64, error) {
	// SELECT seq FROM sqlite_sequence WHERE name = 'tags' LIMIT 1;
	next, err := bob.One(ctx, exec,
		sqlite.Select(
			sm.Columns(sqlite.Quote("seq")),
			sm.From(sqlite.Quote("sqlite_sequence")),
			sm.Where(sqlite.Quote("name").EQ(sqlite.Quote(models.TableNames.Tags))),
		),
		scan.SingleColumnMapper[int64],
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 1, nil
		}
		return 0, err
	}

	return next + 1, nil
}
