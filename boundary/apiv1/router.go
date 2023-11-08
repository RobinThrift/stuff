//lint:file-ignore SA1019 Ignore because generated code produces these
package apiv1

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/RobinThrift/stuff/auth"
	"github.com/RobinThrift/stuff/control"
	"github.com/RobinThrift/stuff/entities"
	"github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/go-chi/chi/v5"
)

type Router struct {
	assets        AssetCtrl
	customAttrs   CustomAttrCtrl
	categories    CategoryCtrl
	suppliers     SupplierCtrl
	locations     LocationCtrl
	models        ModelCtrl
	manufacturers ManufacturerCtrl

	tags  TagCtrl
	users UserCtrl
}

type AssetCtrl interface {
	Get(ctx context.Context, query control.GetAssetQuery) (*entities.Asset, error)
	List(ctx context.Context, query control.ListAssetsQuery) (*entities.ListPage[*entities.Asset], error)
	Create(ctx context.Context, cmd control.CreateAssetCmd) (*entities.Asset, error)
	Update(ctx context.Context, cmd control.UpdateAssetCmd) (*entities.Asset, error)
	Delete(ctx context.Context, asset *entities.Asset) error
}

type TagCtrl interface {
	List(ctx context.Context, query control.ListTagsQuery) (*entities.ListPage[*entities.Tag], error)
}

type CustomAttrCtrl interface {
	List(ctx context.Context, query control.ListCustomAttrsQuery) (*entities.ListPage[*entities.CustomAttr], error)
}

type SupplierCtrl interface {
	List(ctx context.Context, query control.ListSuppliersQuery) (*entities.ListPage[*entities.Supplier], error)
}

type CategoryCtrl interface {
	List(ctx context.Context, query control.ListCategoriesQuery) (*entities.ListPage[*entities.Category], error)
}

type LocationCtrl interface {
	ListLocations(ctx context.Context, query control.ListLocationsQuery) (*entities.ListPage[*entities.Location], error)
	ListPositionCodes(ctx context.Context, query control.ListPositionCodesQuery) (*entities.ListPage[*entities.PositionCode], error)
}

type ModelCtrl interface {
	List(ctx context.Context, query control.ListModelsQuery) (*entities.ListPage[*entities.Model], error)
}

type ManufacturerCtrl interface {
	List(ctx context.Context, query control.ListManufacturersQuery) (*entities.ListPage[*entities.Manufacturer], error)
}

type UserCtrl interface {
	List(ctx context.Context, query control.ListUsersQuery) (*entities.ListPage[*auth.User], error)
}

func NewRouter(
	mux chi.Router,

	assets AssetCtrl,
	categories CategoryCtrl,
	customAttrs CustomAttrCtrl,
	suppliers SupplierCtrl,
	locations LocationCtrl,
	models ModelCtrl,
	manufacturers ManufacturerCtrl,

	tags TagCtrl,
	users UserCtrl,
) *Router {
	r := &Router{ //nolint: varnamelen
		assets:        assets,
		customAttrs:   customAttrs,
		categories:    categories,
		suppliers:     suppliers,
		locations:     locations,
		models:        models,
		manufacturers: manufacturers,
		tags:          tags,
		users:         users,
	}

	errorHandlerFunc := func(w http.ResponseWriter, r *http.Request, err error) {
		slog.ErrorContext(r.Context(), err.Error(), "error", err)

		code := http.StatusInternalServerError

		apiErr := Error{
			Code:   http.StatusInternalServerError,
			Detail: err.Error(),
			Title:  http.StatusText(http.StatusInternalServerError),
			Type:   "stuff/api/v1/internalServerError",
		}

		if errors.Is(err, auth.ErrUnauthorized) {
			code = http.StatusUnauthorized
			apiErr = Error{
				Code:   http.StatusUnauthorized,
				Title:  http.StatusText(http.StatusUnauthorized),
				Detail: err.Error(),
				Type:   "stuff/api/v1/Unauthorized",
			}
		}

		w.WriteHeader(code)

		b, err := json.Marshal(apiErr)
		if err != nil {
			slog.ErrorContext(r.Context(), "error while trying to marshal api error to json", "error", err)
			return
		}

		_, err = w.Write(b)
		if err != nil {
			slog.ErrorContext(r.Context(), "error while writing http response", "error", err)
			return
		}
	}

	HandlerWithOptions(NewStrictHandlerWithOptions(r, nil, StrictHTTPServerOptions{
		RequestErrorHandlerFunc:  errorHandlerFunc,
		ResponseErrorHandlerFunc: errorHandlerFunc,
	}), ChiServerOptions{
		BaseRouter:       mux,
		ErrorHandlerFunc: errorHandlerFunc,
	})

	return r
}

