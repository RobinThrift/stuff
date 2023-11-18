package htmlui

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/RobinThrift/stuff/auth"
	"github.com/RobinThrift/stuff/control"
	"github.com/RobinThrift/stuff/internal/server/session"
	"github.com/RobinThrift/stuff/views"
	"github.com/RobinThrift/stuff/views/pages"
)

type usersListParams struct {
	Query    string `query:"query"`
	Page     int    `query:"page"`
	PageSize int    `query:"page_size"`
	OrderBy  string `query:"order_by"`
	OrderDir string `query:"order_dir"`
}

// [GET] /users
func (rt *Router) usersListHandler(w http.ResponseWriter, r *http.Request, params usersListParams) error {
	if params.PageSize == 0 {
		params.PageSize = 25
	}

	users, err := rt.users.List(r.Context(), control.ListUsersQuery{
		Search:   params.Query,
		Page:     params.Page,
		PageSize: params.PageSize,
	})
	if err != nil {
		return err
	}

	page := pages.UsersListPage{Users: &views.Pagination[*auth.User]{
		ListPage: users,
		URL:      r.URL,
	}}

	return page.Render(w, r)
}

// [GET] /users/new
func (rt *Router) usersNewHandler(w http.ResponseWriter, r *http.Request, params struct{}) error {
	if user, ok := session.Get[*auth.User](r.Context(), "user"); !ok {
		return errors.New("can't find user in session")
	} else if !user.IsAdmin {
		http.Redirect(w, r, "/", http.StatusFound)
	}

	page := pages.UsersNewPage{User: &auth.User{}, ValidationErrs: map[string]string{}}
	return page.Render(w, r)
}

// [POST] /users/new
func (rt *Router) usersNewSubmitHandler(w http.ResponseWriter, r *http.Request, params struct{}) error {
	if user, ok := session.Get[*auth.User](r.Context(), "user"); !ok {
		return errors.New("can't find user in session")
	} else if !user.IsAdmin {
		http.Redirect(w, r, "/", http.StatusFound)
	}

	page := pages.UsersNewPage{User: &auth.User{}, ValidationErrs: map[string]string{}}

	err := rt.forms.Decode(page.User, r.PostForm)
	if err != nil {
		return err
	}

	if page.User.Username == "" {
		page.ValidationErrs["username"] = control.ErrUsernameEmpty.Error()
	}

	if page.User.DisplayName == "" {
		page.ValidationErrs["display_name"] = "Display Name must not be empty"
	}

	initPasswd := r.PostForm.Get("init_password")
	if initPasswd == "" {
		page.ValidationErrs["init_password"] = "Initial Password must not be empty"
	}

	if len(page.ValidationErrs) != 0 {
		return page.Render(w, r)
	}

	err = rt.auth.CreateUser(r.Context(), control.CreateUserCmd{User: page.User, PlaintextPasswd: initPasswd})
	if err != nil {
		slog.ErrorContext(r.Context(), "error creating user", "error", err)
		page.ValidationErrs["general"] = fmt.Sprintf("error creating user: %v", err)
		return page.Render(w, r)
	}

	views.SetFlashMessage(r.Context(), views.FlashMessageSuccess, fmt.Sprintf("Created user %s", page.User.Username))

	http.Redirect(w, r, "/users", http.StatusFound)
	return nil
}

// [GET] /users/me
func (rt *Router) usersCurrentHandler(w http.ResponseWriter, r *http.Request, params struct{}) error {
	user, ok := session.Get[*auth.User](r.Context(), "user")
	if !ok {
		return errors.New("can't find user in session")
	}

	page := pages.UsersCurrentPage{User: user, ValidationErrs: map[string]string{}}
	return page.Render(w, r)
}

// [POST] /users/me
func (rt *Router) usersCurrentSubmitHandler(w http.ResponseWriter, r *http.Request, params struct{}) error {
	user, ok := session.Get[*auth.User](r.Context(), "user")
	if !ok {
		return errors.New("can't find user in session")
	}

	page := pages.UsersCurrentPage{User: user, ValidationErrs: map[string]string{}}

	err := rt.forms.Decode(page.User, r.PostForm)
	if err != nil {
		return err
	}

	if page.User.DisplayName == "" {
		page.ValidationErrs["display_name"] = "Display Name must not be empty"
	}

	if len(page.ValidationErrs) != 0 {
		page.User, ok = session.Get[*auth.User](r.Context(), "user")
		if !ok {
			return errors.New("can't find user in session")
		}

		return page.Render(w, r)
	}

	err = rt.users.Update(r.Context(), page.User)
	if err != nil {
		slog.ErrorContext(r.Context(), "error creating user", "error", err)
		page.ValidationErrs["general"] = fmt.Sprintf("error creating user: %v", err)
		return page.Render(w, r)
	}

	views.SetFlashMessage(r.Context(), views.FlashMessageSuccess, fmt.Sprintf("Updated user %s", page.User.Username))

	http.Redirect(w, r, "/users/me", http.StatusFound)
	return nil
}

