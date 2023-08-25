package auth

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/csrf"
	"github.com/kodeshack/stuff/server/session"
	"github.com/kodeshack/stuff/users"
	"github.com/kodeshack/stuff/views"
)

type Router struct {
	Control *Control
}

const userForPasswordChangeKey = "user_for_password_change"

func (rt *Router) RegisterRoutes(mux *chi.Mux) {
	mux.Get("/login", views.HTTPHandlerFuncErr(rt.handleLoginGet))
	mux.Post("/login", views.HTTPHandlerFuncErr(rt.handleLoginPost))
	mux.Post("/logout", views.HTTPHandlerFuncErr(rt.handleLogoutGet))
	mux.Get("/auth/changepassword", views.HTTPHandlerFuncErr(rt.handleChangePasswordGet))
	mux.Post("/auth/changepassword", views.HTTPHandlerFuncErr(rt.handleChangePasswordPost))
}

// [GET] /login
func (rt *Router) handleLoginGet(w http.ResponseWriter, r *http.Request) error {
	_, ok := session.Get[any](r.Context(), userForPasswordChangeKey)
	if ok {
		http.Redirect(w, r, "/auth/changepassword", http.StatusFound)
		return nil
	}

	return renderLoginPage(w, r, map[string]string{})
}

// [POST] /login
func (rt *Router) handleLoginPost(w http.ResponseWriter, r *http.Request) error {
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

	http.Redirect(w, r, "/", http.StatusFound)
	return nil
}

// [GET] /logout
func (rt *Router) handleLogoutGet(w http.ResponseWriter, r *http.Request) error {
	err := session.Destroy(r.Context())
	if err != nil {
		return fmt.Errorf("error destroying sessions: %w", err)
	}

	http.Redirect(w, r, "/", http.StatusFound)
	return nil
}

// [GET] /auth/changepassword
func (rt *Router) handleChangePasswordGet(w http.ResponseWriter, r *http.Request) error {
	_, ok := session.Get[*users.User](r.Context(), userForPasswordChangeKey)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return nil
	}

	return renderChangePasswordPage(w, r, map[string]string{})
}

// [POST] /auth/changepassword
func (rt *Router) handleChangePasswordPost(w http.ResponseWriter, r *http.Request) error {
	user, ok := session.Get[*users.User](r.Context(), userForPasswordChangeKey)
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

func renderLoginPage(w http.ResponseWriter, r *http.Request, validationErrors map[string]string) error {
	loginPage := views.LoginPage(csrf.Token(r), validationErrors)
	page := views.Document("Login", loginPage)

	err := page.Render(r.Context(), w)
	if err != nil {
		return fmt.Errorf("error rendering login page: %w", err)
	}

	return nil
}

func renderChangePasswordPage(w http.ResponseWriter, r *http.Request, validationErrors map[string]string) error {
	changePasswordPage := views.ChangePasswordPage(csrf.Token(r), validationErrors)
	page := views.Document("Change Password", changePasswordPage)

	err := page.Render(r.Context(), w)
	if err != nil {
		return fmt.Errorf("error rendering change password page: %w", err)
	}

	return nil
}