// (GET /v1/assets)
func (r *Router) ListAssets(ctx context.Context, req ListAssetsRequestObject) (ListAssetsResponseObject, error) {
	list, err := r.assets.List(ctx, control.ListAssetsQuery{
		SearchRaw:    valFromPtr(req.Params.Query),
		SearchFields: decodeSearchQuery(req.Params.Query),
		Page:         valFromPtr(req.Params.Page),
		PageSize:     valFromPtr(req.Params.PageSize),
		OrderBy:      valFromPtr(req.Params.OrderBy),
		OrderDir:     valFromPtr(req.Params.OrderDir),
		AssetType:    entities.AssetType(valFromPtr(req.Params.Type)),
	})
	if err != nil {
		return nil, err
	}

	assets := make([]Asset, 0, len(list.Items))
	for _, asset := range list.Items {
		assets = append(assets, mapAssetToAPI(asset))
	}

	return ListAssets200JSONResponse{
		Assets:   assets,
		NumPages: list.NumPages,
		Page:     list.Page,
		PageSize: list.PageSize,
		Total:    list.Total,
	}, nil
}

// (POST /v1/assets)
func (r *Router) CreateAsset(ctx context.Context, req CreateAssetRequestObject) (CreateAssetResponseObject, error) {
	asset := mapCreateAssetBodyToAsset(req.Body)

	created, err := r.assets.Create(ctx, control.CreateAssetCmd{Asset: asset})
	if err != nil {
		return nil, err
	}

	return CreateAsset201JSONResponse(mapAssetToAPI(created)), nil
}

// (GET /v1/assets/{tagOrID})
func (r *Router) GetAsset(ctx context.Context, req GetAssetRequestObject) (GetAssetResponseObject, error) {
	query := control.GetAssetQuery{
		IncludeParts:     true,
		IncludePurchases: true,
		IncludeFiles:     true,
		IncludeChildren:  valFromPtr(req.Params.IncludeChildren),
	}

	if id, err := strconv.ParseInt(req.TagOrID, 10, 64); err == nil {
		query.Tag = req.TagOrID
		query.ID = id
	} else {
		query.Tag = req.TagOrID
	}

	asset, err := r.assets.Get(ctx, query)
	if err != nil {
		return nil, err
	}

	return GetAsset200JSONResponse(mapAssetToAPI(asset)), nil
}

// (PUT /v1/assets/{tagOrID})
func (r *Router) UpdateAsset(ctx context.Context, req UpdateAssetRequestObject) (UpdateAssetResponseObject, error) {
	var query control.GetAssetQuery
	if id, err := strconv.ParseInt(req.TagOrID, 10, 64); err == nil {
		query.Tag = req.TagOrID
		query.ID = id
	} else {
		query.Tag = req.TagOrID
	}

	asset, err := r.assets.Get(ctx, query)
	if err != nil {
		return nil, err
	}

	mapUpdateIntoAsset(asset, req.Body)

	updated, err := r.assets.Update(ctx, control.UpdateAssetCmd{Asset: asset})
	if err != nil {
		return nil, err
	}

	return UpdateAsset200JSONResponse(mapAssetToAPI(updated)), nil
}

