package code

import (
	"regexp"

	"github.com/errata-ai/vale/v3/internal/core"
	"github.com/smacker/go-tree-sitter/css"
)

func CSS() *Language {
	return &Language{
		Delims:  regexp.MustCompile(`/\*!?|\*/`),
		Parser:  css.GetLanguage(),
		Queries: []core.Scope{{Name: "", Expr: "(comment) @comment", Type: ""}},
		Padding: func(s string) int {
			return computePadding(s, []string{"/*"})
		},
	}
}
