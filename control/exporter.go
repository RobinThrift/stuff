package control

import (
	"context"
	"fmt"
	"io"

	"github.com/RobinThrift/stuff/internal/exporter"
	"github.com/RobinThrift/stuff/storage/database"
)

type ExporterCtrl struct {
	db     *database.Database
	assets *AssetControl
}

func NewExporterCtrl(db *database.Database, assets *AssetControl) *ExporterCtrl {
	return &ExporterCtrl{db: db, assets: assets}
}

type ExportCmd struct {
	Format string
}

func (ec *ExporterCtrl) Export(ctx context.Context, w io.Writer, cmd ExportCmd) error {
	assets, err := ec.assets.List(ctx, ListAssetsQuery{})
	if err != nil {
		return err
	}

	switch cmd.Format {
	case "json":
		return exporter.ExportAssetsAsJSON(w, assets.Items)
	case "csv":
		return exporter.ExportAssetsAsCSV(w, assets.Items)
	}

	return fmt.Errorf("unknown export format: %s", cmd.Format)
}
