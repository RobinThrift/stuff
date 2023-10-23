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
	"github.com/stephenafamo/bob/dialect/sqlite/dialect"
	"github.com/stephenafamo/bob/dialect/sqlite/sm"
	bmods "github.com/stephenafamo/bob/mods"
)

type UserRepoSQLite struct{}

func (*UserRepoSQLite) List(ctx context.Context, exec bob.Executor, query ListUsersQuery) (*UserListPage, error) {
	limit := query.PageSize

	if limit == 0 {
		limit = 50
	}

	if limit > 100 {
		limit = 100
	}

	offset := limit * query.Page

	mods := []bob.Mod[*dialect.SelectQuery]{
		sm.Limit(limit),
		sm.Offset(offset),
	}

	if query.OrderBy != "" {
		if query.OrderDir == "" {
			query.OrderDir = "ASC"
		}

		mods = append(mods, bmods.OrderBy[*dialect.SelectQuery]{
			Expression: query.OrderBy,
			Direction:  query.OrderDir,
		})
	}

	if query.Search != "" {
		mods = append(mods, models.SelectWhere.Users.Username.Like("%"+query.Search+"%"))
	}

	users, err := models.Users.Query(ctx, exec, mods...).All()
	if err != nil {
		return nil, fmt.Errorf("error getting users: %w", err)
	}

	count, err := models.Users.Query(ctx, exec, mods...).Count()
	if err != nil {
		return nil, fmt.Errorf("error counting users: %w", err)
	}

	page := &UserListPage{
		Users:    make([]*User, 0, len(users)),
		Total:    int(count),
		Page:     query.Page,
		PageSize: query.PageSize,
		NumPages: int(count) / query.PageSize,
	}

	for i := range users {
		page.Users = append(page.Users, &User{
			ID:          users[i].ID,
			Username:    users[i].Username,
			DisplayName: users[i].DisplayName,
			IsAdmin:     users[i].IsAdmin,
			CreatedAt:   users[i].CreatedAt.Time,
			UpdatedAt:   users[i].UpdatedAt.Time,
		})
	}

	return page, nil
}

func (*UserRepoSQLite) Create(ctx context.Context, tx bob.Executor, toCreate *User) (*User, error) {
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

func (*UserRepoSQLite) Update(ctx context.Context, tx bob.Executor, toUpdate *User) (*User, error) {
	model := &models.User{
		ID:          toUpdate.ID,
		Username:    toUpdate.Username,
		DisplayName: toUpdate.DisplayName,
		IsAdmin:     toUpdate.IsAdmin,
		AuthRef:     toUpdate.AuthRef,
	}

	err := models.Users.Update(ctx, tx, &models.UserSetter{
		Username:    omit.From(toUpdate.Username),
		DisplayName: omit.From(toUpdate.DisplayName),
		IsAdmin:     omit.From(toUpdate.IsAdmin),
		AuthRef:     omit.From(toUpdate.AuthRef),
		UpdatedAt:   omit.From(types.NewSQLiteDatetime(time.Now())),
	}, model)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:          model.ID,
		Username:    model.Username,
		DisplayName: model.DisplayName,
		IsAdmin:     model.IsAdmin,
		AuthRef:     model.AuthRef,
		CreatedAt:   model.CreatedAt.Time,
		UpdatedAt:   model.UpdatedAt.Time,
	}, nil
}

func (*UserRepoSQLite) Get(ctx context.Context, tx bob.Executor, id int64) (*User, error) {
	query := models.Users.Query(ctx, tx, models.SelectWhere.Users.ID.EQ(id))
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

func (*UserRepoSQLite) GetByRef(ctx context.Context, tx bob.Executor, ref string) (*User, error) {
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

func (*UserRepoSQLite) CountAdmins(ctx context.Context, tx bob.Executor) (int64, error) {
	query := models.Users.Query(ctx, tx, models.SelectWhere.Users.IsAdmin.EQ(true))
	return query.Count()
}

func (*UserRepoSQLite) Delete(ctx context.Context, tx bob.Executor, id int64) error {
	_, err := models.Users.DeleteQ(ctx, tx, models.DeleteWhere.Users.ID.EQ(id)).Exec()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("error deleting user: %w", ErrUserNotFound)
		}
		return err
	}

	return nil
}
