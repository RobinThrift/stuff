package tags

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/RobinThrift/stuff/api"
)

type APIRouter struct {
	Control *Control
}

func (rt *APIRouter) RegisterRoutes(mux *chi.Mux) {
	mux.Get("/api/v1/tags", rt.apiListTags)
}

// [GET] /api/v1/tags
func (rt *APIRouter) apiListTags(w http.ResponseWriter, r *http.Request) {
	type tag struct {
		Tag string `json:"tag"`
	}

	type pageJSON struct {
		Tags     []tag `json:"tags"`
		Total    int   `json:"total"`
		NumPages int   `json:"numPages"`
		Page     int   `json:"page"`
		PageSize int   `json:"pageSize"`
	}

	query := listTagsQueryFromURL(r.URL.Query())
	page, err := rt.Control.listTags(r.Context(), query)
	if err != nil {
		api.RespondWithError(r.Context(), w, err)
		return
	}

	res := pageJSON{
		Tags:     make([]tag, 0, len(page.Tags)),
		Total:    page.Total,
		NumPages: page.NumPages,
		Page:     page.Page,
		PageSize: page.PageSize,
	}

	for _, t := range page.Tags {
		res.Tags = append(res.Tags, tag{Tag: t.Tag})
	}

	b, err := json.Marshal(res)
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
	q := ListTagsQuery{ //nolint: varnamelen
		Search:   params.Get("query"),
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

	if inUseStr := params.Get("in_use"); inUseStr != "" {
		inUse, err := strconv.ParseBool(inUseStr)
		if err == nil {
			q.InUse = &inUse
		}
	}

	return q
}
