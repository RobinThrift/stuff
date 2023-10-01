//go:build !dev
// +build !dev

package views

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"path"
	"strings"
)

//go:embed templates/pages/*.tmpl
//go:embed templates/partials/*.tmpl
var templateFS embed.FS

var templates = func() map[string]*template.Template {
	tmpls := make(map[string]*template.Template)

	pages, err := fs.Glob(templateFS, "templates/pages/*.html.tmpl")
	if err != nil {
		panic(fmt.Errorf("error globbing pages template FS: %w", err))
	}

	cfs := &componentFS{fs: templateFS}

	for _, page := range pages {
		name := strings.ReplaceAll(path.Base(page), ".html.tmpl", "")

		templateFuncs["children"] = func(childname string, data any) template.HTML {
			var b bytes.Buffer

			err := tmpls[name].ExecuteTemplate(&b, childname, data)
			if err != nil {
				panic(err)
			}

			return template.HTML(b.Bytes())
		}

		tmpls[name] = template.Must(template.New(name).Funcs(templateFuncs).ParseFS(cfs, page, "templates/partials/*.html.tmpl"))
	}

	return tmpls
}()

func execTemplate(w io.Writer, name string, data any) error {
	return templates[name].ExecuteTemplate(w, name+".html.tmpl", data)
}
