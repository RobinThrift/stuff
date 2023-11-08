package control

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"testing"

	"github.com/RobinThrift/stuff/entities"
	"github.com/RobinThrift/stuff/storage/blobs"
	"github.com/RobinThrift/stuff/storage/database"
	"github.com/RobinThrift/stuff/storage/database/sqlite"
	"github.com/stephenafamo/bob"
	"github.com/stretchr/testify/assert"
)

func TestFileControl_CRUD(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	fileCtrl := newTestFileControl(t)

	for i := 0; i < 5; i++ {
		f := newTestFile(t, i, 10)
		_, err := fileCtrl.WriteFile(ctx, f)
		assert.NoError(t, err)
		fileExitsts(t, f.FullPath)
	}

	for i := 5; i < 10; i++ {
		f := newTestFile(t, i, 20)
		_, err := fileCtrl.WriteFile(ctx, f)
		assert.NoError(t, err)
		fileExitsts(t, f.FullPath)
	}

	list, err := fileCtrl.List(ctx, ListFilesQuery{AssetID: 10})
	assert.NoError(t, err)
	assert.Len(t, list.Items, 5)

	err = fileCtrl.DeleteAllForAsset(ctx, 10)
	assert.NoError(t, err)

	list, err = fileCtrl.List(ctx, ListFilesQuery{})
	assert.NoError(t, err)
	assert.Len(t, list.Items, 5)
	for _, f := range list.Items {
		assert.NotEqual(t, 10, f.AssetID)
	}

	fileWithDuplicateHash := newTestFile(t, 9, 10)
	_, err = fileCtrl.WriteFile(ctx, fileWithDuplicateHash)
	assert.NoError(t, err)

	err = fileCtrl.DeleteAllForAsset(ctx, 20)
	assert.NoError(t, err)

	list, err = fileCtrl.List(ctx, ListFilesQuery{})
	assert.NoError(t, err)
	assert.Len(t, list.Items, 1)
}

func newTestFile(t *testing.T, i int, assetID int64) *entities.File {
	name := fmt.Sprintf("File-%d", i)

	h := sha256.New()
	_, err := h.Write([]byte(name))
	assert.NoError(t, err)

	filetype := randFrom([][]string{{"image/png", ".png"}, {"application/pdf", ".pdf"}, {"plain/text", ".txt"}})

	return &entities.File{
		Reader:    bytes.NewBuffer([]byte(name)),
		AssetID:   assetID,
		Name:      name + filetype[1],
		Filetype:  filetype[0],
		CreatedBy: 1,
	}
}

func randFrom[T any](items []T) T {
	i := rand.Intn(len(items))
	return items[i]
}

func newTestFileControl(t *testing.T) *FileControl {
	db, err := sqlite.NewSQLiteDB(":memory:")
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err = db.Close(); err != nil {
			t.Error(err)
		}
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = sqlite.RunMigrations(ctx, db)
	if err != nil {
		t.Fatal(err)
	}

	database := &database.Database{DB: bob.NewDB(db)}

	return NewFileControl(
		database,
		&sqlite.FileRepo{},
		&blobs.LocalFS{
			RootDir: t.TempDir(),
			TmpDir:  t.TempDir(),
		},
	)
}

func fileExitsts(t *testing.T, path string) {
	_, err := os.Stat(path)
	if err != nil {
		t.Error(err)
	}
}

func fileNotExitsts(t *testing.T, path string) {
	_, err := os.Stat(path)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			t.Error(err)
		}
	}
}
