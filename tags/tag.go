package tags

import (
	"time"
)

type Tag struct {
	ID        int64
	Tag       string
	InUse     bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type TagList struct {
	Tags  []*Tag
	Total int
}
