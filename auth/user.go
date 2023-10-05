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

	CreatedAt time.Time `form:"-"`
	UpdatedAt time.Time `form:"-"`
}

type UserListPage struct {
	Users    []*User
	Total    int
	NumPages int
	Page     int
	PageSize int
}
