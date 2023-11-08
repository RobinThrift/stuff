package htmlui

import (
	"log/slog"
	"net/http"

	"github.com/RobinThrift/stuff/control"
	"github.com/RobinThrift/stuff/views/pages"
)

// [GET] /assets/import
func (rt *Router) importAssetsHandler(w http.ResponseWriter, r *http.Request, params struct{}) error {
	page := pages.ImportPage{ValidationErrs: map[string]string{}}
	return page.Render(w, r)
}

// [POST] /assets/import
func (rt *Router) importAssetsSubmitHandler(w http.ResponseWriter, r *http.Request, params struct{}) error {
	page := pages.ImportPage{ValidationErrs: map[string]string{}}

	err := r.ParseMultipartForm(defaultMaxMemory)
	if err != nil {
		return err
	}

	err = rt.forms.Decode(&page, r.PostForm)
	if err != nil {
		return err
	}

	page.ValidationErrs, err = rt.importer.Import(r, control.ImportCmd{
		IgnoreDuplicates: page.IgnoreDuplicates,
		Format:           page.Format,
		SnipeITURL:       page.SnipeITURL,
		SnipeITAPIKey:    page.SnipeITAPIKey,
	})
	if len(page.ValidationErrs) != 0 {
		return page.Render(w, r)
	}

	if err != nil {
		slog.ErrorContext(r.Context(), "error importing assets", "error", err)
		return err
	}

	http.Redirect(w, r, "/assets", http.StatusFound)
	return nil
}
