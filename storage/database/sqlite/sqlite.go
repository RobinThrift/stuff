package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log/slog"
	"time"

	"github.com/pressly/goose/v3"
	bobsqlite "github.com/stephenafamo/bob/dialect/sqlite"
	"github.com/stephenafamo/bob/dialect/sqlite/dialect"
	"github.com/stephenafamo/bob/mods"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed migrations/*.sql
var migrations embed.FS

type Config struct {
	File      string
	Timeout   time.Duration
	EnableWAL bool
}

func NewSQLiteDB(config *Config) (*sql.DB, error) {
	slog.Info("opening SQLite database at " + config.File)

	journalMode := ""
	if config.EnableWAL {
		journalMode = "&_journal_mode=wal"
	}

	connStr := fmt.Sprintf("%s?mode=rwc&cache=shared&_busy_timeout=%d&_foreign_keys=1&_txlock=immediate%s", config.File, config.Timeout.Milliseconds(), journalMode)

	return sql.Open("sqlite3", connStr)
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

func orderByClause(table string, column string, dir string) mods.OrderBy[*dialect.SelectQuery] {
	return mods.OrderBy[*dialect.SelectQuery]{
		Expression: bobsqlite.Quote(table, column).String() + " COLLATE NOCASE",
		Direction:  dir,
	}
}
