package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/RobinThrift/stuff/auth"
	"github.com/RobinThrift/stuff/entities"
	"github.com/RobinThrift/stuff/storage/database"
	"github.com/RobinThrift/stuff/storage/database/sqlite/models"
	"github.com/RobinThrift/stuff/storage/database/sqlite/types"
	"github.com/aarondl/opt/omit"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/sqlite/dialect"
	"github.com/stephenafamo/bob/dialect/sqlite/sm"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepo struct{}

func (*UserRepo) Get(ctx context.Context, exec bob.Executor, id int64) (*auth.User, error) {
	query := models.Users.Query(ctx, exec, models.SelectWhere.Users.ID.EQ(id))
	user, err := query.One()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
	}

	return &auth.User{
		ID:          user.ID,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		IsAdmin:     user.IsAdmin,
		AuthRef:     user.AuthRef,
		CreatedAt:   user.CreatedAt.Time,
		UpdatedAt:   user.UpdatedAt.Time,
	}, nil
}

func (*UserRepo) GetByUsername(ctx context.Context, exec bob.Executor, username string) (*auth.User, error) {
	query := models.Users.Query(ctx, exec, models.SelectWhere.Users.Username.EQ(username))
	user, err := query.One()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
	}

	return &auth.User{
		ID:          user.ID,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		IsAdmin:     user.IsAdmin,
		AuthRef:     user.AuthRef,
		CreatedAt:   user.CreatedAt.Time,
		UpdatedAt:   user.UpdatedAt.Time,
	}, nil
}

func (*UserRepo) GetByRef(ctx context.Context, exec bob.Executor, ref string) (*auth.User, error) {
	query := models.Users.Query(ctx, exec, models.SelectWhere.Users.AuthRef.EQ(ref))
	user, err := query.One()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
	}

	return &auth.User{
		ID:          user.ID,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		IsAdmin:     user.IsAdmin,
		AuthRef:     user.AuthRef,
		CreatedAt:   user.CreatedAt.Time,
		UpdatedAt:   user.UpdatedAt.Time,
	}, nil
}

func (*UserRepo) List(ctx context.Context, exec bob.Executor, query database.ListUsersQuery) (*entities.ListPage[*auth.User], error) {
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

	if query.Search != "" {
		mods = append(mods, models.SelectWhere.Users.Username.Like("%"+query.Search+"%"))
	}

	count, err := models.Users.Query(ctx, exec, mods...).Count()
	if err != nil {
		return nil, fmt.Errorf("error counting users: %w", err)
	}

	if query.OrderBy != "" {
		if query.OrderDir == "" {
			query.OrderDir = "ASC"
		}

		mods = append(mods, orderByClause(models.TableNames.Users, query.OrderBy, query.OrderDir))
	}

	users, err := models.Users.Query(ctx, exec, mods...).All()
	if err != nil {
		return nil, fmt.Errorf("error getting users: %w", err)
	}

	pageSize := limit
	page := &entities.ListPage[*auth.User]{
		Items:    make([]*auth.User, 0, len(users)),
		Total:    int(count),
		Page:     query.Page,
		PageSize: query.PageSize,
		NumPages: int(count) / pageSize,
	}

	for i := range users {
		page.Items = append(page.Items, &auth.User{
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

func (*UserRepo) CountAdmins(ctx context.Context, exec bob.Executor) (int64, error) {
	query := models.Users.Query(ctx, exec, models.SelectWhere.Users.IsAdmin.EQ(true))
	return query.Count()
}

func (*UserRepo) Create(ctx context.Context, exec bob.Executor, toCreate *auth.User) error {
	inserted, err := models.Users.Insert(ctx, exec, &models.UserSetter{
		Username:    omit.From[string](toCreate.Username),
		DisplayName: omit.From[string](toCreate.DisplayName),
		IsAdmin:     omit.From[bool](toCreate.IsAdmin),
		AuthRef:     omit.From[string](toCreate.AuthRef),
	})
	if err != nil {
		return err
	}

	toCreate.ID = inserted.ID

	return nil
}

func (*UserRepo) Update(ctx context.Context, exec bob.Executor, toUpdate *auth.User) error {
	model := &models.User{
		ID:          toUpdate.ID,
		Username:    toUpdate.Username,
		DisplayName: toUpdate.DisplayName,
		IsAdmin:     toUpdate.IsAdmin,
		AuthRef:     toUpdate.AuthRef,
	}

	err := models.Users.Update(ctx, exec, &models.UserSetter{
		Username:    omit.From[string](toUpdate.Username),
		DisplayName: omit.From[string](toUpdate.DisplayName),
		IsAdmin:     omit.From[bool](toUpdate.IsAdmin),
		AuthRef:     omit.From[string](toUpdate.AuthRef),
		UpdatedAt:   omit.From(types.NewSQLiteDatetime(time.Now())),
	}, model)
	if err != nil {
		return err
	}

	toUpdate.UpdatedAt = model.UpdatedAt.Time

	return nil
}

func (*UserRepo) Delete(ctx context.Context, exec bob.Executor, id int64) error {
	_, err := models.Users.DeleteQ(ctx, exec, models.DeleteWhere.Users.ID.EQ(id)).Exec()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("error deleting user: %w", ErrUserNotFound)
		}
		return err
	}

	return nil
}
