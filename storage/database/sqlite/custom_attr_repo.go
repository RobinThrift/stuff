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

type CustomAttrRepo struct{}

func (cr *CustomAttrRepo) List(ctx context.Context, exec bob.Executor, query database.ListCustomAttrsQuery) (*entities.ListPage[*entities.CustomAttr], error) {
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
		qmods = append(qmods, models.SelectWhere.CustomAttrNames.AttrName.Like("%"+query.Search+"%"))
	}

	count, err := models.CustomAttrNames.Query(ctx, exec, qmods...).Count()
	if err != nil {
		return nil, fmt.Errorf("error counting locations: %w", err)
	}

	customAttrs, err := models.CustomAttrNames.Query(ctx, exec, qmods...).All()
	if err != nil {
		return nil, err
	}

	numPages, pageSize := calcNumPages(query.PageSize, count)
	page := &entities.ListPage[*entities.CustomAttr]{
		Items:    make([]*entities.CustomAttr, 0, len(customAttrs)),
		Total:    int(count),
		Page:     query.Page,
		PageSize: pageSize,
		NumPages: numPages,
	}

	for _, ca := range customAttrs {
		if ca.AttrName.IsSet() {
			page.Items = append(page.Items, &entities.CustomAttr{Name: ca.AttrName.GetOrZero()})
		}
	}

	return page, nil
}
