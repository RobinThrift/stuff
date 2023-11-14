package pages

import (
	"net/http"

	"github.com/RobinThrift/stuff/entities"
	"github.com/RobinThrift/stuff/views"
)

type AssetFileDeletePage struct {
	File *entities.File
}

func (m *AssetFileDeletePage) Render(w http.ResponseWriter, r *http.Request) error {
	return views.Render(w, "assets_file_delete_page", views.Model[*AssetFileDeletePage]{
		Global: views.NewGlobal("Confirm deletion of"+m.File.Name, r),
		Data:   m,
	})
}
