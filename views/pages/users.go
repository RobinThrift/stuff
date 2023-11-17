package pages

import (
	"net/http"

	"github.com/RobinThrift/stuff/auth"
	"github.com/RobinThrift/stuff/internal/server/session"
	"github.com/RobinThrift/stuff/views"
)

type UsersListPage struct {
	Users *views.Pagination[*auth.User]
}

func (p *UsersListPage) Render(w http.ResponseWriter, r *http.Request) error {
	return views.Render(w, "users_list_page", views.Model[*UsersListPage]{
		Global: views.NewGlobal("Users", r),
		Data:   p,
	})
}

type UsersNewPage struct {
	User           *auth.User
	ValidationErrs map[string]string
}

func (p *UsersNewPage) Render(w http.ResponseWriter, r *http.Request) error {
	csrfErr, ok := session.Pop[string](r.Context(), "csrf_error")
	if ok {
		p.ValidationErrs["general"] = csrfErr
	}

	return views.Render(w, "users_new_page", views.Model[*UsersNewPage]{
		Global: views.NewGlobal("New User", r),
		Data:   p,
	})
}

type UsersCurrentPage struct {
	User           *auth.User
	ValidationErrs map[string]string
}

func (p *UsersCurrentPage) Render(w http.ResponseWriter, r *http.Request) error {
	csrfErr, ok := session.Pop[string](r.Context(), "csrf_error")
	if ok {
		p.ValidationErrs["general"] = csrfErr
	}

	return views.Render(w, "users_me_page", views.Model[*UsersCurrentPage]{
		Global: views.NewGlobal(p.User.DisplayName+" Settings", r),
		Data:   p,
	})
}

type UserDeletePage struct {
	User           *auth.User
	ValidationErrs map[string]string
}

func (p *UserDeletePage) Render(w http.ResponseWriter, r *http.Request) error {
	csrfErr, ok := session.Pop[string](r.Context(), "csrf_error")
	if ok {
		p.ValidationErrs["general"] = csrfErr
	}

	return views.Render(w, "users_delete_page", views.Model[*UserDeletePage]{
		Global: views.NewGlobal("Delete "+p.User.DisplayName, r),
		Data:   p,
	})
}
