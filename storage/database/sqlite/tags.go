package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/aarondl/opt/omit"
	"github.com/kodeshack/stuff/storage/database"
	"github.com/kodeshack/stuff/storage/database/sqlite/models"
	"github.com/kodeshack/stuff/storage/database/sqlite/types"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/sqlite"
	"github.com/stephenafamo/bob/dialect/sqlite/dialect"
	"github.com/stephenafamo/bob/dialect/sqlite/sm"
	bmods "github.com/stephenafamo/bob/mods"
	"github.com/stephenafamo/scan"
)

type TagRepo struct{}

func (*TagRepo) List(ctx context.Context, exec bob.Executor, query database.ListTagsQuery) (*database.TagList, error) {
	if query.Limit == 0 {
		query.Limit = 50
	}

	mods := []bob.Mod[*dialect.SelectQuery]{
		sm.Limit(query.Limit),
		sm.Offset(query.Offset),
	}

	if query.OrderBy != "" {
		if query.OrderDir == "" {
			query.OrderDir = "ASC"
		}

		mods = append(mods, bmods.OrderBy[*dialect.SelectQuery]{
			Expression: query.OrderBy,
			Direction:  query.OrderDir,
		})
	}

	tags, err := models.Tags.Query(ctx, exec, mods...).All()
	if err != nil {
		return nil, fmt.Errorf("error getting tags: %w", err)
	}

	count, err := models.Tags.Query(ctx, exec, mods...).Count()
	if err != nil {
		return nil, fmt.Errorf("error counting tags: %w", err)
	}

	tagList := &database.TagList{
		Tags:  make([]*database.Tag, 0, len(tags)),
		Total: int(count),
	}

	for i := range tags {
		tagList.Tags = append(tagList.Tags, &database.Tag{
			ID:        tags[i].ID,
			Tag:       tags[i].Tag,
			InUse:     tags[i].InUse,
			CreatedAt: tags[i].CreatedAt.Time,
			UpdatedAt: tags[i].UpdatedAt.Time,
		})
	}

	return tagList, nil
}

func (*TagRepo) GetUnused(ctx context.Context, exec bob.Executor) (*database.Tag, error) {
	model, err := models.Tags.Query(ctx, exec, models.SelectWhere.Tags.InUse.EQ(false)).One()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
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
		Tag:   omit.From(tag.Tag),
		InUse: omit.From(true),
	}

	inserted, err := models.Tags.Insert(ctx, exec, model)
	if err != nil {
		return nil, err
	}

	return &database.Tag{
		ID:        inserted.ID,
		Tag:       inserted.Tag,
		InUse:     inserted.InUse,
		CreatedAt: inserted.CreatedAt.Time,
		UpdatedAt: inserted.UpdatedAt.Time,
	}, nil
}

func (*TagRepo) MarkTagUsed(ctx context.Context, exec bob.Executor, tag string) error {
	setter := models.TagSetter{
		InUse:     omit.From(true),
		UpdatedAt: omit.From(types.NewSQLiteDatetime(time.Now())),
	}

	_, err := models.Tags.UpdateQ(ctx, exec, models.UpdateWhere.Tags.Tag.EQ(tag), setter).Exec()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return database.ErrTagNotFound
		}
		return err
	}

	return nil
}

func (*TagRepo) MarkTagUnused(ctx context.Context, exec bob.Executor, tag string) error {
	setter := models.TagSetter{
		InUse:     omit.From(false),
		UpdatedAt: omit.From(types.NewSQLiteDatetime(time.Now())),
	}

	_, err := models.Tags.UpdateQ(ctx, exec, models.UpdateWhere.Tags.Tag.EQ(tag), setter).Exec()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return database.ErrTagNotFound
		}
		return err
	}

	return nil
}

func (*TagRepo) Delete(ctx context.Context, exec bob.Executor, tag string) error {
	_, err := models.Tags.DeleteQ(ctx, exec, models.DeleteWhere.Tags.Tag.EQ(tag), models.DeleteWhere.Tags.InUse.EQ(false)).Exec()
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
