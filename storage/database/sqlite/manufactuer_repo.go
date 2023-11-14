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

type ManufacturerRepo struct{}

func (cr *ManufacturerRepo) List(ctx context.Context, exec bob.Executor, query database.ListManufacturersQuery) (*entities.ListPage[*entities.Manufacturer], error) {
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
		qmods = append(qmods, models.SelectWhere.Manufacturers.Manufacturer.Like("%"+query.Search+"%"))
	}

	count, err := models.Manufacturers.Query(ctx, exec, qmods...).Count()
	if err != nil {
		return nil, fmt.Errorf("error counting manufacturers: %w", err)
	}

	manufacturers, err := models.Manufacturers.Query(ctx, exec, qmods...).All()
	if err != nil {
		return nil, err
	}

	numPages, pageSize := calcNumPages(query.PageSize, count)
	page := &entities.ListPage[*entities.Manufacturer]{
		Items:    make([]*entities.Manufacturer, 0, len(manufacturers)),
		Total:    int(count),
		Page:     query.Page,
		PageSize: pageSize,
		NumPages: numPages,
	}

	for _, l := range manufacturers {
		if l.Manufacturer.IsSet() {
			page.Items = append(page.Items, &entities.Manufacturer{Name: l.Manufacturer.GetOrZero()})
		}
	}

	return page, nil
}
