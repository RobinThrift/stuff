package sqlite

import (
	"context"
	"crypto/sha256"
	"fmt"
	"testing"

	"github.com/RobinThrift/stuff/entities"
	"github.com/RobinThrift/stuff/storage/database"
	"github.com/stephenafamo/bob"
	"github.com/stretchr/testify/assert"
)

func TestFileRepo_CRUD(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	fr, exec := newTestFileRepo(t) // nolint: varnamelen // this is fine for the test

	for i := 0; i < 5; i++ {
		f := newTestFile(t, i, 10)
		_, err := fr.Create(ctx, exec, f)
		assert.NoError(t, err)
	}

	for i := 5; i < 10; i++ {
		f := newTestFile(t, i, 20)
		_, err := fr.Create(ctx, exec, f)
		assert.NoError(t, err)
	}

	publicPath := newTestFile(t, 0, 10).PublicPath
	fetchedByPublicPath, err := fr.GetByPublicPath(ctx, exec, publicPath)
	assert.NoError(t, err)
	assert.Equal(t, publicPath, fetchedByPublicPath.PublicPath)

	list, err := fr.List(ctx, exec, database.ListFilesQuery{AssetID: 10})
	assert.NoError(t, err)
	assert.Len(t, list.Items, 5)

	ids := make([]int64, 0, len(list.Items))
	for _, f := range list.Items {
		ids = append(ids, f.ID)
	}

	err = fr.Delete(ctx, exec, ids)
	assert.NoError(t, err)

	list, err = fr.List(ctx, exec, database.ListFilesQuery{})
	assert.NoError(t, err)
	assert.Len(t, list.Items, 5)
	for _, f := range list.Items {
		assert.NotEqual(t, 10, f.AssetID)
	}

	fileWithDuplicateHash := newTestFile(t, 9, 10)
	insertedID, err := fr.Create(ctx, exec, fileWithDuplicateHash)
	assert.NoError(t, err)

	list, err = fr.List(ctx, exec, database.ListFilesQuery{Hashes: [][]byte{fileWithDuplicateHash.Sha256}})
	assert.NoError(t, err)
	assert.Len(t, list.Items, 2)
	assert.Equal(t, list.Items[0].Sha256, list.Items[1].Sha256)

	fetched, err := fr.Get(ctx, exec, insertedID)
	assert.NoError(t, err)

	fileWithDuplicateHash.ID = insertedID
	fileWithDuplicateHash.CreatedAt = fetched.CreatedAt
	fileWithDuplicateHash.UpdatedAt = fetched.UpdatedAt
	assert.Equal(t, fileWithDuplicateHash, fetched)
}

var filetypes = [][]string{{"image/png", ".png"}, {"application/pdf", ".pdf"}, {"plain/text", ".txt"}}

func newTestFile(t *testing.T, i int, assetID int64) *entities.File {
	name := fmt.Sprintf("File-%d", i)

	h := sha256.New()
	_, err := h.Write([]byte(name))
	assert.NoError(t, err)

	hash := h.Sum(nil)

	filetype := filetypes[i%len(filetypes)]

	return &entities.File{
		AssetID:    assetID,
		Name:       name,
		Filetype:   filetype[0],
		SizeBytes:  int64(len(name)),
		PublicPath: "/assets/files/" + name + filetype[1],
		FullPath:   "/var/run/stuff/files/" + name + filetype[1],
		Sha256:     hash,
		CreatedBy:  1,
	}
}

func newTestFileRepo(t *testing.T) (*FileRepo, bob.Executor) {
	db, err := NewSQLiteDB(":memory:")
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

	err = RunMigrations(ctx, db)
	if err != nil {
		t.Fatal(err)
	}

	return &FileRepo{}, bob.NewDB(db)
}
