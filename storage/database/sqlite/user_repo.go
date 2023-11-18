package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/RobinThrift/stuff/auth"
	"github.com/RobinThrift/stuff/entities"
	"github.com/RobinThrift/stuff/storage/database"
	"github.com/RobinThrift/stuff/storage/database/sqlite/models"
	"github.com/RobinThrift/stuff/storage/database/sqlite/types"
	"github.com/aarondl/opt/omit"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/sqlite/dialect"
	"github.com/stephenafamo/bob/dialect/sqlite/im"
	"github.com/stephenafamo/bob/dialect/sqlite/sm"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepo struct{}

func (*UserRepo) Get(ctx context.Context, exec bob.Executor, id int64) (*auth.User, error) {
	query := models.Users.Query(ctx, exec, models.SelectWhere.Users.ID.EQ(id), models.ThenLoadUserUserPreferences())
	user, err := query.One()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: id %d", ErrUserNotFound, id)
		}

		return nil, fmt.Errorf("error getting user by id: id %d: %w", id, err)
	}

	return mapUserModelToEntity(ctx, user)
}

func (*UserRepo) GetByUsername(ctx context.Context, exec bob.Executor, username string) (*auth.User, error) {
	query := models.Users.Query(ctx, exec, models.SelectWhere.Users.Username.EQ(username), models.ThenLoadUserUserPreferences())
	user, err := query.One()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: username %s", ErrUserNotFound, username)
		}

		return nil, fmt.Errorf("error getting user by username: username %s: %w", username, err)
	}

	return mapUserModelToEntity(ctx, user)
}

func (*UserRepo) GetByRef(ctx context.Context, exec bob.Executor, ref string) (*auth.User, error) {
	query := models.Users.Query(ctx, exec, models.SelectWhere.Users.AuthRef.EQ(ref), models.ThenLoadUserUserPreferences())
	user, err := query.One()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: ref %s", ErrUserNotFound, ref)
		}

		return nil, fmt.Errorf("error getting user by ref: ref %s: %w", ref, err)
	}

	return mapUserModelToEntity(ctx, user)
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

func (*UserRepo) UpsertPreferences(ctx context.Context, exec bob.Executor, user *auth.User) error {
	inserts := make([]bob.Mod[*dialect.InsertQuery], 0, 8)

	inserts = append(
		inserts,
		im.IntoAs(models.TableNames.UserPreferences, models.TableNames.UserPreferences,
			models.ColumnNames.UserPreferences.UserID,
			models.ColumnNames.UserPreferences.Key,
			models.ColumnNames.UserPreferences.Value,
			models.ColumnNames.UserPreferences.CreatedAt,
			models.ColumnNames.UserPreferences.UpdatedAt,
		),
		im.OnConflict(
			models.ColumnNames.UserPreferences.UserID,
			models.ColumnNames.UserPreferences.Key,
		).SetExcluded(
			models.ColumnNames.UserPreferences.Value,
			models.ColumnNames.UserPreferences.UpdatedAt,
		).DoUpdate(),
	)

	setters, err := mapUserPrefsToInsert(user.ID, user.Preferences)
	if err != nil {
		return err
	}

	inserts = append(inserts, setters...)

	_, err = models.UserPreferences.InsertQ(ctx, exec, inserts...).Exec()
	if err != nil {
		return err
	}

	return nil
}

func mapUserModelToEntity(ctx context.Context, user *models.User) (*auth.User, error) {
	var prefs auth.UserPreferences
	for _, pref := range user.R.UserPreferences {
		err := mapUserPrefValueToEntity(ctx, &prefs, pref)
		if err != nil {
			return nil, err
		}
	}

	return &auth.User{
		ID:          user.ID,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		IsAdmin:     user.IsAdmin,
		AuthRef:     user.AuthRef,
		Preferences: prefs,
		CreatedAt:   user.CreatedAt.Time,
		UpdatedAt:   user.UpdatedAt.Time,
	}, nil
}

