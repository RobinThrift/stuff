package sqlite

import (
	"github.com/stephenafamo/bob"
	_ "modernc.org/sqlite"
)

func NewSQLiteDB(path string) (bob.DB, error) {
	return bob.Open("sqlite", path)
}
