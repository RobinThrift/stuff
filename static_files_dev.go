//go:build dev
// +build dev

package stuff

import (
	"net/http"
)

func StaticFiles(prefix string) http.Handler {
	return http.StripPrefix(prefix, http.FileServer(http.Dir("./build")))
}
