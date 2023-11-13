package htmlui

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/RobinThrift/stuff/auth"
	"github.com/RobinThrift/stuff/control"
	"github.com/RobinThrift/stuff/entities"
	"github.com/RobinThrift/stuff/internal/server/session"
	"github.com/RobinThrift/stuff/views"
	"github.com/RobinThrift/stuff/views/pages"
)

type assetsListParams struct {
	Query     string `query:"query"`
	Page      int    `query:"page"`
	PageSize  int    `query:"page_size"`
	OrderBy   string `query:"order_by"`
	OrderDir  string `query:"order_dir"`
	AssetType string `query:"type"`
}

// [GET] /assets
func (rt *Router) assetsListHandler(w http.ResponseWriter, r *http.Request, params assetsListParams) error {
	if params.PageSize == 0 {
		params.PageSize = 25
	}

	list, err := rt.assets.List(r.Context(), control.ListAssetsQuery{
		SearchRaw:    params.Query,
		SearchFields: decodeSearchQuery(params.Query),
		Page:         params.Page,
		PageSize:     params.PageSize,
		OrderBy:      params.OrderBy,
		OrderDir:     params.OrderDir,
		AssetType:    entities.AssetType(strings.ToUpper(params.AssetType)),
	})
	if err != nil {
		return err
	}

	page := &pages.AssetListPage{
		Assets: list,
		Search: params.Query,
	}

	return page.Render(w, r)
}

type assetsGetParams struct {
	TagOrID string `url:"id"`
}

// [GET] /assets/{id}
func (rt *Router) assetsGetHandler(w http.ResponseWriter, r *http.Request, params assetsGetParams) error {
	query := getAssetQuery(params.TagOrID)
	query.IncludeParent = true
	query.IncludeParts = true
	query.IncludePurchases = true
	query.IncludeChildren = true
	asset, err := rt.assets.Get(r.Context(), query)
	if err != nil {
		return err
	}

	page := &pages.AssetViewPage{
		Asset:            asset,
		DecimalSeparator: rt.config.DecimalSeparator,
	}

	return page.Render(w, r)
}

// [GET] /assets/{id}/files
func (rt *Router) assetFilesHandler(w http.ResponseWriter, r *http.Request, params assetsGetParams) error {
	query := getAssetQuery(params.TagOrID)
	query.IncludeFiles = true
	asset, err := rt.assets.Get(r.Context(), query)
	if err != nil {
		return err
	}

	page := &pages.AssetFilesPage{
		Asset: asset,
	}

	return page.Render(w, r)
}

// [POST] /assets/{id}/files
func (rt *Router) assetFilesNewSubmitHandler(w http.ResponseWriter, r *http.Request, params assetsGetParams) error {
	user, ok := session.Get[*auth.User](r.Context(), "user")
	if !ok {
		return errors.New("can't find user in session")
	}

	asset, err := rt.getAsset(r.Context(), params.TagOrID)
	if err != nil {
		return err
	}

	files, err := fileUploads(r, asset.ID, user.ID)
	if err != nil {
		return err
	}

	for _, file := range files {
		_, err = rt.files.WriteFile(r.Context(), file)
		if err != nil {
			return err
		}
	}

	return nil
}

type fileDeleteParams struct {
	TagOrID string `url:"id"`
	FileID  int64  `url:"fileID"`
}

// [GET] /assets/{id}/files/{fileID}/delete
func (rt *Router) assetFilesDeleteHandler(w http.ResponseWriter, r *http.Request, params fileDeleteParams) error {
	file, err := rt.files.Get(r.Context(), params.FileID)
	if err != nil {
		if errors.Is(err, control.ErrFileNotFound) {
			return views.ErrorPageErr{Err: err, Code: http.StatusNotFound}
		}
		return err
	}

	page := &pages.AssetFileDeletePage{File: file}

	return page.Render(w, r)
}

// [POST] /assets/{id}/files/{fileID}/delete
func (rt *Router) assetFilesDeleteSubmitHandler(w http.ResponseWriter, r *http.Request, params fileDeleteParams) error {
	err := rt.files.Delete(r.Context(), params.FileID)
	if err != nil {
		return err
	}

	http.Redirect(w, r, "/assets/"+params.TagOrID+"/files", http.StatusFound)
	return nil
}

