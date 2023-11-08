package control

import (
	"context"
	"net/url"

	"github.com/RobinThrift/stuff/entities"
)

type LabelController struct {
	assets *AssetControl
}

func NewLabelController(assets *AssetControl) *LabelController {
	return &LabelController{assets: assets}
}

type GenerateLabelSheetQuery struct {
	BaseURL *url.URL
	IDs     []int64
	Sheet   *entities.Sheet
}

func (lc *LabelController) GenerateLabelSheet(ctx context.Context, query GenerateLabelSheetQuery) ([]byte, error) {
	assets, err := lc.assets.List(ctx, ListAssetsQuery{IDs: query.IDs})
	if err != nil {
		return nil, err
	}

	labels := make([]entities.Label, 0, len(assets.Items))
	for _, a := range assets.Items {
		l, err := a.Labels(query.BaseURL, 200)
		if err != nil {
			return nil, err
		}
		labels = append(labels, l...)
	}

	query.Sheet.Labels = labels

	return query.Sheet.Generate()
}
