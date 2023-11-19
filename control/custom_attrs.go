package control

import (
	"context"

	"github.com/RobinThrift/stuff/entities"
	"github.com/RobinThrift/stuff/storage/database"
	"github.com/stephenafamo/bob"
)

type CustomAttrCtrl struct {
	db   *database.Database
	repo CustomAttrRepo
}

type CustomAttrRepo interface {
	List(ctx context.Context, exec bob.Executor, query database.ListCustomAttrsQuery) (*entities.ListPage[*entities.CustomAttr], error)
}

func NewCustomAttrCtrl(db *database.Database, repo CustomAttrRepo) *CustomAttrCtrl {
	return &CustomAttrCtrl{db: db, repo: repo}
}

type ListCustomAttrsQuery struct {
	Search   string
	Page     int
	PageSize int
}

func (cac *CustomAttrCtrl) List(ctx context.Context, query ListCustomAttrsQuery) (*entities.ListPage[*entities.CustomAttr], error) {
	return database.InTransaction(ctx, cac.db, func(ctx context.Context, tx database.Executor) (*entities.ListPage[*entities.CustomAttr], error) {
		return cac.repo.List(ctx, tx, database.ListCustomAttrsQuery{
			Search:   query.Search,
			Page:     query.Page,
			PageSize: query.PageSize,
		})
	})
}
