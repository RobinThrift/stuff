package htmlui

import (
	"log/slog"
	"net/http"

	"github.com/RobinThrift/stuff/views"
)

func viewRenderHandler[T any](h func(w http.ResponseWriter, r *http.Request, params T) error) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var params T
		err := decodeParams(&params, r)
		if err != nil {
			slog.ErrorContext(r.Context(), r.URL.Path+" error decoding url parameters", "error", err)
			views.RenderErrorPage(w, r, err)
			return
		}

		err = h(w, r, params)
		if err != nil {
			slog.ErrorContext(r.Context(), r.URL.Path, "error", err)
			views.RenderErrorPage(w, r, err)
			return
		}
	})
}
