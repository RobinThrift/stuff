package htmlui

import (
	"net/http"

	"github.com/RobinThrift/stuff/control"
)

type exportAssetsParams struct {
	Format string `url:"format"`
}

func (rt *Router) exportAssetsHandler(w http.ResponseWriter, r *http.Request, params exportAssetsParams) error {
	switch params.Format {
	case "json":
		w.Header().Add("content-disposition", `attachment; filename="assets_export.json"`)
		w.Header().Add("content-type", "application/json; charset=utf-8")
	case "csv":
		w.Header().Add("content-disposition", `attachment; filename="assets_export.csv"`)
		w.Header().Add("content-type", "text/csv; charset=utf-8")
	}

	return rt.exporter.Export(r.Context(), w, control.ExportCmd{Format: params.Format})
}
