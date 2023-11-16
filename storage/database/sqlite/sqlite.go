package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"log/slog"

	"github.com/pressly/goose/v3"
	bobsqlite "github.com/stephenafamo/bob/dialect/sqlite"
	"github.com/stephenafamo/bob/dialect/sqlite/dialect"
	"github.com/stephenafamo/bob/mods"
	sqlite "modernc.org/sqlite"
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

func wrapSqliteErr(err error) error {
	var sqliteErr *sqlite.Error
	if errors.As(err, &sqliteErr) {
		return fmt.Errorf("%s: %s", sqlite.ErrorCodeString[sqliteErr.Code()], sqliteErr.Error())
	}

	return err

}

func orderByClause(table string, column string, dir string) mods.OrderBy[*dialect.SelectQuery] {
	return mods.OrderBy[*dialect.SelectQuery]{
		Expression: bobsqlite.Quote(table, column).String() + " COLLATE NOCASE",
		Direction:  dir,
	}
}
