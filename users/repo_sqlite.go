package users

import (
	"context"
	"database/sql"
	"errors"

	"github.com/aarondl/opt/omit"
	"github.com/kodeshack/stuff/storage/database/sqlite/models"
	"github.com/stephenafamo/bob"
)

type RepoSQLite struct{}

func (*RepoSQLite) Create(ctx context.Context, tx bob.Executor, toCreate *User) (*User, error) {
	inserted, err := models.Users.Insert(ctx, tx, &models.UserSetter{
		Username:    omit.From(toCreate.Username),
		DisplayName: omit.From(toCreate.DisplayName),
		IsAdmin:     omit.From(toCreate.IsAdmin),
		AuthRef:     omit.From(toCreate.AuthRef),
	})
	if err != nil {
		return nil, err
	}

	return &User{
		ID:          inserted.ID,
		Username:    inserted.Username,
		DisplayName: inserted.DisplayName,
		IsAdmin:     inserted.IsAdmin,
		AuthRef:     inserted.AuthRef,
		CreatedAt:   inserted.CreatedAt.Time,
		UpdatedAt:   inserted.UpdatedAt.Time,
	}, nil
}

func (*RepoSQLite) GetByRef(ctx context.Context, tx bob.Executor, ref string) (*User, error) {
	query := models.Users.Query(ctx, tx, models.SelectWhere.Users.AuthRef.EQ(ref))
	user, err := query.One()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
	}

	return &User{
		ID:          user.ID,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		IsAdmin:     user.IsAdmin,
		AuthRef:     user.AuthRef,
		CreatedAt:   user.CreatedAt.Time,
		UpdatedAt:   user.UpdatedAt.Time,
	}, nil
}
