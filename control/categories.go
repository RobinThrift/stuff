package control

import (
	"context"

	"github.com/RobinThrift/stuff/entities"
	"github.com/RobinThrift/stuff/storage/database"
	"github.com/stephenafamo/bob"
)

type CategoryCtrl struct {
	db *database.Database

	repo CategoryRepo
}

type CategoryRepo interface {
	List(ctx context.Context, exec bob.Executor, query database.ListCategoriesQuery) (*entities.ListPage[*entities.Category], error)
}

func NewCategoryCtrl(db *database.Database, repo CategoryRepo) *CategoryCtrl {
	return &CategoryCtrl{db: db, repo: repo}
}

type ListCategoriesQuery struct {
	Search   string
	Page     int
	PageSize int
}

func (cc *CategoryCtrl) List(ctx context.Context, query ListCategoriesQuery) (*entities.ListPage[*entities.Category], error) {
	return database.InTransaction(ctx, cc.db, func(ctx context.Context, tx bob.Tx) (*entities.ListPage[*entities.Category], error) {
		return cc.repo.List(ctx, tx, database.ListCategoriesQuery(query))
	})
}
