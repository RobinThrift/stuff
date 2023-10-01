package assets

import (
	"errors"
	"fmt"
	"log/slog"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/form/v4"
	"github.com/kodeshack/stuff/server/session"
	"github.com/kodeshack/stuff/users"
	"github.com/kodeshack/stuff/views"
	"github.com/microcosm-cc/bluemonday"
)

var policy = bluemonday.StrictPolicy()

type UIRouter struct {
	Control          *Control
	Decoder          *form.Decoder
	DefaultCurrency  string
	DecimalSeparator string
	FileDir          string
}

func (rt *UIRouter) RegisterRoutes(mux *chi.Mux) {
	mux.Handle("/assets/files/*", http.StripPrefix("/assets/files/", http.FileServer(http.Dir(rt.FileDir))))

	mux.Get("/", views.HTTPHandlerFuncErr(rt.handleAssetsListGet))
	mux.Get("/assets", views.HTTPHandlerFuncErr(rt.handleAssetsListGet))
	mux.Get("/assets/new", views.HTTPHandlerFuncErr(rt.handleAssetsNewGet))
	mux.Post("/assets/new", views.HTTPHandlerFuncErr(rt.handleAssetsNewPost))
	mux.Get("/assets/{id}", views.HTTPHandlerFuncErr(rt.handleAssetsViewGet))
	mux.Get("/assets/{id}/edit", views.HTTPHandlerFuncErr(rt.handleAssetsEditGet))
	mux.Post("/assets/{id}/edit", views.HTTPHandlerFuncErr(rt.handleAssetsEditPost))

	mux.Get("/assets/{id}/delete", views.HTTPHandlerFuncErr(rt.handleAssetsDeleteGet))
	mux.Post("/assets/{id}/delete", views.HTTPHandlerFuncErr(rt.handleAssetsDeleteDelete))

	mux.Get("/assets/export/json", views.HTTPHandlerFuncErr(rt.handleAssetsExportJSON))
	mux.Get("/assets/export/csv", views.HTTPHandlerFuncErr(rt.handleAssetsExportCSV))
}

// [GET] /
// [GET] /assets
func (rt *UIRouter) handleAssetsListGet(w http.ResponseWriter, r *http.Request) error {
	query := listAssetsQueryFromURL(r.URL.Query())

	page, err := rt.Control.listAssets(r.Context(), query)
	if err != nil {
		return err
	}

	return renderListAssetsPage(w, r, query, page)
}

// [GET] /assets/{id}
func (rt *UIRouter) handleAssetsViewGet(w http.ResponseWriter, r *http.Request) error {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Redirect(w, r, "/assets", http.StatusFound)
		return nil
	}

	asset, err := rt.Control.getAsset(r.Context(), id)
	if err != nil {
		return err
	}

	return renderViewAssetPage(w, r, ViewAssetsPageViewModel{
		Asset:            asset,
		DecimalSeparator: rt.DecimalSeparator,
	})
}

// [GET] /assets/new
func (rt *UIRouter) handleAssetsNewGet(w http.ResponseWriter, r *http.Request) error {
	return rt.renderEditAssetPage(w, r, EditAssetsPageViewModel{Asset: &Asset{}, IsNew: true, ValidationErrs: map[string]string{}})
}

// [POST] /assets/new
func (rt *UIRouter) handleAssetsNewPost(w http.ResponseWriter, r *http.Request) error {
	user, ok := session.Get[*users.User](r.Context(), "user")
	if !ok {
		return errors.New("can't find user in session")
	}

	var asset Asset

	err := rt.Decoder.Decode(&asset, r.PostForm)
	if err != nil {
		return err
	}

	validationErrs := map[string]string{}

	if asset.Name == "" {
		validationErrs["name"] = "Name must not be empty"
	}

	if asset.Tag == "" {
		validationErrs["tag"] = "Tag must not be empty"
	}

	if asset.Category == "" {
		validationErrs["category"] = "Category must not be empty"
	}

	sanitizeAssetFields(&asset)

	if len(validationErrs) != 0 {
		return rt.renderEditAssetPage(w, r, EditAssetsPageViewModel{Asset: &asset, IsNew: true, ValidationErrs: validationErrs})
	}

	file, err := handleFileUpload(r, "image")
	if err != nil {
		validationErrs["general"] = err.Error()
		return rt.renderEditAssetPage(w, r, EditAssetsPageViewModel{Asset: &asset, IsNew: true, ValidationErrs: validationErrs})
	}

	asset.MetaInfo.CreatedBy = user.ID

	created, err := rt.Control.createAsset(r.Context(), &asset, file)
	if err != nil {
		return err
	}

	session.Put(r.Context(), "info_message", fmt.Sprintf("New asset '%s' created", created.Name))

	http.Redirect(w, r, "/assets", http.StatusFound)
	return nil
}

