package pages

import (
	"net/http"

	"github.com/RobinThrift/stuff/internal/server/session"
	"github.com/RobinThrift/stuff/views"
)

type LoginPage struct {
	ValidationErrs map[string]string
}

func (p *LoginPage) Render(w http.ResponseWriter, r *http.Request) error {
	csrfErr, ok := session.Pop[string](r.Context(), "csrf_error")
	if ok {
		p.ValidationErrs["general"] = csrfErr
	}

	return views.Render(w, "login_page", views.Model[*LoginPage]{
		Global: views.NewGlobal("Login", r),
		Data:   p,
	})
}

type ChangePasswordPage struct {
	ValidationErrs map[string]string
}

func (p *ChangePasswordPage) Render(w http.ResponseWriter, r *http.Request) error {
	csrfErr, ok := session.Pop[string](r.Context(), "csrf_error")
	if ok {
		p.ValidationErrs["general"] = csrfErr
	}

	return views.Render(w, "change_password_page", views.Model[*ChangePasswordPage]{
		Global: views.NewGlobal("Change Password", r),
		Data:   p,
	})
}

type AuthPasswordResetPage struct {
	ValidationErrs map[string]string
}

func (p *AuthPasswordResetPage) Render(w http.ResponseWriter, r *http.Request) error {
	csrfErr, ok := session.Pop[string](r.Context(), "csrf_error")
	if ok {
		p.ValidationErrs["general"] = csrfErr
	}

	return views.Render(w, "users_reset_pw_page", views.Model[*AuthPasswordResetPage]{
		Global: views.NewGlobal("Reset Password", r),
		Data:   p,
	})
}
