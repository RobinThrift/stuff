package htmlui

import (
	"fmt"
	"net/http"

	"github.com/RobinThrift/stuff/auth"
	"github.com/RobinThrift/stuff/control"
	"github.com/RobinThrift/stuff/internal/server/session"
	"github.com/RobinThrift/stuff/views/pages"
)

const userForPasswordChangeKey = "user_for_password_change"

// [GET] /login
func (rt *Router) authLoginHandler(w http.ResponseWriter, r *http.Request, params struct{}) error {
	_, ok := session.Get[*auth.User](r.Context(), "user")
	if ok {
		http.Redirect(w, r, "/", http.StatusFound)
		return nil
	}

	_, ok = session.Get[any](r.Context(), userForPasswordChangeKey)
	if ok {
		http.Redirect(w, r, "/auth/changepassword", http.StatusFound)
		return nil
	}

	page := pages.LoginPage{ValidationErrs: map[string]string{}}
	return page.Render(w, r)
}

// [POST] /login
func (rt *Router) authLoginSubmitHandler(w http.ResponseWriter, r *http.Request, params struct{}) error {
	user, validationErrs, err := rt.auth.GetUserForCredentials(r.Context(), control.GetUserForCredentialsQuery{
		Username: r.PostForm.Get("username"), PlaintextPasswd: r.PostForm.Get("password"),
	})
	if err != nil {
		return err
	}

	err = session.RenewToken(r.Context())
	if err != nil {
		return err
	}

	page := pages.LoginPage{ValidationErrs: validationErrs}

	if len(validationErrs) != 0 {
		return page.Render(w, r)
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
func (rt *Router) authLogoutHandler(w http.ResponseWriter, r *http.Request, params struct{}) error {
	err := session.Destroy(r.Context())
	if err != nil {
		return fmt.Errorf("error destroying sessions: %w", err)
	}

	http.Redirect(w, r, "/", http.StatusFound)
	return nil
}

// [GET] /auth/changepassword
func (rt *Router) authChangePasswordHandler(w http.ResponseWriter, r *http.Request, params struct{}) error {
	_, ok := session.Get[*auth.User](r.Context(), userForPasswordChangeKey)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return nil
	}

	page := pages.ChangePasswordPage{ValidationErrs: map[string]string{}}
	return page.Render(w, r)
}

// [POST] /auth/changepassword
func (rt *Router) authChangePasswordSubmitHandler(w http.ResponseWriter, r *http.Request, params struct{}) error {
	user, ok := session.Get[*auth.User](r.Context(), userForPasswordChangeKey)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return nil
	}

	var err error
	page := pages.ChangePasswordPage{}

	page.ValidationErrs, err = rt.auth.ChangeUserCredentials(r.Context(), control.ChangeUserCredentialsCmd{
		User:                     user,
		CurrPasswdPlaintext:      r.PostForm.Get("current_password"),
		NewPasswdPlaintext:       r.PostForm.Get("new_password"),
		NewPasswdRepeatPlaintext: r.PostForm.Get("new_password_repeat"),
	})

	if err != nil {
		return err
	}

	err = session.RenewToken(r.Context())
	if err != nil {
		return fmt.Errorf("error renewing session token: %w", err)
	}

	if len(page.ValidationErrs) != 0 {
		return page.Render(w, r)
	}

	err = session.Destroy(r.Context())
	if err != nil {
		return fmt.Errorf("error destroying session: %w", err)
	}

	http.Redirect(w, r, "/login", http.StatusFound)

	return nil
}
