package assets

import (
	"fmt"
	"net/http"

	"github.com/RobinThrift/stuff/server/session"
	"github.com/RobinThrift/stuff/views"
)

type ListAssetsPageViewModel struct {
	Page    *AssetListPage
	Query   ListAssetsQuery
	Columns map[string]bool
	Compact bool
}

type EditAssetsPageViewModel struct {
	IsNew            bool
	Asset            *Asset
	ValidationErrs   map[string]string
	DefaultCurrency  string
	DecimalSeparator string
}

type ViewAssetsPageViewModel struct {
	Asset            *Asset
	DecimalSeparator string
	ValidationErrs   map[string]string
}

type DeleteAssetsPageViewModel struct {
	Asset   *Asset
	Message string
}

func renderListAssetsPage(w http.ResponseWriter, r *http.Request, query ListAssetsQuery, page *AssetListPage) error {
	search := ""
	if query.Search != nil {
		search = query.Search.Raw
	}

	global := views.NewGlobal("Assets", r)
	global.Search = search

	columns, _ := session.Get[map[string]bool](r.Context(), "assets_list_columns")
	if len(columns) == 0 {
		columns = map[string]bool{
			"Tag": true, "Image": true, "Name": true, "Type": true, "Category": true, "Location": true, "Status": true,
		}
	}

	compact, _ := session.Get[bool](r.Context(), "assets_lists_compact")

	err := views.Render(w, "assets_list_page", views.Model[ListAssetsPageViewModel]{
		Global: global,
		Data: ListAssetsPageViewModel{
			Page:    page,
			Query:   query,
			Columns: columns,
			Compact: compact,
		},
	})
	if err != nil {
		return fmt.Errorf("error rendering list assets page: %w", err)
	}

	return nil
}

func (rt *UIRouter) renderEditAssetPage(w http.ResponseWriter, r *http.Request, model EditAssetsPageViewModel) error {
	model.DefaultCurrency = rt.DefaultCurrency
	model.DecimalSeparator = rt.DecimalSeparator

	if len(model.Asset.Purchases) == 0 {
		model.Asset.Purchases = []*Purchase{{Currency: rt.DefaultCurrency}}
	}

	title := "New Asset"
	if !model.IsNew {
		title = "Edit Asset"
	}

	if model.Asset.Tag == "" {
		tag, err := rt.Control.generateTag(r.Context())
		if err != nil {
			return err
		}
		model.Asset.Tag = tag
	}

	csrfErr, ok := session.Pop[string](r.Context(), "csrf_error")
	if ok {
		model.ValidationErrs["general"] = csrfErr
	}

	err := views.Render(w, "assets_edit_page", views.Model[EditAssetsPageViewModel]{
		Global: views.NewGlobal(title, r),
		Data:   model,
	})
	if err != nil {
		return fmt.Errorf("error rendering edit asset page: %w", err)
	}

	return nil
}

func renderViewAssetPage(w http.ResponseWriter, r *http.Request, model ViewAssetsPageViewModel) error {
	err := views.Render(w, "assets_view_page", views.Model[ViewAssetsPageViewModel]{
		Global: views.NewGlobal(model.Asset.Name, r),
		Data:   model,
	})
	if err != nil {
		return fmt.Errorf("error rendering view assets page: %w", err)
	}

	return nil
}

type AssetFilesPageViewModel struct {
	Asset *Asset
}

func renderAssetFilesPage(w http.ResponseWriter, r *http.Request, model AssetFilesPageViewModel) error {
	err := views.Render(w, "assets_files_page", views.Model[AssetFilesPageViewModel]{
		Global: views.NewGlobal(model.Asset.Name+": Files", r),
		Data:   model,
	})
	if err != nil {
		return fmt.Errorf("error rendering asset files page: %w", err)
	}

	return nil
}

type AssetFileDeletePageViewModel struct {
	File *File
}

func renderAssetFileDeletePage(w http.ResponseWriter, r *http.Request, model AssetFileDeletePageViewModel) error {
	err := views.Render(w, "assets_file_delete_page", views.Model[AssetFileDeletePageViewModel]{
		Global: views.NewGlobal("Confirm deletion of"+model.File.Name, r),
		Data:   model,
	})
	if err != nil {
		return fmt.Errorf("error rendering delte asset file page: %w", err)
	}

	return nil
}

func renderDeleteAssetPage(w http.ResponseWriter, r *http.Request, model DeleteAssetsPageViewModel) error {
	csrfErr, ok := session.Pop[string](r.Context(), "csrf_error")
	if ok {
		if model.Message != "" {
			model.Message += "\n"
		}
		model.Message += csrfErr
	}

	err := views.Render(w, "assets_delete_page", views.Model[DeleteAssetsPageViewModel]{
		Global: views.NewGlobal(model.Asset.Name, r),
		Data:   model,
	})
	if err != nil {
		return fmt.Errorf("error rendering delete assets page: %w", err)
	}

	return nil
}

type LabelSheetCreatorPageViewModel struct {
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

	Assets []*Asset `form:"-"`

	ValidationErrs map[string]string `form:"-"`
}

func renderLabelSheetCreatorPage(w http.ResponseWriter, r *http.Request, model LabelSheetCreatorPageViewModel) error {
	csrfErr, ok := session.Pop[string](r.Context(), "csrf_error")
	if ok {
		model.ValidationErrs["general"] = csrfErr
	}

	err := views.Render(w, "assets_label_sheet_creator", views.Model[LabelSheetCreatorPageViewModel]{
		Global: views.NewGlobal("Create Label Sheet", r),
		Data:   model,
	})
	if err != nil {
		return fmt.Errorf("error rendering create label sheet page: %w", err)
	}

	return nil
}

type ImportPageViewModel struct {
	Format           string `form:"format"`
	IgnoreDuplicates bool   `form:"ignore_duplicates"`

	SnipeITURL    string `form:"snipeit_url"`
	SnipeITAPIKey string `form:"snipeit_api_key"`

	ValidationErrs map[string]string `form:"-"`
}

func renderImportPage(w http.ResponseWriter, r *http.Request, model ImportPageViewModel) error {
	csrfErr, ok := session.Pop[string](r.Context(), "csrf_error")
	if ok {
		model.ValidationErrs["general"] = csrfErr
	}

	err := views.Render(w, "assets_import", views.Model[ImportPageViewModel]{
		Global: views.NewGlobal("Import Assets", r),
		Data:   model,
	})
	if err != nil {
		return fmt.Errorf("error rendering import page: %w", err)
	}

	return nil
}
