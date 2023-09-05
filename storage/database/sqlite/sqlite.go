package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log/slog"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrations embed.FS

func NewSQLiteDB(path string) (*sql.DB, error) {
	slog.Info("opening SQLite database at " + path)
	return sql.Open("sqlite", path)
}

func RunMigrations(ctx context.Context, db *sql.DB) error {
	slog.InfoContext(ctx, "running migrations")

	goose.SetBaseFS(migrations)
	err := goose.SetDialect("sqlite3")
	if err != nil {
		panic(err)
	}

	goose.SetTableName("migrations")

	err = goose.Up(db, "migrations")
	if err != nil {
		return fmt.Errorf("error running migrations: %w", err)
	}

	slog.InfoContext(ctx, "successfully ran migrations")
	return nil
}
