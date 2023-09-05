package assets

import (
	"encoding/json"
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
	"github.com/gorilla/csrf"
	"github.com/kodeshack/stuff/api"
	"github.com/kodeshack/stuff/server/session"
	"github.com/kodeshack/stuff/users"
	"github.com/kodeshack/stuff/views"
	"github.com/microcosm-cc/bluemonday"
)

var policy = bluemonday.StrictPolicy()

type Router struct {
	Control          *Control
	Decoder          *form.Decoder
	DefaultCurrency  string
	DecimalSeparator string
	FileDir          string
}

func (rt *Router) RegisterRoutes(mux *chi.Mux) {
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

	mux.Get("/api/v1/assets/categories", rt.apiListCategories)
}

// [GET] /
// [GET] /assets
func (rt *Router) handleAssetsListGet(w http.ResponseWriter, r *http.Request) error {
	query := listAssetsQueryFromURL(r.URL.Query())
	assetList, err := rt.Control.listAssets(r.Context(), query)
	if err != nil {
		return err
	}

	return renderListAssetsPage(w, r, assetList, query)
}

// [GET] /assets/{id}
func (rt *Router) handleAssetsViewGet(w http.ResponseWriter, r *http.Request) error {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		http.Redirect(w, r, "/assets", http.StatusFound)
		return nil
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return err
	}

	asset, err := rt.Control.getAsset(r.Context(), id)
	if err != nil {
		return err
	}

	viewAssetPage := viewAssetPage(asset, rt.DecimalSeparator)

	page := views.Document(asset.Name, viewAssetPage)

	err = page.Render(r.Context(), w)
	if err != nil {
		return fmt.Errorf("error rendering view asset page: %w", err)
	}

	return nil
}

// [GET] /assets/new
func (rt *Router) handleAssetsNewGet(w http.ResponseWriter, r *http.Request) error {
	return rt.renderEditAssetPage(w, r, editAssetPageProps{asset: &Asset{}, isNewAsset: true, validationErrs: map[string]string{}})
}

// [POST] /assets/new
func (rt *Router) handleAssetsNewPost(w http.ResponseWriter, r *http.Request) error {
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

	asset.Notes = policy.Sanitize(asset.Notes)

	if len(validationErrs) != 0 {
		return rt.renderEditAssetPage(w, r, editAssetPageProps{isNewAsset: true, asset: &asset, validationErrs: validationErrs})
	}

	file, err := handleFileUpload(r, "image")
	if err != nil {
		validationErrs["general"] = err.Error()
		return rt.renderEditAssetPage(w, r, editAssetPageProps{isNewAsset: true, asset: &asset, validationErrs: validationErrs})
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
func (rt *Router) handleAssetsEditGet(w http.ResponseWriter, r *http.Request) error {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		http.Redirect(w, r, "/assets", http.StatusFound)
		return nil
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return err
	}

	asset, err := rt.Control.getAsset(r.Context(), id)
	if err != nil {
		return err
	}

	return rt.renderEditAssetPage(w, r, editAssetPageProps{isNewAsset: false, asset: asset, validationErrs: map[string]string{}})
}

// [POST] /assets/{id}/edit
func (rt *Router) handleAssetsEditPost(w http.ResponseWriter, r *http.Request) error {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		http.Redirect(w, r, "/assets", http.StatusFound)
		return nil
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return err
	}

	asset, err := rt.Control.getAsset(r.Context(), id)
	if err != nil {
		return err
	}

	err = rt.Decoder.Decode(asset, r.PostForm)
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

	asset.Notes = policy.Sanitize(asset.Notes)

	if len(validationErrs) != 0 {
		return rt.renderEditAssetPage(w, r, editAssetPageProps{isNewAsset: false, asset: asset, validationErrs: validationErrs})
	}

	file, err := handleFileUpload(r, "image")
	if err != nil {
		validationErrs["general"] = err.Error()
		return rt.renderEditAssetPage(w, r, editAssetPageProps{isNewAsset: false, asset: asset, validationErrs: validationErrs})
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
func (rt *Router) handleAssetsDeleteGet(w http.ResponseWriter, r *http.Request) error {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		http.Redirect(w, r, "/assets", http.StatusFound)
		return nil
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return err
	}

	asset, err := rt.Control.getAsset(r.Context(), id)
	if err != nil {
		return err
	}

	return renderDeleteAssetPage(w, r, asset, "")
}

// [DELETE] /assets/{id}/delete
func (rt *Router) handleAssetsDeleteDelete(w http.ResponseWriter, r *http.Request) error {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		http.Redirect(w, r, "/assets", http.StatusFound)
		return nil
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return err
	}

	asset, err := rt.Control.getAsset(r.Context(), id)
	if err != nil {
		return err
	}

	err = rt.Control.deleteAsset(r.Context(), asset)
	if err != nil {
		slog.ErrorContext(r.Context(), "error deleting asset", "error", err)
		return renderDeleteAssetPage(w, r, asset, err.Error())
	}

	session.Put(r.Context(), "info_message", fmt.Sprintf("Asset '%s' deleted", asset.Name))

	http.Redirect(w, r, "/assets", http.StatusFound)
	return nil
}

