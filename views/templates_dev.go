//go:build dev
// +build dev

package views

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"os"
	"path"
)

var templateDir = path.Join("views", "templates")

var templateFS = os.DirFS(".")

func execTemplate(w io.Writer, name string, data any) error {
	cfs := &componentFS{fs: templateFS}

	templates, err := template.New(name).Funcs(templateFuncs).ParseFS(cfs, path.Join(templateDir, "pages", name+".html.tmpl"), "views/templates/partials/*.html.tmpl")
	if err != nil {
		printTemplate(cfs, name)
		return err
	}

	err = templates.ExecuteTemplate(w, name+".html.tmpl", data)
	if err != nil {
		printTemplate(cfs, name)
		return err
	}

	return nil
}

func printTemplate(fs fs.FS, name string) {
	file, err := fs.Open(path.Join(templateDir, "pages", name+".html.tmpl"))
	if err != nil {
		panic(err)
	}
	defer file.Close()

	contents, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}

	fmt.Println(name + ".html.tmpl:")
	fmt.Println(string(contents))
}
