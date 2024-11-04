package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFormatFromExt(t *testing.T) {
	extToFormat := map[string][]string{
		".py":    {".py", "code"},
		".cxx":   {".cpp", "code"},
		".mdown": {".md", "markup"},
	}
	m := map[string]string{}
	for ext, format := range extToFormat {
		normExt, f := FormatFromExt(ext, m)
		if format[0] != normExt {
			t.Errorf("expected = %v, got = %v", format[0], normExt)
		}
		if format[1] != f {
			t.Errorf("expected = %v, got = %v", format[1], f)
		}
	}
}

func TestPrepText(t *testing.T) {
	rawToPrepped := map[string]string{
		"foo\r\nbar":     "foo\nbar",
		"foo\r\n\r\nbar": "foo\n\nbar",
	}
	for raw, prepped := range rawToPrepped {
		if prepped != Sanitize(raw) {
			t.Errorf("expected = %v, got = %v", prepped, Sanitize(raw))
		}
	}
}

func TestPhrase(t *testing.T) {
	rawToPrepped := map[string]bool{
		"test suite":    true,
		"test[ ]?suite": false,
	}
	for input, output := range rawToPrepped {
		result := IsPhrase(input)
		if result != output {
			t.Errorf("expected = %v, got = %v", output, result)
		}
	}
}

func TestNormalizePath(t *testing.T) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		t.Log("os.UserHomeDir failed, will not proceed with tests")
		return
	}
	stylesPathInput := filepath.FromSlash("~/.vale")
	expectedOutput := filepath.Join(homedir, ".vale")
	result := normalizePath(stylesPathInput)
	if result != expectedOutput {
		t.Errorf("expected = %v, got = %v", expectedOutput, result)
	}
	stylesPathInput, err = os.MkdirTemp("", "vale_test")
	if err != nil {
		t.Log("os.MkdirTemp failed, will not proceed with tests")
		return
	}
	expectedOutput = stylesPathInput
	result = normalizePath(stylesPathInput)
	if result != expectedOutput {
		t.Errorf("expected = %v, got = %v", expectedOutput, result)
	}
	stylesPathInput, err = os.MkdirTemp("", "vale~test")
	if err != nil {
		t.Log("os.MkdirTemp failed in second case, will not proceed with tests")
		return
	}
	expectedOutput = stylesPathInput
	result = normalizePath(stylesPathInput)
	if result != expectedOutput {
		t.Errorf("expected = %v, got = %v", expectedOutput, result)
	}
}
