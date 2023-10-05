package auth

import (
	"fmt"
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/kodeshack/stuff/server/session"
	"github.com/kodeshack/stuff/views"
)

type ListUsersPageViewModel struct {
	Page    *UserListPage
	Query   ListUsersQuery
	Compact bool
}

func renderListUsersPage(w http.ResponseWriter, r *http.Request, query ListUsersQuery, page *UserListPage) error {
	global := views.NewGlobal("Users", r)

	compact, _ := session.Get[bool](r.Context(), "users_lists_compact")

	err := views.Render(w, "users_list_page", views.Model[ListUsersPageViewModel]{
		Global: global,
		Data: ListUsersPageViewModel{
			Page:    page,
			Query:   query,
			Compact: compact,
		},
	})
	if err != nil {
		return fmt.Errorf("error rendering list users page: %w", err)
	}

	return nil
}

type NewUserPageViewModel struct {
	User           *User
	ValidationErrs map[string]string
}

func renderNewUserPage(w http.ResponseWriter, r *http.Request, model NewUserPageViewModel) error {
	csrfErr, ok := session.Pop[string](r.Context(), "csrf_error")
	if ok {
		model.ValidationErrs["general"] = csrfErr
	}

	err := views.Render(w, "users_new_page", views.Model[NewUserPageViewModel]{
		Global: views.NewGlobal("New User", r),
		Data:   model,
	})
	if err != nil {
		return fmt.Errorf("error rendering new user page: %w", err)
	}

	return nil
}

type CurrentUserPageViewModel struct {
	User           *User
	ValidationErrs map[string]string
}

func renderCurrentUserPage(w http.ResponseWriter, r *http.Request, model CurrentUserPageViewModel) error {
	csrfErr, ok := session.Pop[string](r.Context(), "csrf_error")
	if ok {
		model.ValidationErrs["general"] = csrfErr
	}

	err := views.Render(w, "users_me_page", views.Model[CurrentUserPageViewModel]{
		Global: views.NewGlobal(model.User.DisplayName+" Settings", r),
		Data:   model,
	})
	if err != nil {
		return fmt.Errorf("error rendering user me page: %w", err)
	}

	return nil
}

type ResetPasswordPageViewModel struct {
	ValidationErrs map[string]string
}

func renderResetPasswordPage(w http.ResponseWriter, r *http.Request, model ResetPasswordPageViewModel) error {
	csrfErr, ok := session.Pop[string](r.Context(), "csrf_error")
	if ok {
		model.ValidationErrs["general"] = csrfErr
	}

	err := views.Render(w, "users_reset_pw_page", views.Model[ResetPasswordPageViewModel]{
		Global: views.NewGlobal("Reset Password", r),
		Data:   model,
	})
	if err != nil {
		return fmt.Errorf("error rendering reset password page: %w", err)
	}

	return nil
}

type DeleteUserPageViewModel struct {
	User *User
}

func renderDeleteUserPage(w http.ResponseWriter, r *http.Request, model DeleteUserPageViewModel) error {
	err := views.Render(w, "users_delete_page", views.Model[DeleteUserPageViewModel]{
		Global: views.NewGlobal("Delete "+model.User.Username, r),
		Data:   model,
	})
	if err != nil {
		return fmt.Errorf("error rendering user delete page: %w", err)
	}

	return nil
}

func renderLoginPage(w http.ResponseWriter, r *http.Request, validationErrors map[string]string) error {
	csrfErr, ok := session.Pop[string](r.Context(), "csrf_error")
	if ok {
		validationErrors["general"] = csrfErr
	}

	err := views.RenderLoginPage(w, views.Model[views.LoginPageViewModel]{
		Global: views.Global{
			Title:     "Login",
			CSRFToken: csrf.Token(r),
		},
		Data: views.LoginPageViewModel{
			ValidationErrs: validationErrors,
		},
	})
	if err != nil {
		return fmt.Errorf("error rendering login page: %w", err)
	}

	return nil
}

func renderChangePasswordPage(w http.ResponseWriter, r *http.Request, validationErrors map[string]string) error {
	csrfErr, ok := session.Pop[string](r.Context(), "csrf_error")
	if ok {
		validationErrors["general"] = csrfErr
	}

	err := views.RenderChangePasswordPage(w, views.Model[views.ChangePasswordPageViewModel]{
		Global: views.Global{
			Title:     "Change Password",
			CSRFToken: csrf.Token(r),
		},
		Data: views.ChangePasswordPageViewModel{
			ValidationErrs: validationErrors,
		},
	})
	if err != nil {
		return fmt.Errorf("error rendering change password page: %w", err)
	}

	return nil
}
