package control

import (
	"context"

	"github.com/RobinThrift/stuff/entities"
	"github.com/RobinThrift/stuff/storage/database"
	"github.com/stephenafamo/bob"
)

type SupplierCtrl struct {
	db *database.Database

	repo SupplierRepo
}

type SupplierRepo interface {
	List(ctx context.Context, exec bob.Executor, query database.ListSuppliersQuery) (*entities.ListPage[*entities.Supplier], error)
}

func NewSupplierCtrl(db *database.Database, repo SupplierRepo) *SupplierCtrl {
	return &SupplierCtrl{db: db, repo: repo}
}

type ListSuppliersQuery struct {
	Search   string
	Page     int
	PageSize int
}

func (cc *SupplierCtrl) List(ctx context.Context, query ListSuppliersQuery) (*entities.ListPage[*entities.Supplier], error) {
	return database.InTransaction(ctx, cc.db, func(ctx context.Context, tx database.Executor) (*entities.ListPage[*entities.Supplier], error) {
		return cc.repo.List(ctx, tx, database.ListSuppliersQuery(query))
	})
}
