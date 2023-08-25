//go:build !dev
// +build !dev

package stuff

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed build
var _assets embed.FS

var _corrected, _ = fs.Sub(_assets, "build")

func StaticFiles(prefix string) http.Handler {
	return http.StripPrefix(prefix, http.FileServer(http.FS(_corrected)))
}
