package sqlite

import (
	"context"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunMigrations(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	file := path.Join(t.TempDir(), t.Name()+".db")

	db, err := NewSQLiteDB(file)
	assert.NoError(t, err)

	err = RunMigrations(ctx, db)
	assert.NoError(t, err)
}