// (DELETE /v1/assets/{tagOrID})
func (r *Router) DeleteAsset(ctx context.Context, req DeleteAssetRequestObject) (DeleteAssetResponseObject, error) {
	var query control.GetAssetQuery
	if id, err := strconv.ParseInt(req.TagOrID, 10, 64); err == nil {
		query.Tag = req.TagOrID
		query.ID = id
	} else {
		query.Tag = req.TagOrID
	}

	asset, err := r.assets.Get(ctx, query)
	if err != nil {
		return nil, err
	}

	err = r.assets.Delete(ctx, asset)
	if err != nil {
		return nil, err
	}

	return DeleteAsset204Response{}, nil
}

// (GET /v1/categories)
func (r *Router) ListCategories(ctx context.Context, req ListCategoriesRequestObject) (ListCategoriesResponseObject, error) {
	categories, err := r.categories.List(ctx, control.ListCategoriesQuery{
		Search:   valFromPtr(req.Params.Query),
		Page:     valFromPtr(req.Params.Page),
		PageSize: valFromPtr(req.Params.PageSize),
	})
	if err != nil {
		return nil, err
	}

	cats := make([]Category, 0, len(categories.Items))
	for _, cat := range categories.Items {
		cats = append(cats, Category{Name: cat.Name})
	}

	return ListCategories200JSONResponse{
		Categories: cats,
		NumPages:   categories.NumPages,
		Page:       categories.Page,
		PageSize:   categories.PageSize,
		Total:      categories.Total,
	}, nil
}

// (GET /v1/locations)
func (r *Router) ListLocations(ctx context.Context, req ListLocationsRequestObject) (ListLocationsResponseObject, error) {
	list, err := r.locations.ListLocations(ctx, control.ListLocationsQuery{
		Search:   valFromPtr(req.Params.Query),
		Page:     valFromPtr(req.Params.Page),
		PageSize: valFromPtr(req.Params.PageSize),
	})
	if err != nil {
		return nil, err
	}

	locations := make([]Location, 0, len(list.Items))
	for _, location := range list.Items {
		locations = append(locations, Location{Name: location.Name})
	}

	return ListLocations200JSONResponse{
		Locations: locations,
		NumPages:  list.NumPages,
		Page:      list.Page,
		PageSize:  list.PageSize,
		Total:     list.Total,
	}, nil
}

// (GET /v1/locations/position_codes)
func (r *Router) ListPositionCodes(ctx context.Context, req ListPositionCodesRequestObject) (ListPositionCodesResponseObject, error) {
	list, err := r.locations.ListPositionCodes(ctx, control.ListPositionCodesQuery{
		Search:   valFromPtr(req.Params.Query),
		Page:     valFromPtr(req.Params.Page),
		PageSize: valFromPtr(req.Params.PageSize),
	})
	if err != nil {
		return nil, err
	}

	posCodes := make([]PositionCode, 0, len(list.Items))
	for _, pos := range list.Items {
		posCodes = append(posCodes, PositionCode{Code: pos.Code})
	}

	return ListPositionCodes200JSONResponse{
		PositionCodes: posCodes,
		NumPages:      list.NumPages,
		Page:          list.Page,
		PageSize:      list.PageSize,
		Total:         list.Total,
	}, nil
}

// (GET /v1/manufacturers)
func (r *Router) ListManufacturers(ctx context.Context, req ListManufacturersRequestObject) (ListManufacturersResponseObject, error) {
	list, err := r.manufacturers.List(ctx, control.ListManufacturersQuery{
		Search:   valFromPtr(req.Params.Query),
		Page:     valFromPtr(req.Params.Page),
		PageSize: valFromPtr(req.Params.PageSize),
	})
	if err != nil {
		return nil, err
	}

	manufacturers := make([]Manufacturer, 0, len(list.Items))
	for _, model := range list.Items {
		manufacturers = append(manufacturers, Manufacturer{Name: model.Name})
	}

	return ListManufacturers200JSONResponse{
		Manufacturers: manufacturers,
		NumPages:      list.NumPages,
		Page:          list.Page,
		PageSize:      list.PageSize,
		Total:         list.Total,
	}, nil
}

