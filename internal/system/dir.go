package system

import "os"

// Mkdir creates a directory at the given path.
func Mkdir(dir string) error {
	return os.MkdirAll(dir, os.ModeDir|0700)
}

// IsDir determines if the path given by `filename` is a directory.
func IsDir(filename string) bool {
	fi, err := os.Stat(filename)
	return err == nil && fi.IsDir()
}
