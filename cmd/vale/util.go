package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

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

func unarchive(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}

	for _, file := range r.File {
		destPath := filepath.Join(dest, file.Name)
		if file.FileInfo().IsDir() {
			os.MkdirAll(destPath, os.ModePerm)
			continue
		}
		if err := os.MkdirAll(filepath.Dir(destPath), os.ModePerm); err != nil {
			return err
		}

		dstFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}
		defer dstFile.Close()

		srcFile, err := file.Open()
		if err != nil {
			return err
		}
		defer srcFile.Close()

		_, err = io.Copy(dstFile, srcFile)
		if err != nil {
			return err
		}
	}
	return nil
}
