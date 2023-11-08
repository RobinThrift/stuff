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

type LocationRepo struct{}

func (cr *LocationRepo) ListLocations(ctx context.Context, exec bob.Executor, query database.ListLocationsQuery) (*entities.ListPage[*entities.Location], error) {
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
		qmods = append(qmods, models.SelectWhere.Locations.LocName.Like("%"+query.Search+"%"))
	}

	count, err := models.Locations.Query(ctx, exec, qmods...).Count()
	if err != nil {
		return nil, fmt.Errorf("error counting locations: %w", err)
	}

	locations, err := models.Locations.Query(ctx, exec, qmods...).All()
	if err != nil {
		return nil, err
	}

	numPages, pageSize := calcNumPages(query.PageSize, count)
	page := &entities.ListPage[*entities.Location]{
		Items:    make([]*entities.Location, 0, len(locations)),
		Total:    int(count),
		Page:     query.Page,
		PageSize: pageSize,
		NumPages: numPages,
	}

	for _, l := range locations {
		if l.LocName.IsSet() {
			page.Items = append(page.Items, &entities.Location{Name: l.LocName.GetOrZero()})
		}
	}

	return page, nil
}

func (cr *LocationRepo) ListPositionCodes(ctx context.Context, exec bob.Executor, query database.ListPositionCodesQuery) (*entities.ListPage[*entities.PositionCode], error) {
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
		qmods = append(qmods, models.SelectWhere.PositionCodes.PosCode.Like("%"+query.Search+"%"))
	}

	count, err := models.PositionCodes.Query(ctx, exec, qmods...).Count()
	if err != nil {
		return nil, fmt.Errorf("error counting position codes: %w", err)
	}

	poscodes, err := models.PositionCodes.Query(ctx, exec, qmods...).All()
	if err != nil {
		return nil, err
	}

	numPages, pageSize := calcNumPages(query.PageSize, count)
	page := &entities.ListPage[*entities.PositionCode]{
		Items:    make([]*entities.PositionCode, 0, len(poscodes)),
		Total:    int(count),
		Page:     query.Page,
		PageSize: pageSize,
		NumPages: numPages,
	}

	for _, p := range poscodes {
		if p.PosCode.IsSet() {
			page.Items = append(page.Items, &entities.PositionCode{Code: p.PosCode.GetOrZero()})
		}
	}

	return page, nil
}