// [GET] /assets/{id}/edit
func (rt *UIRouter) handleAssetsEditGet(w http.ResponseWriter, r *http.Request) error {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Redirect(w, r, "/assets", http.StatusFound)
		return nil
	}

	asset, err := rt.Control.getAsset(r.Context(), id)
	if err != nil {
		return err
	}

	return rt.renderEditAssetPage(w, r, EditAssetsPageViewModel{Asset: asset, ValidationErrs: map[string]string{}})
}

// [POST] /assets/{id}/edit
func (rt *UIRouter) handleAssetsEditPost(w http.ResponseWriter, r *http.Request) error {
	user, ok := session.Get[*users.User](r.Context(), "user")
	if !ok {
		return errors.New("can't find user in session")
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		http.Redirect(w, r, "/assets", http.StatusFound)
		return nil
	}

	asset, err := rt.Control.getAsset(r.Context(), id)
	if err != nil {
		return err
	}

	asset.Parts = nil

	err = rt.Decoder.Decode(asset, r.PostForm)
	if err != nil {
		return err
	}

	// manually set to allow the removal of the parent ID when sending an empty string
	if r.PostForm.Get("parent_asset_id") == "" {
		asset.ParentAssetID = 0
	}

	validationErrs := map[string]string{}

	if asset.Name == "" {
		validationErrs["name"] = "Name must not be empty"
	}

	if asset.Tag == "" {
		validationErrs["tag"] = "Tag must not be empty"
	}

	if asset.Category == "" {
		validationErrs["category"] = "Category must not be empty"
	}

	sanitizeAssetFields(asset)

	if len(validationErrs) != 0 {
		return rt.renderEditAssetPage(w, r, EditAssetsPageViewModel{Asset: asset, IsNew: true, ValidationErrs: validationErrs})
	}

	file, err := handleFileUpload(r, "image")
	if err != nil {
		validationErrs["general"] = err.Error()
		return rt.renderEditAssetPage(w, r, EditAssetsPageViewModel{Asset: asset, IsNew: true, ValidationErrs: validationErrs})
	}

	for i := range asset.Parts {
		if asset.Parts[i].CreatedBy == 0 {
			asset.Parts[i].CreatedBy = user.ID
		}
	}

	updated, err := rt.Control.updateAsset(r.Context(), asset, file)
	if err != nil {
		return err
	}

	session.Put(r.Context(), "info_message", fmt.Sprintf("Asset '%s' updated", updated.Name))

	http.Redirect(w, r, "/assets", http.StatusFound)
	return nil
}

// [GET] /assets/{id}/delete
func (rt *UIRouter) handleAssetsDeleteGet(w http.ResponseWriter, r *http.Request) error {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Redirect(w, r, "/assets", http.StatusFound)
		return nil
	}

	asset, err := rt.Control.getAsset(r.Context(), id)
	if err != nil {
		return err
	}

	return renderDeleteAssetPage(w, r, DeleteAssetsPageViewModel{Asset: asset})
}

// [DELETE] /assets/{id}/delete
func (rt *UIRouter) handleAssetsDeleteDelete(w http.ResponseWriter, r *http.Request) error {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Redirect(w, r, "/assets", http.StatusFound)
		return nil
	}

	asset, err := rt.Control.getAsset(r.Context(), id)
	if err != nil {
		return err
	}

	err = rt.Control.deleteAsset(r.Context(), asset)
	if err != nil {
		slog.ErrorContext(r.Context(), "error deleting asset", "error", err)
		return renderDeleteAssetPage(w, r, DeleteAssetsPageViewModel{Asset: asset, Message: err.Error()})
	}

	session.Put(r.Context(), "info_message", fmt.Sprintf("Asset '%s' deleted", asset.Name))

	http.Redirect(w, r, "/assets", http.StatusFound)
	return nil
}

// [GET] /assets/export/json
func (rt *UIRouter) handleAssetsExportJSON(w http.ResponseWriter, r *http.Request) error {
	assets, err := rt.Control.listAssets(r.Context(), ListAssetsQuery{})
	if err != nil {
		return err
	}

	w.Header().Add("content-disposition", `attachment; filename="assets_export.json"`)
	w.Header().Add("content-type", "application/json; charset=utf-8")

	return exportAssetsAsJSON(w, assets.Assets)
}

// [GET] /assets/export/csv
func (rt *UIRouter) handleAssetsExportCSV(w http.ResponseWriter, r *http.Request) error {
	assets, err := rt.Control.listAssets(r.Context(), ListAssetsQuery{})
	if err != nil {
		return err
	}

	w.Header().Add("content-disposition", `attachment; filename="assets_export.csv"`)
	w.Header().Add("content-type", "text/csv; charset=utf-8")
	return exportAssetsAsCSV(w, assets.Assets)
}

