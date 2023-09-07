package views

import (
	"errors"
	"log/slog"
	"net/http"
)

type ErrorPageErr struct {
	error
	Code    int
	Title   string
	Message string
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

func HTTPHandlerFuncErr(handler func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := handler(w, r)
		if err != nil {
			slog.ErrorContext(r.Context(), r.URL.Path, "error", err)
			RenderErrorPage(w, r, err)
		}
	})
}
