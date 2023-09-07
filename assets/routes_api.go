package assets

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kodeshack/stuff/api"
)

type APIRouter struct {
	Control *Control
}

func (rt *APIRouter) RegisterRoutes(mux *chi.Mux) {
	mux.Get("/api/v1/assets", rt.apiListAssets)
	mux.Get("/api/v1/assets/categories", rt.apiListCategories)
}

// [GET] /api/v1/assets/categories
func (rt *APIRouter) apiListCategories(w http.ResponseWriter, r *http.Request) {
	type category struct {
		Name string `json:"name"`
	}

	type page struct {
		Categories []category `json:"categories"`
	}

	cats, err := rt.Control.listCategories(r.Context(), ListCategoriesQuery{Search: r.URL.Query().Get("query")})
	if err != nil {
		api.RespondWithError(r.Context(), w, err)
		return
	}

	res := page{
		Categories: make([]category, 0, len(cats)),
	}

	for _, c := range cats {
		res.Categories = append(res.Categories, category(c))
	}

	b, err := json.Marshal(res)
	if err != nil {
		slog.ErrorContext(r.Context(), "error marshalling categories JSON", "error", err)
		return
	}

	api.AddJSONContentType(w)
	_, err = w.Write(b)
	if err != nil {
		slog.ErrorContext(r.Context(), "error writing to HTTP response", "error", err)
	}
}

// [GET] /api/v1/assets
func (rt *APIRouter) apiListAssets(w http.ResponseWriter, r *http.Request) {
	query := listAssetsQueryFromURL(r.URL.Query())
	page, err := rt.Control.listAssets(r.Context(), query)
	if err != nil {
		api.RespondWithError(r.Context(), w, err)
		return
	}

	b, err := json.Marshal(page.Assets)
	if err != nil {
		slog.ErrorContext(r.Context(), "error marshalling assets JSON", "error", err)
		return
	}

	api.AddJSONContentType(w)
	_, err = w.Write(b)
	if err != nil {
		slog.ErrorContext(r.Context(), "error writing to HTTP response", "error", err)
	}
}
