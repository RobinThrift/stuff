package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/RobinThrift/stuff/auth"
	"github.com/RobinThrift/stuff/storage/database/sqlite/models"
	"github.com/RobinThrift/stuff/storage/database/sqlite/types"
	"github.com/aarondl/opt/omit"
	"github.com/stephenafamo/bob"
)

var ErrLocalAuthUserNotFound = errors.New("user for local auth not found")

type LocalAuthRepo struct{}

func (*LocalAuthRepo) Get(ctx context.Context, exec bob.Executor, username string) (*auth.LocalAuthUser, error) {
	query := models.LocalAuthUsers.Query(ctx, exec, models.SelectWhere.LocalAuthUsers.Username.EQ(username))
	user, err := query.One()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrLocalAuthUserNotFound
		}

		return nil, err
	}

	return &auth.LocalAuthUser{
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

func (*LocalAuthRepo) Create(ctx context.Context, exec bob.Executor, user *auth.LocalAuthUser) error {
	inserted, err := models.LocalAuthUsers.Insert(ctx, exec, &models.LocalAuthUserSetter{
		Username:               omit.From[string](user.Username),
		Algorithm:              omit.From[string](user.Algorithm),
		Params:                 omit.From[string](user.Params),
		Salt:                   omit.From[[]byte](user.Salt),
		Password:               omit.From[[]byte](user.Password),
		RequiresPasswordChange: omit.From[bool](user.RequiresPasswordChange),
	})
	if err != nil {
		return err
	}

	user.ID = inserted.ID
	user.CreatedAt = inserted.CreatedAt.Time
	user.UpdatedAt = inserted.UpdatedAt.Time

	return nil
}

func (*LocalAuthRepo) Update(ctx context.Context, exec bob.Executor, user *auth.LocalAuthUser) error {
	_, err := models.LocalAuthUsers.UpdateQ(ctx, exec, models.UpdateWhere.LocalAuthUsers.ID.EQ(user.ID), &models.LocalAuthUserSetter{
		Algorithm:              omit.From[string](user.Algorithm),
		Params:                 omit.From[string](user.Params),
		Salt:                   omit.From[[]byte](user.Salt),
		Password:               omit.From[[]byte](user.Password),
		RequiresPasswordChange: omit.From[bool](user.RequiresPasswordChange),
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

func (*LocalAuthRepo) DeleteByUsername(ctx context.Context, exec bob.Executor, username string) error {
	_, err := models.LocalAuthUsers.DeleteQ(ctx, exec, models.DeleteWhere.LocalAuthUsers.Username.EQ(username)).Exec()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("error deleting local auth user: %w", ErrLocalAuthUserNotFound)
		}
		return err
	}

	return nil
}
