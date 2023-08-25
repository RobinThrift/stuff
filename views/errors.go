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
	statusCode := http.StatusInternalServerError
	title := "Unknown Error"
	message := err.Error()

	var errPageErr ErrorPageErr
	if errors.As(err, &errPageErr) {
		statusCode = errPageErr.Code
		title = errPageErr.Title
		if errPageErr.Message != "" {
			message = errPageErr.Message
		}
	}

	errorPage := ErrorPage(statusCode, title, message)
	page := Document(title, errorPage)

	renderErr := page.Render(r.Context(), w)
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
