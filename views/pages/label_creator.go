package pages

import (
	"net/http"

	"github.com/RobinThrift/stuff/entities"
	"github.com/RobinThrift/stuff/internal/server/session"
	"github.com/RobinThrift/stuff/views"
)

type LabelSheetCreatorPage struct {
	Template string `form:"template"`

	SelectedAssetIDs []int64 `form:"-"`

	PageSize   string `form:"page_size"`
	SkipLabels int    `form:"skip_labels"`

	NumColumns   int     `form:"page_cols"`
	NumRows      int     `form:"page_rows"`
	MarginLeft   float64 `form:"page_margin_left"`
	MarginTop    float64 `form:"page_margin_top"`
	MarginRight  float64 `form:"page_margin_right"`
	MarginBottom float64 `form:"page_margin_bottom"`

	ShowBorders       bool    `form:"label_show_borders"`
	FontSize          float64 `form:"label_font_size"`
	Width             float64 `form:"label_width"`
	Height            float64 `form:"label_height"`
	VerticalPadding   float64 `form:"label_vertical_padding"`
	HorizontalPadding float64 `form:"label_horizontal_padding"`
	VerticalSpacing   float64 `form:"label_vertical_spacing"`
	HorizontalSpacing float64 `form:"label_horizontal_spacing"`

	Assets []*entities.Asset `form:"-"`

	ValidationErrs map[string]string `form:"-"`
}

func (m *LabelSheetCreatorPage) Render(w http.ResponseWriter, r *http.Request) error {
	csrfErr, ok := session.Pop[string](r.Context(), "csrf_error")
	if ok {
		m.ValidationErrs["general"] = csrfErr
	}

	return views.Render(w, "assets_label_sheet_creator", views.Model[*LabelSheetCreatorPage]{
		Global: views.NewGlobal("Create Label Sheet", r),
		Data:   m,
	})
}
