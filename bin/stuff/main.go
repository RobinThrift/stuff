package main

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

	"github.com/RobinThrift/stuff/assets"
	"github.com/RobinThrift/stuff/auth"
	"github.com/RobinThrift/stuff/config"
	"github.com/RobinThrift/stuff/log"
	"github.com/RobinThrift/stuff/server"
	"github.com/RobinThrift/stuff/storage/database"
	"github.com/RobinThrift/stuff/storage/database/sqlite"
	"github.com/RobinThrift/stuff/tags"
	"github.com/alexedwards/scs/v2"
	"github.com/stephenafamo/bob"
)

func main() {
	if err := run(); err != nil {
		fmt.Println("error starting stuff", err)
		os.Exit(1)
	}
}

func run() error {
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
	config, err := config.NewConfigFromEnv()
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

	tagCtrl := &tags.Control{Algorithm: config.TagAlgorithm, DB: database, TagRepo: &tags.RepoSQLite{}}
	authCtrl := &auth.Control{
		DB:            database,
		UserRepo:      &auth.UserRepoSQLite{},
		LocalAuthRepo: &auth.LocalAuthRepoSQLite{},
		Argon2Params:  auth.Argon2Params(config.Auth.Local.Argon2Params),
	}
	assetsCtrl := &assets.Control{
		DB:        database,
		AssetRepo: &assets.RepoSQLite{},
		TagCtrl:   tagCtrl,
		FileDir:   config.FileDir,
		TmpDir:    config.TmpDir,
	}

	err = authCtrl.RunInitSetup(ctx, "admin", config.Auth.Local.InitialAdminPassword)
	if err != nil {
		return nil, nil, errors.Join(db.Close(), err)
	}

	authRouter := &auth.UIRouter{Control: authCtrl, Decoder: auth.NewDecoder()}
	assetsUIRouter := &assets.UIRouter{
		BaseURL:          baseURL,
		Control:          assetsCtrl,
		Decoder:          assets.NewDecoder(config.DecimalSeparator),
		DefaultCurrency:  config.DefaultCurrency,
		DecimalSeparator: config.DecimalSeparator,
		FileDir:          config.FileDir,
	}

	assetsAPIRouter := &assets.APIRouter{
		Control: assetsCtrl,
	}

	tagsUIRouter := &tags.UIRouter{Control: tagCtrl}
	tagAPIRouter := &tags.APIRouter{Control: tagCtrl}

	sm := scs.New()
	sm.Store = sqlite.NewSQLiteSessionStore(database.DB) //nolint:contextcheck // false positive IMO
	sm.Lifetime = 24 * time.Hour
	sm.Cookie.HttpOnly = true
	sm.Cookie.Persist = true
	sm.Cookie.SameSite = http.SameSiteStrictMode

	srv, err := server.NewServer(
		config.Addr,
		config.UseSecureCookies,
		sm,
		authRouter.RegisterRoutes,
		assetsUIRouter.RegisterRoutes,
		assetsAPIRouter.RegisterRoutes,
		tagsUIRouter.RegisterRoutes,
		tagAPIRouter.RegisterRoutes,
	)
	if err != nil {
		return nil, nil, err
	}

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
