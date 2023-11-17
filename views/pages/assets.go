package pages

import (
	"net/http"

	"github.com/RobinThrift/stuff/auth"
	"github.com/RobinThrift/stuff/entities"
	"github.com/RobinThrift/stuff/internal/server/session"
	"github.com/RobinThrift/stuff/views"
)

type AssetListPage struct {
	Assets  *views.Pagination[*entities.Asset]
	Search  string
	Columns map[string]bool
}

var defaultAssetListColumns = map[string]bool{
	"Tag": true, "Image": true, "Name": true, "Type": true, "Category": true, "Location": true, "Status": true,
}

func (m *AssetListPage) Render(w http.ResponseWriter, r *http.Request) error {
	global := views.NewGlobal("Assets", r)

	user, _ := session.Get[*auth.User](r.Context(), "user")

	m.Columns = defaultAssetListColumns

	if len(user.Preferences.AssetListColumns) != 0 {
		m.Columns = make(map[string]bool, len(user.Preferences.AssetListColumns))
		for _, c := range user.Preferences.AssetListColumns {
			m.Columns[c] = true
		}
	}

	return views.Render(w, "assets_list_page", views.Model[*AssetListPage]{
		Global: global,
		Data:   m,
	})
}

type AssetEditPage struct {
	Asset            *entities.Asset
	IsNew            bool
	ValidationErrs   map[string]string
	DefaultCurrency  string
	DecimalSeparator string
}

func (m *AssetEditPage) Render(w http.ResponseWriter, r *http.Request) error {
	if len(m.Asset.Purchases) == 0 {
		m.Asset.Purchases = []*entities.Purchase{{Currency: m.DefaultCurrency}}
	}

	title := "New Asset"
	if !m.IsNew {
		title = "Edit Asset"
	}

	csrfErr, ok := session.Pop[string](r.Context(), "csrf_error")
	if ok {
		m.ValidationErrs["general"] = csrfErr
	}

	return views.Render(w, "assets_edit_page", views.Model[*AssetEditPage]{
		Global: views.NewGlobal(title, r),
		Data:   m,
	})
}

type AssetViewPage struct {
	Asset            *entities.Asset
	DecimalSeparator string
}

func (m *AssetViewPage) Render(w http.ResponseWriter, r *http.Request) error {
	return views.Render(w, "assets_view_page", views.Model[*AssetViewPage]{
		Global: views.NewGlobal(m.Asset.Name, r),
		Data:   m,
	})
}

type AssetDeletePage struct {
	Asset   *entities.Asset
	Message string
}

func (m *AssetDeletePage) Render(w http.ResponseWriter, r *http.Request) error {
	csrfErr, ok := session.Pop[string](r.Context(), "csrf_error")
	if ok {
		if m.Message != "" {
			m.Message += "\n"
		}
		m.Message += csrfErr
	}

	return views.Render(w, "assets_delete_page", views.Model[*AssetDeletePage]{
		Global: views.NewGlobal(m.Asset.Name, r),
		Data:   m,
	})
}
