package control

import (
	"context"
	"testing"
	"time"

	"github.com/RobinThrift/stuff/storage/database"
	"github.com/RobinThrift/stuff/storage/database/sqlite"
	"github.com/stephenafamo/bob"
	"github.com/stretchr/testify/assert"
)

func TestTagControl_CRUD(t *testing.T) {
	algorithms := []string{"nanoid", "ksuid", "uuid", "sequential"}

	for _, algo := range algorithms {
		t.Run(algo, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			t.Cleanup(cancel)

			tagCtrl := newTestTagControl(t, algo)

			// create a set of 10 new tags that are marked as used
			tagSet := map[string]int{}
			for i := 0; i < 10; i++ {
				next, err := tagCtrl.GetNext(ctx)
				assert.NoError(t, err)
				tagSet[next] = 1

				_, err = tagCtrl.CreateIfNotExists(ctx, next)
				assert.NoError(t, err)
			}

			assert.Len(t, tagSet, 10)

			// fetch a list of used tags
			inUse := true
			usedTags, err := tagCtrl.List(ctx, ListTagsQuery{InUse: &inUse})
			assert.NoError(t, err)
			assert.Len(t, usedTags.Items, 10)

			// fetch a list of unused tags
			inUse = false
			unusedTags, err := tagCtrl.List(ctx, ListTagsQuery{InUse: &inUse})
			assert.NoError(t, err)
			assert.Len(t, unusedTags.Items, 0)

			// mark all used tags as unused
			for _, tag := range usedTags.Items {
				err = tagCtrl.MarkTagUnused(ctx, tag.Tag)
				assert.NoError(t, err)
			}

			// fetch a list of unused tags
			unusedTags, err = tagCtrl.List(ctx, ListTagsQuery{InUse: &inUse})
			assert.NoError(t, err)
			assert.Len(t, unusedTags.Items, 10)

			// fetch a list of used tags
			inUse = true
			usedTags, err = tagCtrl.List(ctx, ListTagsQuery{InUse: &inUse})
			assert.NoError(t, err)
			assert.Len(t, usedTags.Items, 0)

			// mark all unused tags as used
			for _, tag := range unusedTags.Items {
				_, err = tagCtrl.CreateIfNotExists(ctx, tag.Tag)
				assert.NoError(t, err)
			}

			// fetch a list of used tags
			inUse = true
			usedTags, err = tagCtrl.List(ctx, ListTagsQuery{InUse: &inUse})
			assert.NoError(t, err)
			assert.Len(t, usedTags.Items, 10)

			// fetch a list of unused tags
			inUse = false
			unusedTags, err = tagCtrl.List(ctx, ListTagsQuery{InUse: &inUse})
			assert.NoError(t, err)
			assert.Len(t, unusedTags.Items, 0)

			// try to delete used tags (should not allow the deletion of used tags)
			for _, tag := range usedTags.Items {
				err = tagCtrl.Delete(ctx, tag.Tag)
				assert.NoError(t, err)
			}

			// fetch a list of used tags, the list should not have changed
			inUse = true
			usedTags, err = tagCtrl.List(ctx, ListTagsQuery{InUse: &inUse})
			assert.NoError(t, err)
			assert.Len(t, usedTags.Items, 10)

			// mark all used tags as unused again and then delete
			for _, tag := range usedTags.Items {
				err = tagCtrl.MarkTagUnused(ctx, tag.Tag)
				assert.NoError(t, err)

				err = tagCtrl.Delete(ctx, tag.Tag)
				assert.NoError(t, err)
			}

			// fetch list of all tags
			allTags, err := tagCtrl.List(ctx, ListTagsQuery{})
			assert.NoError(t, err)
			assert.Len(t, allTags.Items, 0)
		})
	}

}

func newTestTagControl(t *testing.T, algorithm string) *TagControl {
	db, err := sqlite.NewSQLiteDB(&sqlite.Config{File: ":memory:", Timeout: time.Millisecond * 500})
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

	return NewTagControl(database, algorithm, &sqlite.TagRepo{})
}
