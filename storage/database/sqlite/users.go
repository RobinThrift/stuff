package sqlite

import (
	"context"
	"database/sql"
	"errors"

	"github.com/aarondl/opt/omit"
	"github.com/kodeshack/stuff/storage/database"
	"github.com/kodeshack/stuff/storage/database/sqlite/models"
	"github.com/stephenafamo/bob"
)

type UserRepo struct{}

func (*UserRepo) CreateUser(ctx context.Context, tx bob.Executor, toCreate *database.User) (*database.User, error) {
	inserted, err := models.Users.Insert(ctx, tx, &models.UserSetter{
		Username:    omit.From(toCreate.Username),
		DisplayName: omit.From(toCreate.DisplayName),
		IsAdmin:     omit.From(toCreate.IsAdmin),
		AuthRef:     omit.From(toCreate.AuthRef),
	})
	if err != nil {
		return nil, err
	}

	return &database.User{
		ID:          inserted.ID,
		Username:    inserted.Username,
		DisplayName: inserted.DisplayName,
		IsAdmin:     inserted.IsAdmin,
		AuthRef:     inserted.AuthRef,
		CreatedAt:   inserted.CreatedAt.Time,
		UpdatedAt:   inserted.UpdatedAt.Time,
	}, nil
}

func (*UserRepo) GetUserByRef(ctx context.Context, tx bob.Executor, ref string) (*database.User, error) {
	query := models.Users.Query(ctx, tx, models.SelectWhere.Users.AuthRef.EQ(ref))
	user, err := query.One()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, database.ErrUserNotFound
		}
	}

	return &database.User{
		ID:          user.ID,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		IsAdmin:     user.IsAdmin,
		AuthRef:     user.AuthRef,
		CreatedAt:   user.CreatedAt.Time,
		UpdatedAt:   user.UpdatedAt.Time,
	}, nil
}
