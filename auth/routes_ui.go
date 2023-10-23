package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/RobinThrift/stuff/server/session"
	"github.com/RobinThrift/stuff/views"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/form/v4"
)

type UIRouter struct {
	Control *Control
	Decoder *form.Decoder
}

const userForPasswordChangeKey = "user_for_password_change"

func (rt *UIRouter) RegisterRoutes(mux *chi.Mux) {
	mux.Get("/users", views.HTTPHandlerFuncErr(rt.handleUsersListGet))

	mux.Get("/users/new", views.HTTPHandlerFuncErr(rt.handleUsersNewGet))
	mux.Post("/users/new", views.HTTPHandlerFuncErr(rt.handleUsersNewPost))

	mux.Get("/users/me", views.HTTPHandlerFuncErr(rt.handleUsersMeGet))
	mux.Post("/users/me", views.HTTPHandlerFuncErr(rt.handleUsersMePost))
	mux.Get("/users/me/changepassword", views.HTTPHandlerFuncErr(rt.handleUsersMeInitChangePassword))

	mux.Get("/users/{id}/reset_password", views.HTTPHandlerFuncErr(rt.handleUsersResetPasswordGet))
	mux.Post("/users/{id}/reset_password", views.HTTPHandlerFuncErr(rt.handleUsersResetPasswordPost))
	mux.Get("/users/{id}/toggle_admin", views.HTTPHandlerFuncErr(rt.handleUsersToogleAdminGet))
	mux.Get("/users/{id}/delete", views.HTTPHandlerFuncErr(rt.handleUsersDeleteGet))
	mux.Post("/users/{id}/delete", views.HTTPHandlerFuncErr(rt.handleUsersDeletePost))

	mux.Post("/users/settings", logError(rt.postUserSettings))

	mux.Get("/login", views.HTTPHandlerFuncErr(rt.handleLoginGet))
	mux.Post("/login", views.HTTPHandlerFuncErr(rt.handleLoginPost))
	mux.Get("/logout", views.HTTPHandlerFuncErr(rt.handleLogoutGet))
	mux.Get("/auth/changepassword", views.HTTPHandlerFuncErr(rt.handleChangePasswordGet))
	mux.Post("/auth/changepassword", views.HTTPHandlerFuncErr(rt.handleChangePasswordPost))
}

// [GET] /users
func (rt *UIRouter) handleUsersListGet(w http.ResponseWriter, r *http.Request) error {
	query := newListUsersQuery(r.URL.Query())

	page, err := rt.Control.listUsers(r.Context(), query)
	if err != nil {
		return err
	}

	return renderListUsersPage(w, r, query, page)
}

// [GET] /users/new
func (rt *UIRouter) handleUsersNewGet(w http.ResponseWriter, r *http.Request) error {
	if user, ok := session.Get[*User](r.Context(), "user"); !ok {
		return errors.New("can't find user in session")
	} else if !user.IsAdmin {
		http.Redirect(w, r, "/", http.StatusFound)
	}

	return renderNewUserPage(w, r, NewUserPageViewModel{User: &User{}, ValidationErrs: map[string]string{}})
}

// [POST] /users/new
func (rt *UIRouter) handleUsersNewPost(w http.ResponseWriter, r *http.Request) error {
	if user, ok := session.Get[*User](r.Context(), "user"); !ok {
		return errors.New("can't find user in session")
	} else if !user.IsAdmin {
		http.Redirect(w, r, "/", http.StatusFound)
	}

	var user User

	err := rt.Decoder.Decode(&user, r.PostForm)
	if err != nil {
		return err
	}

	validationErrs := map[string]string{}

	if user.Username == "" {
		validationErrs["username"] = ErrUsernameEmpty.Error()
	}

	if user.DisplayName == "" {
		validationErrs["display_name"] = "Display Name must not be empty"
	}

	initPasswd := r.PostForm.Get("init_password")
	if initPasswd == "" {
		validationErrs["init_password"] = "Initial Password must not be empty"
	}

	if len(validationErrs) != 0 {
		return renderNewUserPage(w, r, NewUserPageViewModel{User: &user, ValidationErrs: validationErrs})
	}

	created, err := rt.Control.createLocalAuthUser(r.Context(), &user, initPasswd)
	if err != nil {
		return renderNewUserPage(w, r, NewUserPageViewModel{User: &user, ValidationErrs: map[string]string{"general": fmt.Sprintf("error creating user: %v", err)}})
	}

	session.Put(r.Context(), "info_message", fmt.Sprintf("created user %s (%d)", created.Username, created.ID))

	http.Redirect(w, r, "/users", http.StatusFound)
	return nil
}

