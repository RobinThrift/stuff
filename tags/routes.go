package tags

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kodeshack/stuff/views"
)

type UIRouter struct {
	Control *Control
}

func (rt *UIRouter) RegisterRoutes(mux *chi.Mux) {
	mux.Get("/tags", views.HTTPHandlerFuncErr(rt.handleTagsListGet))
}

// [GET] /tags
func (rt *UIRouter) handleTagsListGet(w http.ResponseWriter, r *http.Request) error {
	query := listTagsQueryFromURL(r.URL.Query())
	page, err := rt.Control.listTags(r.Context(), query)
	if err != nil {
		return err
	}

	return renderListTagsPage(w, r, query, page)
}
