//go:build !dev
// +build !dev

package frontend

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed build
var _assets embed.FS

//go:embed manifest
var manifest embed.FS

var _corrected, _ = fs.Sub(_assets, "build")

func Files(prefix string) http.Handler {
	return http.StripPrefix(prefix, http.FileServer(http.FS(_corrected)))
}

func PDFFont() ([]byte, error) {
	return _assets.ReadFile("build/fonts/OpenSans-Regular.ttf")
}

func Manifest() http.Handler {
	return http.FileServer(http.FS(manifest))
}
