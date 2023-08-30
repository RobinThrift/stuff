package tags

import (
	"time"
)

type Tag struct {
	ID        int64
	Tag       string
	CreatedAt time.Time
	UpdatedAt time.Time
}
