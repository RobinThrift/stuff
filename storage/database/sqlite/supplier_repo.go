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

type SupplierRepo struct{}

func (cr *SupplierRepo) List(ctx context.Context, exec bob.Executor, query database.ListSuppliersQuery) (*entities.ListPage[*entities.Supplier], error) {
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
		qmods = append(qmods, models.SelectWhere.Suppliers.Name.Like("%"+query.Search+"%"))
	}

	count, err := models.Suppliers.Query(ctx, exec, qmods...).Count()
	if err != nil {
		return nil, fmt.Errorf("error counting locations: %w", err)
	}

	suppliers, err := models.Suppliers.Query(ctx, exec, qmods...).All()
	if err != nil {
		return nil, err
	}

	numPages, pageSize := calcNumPages(query.PageSize, count)
	page := &entities.ListPage[*entities.Supplier]{
		Items:    make([]*entities.Supplier, 0, len(suppliers)),
		Total:    int(count),
		Page:     query.Page,
		PageSize: pageSize,
		NumPages: numPages,
	}

	for _, l := range suppliers {
		if l.Name.IsSet() {
			page.Items = append(page.Items, &entities.Supplier{Name: l.Name.GetOrZero()})
		}
	}

	return page, nil
}
