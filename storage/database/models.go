package database

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

	AuthRef string

	CreatedAt time.Time
	UpdatedAt time.Time
}

var ErrLocalAuthUserNotFound = errors.New("user for local auth not found")

type LocalAuthUser struct {
	ID                     int64
	Username               string
	Algorithm              string
	Params                 string
	Salt                   []byte
	Password               []byte
	RequiresPasswordChange bool
	CreatedAt              time.Time
	UpdatedAt              time.Time
}
