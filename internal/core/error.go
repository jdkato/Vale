package core

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pterm/pterm"
)

type lineError struct {
	content string
	line    int
	span    []int
}

type errorCondition func(position int, line, target string) bool

func annotate(file []byte, target string, finder errorCondition) (lineError, error) {
	var sb strings.Builder

	scanner := bufio.NewScanner(bytes.NewBuffer(file))
	context := lineError{span: []int{1, 1}}

	idx := 1
	for scanner.Scan() {
		markup := scanner.Text()
		plain := StripANSI(markup)
		if idx-context.line > 2 && context.line != 0 { //nolint:gocritic
			break
		} else if finder(idx, plain, target) && context.line == 0 {
			context.line = idx

			s := strings.Index(plain, target) + 1
			context.span = []int{s, s + len(target)}

			sb.WriteString(
				fmt.Sprintf("\033[1;32m%4d\033[0m* %s\n", idx, markup))
		} else {
			sb.WriteString(
				fmt.Sprintf("\033[1;32m%4d\033[0m  %s\n", idx, markup))
		}
		idx++
	}

	if err := scanner.Err(); err != nil {
		return lineError{}, err
	}

	lines := []string{}
	for i, l := range strings.Split(sb.String(), "\n") {
		if context.line-i < 3 {
			lines = append(lines, l)
		}
	}

	context.content = strings.Join(lines, "\n")
	return context, nil
}

// NewError creates a colored error from the given information.
//
// # The standard format is
//
// ```
// <code> [<context>] <title>
//
// <msg>
// ```
func NewError(code, title, msg string) error {
	return fmt.Errorf(
		"%s %s\n\n%s\n\n%s",
		pterm.BgRed.Sprint(code),
		title,
		msg,
		pterm.Fuzzy.Sprint(pterm.Italic.Sprintf("Execution stopped with code 1.")),
	)
}

// NewE100 creates a new, formatted "unexpected" error.
//
// Since E100 errors can occur anywhere, we include a "context" that makes it
// clear where exactly the error was generated.
func NewE100(context string, err error) error {
	title := fmt.Sprintf("[%s] %s", context, "Runtime error")
	return NewError("E100", title, err.Error())
}

// NewE201 creates a formatted user-generated error.
//
// 201 errors involve a specific configuration asset and should contain
// parsable location information on their last line of the form:
//
// <path>:<line>:<start>:<end>
func NewE201(msg, value, path string, finder errorCondition) error {
	f, err := os.ReadFile(path)
	if err != nil {
		return NewE100("NewE201", errors.New(msg))
	}

	ctx, err := annotate(f, value, finder)
	if err != nil {
		return NewE100("NewE201/annotate", err)
	}

	title := fmt.Sprintf(
		"Invalid value [%s:%d:%d]:",
		filepath.ToSlash(path),
		ctx.line,
		ctx.span[0])

	return NewError(
		"E201",
		title,
		fmt.Sprintf("%s\n%s", ctx.content, msg))
}

// NewE201FromTarget creates a new E201 error from a target string.
func NewE201FromTarget(msg, value, file string) error {
	return NewE201(
		msg,
		value,
		file,
		func(_ int, line, target string) bool {
			return strings.Contains(line, target)
		})
}

// NewE201FromPosition creates a new E201 error from an in-file location.
func NewE201FromPosition(msg, file string, goal int) error {
	return NewE201(
		msg,
		"",
		file,
		func(position int, _, _ string) bool {
			return position == goal
		})
}
