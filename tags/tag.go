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

type TagListPage struct {
	Tags     []*Tag
	Total    int
	NumPages int
	Page     int
	PageSize int
}