// [GET] /users/me
func (rt *Router) usersCurrentInitChangePasswordHandler(w http.ResponseWriter, r *http.Request, params struct{}) error {
	user, ok := session.Get[*auth.User](r.Context(), "user")
	if !ok {
		return errors.New("can't find user in session")
	}

	session.Put(r.Context(), userForPasswordChangeKey, user)
	http.Redirect(w, r, "/auth/changepassword", http.StatusFound)
	return nil
}

type usersResetPasswordParams struct {
	ID int64 `url:"id"`
}

// [GET] /users/{id}/reset_password
func (rt *Router) usersResetPasswordHandler(w http.ResponseWriter, r *http.Request, params usersResetPasswordParams) error {
	user, ok := session.Get[*auth.User](r.Context(), "user")
	if !ok {
		return errors.New("can't find user in session")
	}

	if !user.IsAdmin {
		http.Redirect(w, r, "/", http.StatusFound)
	}

	if params.ID == user.ID {
		views.SetFlashMessage(r.Context(), views.FlashMessageError, "Can't reset own password")
		http.Redirect(w, r, "/users", http.StatusFound)
		return nil
	}

	page := pages.AuthPasswordResetPage{ValidationErrs: map[string]string{}}
	return page.Render(w, r)
}

// [POST] /users/{id}/reset_password
func (rt *Router) usersResetPasswordSubmitHandler(w http.ResponseWriter, r *http.Request, params usersResetPasswordParams) error {
	user, ok := session.Get[*auth.User](r.Context(), "user")
	if !ok {
		return errors.New("can't find user in session")
	}

	if !user.IsAdmin {
		http.Redirect(w, r, "/", http.StatusFound)
	}

	page := pages.AuthPasswordResetPage{ValidationErrs: map[string]string{}}

	if params.ID == user.ID {
		page.ValidationErrs["general"] = "can't reset own password"
		return page.Render(w, r)
	}

	tmpPasswd := r.PostForm.Get("temp_password")
	if tmpPasswd == "" {
		page.ValidationErrs["temp_password"] = "Temporary password must not be empty"
		return page.Render(w, r)
	}

	err := rt.auth.ResetPassword(r.Context(), control.ResetPasswordCmd{UserID: params.ID, PlaintextPasswd: tmpPasswd})
	if err != nil {
		page.ValidationErrs["general"] = fmt.Sprintf("error updating user: %v", err)
		return page.Render(w, r)
	}

	updated, err := rt.users.Get(r.Context(), params.ID)
	if err != nil {
		return err
	}

	views.SetFlashMessage(r.Context(), views.FlashMessageSuccess, fmt.Sprintf("Reset password for user %s (%d)", updated.Username, updated.ID))

	http.Redirect(w, r, "/users", http.StatusFound)
	return nil
}

type usersToggleAdminParams struct {
	ID int64 `url:"id"`
}

// [GET] /users/{id}/toggle_admin
func (rt *Router) usersToggleAdminHandler(w http.ResponseWriter, r *http.Request, params usersToggleAdminParams) error {
	user, ok := session.Get[*auth.User](r.Context(), "user")
	if !ok {
		return errors.New("can't find user in session")
	}

	if !user.IsAdmin {
		http.Redirect(w, r, "/", http.StatusFound)
	}

	if params.ID == user.ID {
		views.SetFlashMessage(r.Context(), views.FlashMessageError, "Can't change own permissions")
		http.Redirect(w, r, "/users", http.StatusFound)
		return nil
	}

	err := rt.auth.ToggleAdmin(r.Context(), params.ID)
	if err != nil {
		views.SetFlashMessage(r.Context(), views.FlashMessageError, fmt.Sprintf("Error updating user: %v", err))
		http.Redirect(w, r, "/users", http.StatusFound)
		return nil
	}

	updated, err := rt.users.Get(r.Context(), params.ID)
	if err != nil {
		return err
	}

	views.SetFlashMessage(r.Context(), views.FlashMessageSuccess, fmt.Sprintf("Updated user %s (%d)", updated.Username, updated.ID))

	http.Redirect(w, r, "/users", http.StatusFound)
	return nil
}

