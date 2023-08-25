package users

import (
	"context"
	"errors"

	"github.com/kodeshack/stuff/storage/database"
	"github.com/stephenafamo/bob"
)

type Control struct {
	DB       *database.Database
	UserRepo UserRepo
}

type UserRepo interface {
	CreateUser(ctx context.Context, tx bob.Executor, toCreate *database.User) (*database.User, error)
	GetUserByRef(ctx context.Context, tx bob.Executor, ref string) (*database.User, error)
}

func (c *Control) CreateUser(ctx context.Context, toCreate *User, authRef string) (*User, error) {
	return database.InTransaction(ctx, c.DB, func(ctx context.Context, tx bob.Tx) (*User, error) {
		user, err := c.UserRepo.CreateUser(ctx, tx, &database.User{
			Username:    toCreate.Username,
			DisplayName: toCreate.DisplayName,
			IsAdmin:     toCreate.IsAdmin,
			AuthRef:     authRef,
		})
		if err != nil {
			return nil, err
		}

		return &User{
			ID:          user.ID,
			Username:    user.Username,
			DisplayName: user.DisplayName,
			IsAdmin:     user.IsAdmin,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
		}, nil
	})
}

func (c *Control) GetUserByRef(ctx context.Context, ref string) (*User, error) {
	return database.InTransaction(ctx, c.DB, func(ctx context.Context, tx bob.Tx) (*User, error) {
		user, err := c.UserRepo.GetUserByRef(ctx, tx, ref)
		if err != nil {
			if errors.Is(err, database.ErrUserNotFound) {
				return nil, ErrUserNotFound
			}

			return nil, err
		}

		return &User{
			ID:          user.ID,
			Username:    user.Username,
			DisplayName: user.DisplayName,
			IsAdmin:     user.IsAdmin,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
		}, nil
	})
}
