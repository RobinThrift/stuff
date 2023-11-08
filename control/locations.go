package control

import (
	"context"

	"github.com/RobinThrift/stuff/entities"
	"github.com/RobinThrift/stuff/storage/database"
	"github.com/stephenafamo/bob"
)

type LocationControl struct {
	db *database.Database

	locations LocationRepo
}

type LocationRepo interface {
	ListLocations(ctx context.Context, exec bob.Executor, query database.ListLocationsQuery) (*entities.ListPage[*entities.Location], error)
	ListPositionCodes(ctx context.Context, exec bob.Executor, query database.ListPositionCodesQuery) (*entities.ListPage[*entities.PositionCode], error)
}

func NewLocationControl(db *database.Database, repo LocationRepo) *LocationControl {
	return &LocationControl{db: db, locations: repo}
}

type ListLocationsQuery struct {
	Search   string
	Page     int
	PageSize int
}

func (lc *LocationControl) ListLocations(ctx context.Context, query ListLocationsQuery) (*entities.ListPage[*entities.Location], error) {
	return database.InTransaction(ctx, lc.db, func(ctx context.Context, tx bob.Tx) (*entities.ListPage[*entities.Location], error) {
		return lc.locations.ListLocations(ctx, tx, database.ListLocationsQuery(query))
	})
}

type ListPositionCodesQuery struct {
	Search   string
	Page     int
	PageSize int
}

func (lc *LocationControl) ListPositionCodes(ctx context.Context, query ListPositionCodesQuery) (*entities.ListPage[*entities.PositionCode], error) {
	return database.InTransaction(ctx, lc.db, func(ctx context.Context, tx bob.Tx) (*entities.ListPage[*entities.PositionCode], error) {
		return lc.locations.ListPositionCodes(ctx, tx, database.ListPositionCodesQuery(query))
	})
}
