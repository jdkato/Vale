package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mholt/archives"
	"github.com/pterm/pterm"

	"github.com/errata-ai/vale/v3/internal/core"
)

// Response is returned after an action.
type Response struct {
	Msg     string
	Error   string
	Success bool
}

func progressError(context string, err error, p *pterm.ProgressbarPrinter) error {
	_, _ = p.Stop()
	return core.NewE100(context, err)
}

func pluralize(s string, n int) string {
	if n != 1 {
		return s + "s"
	}
	return s
}

func getJSON(data interface{}) string {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(b)
}

func fetchJSON(url string) ([]byte, error) {
	resp, err := http.Get(url) //nolint:gosec,noctx
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func printJSON(t interface{}) error {
	b, err := json.MarshalIndent(t, "", "    ")
	if err != nil {
		fmt.Println("{}")
		return err
	}
	fmt.Println(string(b))
	return nil
}

// Send a JSON response after a local action.
func sendResponse(msg string, err error) error {
	resp := Response{Msg: msg, Success: err == nil}
	if !resp.Success {
		resp.Error = err.Error()
	}
	return printJSON(resp)
}

func fileNameWithoutExt(fileName string) string {
	base := filepath.Base(fileName)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

func platformAndArch() string {
	platform := strings.Title(runtime.GOOS) //nolint:staticcheck

	arch := strings.ToLower(runtime.GOARCH)
	if arch == "amd64" {
		arch = "x86_64"
	}

	return fmt.Sprintf("%s_%s", platform, arch)
}

func mkdir(dir string) error {
	return os.MkdirAll(dir, os.ModeDir|0700)
}

func toCodeStyle(s string) string {
	return pterm.Fuzzy.Sprint(s)
}

// handleFile handles the extraction of a file from the archive.
func handleFile(f archives.FileInfo, dst string) error {
	// Validate and construct the destination path
	dstPath, pathErr := securePath(dst, f.NameInArchive)
	if pathErr != nil {
		return pathErr
	}

	// Ensure the parent directory exists
	parentDir := filepath.Dir(dstPath)
	if dirErr := mkdir(parentDir); dirErr != nil {
		return dirErr
	}

	// Handle directories
	if f.IsDir() {
		// Create the directory with permissions from the archive
		if dirErr := mkdir(dstPath); dirErr != nil {
			return fmt.Errorf("creating directory: %w", dirErr)
		}
		return nil
	}

	// Handle symlinks
	if f.LinkTarget != "" {
		targetPath, linkErr := securePath(dst, f.LinkTarget)
		if linkErr != nil {
			return fmt.Errorf("invalid symlink target: %w", linkErr)
		}
		if linkErr := os.Symlink(targetPath, dstPath); linkErr != nil {
			return fmt.Errorf("create symlink: %w", linkErr)
		}
		return nil
	}

	// Check and handle parent directory permissions
	originalMode, statErr := os.Stat(parentDir)
	if statErr != nil {
		return fmt.Errorf("stat parent directory: %w", statErr)
	}

	// If parent directory is read-only, temporarily make it writable
	if originalMode.Mode().Perm()&0o200 == 0 {
		if chmodErr := os.Chmod(parentDir, originalMode.Mode()|0o200); chmodErr != nil {
			return fmt.Errorf("chmod parent directory: %w", chmodErr)
		}
	}

	// Handle regular files
	reader, openErr := f.Open()
	if openErr != nil {
		return fmt.Errorf("open file: %w", openErr)
	}
	defer reader.Close()

	dstFile, createErr := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY, f.Mode())
	if createErr != nil {
		return fmt.Errorf("create file: %w", createErr)
	}
	defer dstFile.Close()

	if _, copyErr := io.Copy(dstFile, reader); copyErr != nil {
		return fmt.Errorf("copy: %w", copyErr)
	}
	return nil
}

func securePath(basePath, relativePath string) (string, error) {
	relativePath = filepath.Clean("/" + relativePath)                         // Normalize path with a leading slash
	relativePath = strings.TrimPrefix(relativePath, string(os.PathSeparator)) // Remove leading separator

	dstPath := filepath.Join(basePath, relativePath)

	if !strings.HasPrefix(filepath.Clean(dstPath)+string(os.PathSeparator), filepath.Clean(basePath)+string(os.PathSeparator)) {
		return "", fmt.Errorf("illegal file path: %s", dstPath)
	}
	return dstPath, nil
}

func unarchive(src, dst string) error {
	archiveFile, openErr := os.Open(src)
	if openErr != nil {
		return fmt.Errorf("open tarball %s: %w", src, openErr)
	}
	defer archiveFile.Close()
	
	format, input, identifyErr := archives.Identify(context.Background(), src, archiveFile)
	if identifyErr != nil {
		return fmt.Errorf("identify format: %w", identifyErr)
	}

	extractor, ok := format.(archives.Extractor)
	if !ok {
		return fmt.Errorf("unsupported format for extraction")
	}
	
	if dirErr := mkdir(dst); dirErr != nil {
		return fmt.Errorf("creating destination directory: %w", dirErr)
	}

	handler := func(ctx context.Context, f archives.FileInfo) error {
		return handleFile(f, dst)
	}

	if extractErr := extractor.Extract(context.Background(), input, handler); extractErr != nil {
		return fmt.Errorf("extracting files: %w", extractErr)
	}

	return nil
}