func mapUserPrefsToInsert(userID int64, prefs auth.UserPreferences) ([]bob.Mod[*dialect.InsertQuery], error) {
	inserts := make([]bob.Mod[*dialect.InsertQuery], 0, 6)

	inserts = append(inserts,
		models.UserPreferenceSetter{
			UserID:    omit.From(userID),
			Key:       omit.From("sidebar_closed_desktop"),
			Value:     omit.From([]byte(fmt.Sprint(prefs.SidebarClosedDesktop))),
			CreatedAt: omit.From(types.NewSQLiteDatetime(time.Now())),
			UpdatedAt: omit.From(types.NewSQLiteDatetime(time.Now())),
		}.Insert(),
		models.UserPreferenceSetter{
			UserID:    omit.From(userID),
			Key:       omit.From("asset_list_compact"),
			Value:     omit.From([]byte(fmt.Sprint(prefs.AssetListCompact))),
			CreatedAt: omit.From(types.NewSQLiteDatetime(time.Now())),
			UpdatedAt: omit.From(types.NewSQLiteDatetime(time.Now())),
		}.Insert(),
		models.UserPreferenceSetter{
			UserID:    omit.From(userID),
			Key:       omit.From("user_list_compact"),
			Value:     omit.From([]byte(fmt.Sprint(prefs.UserListCompact))),
			CreatedAt: omit.From(types.NewSQLiteDatetime(time.Now())),
			UpdatedAt: omit.From(types.NewSQLiteDatetime(time.Now())),
		}.Insert(),
	)

	if prefs.AssetListColumns != nil {
		val, err := json.Marshal(prefs.AssetListColumns)
		if err != nil {
			return nil, err
		}
		inserts = append(inserts,

			models.UserPreferenceSetter{
				UserID:    omit.From(userID),
				Key:       omit.From("asset_list_columns"),
				Value:     omit.From(val),
				CreatedAt: omit.From(types.NewSQLiteDatetime(time.Now())),
				UpdatedAt: omit.From(types.NewSQLiteDatetime(time.Now())),
			}.Insert(),
		)
	}

	if prefs.ThemeName != "" {
		inserts = append(inserts,
			models.UserPreferenceSetter{
				UserID:    omit.From(userID),
				Key:       omit.From("theme_name"),
				Value:     omit.From([]byte(prefs.ThemeName)),
				CreatedAt: omit.From(types.NewSQLiteDatetime(time.Now())),
				UpdatedAt: omit.From(types.NewSQLiteDatetime(time.Now())),
			}.Insert(),
		)
	}

	if prefs.ThemeMode != "" {
		inserts = append(inserts,
			models.UserPreferenceSetter{
				UserID:    omit.From(userID),
				Key:       omit.From("theme_mode"),
				Value:     omit.From([]byte(prefs.ThemeMode)),
				CreatedAt: omit.From(types.NewSQLiteDatetime(time.Now())),
				UpdatedAt: omit.From(types.NewSQLiteDatetime(time.Now())),
			}.Insert(),
		)
	}

	return inserts, nil
}

func mapUserPrefValueToEntity(ctx context.Context, prefs *auth.UserPreferences, pref *models.UserPreference) error {
	switch pref.Key {
	case "sidebar_closed_desktop":
		val, err := strconv.ParseBool(string(pref.Value))
		if err != nil {
			return err
		}
		prefs.SidebarClosedDesktop = val
		return nil
	case "asset_list_compact":
		val, err := strconv.ParseBool(string(pref.Value))
		if err != nil {
			return err
		}
		prefs.AssetListCompact = val
		return nil
	case "user_list_compact":
		val, err := strconv.ParseBool(string(pref.Value))
		if err != nil {
			return err
		}
		prefs.UserListCompact = val
		return nil
	case "theme_name":
		prefs.ThemeName = string(pref.Value)
		return nil
	case "theme_mode":
		prefs.ThemeMode = string(pref.Value)
		return nil
	case "asset_list_columns":
		err := json.Unmarshal(pref.Value, &prefs.AssetListColumns)
		return err
	}

	slog.WarnContext(ctx, fmt.Sprintf("unknown preference key: %s", pref.Key))

	return nil
}
