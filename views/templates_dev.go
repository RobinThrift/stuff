//go:build dev
// +build dev

package views

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path"
	"regexp"
	"sync"
	"time"
)

var templateDir = path.Join("views", "templates")

var partialsDir = path.Join(templateDir, "partials")
var pagesDir = path.Join(templateDir, "pages")

var templateFS = os.DirFS(".")

var execTemplateMutex sync.Mutex

func timing(tmplName string) func() {
	start := time.Now()
	return func() {
		dur := time.Since(start)
		slog.Debug(fmt.Sprintf("rendering time for %s: %v", tmplName, dur))
	}
}

func execTemplate(w io.Writer, name string, data any) error {
	execTemplateMutex.Lock()
	defer execTemplateMutex.Unlock()

	var templates *template.Template
	cfs := &componentFS{fs: templateFS}

	templateFuncs["children"] = func(childname string, data any) template.HTML {
		var b bytes.Buffer

		err := templates.ExecuteTemplate(&b, childname, data)
		if err != nil {
			panic(err)
		}

		return template.HTML(b.Bytes())
	}

	templates, err := template.New(name).Funcs(templateFuncs).ParseFS(cfs, path.Join(pagesDir, name+".html.tmpl"), path.Join(partialsDir, "*.html.tmpl"))
	if err != nil {
		return enhanceErrorMessage(cfs, err)
	}

	defer timing(name)()
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
	if errors.Is(err, os.ErrNotExist) {
		file, err = fs.Open(path.Join(pagesDir, name))
	}

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
