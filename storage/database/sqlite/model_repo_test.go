package sqlite

import (
	"context"
	"fmt"
	"testing"

	"github.com/RobinThrift/stuff/entities"
	"github.com/RobinThrift/stuff/storage/database"
	"github.com/stephenafamo/bob"
	"github.com/stretchr/testify/assert"
)

func TestModelRepoList(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	mr, exec := newTestModelRepo(t) // nolint: varnamelen // this is fine for the test

	models := []struct {
		name string
		num  string
	}{
		{"model-0", "model-num-0"},
		{"model-1", "model-num-1"},
		{"model-2", "model-num-2"},
		{"model-3", "model-num-3"},
		{"model-0", ""},
		{"model-1", ""},
		{"model-2", ""},
		{"model-3", ""},
	}

	ar := &AssetRepo{}
	for i := 0; i <= 100; i++ {
		model := models[i%len(models)]
		err := ar.Create(ctx, exec, &entities.Asset{
			Type:    entities.AssetTypeConsumable,
			Status:  entities.StatusInUse,
			Tag:     fmt.Sprintf("#tag-%d", i),
			Name:    fmt.Sprintf("Test Asset %d", i),
			Model:   model.name,
			ModelNo: model.num,
		})
		assert.NoError(t, err)
	}

	list, err := mr.List(ctx, exec, database.ListModelsQuery{})
	assert.NoError(t, err)
	assert.Len(t, list.Items, 8)
}

func newTestModelRepo(t *testing.T) (*ModelRepo, bob.Executor) {
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

	return &ModelRepo{}, bob.NewDB(db)
}
