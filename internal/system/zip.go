package system

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Unarchive extracts a ZIP archive to a destination directory.
func Unarchive(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	if err = Mkdir(dest); err != nil {
		return err
	}

	for _, file := range r.File {
		destPath := filepath.Join(dest, filepath.Clean(file.Name))
		if !strings.HasPrefix(destPath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			if err = Mkdir(destPath); err != nil {
				return err
			}
			continue
		}
		if err = Mkdir(filepath.Dir(destPath)); err != nil {
			return err
		}

		dstFile, dstErr := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if dstErr != nil {
			return dstErr
		}
		defer dstFile.Close()

		srcFile, srcErr := file.Open()
		if srcErr != nil {
			return srcErr
		}
		defer srcFile.Close()

		if _, err = io.Copy(dstFile, io.LimitReader(srcFile, 1024*1024*1024*10)); err != nil {
			return err
		}
	}
	return nil
}
