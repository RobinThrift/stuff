package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/kodeshack/stuff/assets"
	"github.com/kodeshack/stuff/auth"
	"github.com/kodeshack/stuff/config"
	"github.com/kodeshack/stuff/log"
	"github.com/kodeshack/stuff/server"
	"github.com/kodeshack/stuff/storage/database"
	"github.com/kodeshack/stuff/storage/database/sqlite"
	"github.com/kodeshack/stuff/tags"
	"github.com/kodeshack/stuff/users"
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

	start, err := setup(ctx)
	if err != nil {
		return err
	}

	return start(ctx)
}

func setup(ctx context.Context) (func(context.Context) error, error) {
	config, err := config.NewConfigFromEnv()
	if err != nil {
		return nil, err
	}

	err = log.SetupLogger(config.LogLevel, config.LogFormat)
	if err != nil {
		return nil, err
	}

	db, err := sqlite.NewSQLiteDB(config.Database.Path)
	if err != nil {
		return nil, err
	}

	err = sqlite.RunMigrations(ctx, db)
	if err != nil {
		return nil, err
	}

	database := &database.Database{DB: bob.NewDB(db)}

	tagCtrl := &tags.Control{Algorithm: config.TagAlgorithm, DB: database, TagRepo: &tags.RepoSQLite{}}
	userCtrl := &users.Control{DB: database, UserRepo: &users.RepoSQLite{}}
	authCtrl := &auth.Control{
		DB:            database,
		Users:         userCtrl,
		LocalAuthRepo: &auth.RepoSQLite{},
		Argon2Params:  auth.Argon2Params(config.Auth.Local.Argon2Params),
	}
	assetsCtrl := &assets.Control{
		DB:        database,
		AssetRepo: &assets.RepoSQLite{},
		TagCtrl:   tagCtrl,
		FileDir:   config.FileDir,
	}

	err = authCtrl.RunInitSetup(ctx, "admin", config.Auth.Local.InitialAdminPassword)
	if err != nil {
		return nil, errors.Join(db.Close(), err)
	}

	authRouter := &auth.Router{Control: authCtrl}
	assetsUIRouter := &assets.UIRouter{
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

	usersUIRouter := &users.UIRouter{}

	sm := scs.New()
	sm.Store = sqlite.NewSQLiteSessionStore(database.DB) //nolint:contextcheck // false positive IMO
	sm.Lifetime = 24 * time.Hour
	sm.Cookie.HttpOnly = true
	sm.Cookie.Persist = true
	sm.Cookie.SameSite = http.SameSiteStrictMode

	srv, err := server.NewServer(config.Addr, sm,
		authRouter.RegisterRoutes,
		assetsUIRouter.RegisterRoutes,
		assetsAPIRouter.RegisterRoutes,
		tagsUIRouter.RegisterRoutes,
		tagAPIRouter.RegisterRoutes,
		usersUIRouter.RegisterRoutes,
	)
	if err != nil {
		return nil, err
	}

	return func(ctx context.Context) error {
		defer func() {
			if err := db.Close(); err != nil {
				slog.ErrorContext(ctx, "error closing database", "error", err)
			}
		}()

		return srv.Start(ctx)
	}, nil
}
