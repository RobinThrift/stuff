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

func TestTagRepo_CRUD(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	tr, exec := newTestTagRepo(t) // nolint: varnamelen // this is fine for the test

	for i := 0; i < 10; i++ {
		err := tr.Create(ctx, exec, &entities.Tag{Tag: fmt.Sprintf("tag-%d", i)})
		assert.NoError(t, err)
	}

	unused, err := tr.GetUnused(ctx, exec)
	assert.NoError(t, err)
	assert.NotNil(t, unused)

	for i := 0; i < 10; i++ {
		err = tr.MarkTagUsed(ctx, exec, fmt.Sprintf("tag-%d", i))
		assert.NoError(t, err)
	}

	unused, err = tr.GetUnused(ctx, exec)
	assert.NoError(t, err)
	assert.Nil(t, unused)

	for i := 0; i < 5; i++ {
		err = tr.MarkTagUnused(ctx, exec, fmt.Sprintf("tag-%d", i))
		assert.NoError(t, err)
	}

	unused, err = tr.GetUnused(ctx, exec)
	assert.NoError(t, err)
	assert.NotNil(t, unused)

	list, err := tr.List(ctx, exec, database.ListTagsQuery{})
	assert.NoError(t, err)
	assert.Len(t, list.Items, 10)

	for i := 0; i < 10; i++ {
		err = tr.MarkTagUnused(ctx, exec, fmt.Sprintf("tag-%d", i))
		assert.NoError(t, err)
		err = tr.Delete(ctx, exec, fmt.Sprintf("tag-%d", i))
		assert.NoError(t, err)
	}

	list, err = tr.List(ctx, exec, database.ListTagsQuery{})
	assert.NoError(t, err)
	assert.Len(t, list.Items, 0)
}

func TestTagRepo_NextSequential(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	tr, exec := newTestTagRepo(t)

	for i := 0; i < 10; i++ {
		err := tr.Create(ctx, exec, &entities.Tag{Tag: fmt.Sprintf("tag-%d", i)})
		assert.NoError(t, err)
	}

	next, err := tr.NextSequential(ctx, exec)
	assert.NoError(t, err)
	assert.Equal(t, int64(11), next)
}

func newTestTagRepo(t *testing.T) (*TagRepo, bob.Executor) {
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

	return &TagRepo{}, bob.NewDB(db)
}
