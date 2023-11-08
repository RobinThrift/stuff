package entities

import (
	"errors"
	"time"
)

var ErrTagNotFound = errors.New("tag not found")

type Tag struct {
	ID        int64
	Tag       string
	InUse     bool
	CreatedAt time.Time
	UpdatedAt time.Time
}
