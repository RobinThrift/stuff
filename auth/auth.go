package auth

import (
	"errors"
	"time"
)

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
