package views

import (
	"net/http"
	"net/url"

	"github.com/RobinThrift/stuff"
	"github.com/RobinThrift/stuff/server/session"
	"github.com/gorilla/csrf"
)

type Model[D any] struct {
	Global Global
	Data   D
}

type Global struct {
	Title              string
	CSRFToken          string
	FlashMessage       string
	Search             string
	CurrentPage        string
	SidebarClosed      bool
	CurrentUserIsAdmin bool
	Version            string
	Referer            string
}

func NewGlobal(title string, r *http.Request) Global {
	infomsg, _ := session.Pop[string](r.Context(), "info_message")

	sidebarClosed, _ := session.Get[bool](r.Context(), "sidebar_closed")

	currentUserIsAdmin, _ := session.Get[bool](r.Context(), "user_is_admin")

	var referer string
	if refHeaderURL, err := url.Parse(r.Header.Get("Referer")); err == nil {
		// @TODO: fix with better fallback
		if r.URL.Host == refHeaderURL.Host || r.URL.Host == "" {
			referer = r.Header.Get("Referer")
		}
	}

	return Global{
		Title:              title,
		CSRFToken:          csrf.Token(r),
		FlashMessage:       infomsg,
		CurrentPage:        r.URL.Path,
		SidebarClosed:      sidebarClosed,
		CurrentUserIsAdmin: currentUserIsAdmin,
		Version:            stuff.Version,
		Referer:            referer,
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
