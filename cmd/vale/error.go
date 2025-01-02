package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/errata-ai/vale/v3/internal/core"
)

type valeError struct {
	line int
	path string
	text string
	code string
	span int
}

var logger = log.New(os.Stderr, "", 0)
var header = regexp.MustCompile(`(\w+) .+ \[(.+):(\d+):(\d+)\]`)

func parseError(err error) (valeError, error) {
	var parsed valeError

	plain := core.StripANSI(err.Error())
	lines := strings.Split(plain, "\n\n")

	if len(lines) < 3 {
		return parsed, errors.New("missing body")
	} else if !header.MatchString(lines[0]) {
		return parsed, errors.New("missing header")
	}

	groups := header.FindStringSubmatch(lines[0])

	parsed.code = groups[1]
	parsed.path = groups[2]

	i, err := strconv.Atoi(groups[3])
	if err != nil {
		return parsed, errors.New("missing line")
	}
	parsed.line = i

	i, err = strconv.Atoi(groups[4])
	if err != nil {
		return parsed, errors.New("missing span")
	}
	parsed.span = i

	parsed.text = lines[len(lines)-2]
	return parsed, nil
}

// ShowError displays the given error in the user-specified format.
func ShowError(err error, style string, out io.Writer) {
	parsed, failed := parseError(err)

	logger.SetOutput(out)
	switch style {
	case "JSON":
		var data interface{}

		if failed != nil {
			data = struct {
				Line int
				Path string
				Text string
				Code string
				Span int
			}{
				Line: 0,
				Path: "",
				Text: core.StripANSI(err.Error()),
				Code: "E100",
				Span: 0,
			}
		} else {
			data = struct {
				Line int
				Path string
				Text string
				Code string
				Span int
			}{
				Line: parsed.line,
				Path: parsed.path,
				Text: parsed.text,
				Code: parsed.code,
				Span: parsed.span,
			}
		}

		logger.Println(getJSON(data))
	case "line":
		var data string

		if failed != nil {
			data = err.Error()
		} else {
			data = fmt.Sprintf("%s:%d:%s:%s",
				parsed.path, parsed.line, parsed.code, parsed.text)
		}

		logger.Println(data)
	default:
		logger.Println(err)
	}
}
