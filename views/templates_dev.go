//go:build dev
// +build dev

package views

import (
	"html/template"
	"io"
	"os"
	"path"
)

var templateDir = path.Join("views", "templates")

var templateFS = os.DirFS(".")

func execTemplate(w io.Writer, name string, data any) error {
	templates, err := template.New(name).Funcs(templateFuncs).ParseFS(templateFS, path.Join(templateDir, "pages", name+".html.tmpl"), "views/templates/partials/*.html.tmpl")
	if err != nil {
		return err
	}

	return templates.ExecuteTemplate(w, name+".html.tmpl", data)
}
