package control

import (
	"context"

	"github.com/RobinThrift/stuff/entities"
	"github.com/RobinThrift/stuff/storage/database"
	"github.com/stephenafamo/bob"
)

type ModelCtrl struct {
	db *database.Database

	repo ModelRepo
}

type ModelRepo interface {
	List(ctx context.Context, exec bob.Executor, query database.ListModelsQuery) (*entities.ListPage[*entities.Model], error)
}

func NewModelCtrl(db *database.Database, repo ModelRepo) *ModelCtrl {
	return &ModelCtrl{db: db, repo: repo}
}

type ListModelsQuery struct {
	Search   string
	Page     int
	PageSize int
}

func (cc *ModelCtrl) List(ctx context.Context, query ListModelsQuery) (*entities.ListPage[*entities.Model], error) {
	return database.InTransaction(ctx, cc.db, func(ctx context.Context, tx database.Executor) (*entities.ListPage[*entities.Model], error) {
		return cc.repo.List(ctx, tx, database.ListModelsQuery(query))
	})
}
