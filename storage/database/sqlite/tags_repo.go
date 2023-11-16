package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/RobinThrift/stuff/entities"
	"github.com/RobinThrift/stuff/storage/database"
	"github.com/RobinThrift/stuff/storage/database/sqlite/models"
	"github.com/RobinThrift/stuff/storage/database/sqlite/types"
	"github.com/aarondl/opt/omit"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/sqlite"
	"github.com/stephenafamo/bob/dialect/sqlite/dialect"
	"github.com/stephenafamo/bob/dialect/sqlite/sm"
	"github.com/stephenafamo/scan"
)

var ErrTagNotFound = errors.New("tag not found")

type TagRepo struct{}

func (*TagRepo) List(ctx context.Context, exec bob.Executor, query database.ListTagsQuery) (*entities.ListPage[*entities.Tag], error) {
	limit := query.PageSize

	if limit == 0 {
		limit = 50
	}

	if limit > 100 {
		limit = 100
	}

	offset := limit * query.Page

	qmods := make([]bob.Mod[*dialect.SelectQuery], 0, 3)
	if query.Search != "" {
		qmods = append(qmods, models.SelectWhere.Tags.Tag.Like("%"+query.Search+"%"))
	}

	if query.InUse != nil {
		qmods = append(qmods, models.SelectWhere.Tags.InUse.EQ(*query.InUse))
	}

	count, err := models.Tags.Query(ctx, exec, qmods...).Count()
	if err != nil {
		return nil, fmt.Errorf("error counting tags: %w", err)
	}

	if query.OrderBy != "" {
		if query.OrderDir == "" {
			query.OrderDir = "ASC"
		}

		qmods = append(qmods, orderByClause(models.TableNames.Tags, query.OrderBy, query.OrderDir))
	}

	qmods = append(qmods, sm.Limit(limit), sm.Offset(offset))

	tags, err := models.Tags.Query(ctx, exec, qmods...).All()
	if err != nil {
		return nil, fmt.Errorf("error getting tags: %w", wrapSqliteErr(err))
	}

	numPages, pageSize := calcNumPages(query.PageSize, count)
	page := &entities.ListPage[*entities.Tag]{
		Items:    make([]*entities.Tag, 0, len(tags)),
		Total:    int(count),
		Page:     query.Page,
		PageSize: pageSize,
		NumPages: numPages,
	}

	for i := range tags {
		page.Items = append(page.Items, &entities.Tag{
			ID:        tags[i].ID,
			Tag:       tags[i].Tag,
			InUse:     tags[i].InUse,
			CreatedAt: tags[i].CreatedAt.Time,
			UpdatedAt: tags[i].UpdatedAt.Time,
		})
	}

	return page, nil
}

func (*TagRepo) GetUnused(ctx context.Context, exec bob.Executor) (*entities.Tag, error) {
	model, err := models.Tags.Query(ctx, exec, models.SelectWhere.Tags.InUse.EQ(false)).One()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &entities.Tag{
		ID:        model.ID,
		Tag:       model.Tag,
		CreatedAt: model.CreatedAt.Time,
		UpdatedAt: model.UpdatedAt.Time,
	}, nil
}

func (*TagRepo) Get(ctx context.Context, exec bob.Executor, tag string) (*entities.Tag, error) {
	model, err := models.Tags.Query(ctx, exec, models.SelectWhere.Tags.Tag.EQ(tag)).One()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entities.ErrTagNotFound
		}
		return nil, err
	}

	return &entities.Tag{
		ID:        model.ID,
		Tag:       model.Tag,
		CreatedAt: model.CreatedAt.Time,
		UpdatedAt: model.UpdatedAt.Time,
	}, nil
}

func (*TagRepo) Create(ctx context.Context, exec bob.Executor, tag *entities.Tag) error {
	model := &models.TagSetter{
		Tag:   omit.From(tag.Tag),
		InUse: omit.From(tag.InUse),
	}

	_, err := models.Tags.Insert(ctx, exec, model)
	if err != nil {
		return nil
	}

	return nil
}

func (*TagRepo) MarkTagUsed(ctx context.Context, exec bob.Executor, tag string) error {
	setter := models.TagSetter{
		InUse:     omit.From(true),
		UpdatedAt: omit.From(types.NewSQLiteDatetime(time.Now())),
	}

	_, err := models.Tags.UpdateQ(ctx, exec, models.UpdateWhere.Tags.Tag.EQ(tag), setter).Exec()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entities.ErrTagNotFound
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
			return entities.ErrTagNotFound
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
