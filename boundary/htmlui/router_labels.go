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
		slog.ErrorContext(r.Context(), "error decoding label creation form", "error", err)
		page.ValidationErrs["general"] = err.Error()
		return page.Render(w, r)
	}

	if page.NumColumns == 0 && page.NumRows == 0 && page.Width == 0 && page.Height == 0 {
		errValidation := "must set either Number of Columns/Rows or Label Width/Height"
		page.ValidationErrs["page_cols"] = errValidation
		page.ValidationErrs["page_rows"] = errValidation
		page.ValidationErrs["label_width"] = errValidation
		page.ValidationErrs["label_height"] = errValidation
		return page.Render(w, r)
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
