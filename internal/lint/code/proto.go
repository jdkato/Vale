package code

import (
	"regexp"

	"github.com/errata-ai/vale/v3/internal/core"
	"github.com/smacker/go-tree-sitter/protobuf"
)

func Protobuf() *Language {
	return &Language{
		Delims:  regexp.MustCompile(`//|/\*|\*/`),
		Parser:  protobuf.GetLanguage(),
		Queries: []core.Scope{{Name: "", Expr: "(comment) @comment", Type: ""}},
		Padding: cStyle,
	}
}
