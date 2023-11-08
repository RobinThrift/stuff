package htmlui

import (
	"log/slog"
	"net/http"

	"github.com/RobinThrift/stuff/control"
	"github.com/RobinThrift/stuff/entities"
	"github.com/RobinThrift/stuff/views/pages"
)

// [GET] /assets/labels
func (rt *Router) labelsHandler(w http.ResponseWriter, r *http.Request, params struct{}) error {
	page := pages.LabelSheetCreatorPage{
		Assets:         []*entities.Asset{},
		ValidationErrs: map[string]string{},
	}

	return page.Render(w, r)
}

// [POST] /assets/labels
func (rt *Router) labelsSubmitHandler(w http.ResponseWriter, r *http.Request, params struct{}) error {
	err := r.ParseForm()
	if err != nil {
		return err
	}

	page := pages.LabelSheetCreatorPage{
		Assets:         []*entities.Asset{},
		ValidationErrs: map[string]string{},
	}

	err = rt.forms.Decode(&page, r.PostForm)
	if err != nil {
		return err
	}

	query := control.GenerateLabelSheetQuery{
		BaseURL: rt.config.BaseURL,
		IDs:     page.SelectedAssetIDs,
		Sheet: &entities.Sheet{
			SkipNumLabels: page.SkipLabels,
			PageSize:      entities.PageSize(page.PageSize),
			PageLayout: entities.PageLayout{
				Cols:         page.NumColumns,
				Rows:         page.NumRows,
				MarginLeft:   page.MarginLeft,
				MarginTop:    page.MarginTop,
				MarginRight:  page.MarginRight,
				MarginBottom: page.MarginBottom,
			},

			LabelSize: entities.LabelSize{
				FontSize:          page.FontSize,
				Height:            page.Height,
				Width:             page.Width,
				VerticalPadding:   page.VerticalPadding,
				HorizontalPadding: page.HorizontalPadding,
				VerticalSpacing:   page.VerticalSpacing,
				HorizontalSpacing: page.HorizontalSpacing,
			},
			PrintBorders: page.ShowBorders,
		},
	}

	pdf, err := rt.labels.GenerateLabelSheet(r.Context(), query)
	if err != nil {
		return err
	}

	w.Header().Add("content-disposition", `attachment; filename="labels.pdf"`)
	w.Header().Add("content-type", "application/pdf; charset=utf-8")

	w.WriteHeader(http.StatusOK)

	_, err = w.Write(pdf)
	if err != nil {
		slog.ErrorContext(r.Context(), "error writing to http response", "error", err)
		return err
	}

	return nil
}
