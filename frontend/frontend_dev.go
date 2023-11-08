//go:build dev
// +build dev

package frontend

import (
	"net/http"
	"os"
)

func Files(prefix string) http.Handler {
	return http.StripPrefix(prefix, http.FileServer(http.Dir("./frontend/build")))
}

func PDFFont() ([]byte, error) {
	return os.ReadFile("./frontend/build/fonts/OpenSans-Regular.ttf")
}
