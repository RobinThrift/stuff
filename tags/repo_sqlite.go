package tags

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/RobinThrift/stuff/storage/database/sqlite/models"
	"github.com/RobinThrift/stuff/storage/database/sqlite/types"
	"github.com/aarondl/opt/omit"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/sqlite"
	"github.com/stephenafamo/bob/dialect/sqlite/dialect"
	"github.com/stephenafamo/bob/dialect/sqlite/sm"
	bmods "github.com/stephenafamo/bob/mods"
	"github.com/stephenafamo/scan"
)

type RepoSQLite struct{}

func (*RepoSQLite) List(ctx context.Context, exec bob.Executor, query ListTagsQuery) (*TagListPage, error) {
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

	if query.OrderBy != "" {
		if query.OrderDir == "" {
			query.OrderDir = "ASC"
		}

		mods = append(mods, bmods.OrderBy[*dialect.SelectQuery]{
			Expression: query.OrderBy,
			Direction:  query.OrderDir,
		})
	}

	if query.Search != "" {
		mods = append(mods, models.SelectWhere.Tags.Tag.Like("%"+query.Search+"%"))
	}

	if query.InUse != nil {
		mods = append(mods, models.SelectWhere.Tags.InUse.EQ(*query.InUse))
	}

	tags, err := models.Tags.Query(ctx, exec, mods...).All()
	if err != nil {
		return nil, fmt.Errorf("error getting tags: %w", err)
	}

	count, err := models.Tags.Query(ctx, exec, mods...).Count()
	if err != nil {
		return nil, fmt.Errorf("error counting tags: %w", err)
	}

	page := &TagListPage{
		Tags:     make([]*Tag, 0, len(tags)),
		Total:    int(count),
		Page:     query.Page,
		PageSize: query.PageSize,
		NumPages: int(count) / query.PageSize,
	}

	for i := range tags {
		page.Tags = append(page.Tags, &Tag{
			ID:        tags[i].ID,
			Tag:       tags[i].Tag,
			InUse:     tags[i].InUse,
			CreatedAt: tags[i].CreatedAt.Time,
			UpdatedAt: tags[i].UpdatedAt.Time,
		})
	}

	return page, nil
}

func (*RepoSQLite) GetUnused(ctx context.Context, exec bob.Executor) (*Tag, error) {
	model, err := models.Tags.Query(ctx, exec, models.SelectWhere.Tags.InUse.EQ(false)).One()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &Tag{
		ID:        model.ID,
		Tag:       model.Tag,
		CreatedAt: model.CreatedAt.Time,
		UpdatedAt: model.UpdatedAt.Time,
	}, nil
}

func (*RepoSQLite) Get(ctx context.Context, exec bob.Executor, tag string) (*Tag, error) {
	model, err := models.Tags.Query(ctx, exec, models.SelectWhere.Tags.Tag.EQ(tag)).One()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTagNotFound
		}
		return nil, err
	}

	return &Tag{
		ID:        model.ID,
		Tag:       model.Tag,
		CreatedAt: model.CreatedAt.Time,
		UpdatedAt: model.UpdatedAt.Time,
	}, nil
}

func (*RepoSQLite) Create(ctx context.Context, exec bob.Executor, tag *Tag) (*Tag, error) {
	model := &models.TagSetter{
		Tag:   omit.From(tag.Tag),
		InUse: omit.From(true),
	}

	inserted, err := models.Tags.Insert(ctx, exec, model)
	if err != nil {
		return nil, err
	}

	return &Tag{
		ID:        inserted.ID,
		Tag:       inserted.Tag,
		InUse:     inserted.InUse,
		CreatedAt: inserted.CreatedAt.Time,
		UpdatedAt: inserted.UpdatedAt.Time,
	}, nil
}

func (*RepoSQLite) MarkTagUsed(ctx context.Context, exec bob.Executor, tag string) error {
	setter := models.TagSetter{
		InUse:     omit.From(true),
		UpdatedAt: omit.From(types.NewSQLiteDatetime(time.Now())),
	}

	_, err := models.Tags.UpdateQ(ctx, exec, models.UpdateWhere.Tags.Tag.EQ(tag), setter).Exec()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrTagNotFound
		}
		return err
	}

	return nil
}

func (*RepoSQLite) MarkTagUnused(ctx context.Context, exec bob.Executor, tag string) error {
	setter := models.TagSetter{
		InUse:     omit.From(false),
		UpdatedAt: omit.From(types.NewSQLiteDatetime(time.Now())),
	}

	_, err := models.Tags.UpdateQ(ctx, exec, models.UpdateWhere.Tags.Tag.EQ(tag), setter).Exec()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrTagNotFound
		}
		return err
	}

	return nil
}

func (*RepoSQLite) Delete(ctx context.Context, exec bob.Executor, tag string) error {
	_, err := models.Tags.DeleteQ(ctx, exec, models.DeleteWhere.Tags.Tag.EQ(tag), models.DeleteWhere.Tags.InUse.EQ(false)).Exec()
	return err
}

func (*RepoSQLite) NextSequential(ctx context.Context, exec bob.Executor) (int64, error) {
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
