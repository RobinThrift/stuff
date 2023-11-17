package auth

import (
	"errors"
	"time"
)

var ErrUserNotFound = errors.New("user not found")

type User struct {
	ID       int64  `form:"-"`
	Username string `form:"username"`

	DisplayName string `form:"display_name"`
	IsAdmin     bool   `form:"is_admin"`

	RequiresPasswordChange bool `form:"-"`

	AuthRef string `form:"-"`

	Preferences UserPreferences

	CreatedAt time.Time `form:"-"`
	UpdatedAt time.Time `form:"-"`
}

type UserPreferences struct {
	SidebarClosedDesktop bool

	ThemeName string
	ThemeMode string

	AssetListColumns []string
	AssetListCompact bool

	UserListCompact bool
}
