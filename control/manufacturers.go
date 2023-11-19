package control

import (
	"context"

	"github.com/RobinThrift/stuff/entities"
	"github.com/RobinThrift/stuff/storage/database"
	"github.com/stephenafamo/bob"
)

type ManufactuerCtrl struct {
	db *database.Database

	repo ManufactuerRepo
}

type ManufactuerRepo interface {
	List(ctx context.Context, exec bob.Executor, query database.ListManufacturersQuery) (*entities.ListPage[*entities.Manufacturer], error)
}

func NewManufactuerCtrl(db *database.Database, repo ManufactuerRepo) *ManufactuerCtrl {
	return &ManufactuerCtrl{db: db, repo: repo}
}

type ListManufacturersQuery struct {
	Search   string
	Page     int
	PageSize int
}

func (cc *ManufactuerCtrl) List(ctx context.Context, query ListManufacturersQuery) (*entities.ListPage[*entities.Manufacturer], error) {
	return database.InTransaction(ctx, cc.db, func(ctx context.Context, tx database.Executor) (*entities.ListPage[*entities.Manufacturer], error) {
		return cc.repo.List(ctx, tx, database.ListManufacturersQuery{
			Search:   query.Search,
			Page:     query.Page,
			PageSize: query.PageSize,
		})
	})
}
