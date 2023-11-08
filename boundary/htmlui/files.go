package htmlui

import (
	"errors"
	"fmt"
	"mime"
	"mime/multipart"
	"net/http"

	"github.com/RobinThrift/stuff/entities"
)

const defaultMaxMemory = 32 << 20 // 32 MB same as stdlib

func fileUploads(r *http.Request, assetID int64, userID int64) ([]*entities.File, error) {
	err := r.ParseMultipartForm(defaultMaxMemory)
	if err != nil {
		return nil, err
	}

	files := make([]*entities.File, 0, len(r.MultipartForm.File))

	for k := range r.MultipartForm.File {
		uploaded, header, err := r.FormFile(k)
		if err != nil {
			return nil, err
		}

		files = append(files, &entities.File{
			Reader:    uploaded,
			AssetID:   assetID,
			Name:      k,
			Filetype:  header.Header.Get("content-type"),
			CreatedBy: userID,
		})
	}

	return files, nil
}

var imgAllowList = []string{"image/png", "image/jpeg", "image/webp"}

func handleFileUpload(r *http.Request, key string) (*entities.File, error) {
	err := r.ParseMultipartForm(defaultMaxMemory)
	if err != nil {
		return nil, err
	}

	_, hasFileUpload := r.MultipartForm.File[key]
	if !hasFileUpload {
		return nil, nil
	}

	uploaded, header, err := r.FormFile(key)
	if err != nil {
		return nil, err
	}

	if uploaded != nil {
		err = checkFileType(header, imgAllowList)
		if err != nil {
			return nil, err
		}

		return &entities.File{Name: header.Filename, Reader: uploaded}, nil
	}

	return nil, nil
}

var errFileTypeNotAllowed = errors.New("file type not allowed")

func checkFileType(header *multipart.FileHeader, allowlist []string) error {
	return checkContentTypeAllowed(header.Header.Get("content-type"), allowlist)
}

func checkContentTypeAllowed(ct string, allowlist []string) error {
	mt, _, err := mime.ParseMediaType(ct)
	if err != nil {
		return err
	}

	allowed := false
	for _, m := range allowlist {
		if mt == m {
			allowed = true
			break
		}
	}

	if !allowed {
		return fmt.Errorf("%w: %s", errFileTypeNotAllowed, mt)
	}

	return nil

}
