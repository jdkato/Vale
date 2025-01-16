package system

import (
	"os"
	"os/exec"
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

// Which checks for the existence of any command in `cmds`.
func Which(cmds []string) string {
	for _, cmd := range cmds {
		path, err := exec.LookPath(cmd)
		if err == nil {
			return path
		}
	}
	return ""
}

// NormalizePath expands a tilde and returns the absolute path of `path`.
func NormalizePath(path string) string {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return path
	}

	if path == "~" {
		return homedir
	} else if strings.HasPrefix(path, filepath.FromSlash("~/")) {
		path = filepath.Join(homedir, path[2:])
	}

	return path
}

// DeterminePath determines the path of `keyPath` based on `configPath`.
//
// If `keyPath` is an absolute path, it is returned as is.
//
// If `keyPath` is a relative path, it is joined with `configPath`.
//
// If `configPath` is not a directory, the directory part of `configPath`
// is used.
func DeterminePath(configPath string, keyPath string) string {
	// expand tilde at this point as this is where user-provided paths are provided
	keyPath = NormalizePath(keyPath)
	if !IsDir(configPath) {
		configPath = filepath.Dir(configPath)
	}

	sep := string(filepath.Separator)
	abs := AbsPath(keyPath)

	rel := strings.TrimRight(keyPath, sep)
	if abs != rel || !strings.Contains(keyPath, sep) {
		// The path was relative
		return filepath.Join(configPath, keyPath)
	}

	return abs
}