func listAssetsQueryFromURL(params url.Values) ListAssetsQuery {
	q := ListAssetsQuery{ //nolint: varnamelen
		OrderBy: params.Get("order_by"),
	}

	if query := params.Get("query"); query != "" {
		q.Search = decodeSearchQuery(query)
	}

	if size := params.Get("page_size"); size != "" {
		q.PageSize, _ = strconv.Atoi(size)
	}

	if pageStr := params.Get("page"); pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err == nil {
			q.Page = q.PageSize * page
		}
	}

	if orderDir := params.Get("order_dir"); orderDir != "" {
		orderDir = strings.ToUpper(orderDir)
		if orderDir == "ASC" || orderDir == "DESC" {
			q.OrderDir = orderDir
		}
	}

	if q.PageSize == 0 {
		q.PageSize = 50
	}

	if q.PageSize > 100 {
		q.PageSize = 100
	}

	return q
}

var imgAllowList = []string{"image/png", "image/jpeg", "image/webp"}
var errImgTypeNotAllowed = errors.New("image type not allowed")

// @TODO: implement file size check
func checkImageUpload(header *multipart.FileHeader) error {
	ct := header.Header.Get("content-type")
	mt, _, err := mime.ParseMediaType(ct)
	if err != nil {
		return err
	}

	allowed := false
	for _, m := range imgAllowList {
		if mt == m {
			allowed = true
			break
		}
	}

	if !allowed {
		return errImgTypeNotAllowed
	}

	return nil
}

func handleFileUpload(r *http.Request, key string) (*File, error) {
	_, hasFileUpload := r.MultipartForm.File[key]
	if !hasFileUpload {
		return nil, nil
	}

	uploaded, header, err := r.FormFile(key)
	if err != nil {
		return nil, err
	}

	if uploaded != nil {
		err = checkImageUpload(header)
		if err != nil {
			return nil, err
		}

		return &File{Name: header.Filename, r: uploaded}, nil
	}

	return nil, nil
}

func NewDecoder(decimalSeparator string) *form.Decoder {
	decoder := form.NewDecoder()

	decoder.SetMode(form.ModeExplicit)
	decoder.RegisterCustomTypeFunc(func(s []string) (interface{}, error) {
		if len(s) == 0 || len(s[0]) == 0 {
			return time.Time{}, nil
		}
		return time.Parse("2006-01-02", s[0])
	}, time.Time{})

	decoder.RegisterCustomTypeFunc(func(s []string) (interface{}, error) {
		if len(s) == 0 || len(s[0]) == 0 {
			return MonetaryAmount(0), nil
		}

		base := 0
		fractional := 0
		var err error

		split := strings.SplitN(s[0], decimalSeparator, 2)
		base, err = strconv.Atoi(split[0])
		if err != nil {
			return nil, err
		}

		if len(split) == 2 {
			fractional, err = strconv.Atoi(split[1])
			if err != nil {
				return nil, err
			}
		}

		return MonetaryAmount(base*100 + fractional), nil
	}, MonetaryAmount(0))

	return decoder
}

func decodeSearchQuery(queryStr string) *ListAssetsQuerySearch {
	queryStr = strings.TrimPrefix(queryStr, "*")
	q := &ListAssetsQuerySearch{Raw: queryStr, Fields: map[string]string{}} //nolint: varnamelen

	lastWordEnd := 0
	lastNameEnd := 0
	name := ""
	value := ""
	for i := 0; i < len(queryStr)-1; i++ {
		switch queryStr[i] {
		case ':':
			value = queryStr[lastNameEnd:lastWordEnd]
			if name != "" {
				q.Fields[strings.ToLower(name)] = value
			}
			if queryStr[lastWordEnd] == ' ' {
				lastWordEnd += 1
			}
			name = queryStr[lastWordEnd:i]
			lastNameEnd = i + 1
			if i+1 < len(queryStr) && queryStr[i+1] == ' ' {
				lastNameEnd = i + 2
			}
		case ' ':
			lastWordEnd = i
		}
	}

	if name != "" {
		value = queryStr[lastNameEnd:]
		if value != "" {
			q.Fields[strings.ToLower(name)] = queryStr[lastNameEnd:]
		}
	}

	return q
}

func sanitizeAssetFields(asset *Asset) {
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
	asset.PurchaseInfo.Supplier = policy.Sanitize(asset.PurchaseInfo.Supplier)
	asset.PurchaseInfo.OrderNo = policy.Sanitize(asset.PurchaseInfo.OrderNo)
	asset.PurchaseInfo.Currency = policy.Sanitize(asset.PurchaseInfo.Currency)
}