// [GET] /assets/new
func (rt *Router) assetsNewHandler(w http.ResponseWriter, r *http.Request, params struct{}) error {
	tag, err := rt.tags.GetNext(r.Context())
	if err != nil {
		return err
	}

	page := &pages.AssetEditPage{
		Asset:            &entities.Asset{Tag: tag},
		IsNew:            true,
		ValidationErrs:   map[string]string{},
		DecimalSeparator: rt.config.DecimalSeparator,
		DefaultCurrency:  rt.config.DefaultCurrency,
	}
	return page.Render(w, r)
}

// [POST] /assets/new
func (rt *Router) assetsNewSubmitHandler(w http.ResponseWriter, r *http.Request, params struct{}) error {
	user, ok := session.Get[*auth.User](r.Context(), "user")
	if !ok {
		return errors.New("can't find user in session")
	}

	page := &pages.AssetEditPage{
		Asset:            &entities.Asset{},
		IsNew:            true,
		ValidationErrs:   map[string]string{},
		DecimalSeparator: rt.config.DecimalSeparator,
		DefaultCurrency:  rt.config.DefaultCurrency,
	}

	err := rt.forms.Decode(page.Asset, r.PostForm)
	if err != nil {
		return err
	}

	validationErrs := map[string]string{}

	if page.Asset.Name == "" {
		validationErrs["name"] = "Name must not be empty"
	}

	if page.Asset.Tag == "" {
		validationErrs["tag"] = "Tag must not be empty"
	}

	if page.Asset.Category == "" {
		validationErrs["category"] = "Category must not be empty"
	}

	sanitizeAssetFields(page.Asset)

	if len(validationErrs) != 0 {
		return page.Render(w, r)
	}

	img, err := handleFileUpload(r, "image")
	if err != nil {
		page.ValidationErrs["general"] = err.Error()
		return page.Render(w, r)
	}

	page.Asset.MetaInfo.CreatedBy = user.ID

	created, err := rt.assets.Create(r.Context(), control.CreateAssetCmd{Asset: page.Asset, Image: img})
	if err != nil {
		return err
	}

	session.Put(r.Context(), "info_message", fmt.Sprintf("New asset '%s' created", created.Name))

	http.Redirect(w, r, fmt.Sprintf("/assets/%v", created.ID), http.StatusFound)
	return nil
}

type editAssetParams struct {
	TagOrID string `url:"id"`
}

// [GET] /assets/{id}/edit
func (rt *Router) assetsEditHandler(w http.ResponseWriter, r *http.Request, params editAssetParams) error {
	query := getAssetQuery(params.TagOrID)
	query.IncludeParts = true
	query.IncludePurchases = true
	asset, err := rt.assets.Get(r.Context(), query)
	if err != nil {
		if errors.Is(err, control.ErrAssetNotFound) {
			return views.ErrorPageErr{Err: err, Code: http.StatusNotFound}
		}

		return err
	}

	page := &pages.AssetEditPage{
		Asset:            asset,
		ValidationErrs:   map[string]string{},
		DecimalSeparator: rt.config.DecimalSeparator,
		DefaultCurrency:  rt.config.DefaultCurrency,
	}
	return page.Render(w, r)
}

// [POST] /assets/{id}/edit
func (rt *Router) assetsEditSubmitHandler(w http.ResponseWriter, r *http.Request, params editAssetParams) error {
	user, ok := session.Get[*auth.User](r.Context(), "user")
	if !ok {
		return errors.New("can't find user in session")
	}

	var err error
	page := &pages.AssetEditPage{ValidationErrs: map[string]string{}, DecimalSeparator: rt.config.DecimalSeparator, DefaultCurrency: rt.config.DefaultCurrency}

	query := getAssetQuery(params.TagOrID)
	query.IncludeParts = true
	query.IncludePurchases = true
	page.Asset, err = rt.assets.Get(r.Context(), query)
	if err != nil {
		if errors.Is(err, control.ErrAssetNotFound) {
			return views.ErrorPageErr{Err: err, Code: http.StatusNotFound}
		}

		return fmt.Errorf("error getting asset %v: %w", params.TagOrID, err)
	}

	// set to nil so items can be deleted
	page.Asset.Parts = nil
	page.Asset.Purchases = nil
	page.Asset.CustomAttrs = nil

	err = rt.forms.Decode(page.Asset, r.PostForm)
	if err != nil {
		return fmt.Errorf("error parsing form: %w", err)
	}

	// manually set to allow the removal of the parent ID when sending an empty string
	if r.PostForm.Get("parent_asset_id") == "" {
		page.Asset.ParentAssetID = 0
	}

	validationErrs := map[string]string{}

	if page.Asset.Name == "" {
		validationErrs["name"] = "Name must not be empty"
	}

	if page.Asset.Tag == "" {
		validationErrs["tag"] = "Tag must not be empty"
	}

	if page.Asset.Category == "" {
		validationErrs["category"] = "Category must not be empty"
	}

	sanitizeAssetFields(page.Asset)

	if len(validationErrs) != 0 {
		return page.Render(w, r)
	}

	image, err := handleFileUpload(r, "image")
	if err != nil {
		page.ValidationErrs["general"] = err.Error()
		return page.Render(w, r)
	}

	for i := range page.Asset.Parts {
		if page.Asset.Parts[i].CreatedBy == 0 {
			page.Asset.Parts[i].CreatedBy = user.ID
		}
	}

	updated, err := rt.assets.Update(r.Context(), control.UpdateAssetCmd{Asset: page.Asset, Image: image})
	if err != nil {
		return fmt.Errorf("error updating asset: %w", err)
	}

	session.Put(r.Context(), "info_message", fmt.Sprintf("Asset '%s' updated", updated.Name))

	http.Redirect(w, r, fmt.Sprintf("/assets/%v", updated.ID), http.StatusFound)
	return nil
}

