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
	"regexp"
)

var templateDir = path.Join("views", "templates")

var partialsDir = path.Join(templateDir, "partials")

var templateFS = os.DirFS(".")

func execTemplate(w io.Writer, name string, data any) error {
	cfs := &componentFS{fs: templateFS}

	templates, err := template.New(name).Funcs(templateFuncs).ParseFS(cfs, path.Join(templateDir, "pages", name+".html.tmpl"), path.Join(partialsDir, "*.html.tmpl"))
	if err != nil {
		return enhanceErrorMessage(cfs, err)
	}

	err = templates.ExecuteTemplate(w, name+".html.tmpl", data)
	if err != nil {
		return enhanceErrorMessage(cfs, err)
	}

	return nil
}

var extractFileLineRegex = regexp.MustCompile(`template: (.*):(\d+)`)

func enhanceErrorMessage(fs fs.FS, tmplErr error) error {
	matches := extractFileLineRegex.FindStringSubmatch(tmplErr.Error())
	if len(matches) == 0 {
		return tmplErr
	}

	name := matches[1]

	file, err := fs.Open(path.Join(partialsDir, name))
	if err != nil {
		// ignore error
		return tmplErr
	}
	defer file.Close()

	contents, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}

	return fmt.Errorf("%w: %s:\n%s", tmplErr, name, contents)
}
