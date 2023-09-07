package tags

import (
	"fmt"
	"net/http"

	"github.com/kodeshack/stuff/server/session"
	"github.com/kodeshack/stuff/views"
)

type ListTagsPageViewModel struct {
	Page  *TagListPage
	Query ListTagsQuery
}

func renderListTagsPage(w http.ResponseWriter, r *http.Request, query ListTagsQuery, page *TagListPage) error {
	infomsg, _ := session.Pop[string](r.Context(), "info_message")

	err := views.Render(w, "tags_list_page", views.Model[ListTagsPageViewModel]{
		Global: views.Global{
			Title:        "Tags",
			FlashMessage: infomsg,
		},
		Data: ListTagsPageViewModel{
			Page:  page,
			Query: query,
		},
	})
	if err != nil {
		return fmt.Errorf("error rendering list tags page: %w", err)
	}

	return nil
}
