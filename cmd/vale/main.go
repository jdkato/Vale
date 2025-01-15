package main

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/pflag"

	"github.com/errata-ai/vale/v3/internal/core"
	"github.com/errata-ai/vale/v3/internal/lint"
	"github.com/errata-ai/vale/v3/internal/system"
)

// version is set during the release build process.
var version = "master"

func stat() bool {
	stat, err := os.Stdin.Stat()
	if err != nil || (stat.Mode()&os.ModeCharDevice) != 0 {
		return false
	}
	return true
}

func looksLikeStdin(s string) int {
	isDir := system.IsDir(s)
	if !(system.FileExists(s) || isDir) && s != "" {
		return 1
	} else if isDir {
		return 0
	}
	return -1
}

func doLint(args []string, l *lint.Linter, glob string) ([]*core.File, error) {
	var linted []*core.File
	var err error

	length := len(args)
	if length == 1 && looksLikeStdin(args[0]) == 1 { //nolint:gocritic
		// Case 1:
		//
		// $ vale "some text in a string"
		linted, err = l.LintString(args[0])
	} else if length > 0 {
		// Case 2:
		//
		// $ vale file1 dir1 file2
		input := []string{}
		for _, file := range args {
			status := looksLikeStdin(file)
			if status == 1 {
				return linted, core.NewE100(
					"doLint",
					fmt.Errorf("argument '%s' does not exist", file),
				)
			}
			l.HasDir = status == 0
			input = append(input, file)
		}
		linted, err = l.Lint(input, glob)
	} else {
		// Case 3:
		//
		// $ cat file.md | vale
		stdin, readErr := io.ReadAll(os.Stdin)
		if readErr != nil {
			return linted, core.NewE100("doLint", readErr)
		}
		linted, err = l.LintString(string(stdin))
		if err != nil {
			return linted, core.NewE100("doLint", err)
		}
	}

	return linted, err
}

func handleError(err error) {
	ShowError(err, Flags.Output, os.Stderr)
	os.Exit(2)
}

func main() {
	pflag.Parse()

	args := pflag.Args()
	argc := len(args)

	if Flags.Version { //nolint:gocritic
		fmt.Println("vale version " + version)
		os.Exit(0)
	} else if Flags.Help {
		pflag.Usage()
	} else if argc == 0 && !stat() {
		PrintIntro()
	}

	if argc > 0 {
		cmd, exists := Actions[args[0]]
		if exists {
			if err := cmd(args[1:], &Flags); err != nil {
				handleError(err)
			}
			os.Exit(0)
		}
	}

	config, err := core.ReadPipeline(&Flags, false)
	if err != nil {
		handleError(err)
	}

	linter, err := lint.NewLinter(config)
	if err != nil {
		handleError(err)
	}

	linted, err := doLint(args, linter, Flags.Glob)
	if err != nil {
		handleError(err)
	}

	hasErrors, err := PrintAlerts(linted, config)
	if err != nil {
		handleError(err)
	} else if hasErrors && !Flags.NoExit {
		os.Exit(1)
	}

	os.Exit(0)
}
