package tags

import (
	"fmt"
	"net/http"

	"github.com/kodeshack/stuff/views"
)

type ListTagsPageViewModel struct {
	Page  *TagListPage
	Query ListTagsQuery
}

func renderListTagsPage(w http.ResponseWriter, r *http.Request, query ListTagsQuery, page *TagListPage) error {
	err := views.Render(w, "tags_list_page", views.Model[ListTagsPageViewModel]{
		Global: views.NewGlobal("Tagsl", r),
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
