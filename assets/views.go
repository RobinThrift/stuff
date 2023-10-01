package assets

import (
	"fmt"
	"net/http"

	"github.com/kodeshack/stuff/server/session"
	"github.com/kodeshack/stuff/views"
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
			"Tag": true, "Image": true, "Name": true, "Category": true, "Location": true, "Status": true,
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
	model.DecimalSeparator = rt.DecimalSeparator

	if model.Asset.PurchaseInfo.Currency == "" {
		model.Asset.PurchaseInfo.Currency = rt.DefaultCurrency
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
