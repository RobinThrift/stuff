package entities

import (
	"io"
	"time"
)

type File struct {
	io.Reader

	ID      int64
	AssetID int64

	Name      string
	Filetype  string
	SizeBytes int64

	PublicPath string
	FullPath   string

	Sha256 []byte

	CreatedBy int64
	CreatedAt time.Time
	UpdatedAt time.Time
}
