package views

import (
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/kodeshack/stuff/server/session"
)

type Model[D any] struct {
	Global Global
	Data   D
}

type Global struct {
	Title         string
	CSRFToken     string
	FlashMessage  string
	Search        string
	CurrentPage   string
	SidebarClosed bool
}

func NewGlobal(title string, r *http.Request) Global {
	infomsg, _ := session.Pop[string](r.Context(), "info_message")

	sidebarClosed, _ := session.Get[bool](r.Context(), "sidebar_closed")

	return Global{
		Title:         title,
		CSRFToken:     csrf.Token(r),
		FlashMessage:  infomsg,
		CurrentPage:   r.URL.Path,
		SidebarClosed: sidebarClosed,
	}
}

type DocumentViewModel struct {
	Global Global
	Body   string
	Data   any
}

type LoginPageViewModel struct {
	ValidationErrs map[string]string
}

type ChangePasswordPageViewModel struct {
	ValidationErrs map[string]string
}
