//go:build dev
// +build dev

package static

import (
	"net/http"
)

func Files(prefix string) http.Handler {
	return http.StripPrefix(prefix, http.FileServer(http.Dir("./build")))
}
