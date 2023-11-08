package sqlite

import (
	"context"
	"fmt"

	"github.com/RobinThrift/stuff/entities"
	"github.com/RobinThrift/stuff/storage/database"
	"github.com/RobinThrift/stuff/storage/database/sqlite/models"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/sqlite/dialect"
	"github.com/stephenafamo/bob/dialect/sqlite/sm"
)

type CategoryRepo struct{}

func (cr *CategoryRepo) List(ctx context.Context, exec bob.Executor, query database.ListCategoriesQuery) (*entities.ListPage[*entities.Category], error) {
	limit := query.PageSize
	if limit == 0 {
		limit = 25
	}
	offset := limit * query.Page

	qmods := []bob.Mod[*dialect.SelectQuery]{
		sm.Limit(limit),
		sm.Offset(offset),
		sm.Distinct(),
	}

	if query.Search != "" {
		qmods = append(qmods, models.SelectWhere.Categories.CatName.Like("%"+query.Search+"%"))
	}

	count, err := models.Categories.Query(ctx, exec, qmods...).Count()
	if err != nil {
		return nil, fmt.Errorf("error counting categories: %w", err)
	}

	categories, err := models.Categories.Query(ctx, exec, qmods...).All()
	if err != nil {
		return nil, err
	}

	numPages, pageSize := calcNumPages(query.PageSize, count)
	page := &entities.ListPage[*entities.Category]{
		Items:    make([]*entities.Category, 0, len(categories)),
		Total:    int(count),
		Page:     query.Page,
		PageSize: pageSize,
		NumPages: numPages,
	}

	for _, c := range categories {
		if c.CatName.IsSet() {
			page.Items = append(page.Items, &entities.Category{Name: c.CatName.GetOrZero()})
		}
	}

	return page, nil
}
