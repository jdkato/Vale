package code

import (
	"regexp"

	"github.com/errata-ai/vale/v3/internal/core"
	"github.com/smacker/go-tree-sitter/ruby"
)

func Ruby() *Language {
	return &Language{
		Delims:  regexp.MustCompile(`#|=begin|=end`),
		Parser:  ruby.GetLanguage(),
		Queries: []core.Scope{{Name: "", Expr: "(comment) @comment", Type: ""}},
		Padding: func(s string) int {
			return computePadding(s, []string{"#", `=begin`, `=end`})
		},
	}
}