type usersDeleteParams struct {
	ID int64 `url:"id"`
}

// [GET] /users/{id}/delete
func (rt *Router) usersDeleteHandler(w http.ResponseWriter, r *http.Request, params usersDeleteParams) error {
	user, ok := session.Get[*auth.User](r.Context(), "user")
	if !ok {
		return errors.New("can't find user in session")
	}

	if !user.IsAdmin {
		http.Redirect(w, r, "/", http.StatusFound)
	}

	if params.ID == user.ID {
		views.SetFlashMessage(r.Context(), views.FlashMessageError, "Can't delete self")
		http.Redirect(w, r, "/users", http.StatusFound)
		return nil
	}

	toDelete, err := rt.users.Get(r.Context(), params.ID)
	if err != nil {
		return err
	}

	page := pages.UserDeletePage{User: toDelete, ValidationErrs: map[string]string{}}
	return page.Render(w, r)
}

// [POST] /users/{id}/delete
func (rt *Router) usersDeleteSubmitHandler(w http.ResponseWriter, r *http.Request, params usersDeleteParams) error {
	user, ok := session.Get[*auth.User](r.Context(), "user")
	if !ok {
		return errors.New("can't find user in session")
	}

	if !user.IsAdmin {
		http.Redirect(w, r, "/", http.StatusFound)
	}

	if params.ID == user.ID {
		views.SetFlashMessage(r.Context(), views.FlashMessageError, "Can't delete self")
		http.Redirect(w, r, "/users", http.StatusFound)
		return nil
	}

	toDelete, err := rt.users.Get(r.Context(), params.ID)
	if err != nil {
		return err
	}

	err = rt.auth.DeleteUser(r.Context(), toDelete.ID)
	if err != nil {
		return err
	}

	views.SetFlashMessage(r.Context(), views.FlashMessageSuccess, "User "+toDelete.Username+" deleted")
	http.Redirect(w, r, "/users", http.StatusFound)
	return nil
}

// [POST] /users/settings
func (rt *Router) usersSettingsSubmitHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := session.Get[*auth.User](r.Context(), "user")
	if !ok {
		slog.ErrorContext(r.Context(), "can't find user in session")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	payload, err := decodeUserSettingsBody(r)
	if err != nil {
		slog.ErrorContext(r.Context(), "error reading request body", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if payload.ThemeName != nil {
		user.Preferences.ThemeName = *payload.ThemeName
	}

	if payload.ThemeMode != nil {
		user.Preferences.ThemeMode = *payload.ThemeMode
	}

	if payload.SidebarClosedDesktop != nil {
		user.Preferences.SidebarClosedDesktop = *payload.SidebarClosedDesktop
	}

	if len(payload.AssetListColumns) != 0 {
		user.Preferences.AssetListColumns = payload.AssetListColumns
	}

	if payload.AssetListCompact != nil {
		user.Preferences.AssetListCompact = *payload.AssetListCompact
	}

	if payload.UserListCompact != nil {
		user.Preferences.UserListCompact = *payload.UserListCompact
	}

	err = rt.users.SetUserPreferences(r.Context(), user)
	if err != nil {
		slog.ErrorContext(r.Context(), "error setting user preferences", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	session.Put(r.Context(), "user", user)
}

type userSettingsBody struct {
	ThemeName *string `json:"theme_name,omitempty"`
	ThemeMode *string `json:"theme_mode,omitempty"`

	SidebarClosedDesktop *bool `json:"sidebar_closed_desktop,omitempty"`

	AssetListColumns []string `json:"asset_list_columns,omitempty"`
	AssetListCompact *bool    `json:"asset_list_compact,omitempty"`

	UserListCompact *bool `json:"user_list_compact,omitempty"`
}

func decodeUserSettingsBody(r *http.Request) (payload *userSettingsBody, err error) {
	body, err := io.ReadAll(r.Body)
	defer func() {
		err = errors.Join(err, r.Body.Close())
	}()

	if err != nil {
		return
	}

	err = json.Unmarshal(body, &payload)
	if err != nil {
		err = fmt.Errorf("error unmarshalling request body as JSON: %w", err)
		return
	}

	return
}
