//go:build dev
// +build dev

package static

import (
	"net/http"
	"os"
)

func Files(prefix string) http.Handler {
	return http.StripPrefix(prefix, http.FileServer(http.Dir("./static/build")))
}

func PDFFont() ([]byte, error) {
	return os.ReadFile("./static/build/fonts/OpenSans-Regular.ttf")
}