// (GET /v1/models)
func (r *Router) ListModels(ctx context.Context, req ListModelsRequestObject) (ListModelsResponseObject, error) {
	list, err := r.models.List(ctx, control.ListModelsQuery{
		Search:   valFromPtr(req.Params.Query),
		Page:     valFromPtr(req.Params.Page),
		PageSize: valFromPtr(req.Params.PageSize),
	})
	if err != nil {
		return nil, err
	}

	models := make([]Model, 0, len(list.Items))
	for _, model := range list.Items {
		models = append(models, Model{Name: model.Name, ModelNo: &model.ModelNo})
	}

	return ListModels200JSONResponse{
		Models:   models,
		NumPages: list.NumPages,
		Page:     list.Page,
		PageSize: list.PageSize,
		Total:    list.Total,
	}, nil
}

// (GET /v1/custom_attrs)
func (r *Router) ListCustomAttrs(ctx context.Context, req ListCustomAttrsRequestObject) (ListCustomAttrsResponseObject, error) {
	list, err := r.customAttrs.List(ctx, control.ListCustomAttrsQuery{
		Search:   valFromPtr(req.Params.Query),
		Page:     valFromPtr(req.Params.Page),
		PageSize: valFromPtr(req.Params.PageSize),
	})
	if err != nil {
		return nil, err
	}

	customAttrs := make([]CustomAttr, 0, len(list.Items))
	for _, supplier := range list.Items {
		customAttrs = append(customAttrs, CustomAttr{Name: supplier.Name})
	}

	return ListCustomAttrs200JSONResponse{
		CustomAttrs: customAttrs,
		NumPages:    list.NumPages,
		Page:        list.Page,
		PageSize:    list.PageSize,
		Total:       list.Total,
	}, nil
}

// (GET /v1/suppliers)
func (r *Router) ListSuppliers(ctx context.Context, req ListSuppliersRequestObject) (ListSuppliersResponseObject, error) {
	list, err := r.suppliers.List(ctx, control.ListSuppliersQuery{
		Search:   valFromPtr(req.Params.Query),
		Page:     valFromPtr(req.Params.Page),
		PageSize: valFromPtr(req.Params.PageSize),
	})
	if err != nil {
		return nil, err
	}

	suppliers := make([]Supplier, 0, len(list.Items))
	for _, supplier := range list.Items {
		suppliers = append(suppliers, Supplier{Name: supplier.Name})
	}

	return ListSuppliers200JSONResponse{
		Suppliers: suppliers,
		NumPages:  list.NumPages,
		Page:      list.Page,
		PageSize:  list.PageSize,
		Total:     list.Total,
	}, nil
}

// (GET /v1/tags)
func (r *Router) ListTags(ctx context.Context, req ListTagsRequestObject) (ListTagsResponseObject, error) {
	list, err := r.tags.List(ctx, control.ListTagsQuery{
		Page:     valFromPtr(req.Params.Page),
		PageSize: valFromPtr(req.Params.PageSize),
	})
	if err != nil {
		return nil, err
	}

	tags := make([]Tag, 0, len(list.Items))
	for _, tag := range list.Items {
		tags = append(tags, mapTagToAPI(tag))
	}

	return ListTags200JSONResponse{
		Tags:     tags,
		NumPages: list.NumPages,
		Page:     list.Page,
		PageSize: list.PageSize,
		Total:    list.Total,
	}, nil
}

// (GET /v1/users)
func (r *Router) ListUsers(ctx context.Context, req ListUsersRequestObject) (ListUsersResponseObject, error) {
	list, err := r.users.List(ctx, control.ListUsersQuery{
		Page:     valFromPtr(req.Params.Page),
		PageSize: valFromPtr(req.Params.PageSize),
	})
	if err != nil {
		return nil, err
	}

	users := make([]User, 0, len(list.Items))
	for _, user := range list.Items {
		users = append(users, User{
			Id:          user.ID,
			Username:    user.Username,
			DisplayName: user.DisplayName,
			IsAdmin:     user.IsAdmin,
			CreatedAt:   types.Date{Time: user.CreatedAt},
			UpdatedAt:   types.Date{Time: user.UpdatedAt},
		})
	}

	return ListUsers200JSONResponse{
		Users:    users,
		NumPages: list.NumPages,
		Page:     list.Page,
		PageSize: list.PageSize,
		Total:    list.Total,
	}, nil
}
