package sqlite

import (
	"context"
	"testing"
	"time"

	"github.com/RobinThrift/stuff/auth"
	"github.com/stephenafamo/bob"
	"github.com/stretchr/testify/assert"
)

func TestUserRepo_UpsertPreferences(t *testing.T) {
	userRepo, exec := newTestUserRepo(t)

	prefs := []auth.UserPreferences{
		{},
		{
			SidebarClosedDesktop: true,
			ThemeName:            "default",
			ThemeMode:            "dark",
			AssetListColumns:     []string{"id", "tag", "name"},
			AssetListCompact:     true,
			UserListCompact:      true,
		},
		{
			SidebarClosedDesktop: false,
			ThemeName:            "retro",
			ThemeMode:            "light",
			AssetListColumns:     []string{"id", "image", "name"},
			AssetListCompact:     false,
			UserListCompact:      false,
		},
		{
			SidebarClosedDesktop: true,
			ThemeName:            "default",
			ThemeMode:            "dark",
			AssetListColumns:     []string{"id", "tag", "name"},
			AssetListCompact:     true,
			UserListCompact:      true,
		},
		{
			ThemeName:        "default",
			ThemeMode:        "dark",
			AssetListColumns: []string{},
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err := userRepo.Create(ctx, exec, &auth.User{
		Username:               "testuser",
		DisplayName:            "testuser",
		IsAdmin:                true,
		RequiresPasswordChange: false,
		AuthRef:                "testuser",
	})
	assert.NoError(t, err)

	for i, prefs := range prefs {
		user, err := userRepo.GetByUsername(ctx, exec, "testuser")
		assert.NoError(t, err)

		user.Preferences = prefs

		err = userRepo.UpsertPreferences(ctx, exec, user)
		assert.NoError(t, err, i)

		user, err = userRepo.GetByUsername(ctx, exec, "testuser")
		assert.NoError(t, err, i)

		assert.Equal(t, prefs, user.Preferences, i)
	}
}

func newTestUserRepo(t *testing.T) (*UserRepo, bob.Executor) {
	db, err := NewSQLiteDB(&Config{File: ":memory:", Timeout: time.Millisecond * 500})
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

	return &UserRepo{}, bob.NewDB(db)
}
