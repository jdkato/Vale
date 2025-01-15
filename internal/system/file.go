package system

import (
	"os"
	"path/filepath"
	"strings"
)

// AbsPath returns the absolute path of `path`.
func AbsPath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return absPath
}

// FileExists determines if the path given by `filename` exists.
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

// FileNameWithoutExt returns the filename without its extension.
func FileNameWithoutExt(fileName string) string {
	base := filepath.Base(fileName)
	return strings.TrimSuffix(base, filepath.Ext(base))
}
