package control

import (
	"context"
	"errors"

	"github.com/RobinThrift/stuff/auth"
	"github.com/RobinThrift/stuff/entities"
	"github.com/RobinThrift/stuff/storage/database"
	"github.com/RobinThrift/stuff/storage/database/sqlite"
	"github.com/stephenafamo/bob"
)

var ErrUserNotFound = errors.New("user not found")

type UserControl struct {
	db *database.Database

	repo UserRepo
}

type UserRepo interface {
	List(ctx context.Context, exec bob.Executor, query database.ListUsersQuery) (*entities.ListPage[*auth.User], error)
	Create(ctx context.Context, exec bob.Executor, user *auth.User) error
	Update(ctx context.Context, exec bob.Executor, user *auth.User) error
	Get(ctx context.Context, exec bob.Executor, id int64) (*auth.User, error)
	GetByUsername(ctx context.Context, exec bob.Executor, username string) (*auth.User, error)
	GetByRef(ctx context.Context, exec bob.Executor, ref string) (*auth.User, error)
	CountAdmins(ctx context.Context, exec bob.Executor) (int64, error)
	Delete(ctx context.Context, exec bob.Executor, id int64) error
}

func NewUserCtrl(db *database.Database, repo UserRepo) *UserControl {
	return &UserControl{db: db, repo: repo}
}

func (cc *UserControl) Get(ctx context.Context, id int64) (*auth.User, error) {
	return database.InTransaction(ctx, cc.db, func(ctx context.Context, tx bob.Tx) (*auth.User, error) {
		user, err := cc.repo.Get(ctx, tx, id)
		if err != nil {
			if errors.Is(err, sqlite.ErrUserNotFound) {
				return nil, ErrUserNotFound
			}
			return nil, err
		}

		return user, nil
	})
}

func (cc *UserControl) GetByUsername(ctx context.Context, username string) (*auth.User, error) {
	return database.InTransaction(ctx, cc.db, func(ctx context.Context, tx bob.Tx) (*auth.User, error) {
		user, err := cc.repo.GetByUsername(ctx, tx, username)
		if err != nil {
			if errors.Is(err, sqlite.ErrUserNotFound) {
				return nil, ErrUserNotFound
			}
			return nil, err
		}

		return user, nil
	})
}

func (cc *UserControl) GetByRef(ctx context.Context, ref string) (*auth.User, error) {
	return database.InTransaction(ctx, cc.db, func(ctx context.Context, tx bob.Tx) (*auth.User, error) {
		user, err := cc.repo.GetByRef(ctx, tx, ref)
		if err != nil {
			if errors.Is(err, sqlite.ErrUserNotFound) {
				return nil, ErrUserNotFound
			}
			return nil, err
		}

		return user, nil
	})
}

type ListUsersQuery struct {
	Search   string
	Page     int
	PageSize int
	OrderBy  string
	OrderDir string
}

func (cc *UserControl) List(ctx context.Context, query ListUsersQuery) (*entities.ListPage[*auth.User], error) {
	return database.InTransaction(ctx, cc.db, func(ctx context.Context, tx bob.Tx) (*entities.ListPage[*auth.User], error) {
		return cc.repo.List(ctx, tx, database.ListUsersQuery(query))
	})
}

func (cc *UserControl) Create(ctx context.Context, user *auth.User) error {
	return cc.db.InTransaction(ctx, func(ctx context.Context, tx bob.Tx) error {
		return cc.repo.Create(ctx, tx, user)
	})
}

func (cc *UserControl) Update(ctx context.Context, user *auth.User) error {
	return cc.db.InTransaction(ctx, func(ctx context.Context, tx bob.Tx) error {
		return cc.repo.Update(ctx, tx, user)
	})
}

func (cc *UserControl) Delete(ctx context.Context, id int64) error {
	return cc.db.InTransaction(ctx, func(ctx context.Context, tx bob.Tx) error {
		return cc.repo.Delete(ctx, tx, id)
	})
}

func (cc *UserControl) CountAdmins(ctx context.Context) (int64, error) {
	return database.InTransaction(ctx, cc.db, func(ctx context.Context, tx bob.Tx) (int64, error) {
		return cc.repo.CountAdmins(ctx, tx)
	})
}
