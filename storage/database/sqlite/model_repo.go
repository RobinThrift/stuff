package sqlite

import (
	"context"
	"fmt"

	"github.com/RobinThrift/stuff/entities"
	"github.com/RobinThrift/stuff/storage/database"
	"github.com/RobinThrift/stuff/storage/database/sqlite/models"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/sqlite"
	"github.com/stephenafamo/bob/dialect/sqlite/dialect"
	"github.com/stephenafamo/bob/dialect/sqlite/sm"
)

type ModelRepo struct{}

func (cr *ModelRepo) List(ctx context.Context, exec bob.Executor, query database.ListModelsQuery) (*entities.ListPage[*entities.Model], error) {
	limit := query.PageSize
	if limit == 0 {
		limit = 25
	}
	offset := limit * query.Page

	qmods := []bob.Mod[*dialect.SelectQuery]{
		sm.Limit(limit),
		sm.Offset(offset),
	}

	if query.Search != "" {
		qmods = append(qmods, sqlite.WhereOr(
			models.SelectWhere.Models.Model.Like("%"+query.Search+"%"),
			models.SelectWhere.Models.ModelNo.Like("%"+query.Search+"%"),
		))
	}

	count, err := models.Models.Query(ctx, exec, qmods...).Count()
	if err != nil {
		return nil, fmt.Errorf("error counting locations: %w", err)
	}

	Models, err := models.Models.Query(ctx, exec, qmods...).All()
	if err != nil {
		return nil, err
	}

	numPages, pageSize := calcNumPages(query.PageSize, count)
	page := &entities.ListPage[*entities.Model]{
		Items:    make([]*entities.Model, 0, len(Models)),
		Total:    int(count),
		Page:     query.Page,
		PageSize: pageSize,
		NumPages: numPages,
	}

	for _, m := range Models {
		model := &entities.Model{
			Name:    m.Model.GetOrZero(),
			ModelNo: m.ModelNo.GetOrZero(),
		}

		if model.Name != "" || model.ModelNo != "" {
			page.Items = append(page.Items, model)
		}
	}

	return page, nil
}
