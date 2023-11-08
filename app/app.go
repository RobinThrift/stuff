package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/RobinThrift/stuff/auth"
	"github.com/RobinThrift/stuff/boundary/apiv1"
	"github.com/RobinThrift/stuff/boundary/htmlui"
	"github.com/RobinThrift/stuff/control"
	"github.com/RobinThrift/stuff/internal/log"
	"github.com/RobinThrift/stuff/internal/server"
	"github.com/RobinThrift/stuff/jobs"
	"github.com/RobinThrift/stuff/storage/blobs"
	"github.com/RobinThrift/stuff/storage/database"
	"github.com/RobinThrift/stuff/storage/database/sqlite"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/stephenafamo/bob"
)

func Start() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	start, stop, err := setup(ctx)
	if err != nil {
		return err
	}

	go func() {
		shutdown := make(chan os.Signal, 1)
		signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

		sig := <-shutdown

		stopCtx, stopCtxCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer stopCtxCancel()

		slog.InfoContext(ctx, fmt.Sprintf("received signal %v: triggering shutdown", sig))

		err := stop(stopCtx) //nolint: contextcheck // false positive
		if err != nil {
			slog.ErrorContext(ctx, "could not stop gracefully", "error", err)
		}
	}()

	return start(ctx)
}

func setup(ctx context.Context) (func(context.Context) error, func(context.Context) error, error) {
	config, err := NewConfigFromEnv()
	if err != nil {
		return nil, nil, err
	}

	baseURL, err := url.Parse(config.BaseURL)
	if err != nil {
		return nil, nil, err
	}

	err = log.SetupLogger(config.LogLevel, config.LogFormat)
	if err != nil {
		return nil, nil, err
	}

	db, err := sqlite.NewSQLiteDB(config.Database.Path)
	if err != nil {
		return nil, nil, err
	}

	err = sqlite.RunMigrations(ctx, db)
	if err != nil {
		return nil, nil, err
	}

	database := &database.Database{DB: bob.NewDB(db)}

	userCtrl := control.NewUserCtrl(database, &sqlite.UserRepo{})
	authCtrl := control.NewAuthController(control.AuthConfig{
		Argon2Params: auth.Argon2Params{
			KeyLen:  config.Auth.Local.Argon2Params.KeyLen,
			Memory:  config.Auth.Local.Argon2Params.Memory,
			Threads: config.Auth.Local.Argon2Params.Threads,
			Time:    config.Auth.Local.Argon2Params.Time,
			Version: config.Auth.Local.Argon2Params.Version,
		},
	}, database, userCtrl, &sqlite.LocalAuthRepo{})
	tagCtrl := control.NewTagControl(database, config.TagAlgorithm, &sqlite.TagRepo{})
	fileCtrl := control.NewFileControl(database, &sqlite.FileRepo{}, &blobs.LocalFS{
		RootDir: config.FileDir,
		TmpDir:  config.TmpDir,
	})
	assetCtrl := control.NewAssetControl(
		database,
		tagCtrl,
		fileCtrl,
		&sqlite.AssetRepo{},
	)
	categoryCtrl := control.NewCategoryCtrl(database, &sqlite.CategoryRepo{})
	locationCtrl := control.NewLocationControl(database, &sqlite.LocationRepo{})
	modelCtrl := control.NewModelCtrl(database, &sqlite.ModelRepo{})
	manufacturerCtrl := control.NewManufactuerCtrl(database, &sqlite.ManufacturerRepo{})
	supplierCtrl := control.NewSupplierCtrl(database, &sqlite.SupplierRepo{})
	customAttrCtrl := control.NewCustomAttrCtrl(database, &sqlite.CustomAttrRepo{})

	importerCtrl := control.NewImporterCtrl(database, assetCtrl, tagCtrl)
	exporterCtrl := control.NewExporterCtrl(database, assetCtrl)
	labelsCtrl := control.NewLabelController(assetCtrl)

	initJob := jobs.NewInitJob(jobs.InitJobConfig{
		Username: "admin",
		Password: config.Auth.Local.InitialAdminPassword,
	}, database, authCtrl, userCtrl)

	if err = initJob.Run(ctx); err != nil {
		return nil, nil, errors.Join(db.Close(), err)
	}

	sm := scs.New()
	sm.Store = sqlite.NewSQLiteSessionStore(database.DB) //nolint:contextcheck // false positive IMO
	sm.Lifetime = 24 * time.Hour
	sm.Cookie.HttpOnly = true
	sm.Cookie.Persist = true
	sm.Cookie.SameSite = http.SameSiteStrictMode

	srv, err := server.NewServer(config.Addr, config.UseSecureCookies, sm)
	if err != nil {
		return nil, nil, err
	}

	srv.Mux.Route("/api", func(r chi.Router) {
		apiv1.NewRouter(
			r,
			assetCtrl,
			categoryCtrl,
			customAttrCtrl,
			supplierCtrl,
			locationCtrl,
			modelCtrl,
			manufacturerCtrl,
			tagCtrl,
			userCtrl,
		)
	})

	htmlui.NewRouter(
		srv.Mux,
		htmlui.Config{
			BaseURL:          baseURL,
			DecimalSeparator: config.DecimalSeparator,
			DefaultCurrency:  config.DefaultCurrency,
			AssetFilesDir:    config.FileDir,
		},
		authCtrl,
		assetCtrl,
		fileCtrl,
		tagCtrl,
		userCtrl,
		importerCtrl,
		exporterCtrl,
		labelsCtrl,
	)

	start := func(ctx context.Context) error {
		defer func() {
			if err := db.Close(); err != nil {
				slog.ErrorContext(ctx, "error closing database", "error", err)
			}
		}()

		return srv.Start(ctx)
	}

	stop := func(ctx context.Context) error {
		slog.InfoContext(ctx, "closing database")
		if err := db.Close(); err != nil {
			return fmt.Errorf("error closing database: %w", err)
		}

		return srv.Stop(ctx)
	}

	return start, stop, nil
}
