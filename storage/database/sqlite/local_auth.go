package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/aarondl/opt/omit"
	"github.com/kodeshack/stuff/storage/database"
	"github.com/kodeshack/stuff/storage/database/sqlite/models"
	"github.com/kodeshack/stuff/storage/database/sqlite/types"
	"github.com/stephenafamo/bob"
)

type LocalAuthRepo struct {
}

func (*LocalAuthRepo) GetLocalUser(ctx context.Context, tx bob.Executor, username string) (*database.LocalAuthUser, error) {
	query := models.LocalAuthUsers.Query(ctx, tx, models.SelectWhere.LocalAuthUsers.Username.EQ(username))
	user, err := query.One()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, database.ErrLocalAuthUserNotFound
		}

		return nil, err
	}

	return &database.LocalAuthUser{
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

func (*LocalAuthRepo) CreateLocalUser(ctx context.Context, tx bob.Executor, user *database.LocalAuthUser) (*database.LocalAuthUser, error) {
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

	return &database.LocalAuthUser{
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

func (*LocalAuthRepo) UpdateLocalUser(ctx context.Context, tx bob.Executor, user *database.LocalAuthUser) (error) {
	_, err := models.LocalAuthUsers.UpdateQ(ctx, tx, &models.LocalAuthUserSetter{
		Algorithm:              omit.From(user.Algorithm),
		Params:                 omit.From(user.Params),
		Salt:                   omit.From(user.Salt),
		Password:               omit.From(user.Password),
		RequiresPasswordChange: omit.From(user.RequiresPasswordChange),
		UpdatedAt:              omit.From(types.NewSQLiteDatetime(time.Now())),
	}, models.UpdateWhere.LocalAuthUsers.ID.EQ(user.ID)).One()
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("error updating local auth user: %w", database.ErrLocalAuthUserNotFound)
		}
	}

	return nil
}
