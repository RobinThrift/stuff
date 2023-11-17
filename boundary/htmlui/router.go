package htmlui

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/RobinThrift/stuff/auth"
	"github.com/RobinThrift/stuff/control"
	"github.com/RobinThrift/stuff/entities"
	"github.com/go-chi/chi/v5"

	"github.com/go-playground/form/v4"
)

type Router struct {
	config   Config
	auth     AuthCtrl
	assets   AssetCtrl
	files    FileCtrl
	tags     TagCtrl
	users    UserCtrl
	importer ImporterCtrl
	exporter ExporterCtrl
	labels   LabelCtrl
	forms    *form.Decoder
}

type Config struct {
	BaseURL          *url.URL
	DecimalSeparator string
	DefaultCurrency  string
	AssetFilesDir    string
}

type AssetCtrl interface {
	Get(ctx context.Context, query control.GetAssetQuery) (*entities.Asset, error)
	List(ctx context.Context, query control.ListAssetsQuery) (*entities.ListPage[*entities.Asset], error)
	Create(ctx context.Context, cmd control.CreateAssetCmd) (*entities.Asset, error)
	Update(ctx context.Context, cmd control.UpdateAssetCmd) (*entities.Asset, error)
	Delete(ctx context.Context, asset *entities.Asset) error
}

type FileCtrl interface {
	Get(ctx context.Context, id int64) (*entities.File, error)
	WriteFile(ctx context.Context, file *entities.File) (*entities.File, error)
	Delete(ctx context.Context, id int64) error
}

type TagCtrl interface {
	List(ctx context.Context, query control.ListTagsQuery) (*entities.ListPage[*entities.Tag], error)
	GetNext(ctx context.Context) (string, error)
}

type ImporterCtrl interface {
	Import(r *http.Request, cmd control.ImportCmd) (map[string]string, error)
}

type ExporterCtrl interface {
	Export(ctx context.Context, w io.Writer, cmd control.ExportCmd) error
}

type UserCtrl interface {
	List(ctx context.Context, query control.ListUsersQuery) (*entities.ListPage[*auth.User], error)
	Update(ctx context.Context, user *auth.User) error
	Get(ctx context.Context, id int64) (*auth.User, error)
	SetUserPreferences(ctx context.Context, user *auth.User) error
}

type AuthCtrl interface {
	GetUserForCredentials(ctx context.Context, query control.GetUserForCredentialsQuery) (*auth.User, map[string]string, error)
	ChangeUserCredentials(ctx context.Context, cmd control.ChangeUserCredentialsCmd) (map[string]string, error)
	CreateUser(ctx context.Context, cmd control.CreateUserCmd) error
	ResetPassword(ctx context.Context, cmd control.ResetPasswordCmd) error
	ToggleAdmin(ctx context.Context, id int64) error
	DeleteUser(ctx context.Context, userID int64) error
}

type LabelCtrl interface {
	GenerateLabelSheet(ctx context.Context, query control.GenerateLabelSheetQuery) ([]byte, error)
}

func NewRouter(
	mux chi.Router,
	config Config,
	auth AuthCtrl,
	assets AssetCtrl,
	files FileCtrl,
	tags TagCtrl,
	users UserCtrl,
	importer ImporterCtrl,
	exporter ExporterCtrl,
	labels LabelCtrl,
) *Router {
	r := &Router{ //nolint: varnamelen
		config:   config,
		auth:     auth,
		assets:   assets,
		files:    files,
		tags:     tags,
		users:    users,
		importer: importer,
		exporter: exporter,
		labels:   labels,
		forms:    newDecoder(config.DecimalSeparator),
	}

	mux.Get("/login", viewRenderHandler(r.authLoginHandler))
	mux.Post("/login", viewRenderHandler(r.authLoginSubmitHandler))
	mux.Get("/logout", viewRenderHandler(r.authLogoutHandler))
	mux.Get("/auth/changepassword", viewRenderHandler(r.authChangePasswordHandler))
	mux.Post("/auth/changepassword", viewRenderHandler(r.authChangePasswordSubmitHandler))

	mux.Handle("/assets/files/*", http.StripPrefix("/assets/files/", http.FileServer(http.Dir(config.AssetFilesDir))))

	mux.Get("/", viewRenderHandler(r.assetsListHandler))
	mux.Get("/assets", viewRenderHandler(r.assetsListHandler))

	mux.Get("/tags", viewRenderHandler(r.tagsListHandler))

	mux.Get("/assets/{id}", viewRenderHandler(r.assetsGetHandler))
	mux.Post("/assets/{id}/files", viewRenderHandler(r.assetFilesNewSubmitHandler))
	mux.Get("/assets/{id}/files/{fileID}/delete", viewRenderHandler(r.assetFilesDeleteHandler))
	mux.Post("/assets/{id}/files/{fileID}/delete", viewRenderHandler(r.assetFilesDeleteSubmitHandler))

	mux.Get("/assets/new", viewRenderHandler(r.assetsNewHandler))
	mux.Post("/assets/new", viewRenderHandler(r.assetsNewSubmitHandler))

	mux.Get("/assets/{id}/edit", viewRenderHandler(r.assetsEditHandler))
	mux.Post("/assets/{id}/edit", viewRenderHandler(r.assetsEditSubmitHandler))

	mux.Get("/assets/{id}/delete", viewRenderHandler(r.assetsDeleteHandler))
	mux.Post("/assets/{id}/delete", viewRenderHandler(r.assetsDeleteSubmitHandler))

	mux.Get("/assets/import", viewRenderHandler(r.importAssetsHandler))
	mux.Post("/assets/import", viewRenderHandler(r.importAssetsSubmitHandler))

	mux.Get("/assets/export/labels", viewRenderHandler(r.labelsHandler))
	mux.Post("/assets/export/labels", viewRenderHandler(r.labelsSubmitHandler))

	mux.Get("/assets/export/{format}", viewRenderHandler(r.exportAssetsHandler))

	mux.Get("/users", viewRenderHandler(r.usersListHandler))

	mux.Get("/users/new", viewRenderHandler(r.usersNewHandler))
	mux.Post("/users/new", viewRenderHandler(r.usersNewSubmitHandler))

	mux.Get("/users/me", viewRenderHandler(r.usersCurrentHandler))
	mux.Post("/users/me", viewRenderHandler(r.usersCurrentSubmitHandler))
	mux.Get("/users/me/changepassword", viewRenderHandler(r.usersCurrentInitChangePasswordHandler))

	mux.Get("/users/{id}/reset_password", viewRenderHandler(r.usersResetPasswordHandler))
	mux.Post("/users/{id}/reset_password", viewRenderHandler(r.usersResetPasswordSubmitHandler))

	mux.Get("/users/{id}/toggle_admin", viewRenderHandler(r.usersToggleAdminHandler))

	mux.Get("/users/{id}/delete", viewRenderHandler(r.usersDeleteHandler))
	mux.Post("/users/{id}/delete", viewRenderHandler(r.usersDeleteSubmitHandler))

	mux.Post("/users/settings", r.usersSettingsSubmitHandler)

	return r
}

func newDecoder(decimalSeparator string) *form.Decoder {
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
			return entities.MonetaryAmount(0), nil
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

		return entities.MonetaryAmount(base*100 + fractional), nil
	}, entities.MonetaryAmount(0))

	return decoder
}
