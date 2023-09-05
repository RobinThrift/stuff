package tags

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/kodeshack/stuff/api"
	"github.com/kodeshack/stuff/server/session"
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
	tags, err := rt.Control.listTags(r.Context(), query)
	if err != nil {
		return err
	}

	return renderListTagsPage(w, r, tags, query)
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

func listTagsQueryFromURL(params url.Values) listTagsQuery {
	q := listTagsQuery{
		limit:   50,
		orderBy: params.Get("order_by"),
	}

	if size := params.Get("page_size"); size != "" {
		q.limit, _ = strconv.Atoi(size)
	}

	if pageStr := params.Get("page"); pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err == nil {
			q.offset = q.limit * page
		}
	}

	if orderDir := params.Get("order_dir"); orderDir != "" {
		orderDir = strings.ToUpper(orderDir)
		if orderDir == "ASC" || orderDir == "DESC" {
			q.orderDir = orderDir
		}
	}

	return q
}

func renderListTagsPage(w http.ResponseWriter, r *http.Request, tagList *TagList, query listTagsQuery) error {
	infomsg, _ := session.Pop[string](r.Context(), "info_message")

	listTagsPage := listTagsPage(listTagsPageProps{
		tags:    tagList.Tags,
		total:   tagList.Total,
		query:   query,
		infomsg: infomsg,
	})
	page := views.Document("Tags", listTagsPage)

	err := page.Render(r.Context(), w)
	if err != nil {
		return fmt.Errorf("error rendering list tags page: %w", err)
	}

	return nil
}