// [GET] /api/v1/assets/categories
func (rt *Router) apiListCategories(w http.ResponseWriter, r *http.Request) {
	cats, err := rt.Control.listCategories(r.Context())
	if err != nil {
		api.RespondWithError(r.Context(), w, err)
		return
	}

	b, err := json.Marshal(cats)
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

func renderListAssetsPage(w http.ResponseWriter, r *http.Request, assetList *AssetList, query listAssetsQuery) error {
	infomsg, _ := session.Pop[string](r.Context(), "info_message")

	listAssetsPage := listAssetsPage(listAssetsPageProps{
		assets:  assetList.Assets,
		total:   assetList.Total,
		query:   query,
		infomsg: infomsg,
	})
	page := views.Document("Assets", listAssetsPage)

	err := page.Render(r.Context(), w)
	if err != nil {
		return fmt.Errorf("error rendering list assets page: %w", err)
	}

	return nil
}

func (rt *Router) renderEditAssetPage(w http.ResponseWriter, r *http.Request, props editAssetPageProps) error {
	props.decimalSeparator = rt.DecimalSeparator

	if props.asset.PurchaseInfo.Currency == "" {
		props.asset.PurchaseInfo.Currency = rt.DefaultCurrency
	}

	props.postTarget = "/assets/new"
	props.csrfToken = csrf.Token(r)

	title := "New Asset"
	if !props.isNewAsset {
		title = "Edit Asset"
		props.postTarget = fmt.Sprintf("/assets/%v/edit", props.asset.ID)
	}

	if props.asset.Tag == "" {
		tag, err := rt.Control.generateTag(r.Context())
		if err != nil {
			return err
		}
		props.asset.Tag = tag
	}

	csrfErr, ok := session.Pop[string](r.Context(), "csrf_error")
	if ok {
		props.validationErrs["general"] = csrfErr
	}

	editAssetPage := editAssetPage(props)

	page := views.Document(title, editAssetPage)

	err := page.Render(r.Context(), w)
	if err != nil {
		return fmt.Errorf("error rendering edit asset page: %w", err)
	}

	return nil
}

func renderDeleteAssetPage(w http.ResponseWriter, r *http.Request, asset *Asset, msg string) error {
	csrfErr, ok := session.Pop[string](r.Context(), "csrf_error")
	if ok {
		if msg != "" {
			msg += "\n"
		}
		msg += csrfErr
	}

	deleteAssetPage := deleteAssetPage(deleteAssetPageProps{asset: asset, errMsg: msg, csrfToken: csrf.Token(r)})

	page := views.Document("Delete Asset "+asset.Name, deleteAssetPage)

	err := page.Render(r.Context(), w)
	if err != nil {
		return fmt.Errorf("error rendering delete asset page: %w", err)
	}

	return nil
}

func listAssetsQueryFromURL(params url.Values) listAssetsQuery {
	q := listAssetsQuery{
		limit:   50,
		orderBy: params.Get("order_by"),
	}

	if size := params.Get("page_size"); size != "" {
		q.limit, _ = strconv.Atoi(size)
	}

	if pageStr := params.Get("page"); pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err == nil {
			q.offset = q.limit * page
		}
	}

	if orderDir := params.Get("order_dir"); orderDir != "" {
		orderDir = strings.ToUpper(orderDir)
		if orderDir == "ASC" || orderDir == "DESC" {
			q.orderDir = orderDir
		}
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
		fmt.Printf("err %#v\n", err)
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
