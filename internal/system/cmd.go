package system

import (
	"bytes"
	"errors"
	"os/exec"
	"strings"
)

// ExecuteWithInput runs a command with the given text as input.
func ExecuteWithInput(exe, text string, args ...string) (string, error) {
	var out bytes.Buffer
	var eut bytes.Buffer

	cmd := exec.Command(exe, args...)
	cmd.Stdin = strings.NewReader(text)
	cmd.Stdout = &out
	cmd.Stderr = &eut

	if err := cmd.Run(); err != nil {
		return "", errors.New(eut.String())
	}

	return out.String(), nil
}
