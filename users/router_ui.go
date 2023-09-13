package users

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kodeshack/stuff/server/session"
)

type UIRouter struct {
}

func (rt *UIRouter) RegisterRoutes(mux *chi.Mux) {
	mux.Post("/users/session/sidebar/open", logError(rt.sessionSetSidebarOpen))
}

// [POST] /users/session/sidebar/open
func (rt *UIRouter) sessionSetSidebarOpen(w http.ResponseWriter, r *http.Request) error {
	var payload struct {
		Closed bool `json:"closed"`
	}

	body, err := io.ReadAll(r.Body)
	defer func() {
		err = errors.Join(err, r.Body.Close())
	}()

	if err != nil {
		return fmt.Errorf("error reading request body: %w", err)
	}

	err = json.Unmarshal(body, &payload)
	if err != nil {
		return fmt.Errorf("error unmarshalling request body as JSON: %w", err)
	}

	session.Put(r.Context(), "sidebar_closed", payload.Closed)

	return nil
}

func logError(handler func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := handler(w, r)
		if err != nil {
			slog.ErrorContext(r.Context(), r.URL.Path, "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			_, writeErr := w.Write([]byte(err.Error()))
			if writeErr != nil {
				slog.ErrorContext(r.Context(), "error writing response", "error", writeErr)
			}
		}
	})
}
