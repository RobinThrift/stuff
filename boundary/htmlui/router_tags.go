package htmlui

import (
	"net/http"

	"github.com/RobinThrift/stuff/control"
	"github.com/RobinThrift/stuff/entities"
	"github.com/RobinThrift/stuff/views"
	"github.com/RobinThrift/stuff/views/pages"
)

type tagsListParams struct {
	Query    string `query:"query"`
	Page     int    `query:"page"`
	PageSize int    `query:"page_size"`
	OrderBy  string `query:"order_by"`
	OrderDir string `query:"order_dir"`
}

// [GET] /tags
func (rt *Router) tagsListHandler(w http.ResponseWriter, r *http.Request, params tagsListParams) error {
	if params.PageSize == 0 {
		params.PageSize = 25
	}

	list, err := rt.tags.List(r.Context(), control.ListTagsQuery{
		Search:   params.Query,
		Page:     params.Page,
		PageSize: params.PageSize,
		OrderBy:  params.OrderBy,
		OrderDir: params.OrderDir,
	})
	if err != nil {
		return err
	}

	page := &pages.TagListPage{
		Tags: &views.Pagination[*entities.Tag]{
			ListPage: list,
			URL:      r.URL,
		},
		Search: params.Query,
	}

	return page.Render(w, r)
}
