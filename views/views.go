package views

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/RobinThrift/stuff"
	"github.com/RobinThrift/stuff/internal/server/session"
	"github.com/gorilla/csrf"
)

type Model[D any] struct {
	Global Global
	Data   D
}

type Global struct {
	Title              string
	CSRFToken          string
	FlashMessage       FlashMessage
	Search             string
	CurrentPage        string
	SidebarClosed      bool
	CurrentUserIsAdmin bool
	Version            string
	Referer            string
}

type FlashMessage struct {
	Type FlashMessageType
	Text string
}

type FlashMessageType string

const (
	FlashMessageSuccess FlashMessageType = "success"
	FlashMessageError   FlashMessageType = "error"
)

func NewGlobal(title string, r *http.Request) Global {
	flashMessage, _ := session.Pop[FlashMessage](r.Context(), "flash_message")

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
		FlashMessage:       flashMessage,
		CurrentPage:        r.URL.Path,
		SidebarClosed:      sidebarClosed,
		CurrentUserIsAdmin: currentUserIsAdmin,
		Version:            stuff.Version,
		Referer:            referer,
	}
}

func SetFlashMessage(ctx context.Context, typ FlashMessageType, text string) {
	session.Put(ctx, "flash_message", FlashMessage{Type: typ, Text: text})
}

type ErrorPageErr struct {
	Err     error
	Code    int
	Title   string
	Message string
}

func (err ErrorPageErr) Error() string {
	return err.Err.Error()
}

func RenderErrorPage(w http.ResponseWriter, r *http.Request, err error) {
	var errPageErr ErrorPageErr

	if !errors.As(err, &errPageErr) {
		errPageErr = ErrorPageErr{
			Code:    http.StatusInternalServerError,
			Title:   "Unknown Error",
			Message: err.Error(),
		}
	}

	if errPageErr.Title == "" && errPageErr.Err != nil {
		errPageErr.Title = errPageErr.Err.Error()
	}

	renderErr := Render(w, "error_page", Model[ErrorPageErr]{
		Global: Global{
			Title: errPageErr.Title,
		},
		Data: errPageErr,
	})
	if renderErr != nil {
		slog.ErrorContext(r.Context(), "error rendering error page", "error", renderErr)
	}
}

type DocumentViewModel struct {
	Global Global
	Body   string
	Data   any
}

func Render[D any](w io.Writer, name string, data Model[D]) error {
	b := bytes.NewBuffer(nil)

	if wb, ok := w.(*bytes.Buffer); ok {
		b = wb
	}

	err := execTemplate(b, name, data)
	if err != nil {
		return fmt.Errorf("error rendering template %s: %w", name, err)
	}

	_, err = b.WriteTo(w)
	if err != nil {
		return fmt.Errorf("error writing template %s: %w", name, err)
	}

	return nil
}
