package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/RobinThrift/stuff/storage/database/sqlite/models"
	"github.com/RobinThrift/stuff/storage/database/sqlite/types"
	"github.com/aarondl/opt/omit"
	"github.com/stephenafamo/bob"
)

type LocalAuthRepoSQLite struct{}

func (*LocalAuthRepoSQLite) GetLocalUser(ctx context.Context, tx bob.Executor, username string) (*LocalAuthUser, error) {
	query := models.LocalAuthUsers.Query(ctx, tx, models.SelectWhere.LocalAuthUsers.Username.EQ(username))
	user, err := query.One()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrLocalAuthUserNotFound
		}

		return nil, err
	}

	return &LocalAuthUser{
		ID:                     user.ID,
		Username:               user.Username,
		Algorithm:              user.Algorithm,
		Params:                 user.Params,
		Salt:                   user.Salt,
		Password:               user.Password,
		RequiresPasswordChange: user.RequiresPasswordChange,
		CreatedAt:              user.CreatedAt.Time,
		UpdatedAt:              user.UpdatedAt.Time,
	}, nil
}

func (*LocalAuthRepoSQLite) CreateLocalUser(ctx context.Context, tx bob.Executor, user *LocalAuthUser) (*LocalAuthUser, error) {
	inserted, err := models.LocalAuthUsers.Insert(ctx, tx, &models.LocalAuthUserSetter{
		Username:               omit.From(user.Username),
		Algorithm:              omit.From(user.Algorithm),
		Params:                 omit.From(user.Params),
		Salt:                   omit.From(user.Salt),
		Password:               omit.From(user.Password),
		RequiresPasswordChange: omit.From(user.RequiresPasswordChange),
	})
	if err != nil {
		return nil, err
	}

	return &LocalAuthUser{
		ID:                     inserted.ID,
		Username:               inserted.Username,
		Algorithm:              inserted.Algorithm,
		Params:                 inserted.Params,
		Salt:                   inserted.Salt,
		Password:               inserted.Password,
		RequiresPasswordChange: inserted.RequiresPasswordChange,
		CreatedAt:              inserted.CreatedAt.Time,
		UpdatedAt:              inserted.UpdatedAt.Time,
	}, nil
}

func (*LocalAuthRepoSQLite) UpdateLocalUser(ctx context.Context, tx bob.Executor, user *LocalAuthUser) error {
	_, err := models.LocalAuthUsers.UpdateQ(ctx, tx, models.UpdateWhere.LocalAuthUsers.ID.EQ(user.ID), &models.LocalAuthUserSetter{
		Algorithm:              omit.From(user.Algorithm),
		Params:                 omit.From(user.Params),
		Salt:                   omit.From(user.Salt),
		Password:               omit.From(user.Password),
		RequiresPasswordChange: omit.From(user.RequiresPasswordChange),
		UpdatedAt:              omit.From(types.NewSQLiteDatetime(time.Now())),
	}).Exec()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("error updating local auth user: %w", ErrLocalAuthUserNotFound)
		}
		return err
	}

	return nil
}

func (*LocalAuthRepoSQLite) DeleteByUsername(ctx context.Context, tx bob.Executor, username string) error {
	_, err := models.LocalAuthUsers.DeleteQ(ctx, tx, models.DeleteWhere.LocalAuthUsers.Username.EQ(username)).Exec()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("error deleting local auth user: %w", ErrLocalAuthUserNotFound)
		}
		return err
	}

	return nil
}
