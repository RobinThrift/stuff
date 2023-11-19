package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/pressly/goose/v3"
	bobsqlite "github.com/stephenafamo/bob/dialect/sqlite"
	"github.com/stephenafamo/bob/dialect/sqlite/dialect"
	"github.com/stephenafamo/bob/mods"
	sqlite "modernc.org/sqlite"
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
		journalMode = "&_pragma=journal_mode(wal)"
	}

	connStr := fmt.Sprintf("%s?&_pragma=busy_timeout(%d)&_pragma=foreign_keys(1)&_txlock=immediate%s", config.File, config.Timeout.Milliseconds(), journalMode)

	return sql.Open("sqlite", connStr)
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

func unwapSQLiteError(err error) error {
	var sqliteErr *sqlite.Error
	if errors.As(err, &sqliteErr) {
		if codeStr, ok := sqlite.ErrorCodeString[sqliteErr.Code()]; ok {
			return fmt.Errorf("%s: %s", codeStr, sqliteErr.Error())
		}
	}

	return err
}

func orderByClause(table string, column string, dir string) mods.OrderBy[*dialect.SelectQuery] {
	return mods.OrderBy[*dialect.SelectQuery]{
		Expression: bobsqlite.Quote(table, column).String() + " COLLATE NOCASE",
		Direction:  dir,
	}
}
