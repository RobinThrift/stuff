package assets

import (
	"errors"
)

var ErrAssetNotFound = errors.New("asset not found")
var ErrFileNotFound = errors.New("asset file not found")