// [GET] /users/me
func (rt *UIRouter) handleUsersMeGet(w http.ResponseWriter, r *http.Request) error {
	user, ok := session.Get[*User](r.Context(), "user")
	if !ok {
		return errors.New("can't find user in session")
	}

	return renderCurrentUserPage(w, r, CurrentUserPageViewModel{User: user, ValidationErrs: map[string]string{}})
}

// [POST] /users/me
func (rt *UIRouter) handleUsersMePost(w http.ResponseWriter, r *http.Request) error {
	user, ok := session.Get[*User](r.Context(), "user")
	if !ok {
		return errors.New("can't find user in session")
	}

	err := rt.Decoder.Decode(&user, r.PostForm)
	if err != nil {
		return err
	}

	validationErrs := map[string]string{}

	if user.Username == "" {
		validationErrs["username"] = ErrUsernameEmpty.Error()
	}

	if user.DisplayName == "" {
		validationErrs["display_name"] = "Display Name must not be empty"
	}

	if len(validationErrs) != 0 {
		return renderCurrentUserPage(w, r, CurrentUserPageViewModel{User: user, ValidationErrs: validationErrs})
	}

	updated, err := rt.Control.updateUser(r.Context(), user)
	if err != nil {
		return renderCurrentUserPage(w, r, CurrentUserPageViewModel{User: user, ValidationErrs: map[string]string{"general": fmt.Sprintf("error creating user: %v", err)}})
	}

	session.Put(r.Context(), "info_message", fmt.Sprintf("updated user %s (%d)", updated.Username, updated.ID))

	http.Redirect(w, r, "/users/me", http.StatusFound)
	return nil
}

// [GET] /users/me
func (rt *UIRouter) handleUsersMeInitChangePassword(w http.ResponseWriter, r *http.Request) error {
	user, ok := session.Get[*User](r.Context(), "user")
	if !ok {
		return errors.New("can't find user in session")
	}

	session.Put(r.Context(), userForPasswordChangeKey, user)
	http.Redirect(w, r, "/auth/changepassword", http.StatusFound)
	return nil
}

