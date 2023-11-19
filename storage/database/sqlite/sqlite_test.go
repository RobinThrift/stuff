package sqlite

import (
	"context"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRunMigrations(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	file := path.Join(t.TempDir(), t.Name()+".db")

	db, err := NewSQLiteDB(&Config{File: file, Timeout: time.Millisecond * 500})
	assert.NoError(t, err)

	err = RunMigrations(ctx, db)
	assert.NoError(t, err)
}
