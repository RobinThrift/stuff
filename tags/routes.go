package tags

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/kodeshack/stuff/api"
	"github.com/kodeshack/stuff/views"
)

type Router struct {
	Control *Control
}

func (rt *Router) RegisterRoutes(mux *chi.Mux) {
	mux.Get("/tags", views.HTTPHandlerFuncErr(rt.handleTagsListGet))

	mux.Get("/api/v1/tags", rt.apiListTags)
}

// [GET] /tags
func (rt *Router) handleTagsListGet(w http.ResponseWriter, r *http.Request) error {
	query := listTagsQueryFromURL(r.URL.Query())
	page, err := rt.Control.listTags(r.Context(), query)
	if err != nil {
		return err
	}

	return renderListTagsPage(w, r, query, page)
}

// [GET] /api/v1/tags
func (rt *Router) apiListTags(w http.ResponseWriter, r *http.Request) {
	query := listTagsQueryFromURL(r.URL.Query())
	tags, err := rt.Control.listTags(r.Context(), query)
	if err != nil {
		api.RespondWithError(r.Context(), w, err)
		return
	}

	b, err := json.Marshal(tags)
	if err != nil {
		slog.ErrorContext(r.Context(), "error marshalling tags JSON", "error", err)
		return
	}

	api.AddJSONContentType(w)
	_, err = w.Write(b)
	if err != nil {
		slog.ErrorContext(r.Context(), "error writing to HTTP response", "error", err)
	}
}

func listTagsQueryFromURL(params url.Values) ListTagsQuery {
	q := ListTagsQuery{
		PageSize: 50,
		OrderBy:  params.Get("order_by"),
	}

	if size := params.Get("page_size"); size != "" {
		q.PageSize, _ = strconv.Atoi(size)
	}

	if pageStr := params.Get("page"); pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err == nil {
			q.Page = q.PageSize * page
		}
	}

	if orderDir := params.Get("order_dir"); orderDir != "" {
		orderDir = strings.ToUpper(orderDir)
		if orderDir == "ASC" || orderDir == "DESC" {
			q.OrderDir = orderDir
		}
	}

	return q
}
