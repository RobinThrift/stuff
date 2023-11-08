package blobs

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/RobinThrift/stuff/entities"
)

type LocalFS struct {
	RootDir string
	TmpDir  string
}

func (fs *LocalFS) WriteFile(file *entities.File) (err error) {
	err = ensureDirExists(fs.RootDir)
	if err != nil {
		return err
	}

	fhandle, err := os.CreateTemp(fs.TmpDir, file.Name)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			err = errors.Join(err, fhandle.Close(), os.Remove(fhandle.Name()))
		}
	}()

	h := sha256.New()

	tee := io.TeeReader(file, h)

	file.SizeBytes, err = io.Copy(fhandle, tee)
	if err != nil {
		return err
	}

	err = fhandle.Close()
	if err != nil {
		return err
	}

	ext := path.Ext(file.Name)
	file.Sha256 = h.Sum(nil)
	for _, b := range file.Sha256 {
		file.FullPath = file.FullPath + "/" + fmt.Sprintf("%x", b)
	}

	file.FullPath += ext

	file.PublicPath = "/assets/files" + file.FullPath
	file.FullPath = path.Join(fs.RootDir, file.FullPath)

	err = ensureDirExists(path.Dir(file.FullPath))
	if err != nil {
		return err
	}

	err = os.Rename(fhandle.Name(), file.FullPath)
	if err != nil {
		return err
	}

	return nil
}

func (fs *LocalFS) RemoveFile(file *entities.File) error {
	if !strings.HasPrefix(file.FullPath, fs.RootDir) {
		return fmt.Errorf("invalid file path for deletion: file path is not in configured file dir: %s", file.FullPath)
	}

	filename := path.Base(file.FullPath)

	err := os.Remove(file.FullPath)
	if err != nil {
		return err
	}

	dir := file.FullPath[:len(file.FullPath)-1-len(filename)]
	for dir != fs.RootDir {
		isEmpty, err := isEmptyDir(dir)
		if err != nil {
			return err
		}

		if !isEmpty {
			return nil
		}

		err = os.RemoveAll(dir)
		if err != nil {
			return err
		}

		slashIndex := strings.LastIndex(dir, "/")
		if slashIndex == -1 {
			return nil
		}

		dir = dir[:slashIndex]
	}

	return nil

}

func ensureDirExists(dir string) error {
	stat, err := os.Stat(dir)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}

	if stat == nil {
		return os.MkdirAll(dir, 0755)
	}

	if !stat.IsDir() {
		return fmt.Errorf("%s exists but is not a directory", dir)
	}

	return nil
}

func isEmptyDir(dir string) (bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false, err
	}

	return len(entries) == 0, nil
}
