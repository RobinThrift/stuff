package pages

import (
	"net/http"

	"github.com/RobinThrift/stuff/internal/server/session"
	"github.com/RobinThrift/stuff/views"
)

type ImportPage struct {
	Format           string `form:"format"`
	IgnoreDuplicates bool   `form:"ignore_duplicates"`

	SnipeITURL    string `form:"snipeit_url"`
	SnipeITAPIKey string `form:"snipeit_api_key"`

	ValidationErrs map[string]string `form:"-"`
}

func (m *ImportPage) Render(w http.ResponseWriter, r *http.Request) error {
	csrfErr, ok := session.Pop[string](r.Context(), "csrf_error")
	if ok {
		m.ValidationErrs["general"] = csrfErr
	}

	return views.Render(w, "assets_import", views.Model[*ImportPage]{
		Global: views.NewGlobal("Import Assets", r),
		Data:   m,
	})
}
