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
	mux.Post("/users/settings", logError(rt.postUserSettings))
}

// [POST] /users/settings
func (rt *UIRouter) postUserSettings(w http.ResponseWriter, r *http.Request) error {
	var payload struct {
		Sidebar *struct {
			Closed bool `json:"closed"`
		} `json:"sidebar,omitempty"`
		Assets *struct {
			Columns map[string]bool `json:"columns,omitempty"`
			Compact *bool           `json:"compact,omitempty"`
		} `json:"assetsList,omitempty"`
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

	if payload.Sidebar != nil {
		session.Put(r.Context(), "sidebar_closed", payload.Sidebar.Closed)
	}

	if payload.Assets != nil {
		if len(payload.Assets.Columns) != 0 {
			session.Put(r.Context(), "assets_list_columns", payload.Assets.Columns)
		}

		if payload.Assets.Compact != nil {
			session.Put(r.Context(), "assets_lists_compact", payload.Assets.Compact)
		}
	}

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