// [GET] /users/{id}/reset_password
func (rt *UIRouter) handleUsersResetPasswordGet(w http.ResponseWriter, r *http.Request) error {
	user, ok := session.Get[*User](r.Context(), "user")
	if !ok {
		return errors.New("can't find user in session")
	}

	if !user.IsAdmin {
		http.Redirect(w, r, "/", http.StatusFound)
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id == 0 {
		http.Redirect(w, r, "/users", http.StatusFound)
		return nil
	}

	if id == user.ID {
		session.Put(r.Context(), "info_message", "can't reset own password")
		http.Redirect(w, r, "/users", http.StatusFound)
		return nil
	}

	return renderResetPasswordPage(w, r, ResetPasswordPageViewModel{ValidationErrs: map[string]string{}})
}

// [POST] /users/{id}/reset_password
func (rt *UIRouter) handleUsersResetPasswordPost(w http.ResponseWriter, r *http.Request) error {
	user, ok := session.Get[*User](r.Context(), "user")
	if !ok {
		return errors.New("can't find user in session")
	}

	if !user.IsAdmin {
		http.Redirect(w, r, "/", http.StatusFound)
	}

	validationErrs := map[string]string{}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id == 0 {
		http.Redirect(w, r, "/users", http.StatusFound)
		return nil
	}

	if id == user.ID {
		validationErrs["general"] = "can't reset own password"
		return renderResetPasswordPage(w, r, ResetPasswordPageViewModel{ValidationErrs: validationErrs})
	}

	tmpPasswd := r.PostForm.Get("temp_password")
	if tmpPasswd == "" {
		validationErrs["temp_password"] = "Temporary password must not be empty"
		return renderResetPasswordPage(w, r, ResetPasswordPageViewModel{ValidationErrs: validationErrs})
	}

	updated, err := rt.Control.resetPassword(r.Context(), id, tmpPasswd)
	if err != nil {
		validationErrs["general"] = fmt.Sprintf("error updating user: %v", err)
		return renderResetPasswordPage(w, r, ResetPasswordPageViewModel{ValidationErrs: validationErrs})
	}

	session.Put(r.Context(), "info_message", fmt.Sprintf("reset password for user %s (%d)", updated.Username, updated.ID))

	http.Redirect(w, r, "/users", http.StatusFound)
	return nil
}

// [GET] /users/{id}/toggle_admin
func (rt *UIRouter) handleUsersToogleAdminGet(w http.ResponseWriter, r *http.Request) error {
	user, ok := session.Get[*User](r.Context(), "user")
	if !ok {
		return errors.New("can't find user in session")
	}

	if !user.IsAdmin {
		http.Redirect(w, r, "/", http.StatusFound)
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id == 0 {
		http.Redirect(w, r, "/users", http.StatusFound)
		return nil
	}

	if id == user.ID {
		session.Put(r.Context(), "info_message", "can't change own permissions")
		http.Redirect(w, r, "/users", http.StatusFound)
		return nil
	}

	updated, err := rt.Control.toggleAdmin(r.Context(), id)
	if err != nil {
		session.Put(r.Context(), "info_message", fmt.Sprintf("error updating user: %v", err))
		http.Redirect(w, r, "/users", http.StatusFound)
		return nil
	}

	session.Put(r.Context(), "info_message", fmt.Sprintf("updated user %s (%d)", updated.Username, updated.ID))

	http.Redirect(w, r, "/users", http.StatusFound)
	return nil
}

// [GET] /users/{id}/delete
func (rt *UIRouter) handleUsersDeleteGet(w http.ResponseWriter, r *http.Request) error {
	user, ok := session.Get[*User](r.Context(), "user")
	if !ok {
		return errors.New("can't find user in session")
	}

	if !user.IsAdmin {
		http.Redirect(w, r, "/", http.StatusFound)
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id == 0 {
		http.Redirect(w, r, "/users", http.StatusFound)
		return nil
	}

	if id == user.ID {
		session.Put(r.Context(), "info_message", "can't delete self")
		http.Redirect(w, r, "/users", http.StatusFound)
		return nil
	}

	toDelete, err := rt.Control.getUser(r.Context(), id)
	if err != nil {
		return err
	}

	return renderDeleteUserPage(w, r, DeleteUserPageViewModel{User: toDelete})
}

// [POST] /users/{id}/delete
func (rt *UIRouter) handleUsersDeletePost(w http.ResponseWriter, r *http.Request) error {
	user, ok := session.Get[*User](r.Context(), "user")
	if !ok {
		return errors.New("can't find user in session")
	}

	if !user.IsAdmin {
		http.Redirect(w, r, "/", http.StatusFound)
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id == 0 {
		http.Redirect(w, r, "/users", http.StatusFound)
		return nil
	}

	if id == user.ID {
		session.Put(r.Context(), "info_message", "can't delete self")
		http.Redirect(w, r, "/users", http.StatusFound)
		return nil
	}

	toDelete, err := rt.Control.getUser(r.Context(), id)
	if err != nil {
		return err
	}

	err = rt.Control.deleteLocalAuthUser(r.Context(), toDelete.ID)
	if err != nil {
		return err
	}

	session.Put(r.Context(), "info_message", "user "+toDelete.Username+" deleted")
	http.Redirect(w, r, "/users", http.StatusFound)
	return nil
}

// [POST] /users/settings
func (rt *UIRouter) postUserSettings(w http.ResponseWriter, r *http.Request) error {
	var payload struct {
		Sidebar *struct {
			Closed bool `json:"closed"`
		} `json:"sidebar,omitempty"`
		Assets *struct {
			Columns map[string]bool `json:"columns,omitempty"`
			Compact *bool           `json:"compact,omitempty"`
		} `json:"assetsList,omitempty"`
		Users *struct {
			Compact *bool `json:"compact,omitempty"`
		} `json:"usersList,omitempty"`
	}

	body, err := io.ReadAll(r.Body)
	defer func() {
		err = errors.Join(err, r.Body.Close())
	}()

	if err != nil {
		return fmt.Errorf("error reading request body: %w", err)
	}

	err = json.Unmarshal(body, &payload)
	if err != nil {
		return fmt.Errorf("error unmarshalling request body as JSON: %w", err)
	}

	if payload.Sidebar != nil {
		session.Put(r.Context(), "sidebar_closed", payload.Sidebar.Closed)
	}

	if payload.Assets != nil {
		if len(payload.Assets.Columns) != 0 {
			session.Put(r.Context(), "assets_list_columns", payload.Assets.Columns)
		}

		if payload.Assets.Compact != nil {
			session.Put(r.Context(), "assets_lists_compact", payload.Assets.Compact)
		}
	}

	if payload.Users != nil {
		if payload.Users.Compact != nil {
			session.Put(r.Context(), "users_lists_compact", payload.Users.Compact)
		}
	}

	return nil
}

// [GET] /login
func (rt *UIRouter) handleLoginGet(w http.ResponseWriter, r *http.Request) error {
	_, ok := session.Get[*User](r.Context(), "user")
	if ok {
		http.Redirect(w, r, "/", http.StatusFound)
		return nil
	}

	_, ok = session.Get[any](r.Context(), userForPasswordChangeKey)
	if ok {
		http.Redirect(w, r, "/auth/changepassword", http.StatusFound)
		return nil
	}

	return renderLoginPage(w, r, map[string]string{})
}

// [POST] /login
func (rt *UIRouter) handleLoginPost(w http.ResponseWriter, r *http.Request) error {
	user, validationErrs, err := rt.Control.getUserForCredentials(r.Context(), r.PostForm.Get("username"), r.PostForm.Get("password"))
	if err != nil {
		return err
	}

	err = session.RenewToken(r.Context())
	if err != nil {
		return err
	}

	if len(validationErrs) != 0 {
		return renderLoginPage(w, r, validationErrs)
	}

	if user.RequiresPasswordChange {
		session.Put(r.Context(), userForPasswordChangeKey, user)
		http.Redirect(w, r, "/auth/changepassword", http.StatusFound)
		return nil
	}

	session.Put(r.Context(), "user", user)
	session.Put(r.Context(), "user_is_admin", user.IsAdmin)

	http.Redirect(w, r, "/", http.StatusFound)
	return nil
}

// [GET] /logout
func (rt *UIRouter) handleLogoutGet(w http.ResponseWriter, r *http.Request) error {
	err := session.Destroy(r.Context())
	if err != nil {
		return fmt.Errorf("error destroying sessions: %w", err)
	}

	http.Redirect(w, r, "/", http.StatusFound)
	return nil
}

// [GET] /auth/changepassword
func (rt *UIRouter) handleChangePasswordGet(w http.ResponseWriter, r *http.Request) error {
	_, ok := session.Get[*User](r.Context(), userForPasswordChangeKey)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return nil
	}

	return renderChangePasswordPage(w, r, map[string]string{})
}

// [POST] /auth/changepassword
func (rt *UIRouter) handleChangePasswordPost(w http.ResponseWriter, r *http.Request) error {
	user, ok := session.Get[*User](r.Context(), userForPasswordChangeKey)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return nil
	}

	validationErrs, err := rt.Control.changeUserCredentials(r.Context(), changeUserCredentialsCmd{
		user:                     user,
		currPasswdPlaintext:      r.PostForm.Get("current_password"),
		newPasswdPlaintext:       r.PostForm.Get("new_password"),
		newPasswdRepeatPlaintext: r.PostForm.Get("new_password_repeat"),
	})

	if err != nil {
		return err
	}

	err = session.RenewToken(r.Context())
	if err != nil {
		return fmt.Errorf("error renewing session token: %w", err)
	}

	if len(validationErrs) != 0 {
		return renderChangePasswordPage(w, r, validationErrs)
	}

	err = session.Destroy(r.Context())
	if err != nil {
		return fmt.Errorf("error destroying session: %w", err)
	}

	http.Redirect(w, r, "/login", http.StatusFound)

	return nil
}

func logError(handler func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := handler(w, r)
		if err != nil {
			slog.ErrorContext(r.Context(), r.URL.Path, "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			_, writeErr := w.Write([]byte(err.Error()))
			if writeErr != nil {
				slog.ErrorContext(r.Context(), "error writing response", "error", writeErr)
			}
		}
	})
}

type ListUsersQuery struct {
	Search   string
	Page     int
	PageSize int
	OrderBy  string
	OrderDir string
}

func newListUsersQuery(params url.Values) ListUsersQuery {
	q := ListUsersQuery{ //nolint: varnamelen
		Search:  params.Get("q"),
		OrderBy: params.Get("order_by"),
	}

	if size := params.Get("page_size"); size != "" {
		q.PageSize, _ = strconv.Atoi(size)
	}

	if pageStr := params.Get("page"); pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err == nil {
			q.Page = q.PageSize * page
		}
	}

	if orderDir := params.Get("order_dir"); orderDir != "" {
		orderDir = strings.ToUpper(orderDir)
		if orderDir == "ASC" || orderDir == "DESC" {
			q.OrderDir = orderDir
		}
	}

	if q.PageSize == 0 {
		q.PageSize = 25
	}

	if q.PageSize > 100 {
		q.PageSize = 100
	}

	return q
}

func NewDecoder() *form.Decoder {
	decoder := form.NewDecoder()

	decoder.SetMode(form.ModeExplicit)
	return decoder
}
