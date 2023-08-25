package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"

	migrate "github.com/rubenv/sql-migrate"
	_ "modernc.org/sqlite"
)

//go:embed migrations
var migrations embed.FS

func NewSQLiteDB(path string) (*sql.DB, error) {
	slog.Info("opening SQLite database at " + path)
	return sql.Open("sqlite", path)
}

func RunMigrations(ctx context.Context, db *sql.DB) error {
	slog.InfoContext(ctx, "running migrations")

	subFS, err := fs.Sub(migrations, "migrations")
	if err != nil {
		return err
	}

	source := &migrate.HttpFileSystemMigrationSource{
		FileSystem: http.FS(subFS),
	}

	_, err = migrate.ExecContext(ctx, db, "sqlite3", source, migrate.Up)
	if err != nil {
		return fmt.Errorf("error running migrations: %w", err)
	}

	slog.InfoContext(ctx, "successfully ran migrations")
	return nil
}
