package assets

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/RobinThrift/stuff/api"
	"github.com/go-chi/chi/v5"
)

type APIRouter struct {
	Control *Control
}

func (rt *APIRouter) RegisterRoutes(mux *chi.Mux) {
	mux.Get("/api/v1/assets", rt.apiListAssets)
	mux.Get("/api/v1/assets/categories", rt.apiListCategories)
}

// [GET] /api/v1/assets/categories
func (rt *APIRouter) apiListCategories(w http.ResponseWriter, r *http.Request) {
	type category struct {
		Name string `json:"name"`
	}

	type page struct {
		Categories []category `json:"categories"`
	}

	cats, err := rt.Control.listCategories(r.Context(), ListCategoriesQuery{Search: r.URL.Query().Get("query")})
	if err != nil {
		api.RespondWithError(r.Context(), w, err)
		return
	}

	res := page{
		Categories: make([]category, 0, len(cats)),
	}

	for _, c := range cats {
		res.Categories = append(res.Categories, category(c))
	}

	b, err := json.Marshal(res)
	if err != nil {
		slog.ErrorContext(r.Context(), "error marshalling categories JSON", "error", err)
		return
	}

	api.AddJSONContentType(w)
	_, err = w.Write(b)
	if err != nil {
		slog.ErrorContext(r.Context(), "error writing to HTTP response", "error", err)
	}
}

type apiPart struct {
	ID           int64  `json:"id"`
	AssetID      int64  `json:"assetID"`
	Tag          string `json:"tag"`
	Name         string `json:"name"`
	Location     string `json:"location,omitempty"`
	PositionCode string `json:"positionCode,omitempty"`
	Notes        string `json:"notes,omitempty"`
}

type apiAsset struct {
	ID            int64           `json:"id,omitempty"`
	ParentAssetID int64           `json:"parentAssetID,omitempty"`
	Tag           string          `json:"tag"`
	Status        Status          `json:"status"`
	Name          string          `json:"name"`
	Category      string          `json:"category"`
	Model         string          `json:"model,omitempty"`
	ModelNo       string          `json:"modelNo,omitempty"`
	SerialNo      string          `json:"serialNo,omitempty"`
	Manufacturer  string          `json:"manufacturer,omitempty"`
	Notes         string          `json:"notes,omitempty"`
	ImageURL      string          `json:"imageURL,omitempty"`
	ThumbnailURL  string          `json:"thumbnailURL,omitempty"`
	WarrantyUntil time.Time       `json:"warrantyUntil,omitempty"`
	CustomAttrs   []apiCustomAttr `json:"customAttrs,omitempty"`

	Location     string `json:"location,omitempty"`
	PositionCode string `json:"positionCode,omitempty"`

	Purchases []*apiPurchase `json:"purchases"`

	PartsTotalCounter int        `json:"partsTotalCounter,omitempty"`
	Parts             []*apiPart `json:"parts,omitempty"`
}

type apiCustomAttr struct {
	Name  string `json:"name,omitempty"`
	Value any    `json:"value,omitempty"`
}

type apiPurchase struct {
	Supplier string         `json:"supplier,omitempty"`
	OrderNo  string         `json:"order_no,omitempty"`
	Date     time.Time      `json:"order_date,omitempty"`
	Amount   MonetaryAmount `json:"amount,omitempty"`
	Currency string         `json:"currency,omitempty"`
}

// [GET] /api/v1/assets
func (rt *APIRouter) apiListAssets(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Assets []*apiAsset `json:"assets"`
	}

	query := listAssetsQueryFromURL(r.URL.Query())
	page, err := rt.Control.listAssets(r.Context(), query)
	if err != nil {
		api.RespondWithError(r.Context(), w, err)
		return
	}

	res := response{Assets: make([]*apiAsset, 0, len(page.Assets))}

	for _, asset := range page.Assets {
		parts := make([]*apiPart, 0, len(asset.Parts))

		for _, p := range parts {
			parts = append(parts, &apiPart{
				ID:           p.ID,
				AssetID:      p.AssetID,
				Tag:          p.Tag,
				Name:         p.Name,
				Location:     p.Location,
				PositionCode: p.PositionCode,
				Notes:        p.Notes,
			})
		}

		purchases := make([]*apiPurchase, 0, len(asset.Purchases))
		for _, p := range asset.Purchases {
			purchases = append(purchases, &apiPurchase{
				Supplier: p.Supplier,
				OrderNo:  p.OrderNo,
				Date:     p.Date,
				Amount:   p.Amount,
				Currency: p.Currency,
			})
		}

		customAttrs := make([]apiCustomAttr, 0, len(asset.CustomAttrs))
		for _, ca := range asset.CustomAttrs {
			customAttrs = append(customAttrs, apiCustomAttr(ca))
		}

		res.Assets = append(res.Assets, &apiAsset{
			ID:                asset.ID,
			ParentAssetID:     asset.ParentAssetID,
			Status:            asset.Status,
			Tag:               asset.Tag,
			Name:              asset.Name,
			Category:          asset.Category,
			Model:             asset.Model,
			ModelNo:           asset.ModelNo,
			SerialNo:          asset.SerialNo,
			Manufacturer:      asset.Manufacturer,
			Notes:             asset.Notes,
			ImageURL:          asset.ImageURL,
			ThumbnailURL:      asset.ThumbnailURL,
			WarrantyUntil:     asset.WarrantyUntil,
			CustomAttrs:       customAttrs,
			Location:          asset.Location,
			PositionCode:      asset.PositionCode,
			Purchases:         purchases,
			PartsTotalCounter: asset.PartsTotalCounter,
			Parts:             parts,
		})
	}

	b, err := json.Marshal(res)
	if err != nil {
		slog.ErrorContext(r.Context(), "error marshalling assets JSON", "error", err)
		return
	}

	api.AddJSONContentType(w)
	_, err = w.Write(b)
	if err != nil {
		slog.ErrorContext(r.Context(), "error writing to HTTP response", "error", err)
	}
}
