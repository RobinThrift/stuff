package pages

import (
	"net/http"

	"github.com/RobinThrift/stuff/entities"
	"github.com/RobinThrift/stuff/views"
)

type TagListPage struct {
	Tags   *views.Pagination[*entities.Tag]
	Search string
}

func (m *TagListPage) Render(w http.ResponseWriter, r *http.Request) error {
	return views.Render(w, "tags_list_page", views.Model[*TagListPage]{
		Global: views.NewGlobal("Tagsl", r),
		Data:   m,
	})
}
