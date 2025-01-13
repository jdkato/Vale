package code

import (
	"regexp"

	"github.com/errata-ai/vale/v3/internal/core"
	"github.com/smacker/go-tree-sitter/cpp"
)

func Cpp() *Language {
	return &Language{
		Delims:  regexp.MustCompile(`//|/\*!?|\*/`),
		Parser:  cpp.GetLanguage(),
		Queries: []core.Scope{{Name: "", Expr: "(comment) @comment", Type: ""}},
		Padding: cStyle,
	}
}