type deleteAssetParams struct {
	TagOrID string `url:"id"`
}

// [GET] /assets/{id}/delete
func (rt *Router) assetsDeleteHandler(w http.ResponseWriter, r *http.Request, params deleteAssetParams) error {
	asset, err := rt.getAsset(r.Context(), params.TagOrID)
	if err != nil {
		return err
	}

	page := pages.AssetDeletePage{
		Asset: asset,
	}

	return page.Render(w, r)
}

// [POST] /assets/{id}/delete
func (rt *Router) assetsDeleteSubmitHandler(w http.ResponseWriter, r *http.Request, params deleteAssetParams) error {
	var err error
	page := pages.AssetDeletePage{}

	page.Asset, err = rt.getAsset(r.Context(), params.TagOrID)
	if err != nil {
		return err
	}

	err = rt.assets.Delete(r.Context(), page.Asset)
	if err != nil {
		slog.ErrorContext(r.Context(), "error deleting asset", "error", err)
		page.Message = err.Error()
		return page.Render(w, r)
	}

	session.Put(r.Context(), "info_message", fmt.Sprintf("Asset '%s' deleted", page.Asset.Name))

	http.Redirect(w, r, "/assets", http.StatusFound)
	return nil
}

func (rt *Router) getAsset(ctx context.Context, tagOrID string) (*entities.Asset, error) {
	var query control.GetAssetQuery
	if id, err := strconv.ParseInt(tagOrID, 10, 64); err == nil {
		query.Tag = tagOrID
		query.ID = id
	} else {
		query.Tag = tagOrID
	}

	asset, err := rt.assets.Get(ctx, getAssetQuery(tagOrID))
	if err != nil {
		if errors.Is(err, control.ErrAssetNotFound) {
			return nil, views.ErrorPageErr{Err: err, Code: http.StatusNotFound}
		}

		return nil, err
	}

	return asset, nil
}

func getAssetQuery(tagOrID string) control.GetAssetQuery {
	var query control.GetAssetQuery
	if id, err := strconv.ParseInt(tagOrID, 10, 64); err == nil {
		query.Tag = tagOrID
		query.ID = id
	} else {
		query.Tag = tagOrID
	}

	return query
}

func sanitizeAssetFields(asset *entities.Asset) {
	asset.Tag = policy.Sanitize(asset.Tag)
	asset.Name = policy.Sanitize(asset.Name)
	asset.Category = policy.Sanitize(asset.Category)
	asset.Model = policy.Sanitize(asset.Model)
	asset.ModelNo = policy.Sanitize(asset.ModelNo)
	asset.SerialNo = policy.Sanitize(asset.SerialNo)
	asset.Manufacturer = policy.Sanitize(asset.Manufacturer)
	asset.Location = policy.Sanitize(asset.Location)
	asset.PositionCode = policy.Sanitize(asset.PositionCode)
	asset.Notes = policy.Sanitize(asset.Notes)
	asset.QuantityUnit = policy.Sanitize(asset.QuantityUnit)
	for i := range asset.Purchases {
		asset.Purchases[i].Supplier = policy.Sanitize(asset.Purchases[i].Supplier)
		asset.Purchases[i].OrderNo = policy.Sanitize(asset.Purchases[i].OrderNo)
		asset.Purchases[i].Currency = policy.Sanitize(asset.Purchases[i].Currency)
	}
}