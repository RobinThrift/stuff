package users

import (
	"errors"
	"time"
)

var ErrUserNotFound = errors.New("user not found")

type User struct {
	ID       int64
	Username string

	DisplayName string
	IsAdmin     bool

	RequiresPasswordChange bool

	AuthRef string

	CreatedAt time.Time
	UpdatedAt time.Time
}
